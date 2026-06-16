import os
import json
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


def get_kafka_config(is_consumer=True):
    brokers = os.getenv("KAFKA_BROKERS", "localhost:9092")
    username = os.getenv("KAFKA_USERNAME", "")
    password = os.getenv("KAFKA_PASSWORD", "")
    group_id = os.getenv("KAFKA_GROUP_ID", "processor-group")

    config = {"bootstrap.servers": brokers}

    if is_consumer:
        config["group.id"] = group_id
        config["auto.offset.reset"] = "earliest"

    if username and password:
        config["security.protocol"] = "SASL_SSL"
        config["sasl.mechanisms"] = "SCRAM-SHA-256"
        config["sasl.username"] = username
        config["sasl.password"] = password

    return config


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

    try:
        while True:
            msg = consumer.poll(timeout=1.0)
            if msg is None:
                continue
            if msg.error():
                logger.error(f"Consumer error: {msg.error()}")
                continue

            # --- consume_event span: covers the entire message processing cycle ---
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

                    # --- persist_to_db span: database write ---
                    with tracer.start_as_current_span(
                        "persist_to_db",
                        kind=SpanKind.CLIENT,
                        attributes={
                            "db.system": "postgresql",
                            "db.operation": "INSERT",
                            "event.user_id": event.user_id,
                            "event.action": event.action,
                        },
                    ) as db_span:
                        session = get_session()
                        try:
                            db_event = AnalyticsEvent(
                                user_id=event.user_id,
                                action=event.action,
                                timestamp=event.timestamp,
                                processed_at=datetime.now(timezone.utc),
                            )
                            session.add(db_event)
                            session.commit()
                            db_span.set_attribute("event.status", "persisted")
                            logger.info(
                                f"Successfully processed event for user {event.user_id}"
                            )
                        except Exception as e:
                            session.rollback()
                            db_span.set_status(
                                StatusCode.ERROR, str(e)
                            )
                            db_span.record_exception(e)
                            db_span.set_attribute("event.status", "db_error")
                            logger.error(f"Database error: {str(e)}")
                            # DLQ for failures
                            producer.produce(dlq_topic, value=raw_value)
                            producer.poll(0)
                        finally:
                            session.close()

                    consume_span.set_attribute("event.status", "success")

                except (json.JSONDecodeError, ValidationError) as e:
                    consume_span.set_status(StatusCode.ERROR, str(e))
                    consume_span.record_exception(e)
                    consume_span.set_attribute("event.status", "validation_failed")
                    logger.warning(
                        f"Malformed message or validation error: {str(e)}"
                    )
                    # DLQ
                    producer.produce(dlq_topic, value=raw_value)
                    producer.poll(0)

    except KeyboardInterrupt:
        logger.info("Processor shutting down...")
    except Exception as e:
        logger.error(f"Unexpected error: {str(e)}")
    finally:
        consumer.close()
        producer.flush(10.0)


if __name__ == "__main__":
    main()
