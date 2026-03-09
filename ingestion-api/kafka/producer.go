package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"

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
		Timeout:   10 * 1000 * 1000 * 1000,
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
	})

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
