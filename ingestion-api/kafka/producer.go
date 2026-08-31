package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

var writer *kafka.Writer
var topic string

// InitProducer initializes the Kafka writer.
func InitProducer() error {
	brokers := os.Getenv("KAFKA_BROKERS")
	username := os.Getenv("KAFKA_USERNAME")
	password := os.Getenv("KAFKA_PASSWORD")
	topicEnv := os.Getenv("KAFKA_TOPIC")

	if brokers == "" {
		brokers = "localhost:9092"
	}
	if topicEnv != "" {
		topic = topicEnv
	} else {
		topic = "user-events" // fallback
	}

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	// Connect securely if credentials are provided (Upstash)
	if username != "" && password != "" {
		mechanism, err := scram.Mechanism(scram.SHA256, username, password)
		if err != nil {
			return fmt.Errorf("failed to create sasl mechanism: %w", err)
		}
		dialer.SASLMechanism = mechanism
		dialer.TLS = &tls.Config{}
	}

	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{brokers},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		Dialer:   dialer,

		// The ingestion endpoint answers 202 Accepted — "we have taken custody
		// of this event", not "it is durably stored". Async lets WriteMessages
		// hand the message to the background batcher and return immediately,
		// which is what keeps the HTTP handler in the sub-millisecond range.
		// Left synchronous, every request blocks for a full BatchTimeout
		// (1 second by default) waiting for the batch to flush.
		Async: true,

		// Flush partial batches quickly so an idle system still delivers
		// promptly instead of sitting on a message for a whole second.
		BatchTimeout: 10 * time.Millisecond,
	})

	// Async writes cannot return an error to the HTTP handler, so surface
	// delivery failures here instead of dropping them silently.
	writer.Completion = func(messages []kafka.Message, err error) {
		if err != nil {
			log.Printf("[kafka] delivery failed for %d message(s): %v", len(messages), err)
		}
	}

	return nil
}

// PublishEvent pushes an event payload to Kafka.
func PublishEvent(event interface{}) error {
	if writer == nil {
		return fmt.Errorf("producer not initialized")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: payload,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

// CloseProducer safely flushes and closes the producer.
func CloseProducer() {
	if writer != nil {
		writer.Close()
	}
}
