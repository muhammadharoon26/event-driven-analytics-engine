import os
import json
import time
import logging
from datetime import datetime, timezone
from confluent_kafka import Consumer, Producer, KafkaException
from pydantic import BaseModel, ValidationError

from opentelemetry import trace, context
from opentelemetry.trace import StatusCode, SpanKind

from database.db import AnalyticsEvent, init_db, get_session
from telemetry.otel import init_tracer

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class EventModel(BaseModel):
    user_id: str
    action: str
    timestamp: datetime


# Postgres does a disk sync on every COMMIT, which costs several milliseconds
# regardless of how much data the transaction carries. Committing one row at a
# time therefore caps the worker at roughly 100 events/sec. Buffering rows and
# committing them together amortises that sync across the whole batch.
BATCH_SIZE = int(os.getenv("BATCH_SIZE", "500"))

# Upper bound on how long an event may sit in the buffer waiting for the batch
# to fill. Without it, a trickle of traffic would never reach BATCH_SIZE and
# events would appear to hang.
BATCH_INTERVAL_SECONDS = float(os.getenv("BATCH_INTERVAL_SECONDS", "0.5"))


def get_kafka_config(is_consumer=True):
    brokers = os.getenv("KAFKA_BROKERS", "localhost:9092")
    username = os.getenv("KAFKA_USERNAME", "")
    password = os.getenv("KAFKA_PASSWORD", "")
    group_id = os.getenv("KAFKA_GROUP_ID", "processor-group")

    config = {"bootstrap.servers": brokers}

    if is_consumer:
        config["group.id"] = group_id
        config["auto.offset.reset"] = "earliest"
        # Offsets are committed by hand, only after the batch is safely in
        # Postgres. Left on auto-commit, a crash between "read" and "written"
        # would silently drop those events instead of redelivering them.
        config["enable.auto.commit"] = False

    if username and password:
        config["security.protocol"] = "SASL_SSL"
        config["sasl.mechanisms"] = "SCRAM-SHA-256"
        config["sasl.username"] = username
        config["sasl.password"] = password

    return config


def flush_batch(tracer, consumer, producer, dlq_topic, batch):
    """Write one buffered batch to Postgres, then advance the Kafka offsets.

    `batch` is a list of (AnalyticsEvent, raw_message_bytes) pairs. The raw
    message is carried alongside the row so a failed batch can be replayed onto
    the dead-letter topic without having to re-serialise it.
    """
    if not batch:
        return

    with tracer.start_as_current_span(
        "persist_batch",
        kind=SpanKind.CLIENT,
        attributes={
            "db.system": "postgresql",
            "db.operation": "INSERT",
            "batch.size": len(batch),
        },
    ) as db_span:
        session = get_session()
        try:
            session.add_all([row for row, _ in batch])
            session.commit()
            db_span.set_attribute("event.status", "persisted")

            # Only now is it safe to acknowledge the messages. Committing
            # offsets after the write is what makes delivery at-least-once.
            consumer.commit(asynchronous=False)
            logger.info(f"Persisted batch of {len(batch)} event(s)")
        except Exception as e:
            session.rollback()
            db_span.set_status(StatusCode.ERROR, str(e))
            db_span.record_exception(e)
            db_span.set_attribute("event.status", "db_error")
            logger.error(f"Database error on batch of {len(batch)}: {str(e)}")

            # Park the whole batch on the dead-letter topic, then acknowledge
            # it so one bad batch cannot block the partition forever.
            for _, raw in batch:
                producer.produce(dlq_topic, value=raw)
            producer.poll(0)
            consumer.commit(asynchronous=False)
        finally:
            session.close()


def main():
    logger.info("Initializing telemetry...")
    tracer = init_tracer()

    logger.info("Initializing database...")
    init_db()

    consumer_config = get_kafka_config(is_consumer=True)
    producer_config = get_kafka_config(is_consumer=False)

    topic = os.getenv("KAFKA_TOPIC", "user-events")
    dlq_topic = os.getenv("KAFKA_DLQ_TOPIC", "user-events-dlq")

    consumer = Consumer(consumer_config)
    producer = Producer(producer_config)

    consumer.subscribe([topic])
    logger.info(f"Subscribed to topic '{topic}'")

    # Rows waiting to be written, as (AnalyticsEvent, raw_bytes) pairs.
    batch = []
    last_flush = time.monotonic()

    try:
        while True:
            # A short poll keeps the loop responsive enough to honour
            # BATCH_INTERVAL_SECONDS even when no messages are arriving.
            #
            # Measured alternative: consumer.consume(num_messages=BATCH_SIZE)
            # fetches many messages per call and looks like it should be
            # faster, but it waits on fetch.wait.max.ms to fill the request and
            # measured roughly half the throughput of poll() on this workload.
            # Kept poll() on the strength of the measurement.
            msg = consumer.poll(timeout=0.2)

            if msg is not None and msg.error():
                logger.error(f"Consumer error: {msg.error()}")
                msg = None

            if msg is not None:

                # --- consume_event span: receive + validate for one message ---
                with tracer.start_as_current_span(
                    "consume_event",
                    kind=SpanKind.CONSUMER,
                    attributes={
                        "messaging.system": "kafka",
                        "messaging.destination": topic,
                        "messaging.operation": "receive",
                    },
                ) as consume_span:
                    raw_value = msg.value().decode("utf-8")
                    consume_span.set_attribute(
                        "messaging.kafka.partition", msg.partition()
                    )
                    consume_span.set_attribute(
                        "messaging.kafka.offset", msg.offset()
                    )

                    try:
                        payload = json.loads(raw_value)

                        # --- validate_event span: validation & deserialization ---
                        with tracer.start_as_current_span(
                            "validate_event",
                            kind=SpanKind.INTERNAL,
                        ) as validate_span:
                            try:
                                event = EventModel(**payload)
                                validate_span.set_attribute(
                                    "event.user_id", event.user_id
                                )
                                validate_span.set_attribute(
                                    "event.action", event.action
                                )
                                validate_span.set_attribute("event.status", "valid")
                            except ValidationError as ve:
                                validate_span.set_attribute("event.status", "invalid")
                                validate_span.set_status(
                                    StatusCode.ERROR, str(ve)
                                )
                                validate_span.record_exception(ve)
                                raise

                        batch.append((
                            AnalyticsEvent(
                                user_id=event.user_id,
                                action=event.action,
                                timestamp=event.timestamp,
                                processed_at=datetime.now(timezone.utc),
                            ),
                            raw_value,
                        ))
                        consume_span.set_attribute("event.status", "buffered")

                    except (json.JSONDecodeError, ValidationError) as e:
                        consume_span.set_status(StatusCode.ERROR, str(e))
                        consume_span.record_exception(e)
                        consume_span.set_attribute("event.status", "validation_failed")
                        logger.warning(
                            f"Malformed message or validation error: {str(e)}"
                        )
                        # A message that can never be parsed goes straight to the
                        # dead-letter topic rather than poisoning the batch.
                        producer.produce(dlq_topic, value=raw_value)
                        producer.poll(0)

            # Flush when the buffer is full, or when the oldest buffered event
            # has waited long enough.
            waited = time.monotonic() - last_flush
            if batch and (len(batch) >= BATCH_SIZE or waited >= BATCH_INTERVAL_SECONDS):
                flush_batch(tracer, consumer, producer, dlq_topic, batch)
                batch = []
                last_flush = time.monotonic()

    except KeyboardInterrupt:
        logger.info("Processor shutting down...")
    except Exception as e:
        logger.error(f"Unexpected error: {str(e)}")
    finally:
        # Never drop what is already buffered on the way out.
        try:
            flush_batch(tracer, consumer, producer, dlq_topic, batch)
        except Exception as e:
            logger.error(f"Failed to flush final batch: {str(e)}")
        consumer.close()
        producer.flush(10.0)


if __name__ == "__main__":
    main()
