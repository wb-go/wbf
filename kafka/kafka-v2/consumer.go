package kafkav2

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/logger"
)

// contextKey is a private type used to avoid key collisions in context.WithValue.
type contextKey string

const kafkaMetadataKey contextKey = "kafka_metadata"

// Consumer wraps kafka.Reader to provide structured logging and error context.
// It automatically enriches log records with Kafka-specific metadata such as topic and group ID.
type Consumer struct {
	reader *kafka.Reader
	log    logger.Logger
}

// NewConsumer creates a new Kafka consumer configured with the given brokers, topic, and group ID.
// It sets up structured logging via the provided logger, injecting Kafka metadata into every log record.
func NewConsumer(brokers []string, topic, groupID string, log logger.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
		Logger: kafka.LoggerFunc(func(msg string, args ...any) {
			ctx := context.WithValue(context.Background(), kafkaMetadataKey, map[string]string{
				"topic":    topic,
				"group_id": groupID,
			})
			log.LogAttrs(ctx, logger.InfoLevel, "consumer info",
				logger.String("message", fmt.Sprintf(msg, args...)),
			)
		}),
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...any) {
			ctx := context.WithValue(context.Background(), kafkaMetadataKey, map[string]string{
				"topic":    topic,
				"group_id": groupID,
			})
			log.LogAttrs(ctx, logger.ErrorLevel, "consumer error",
				logger.String("error", fmt.Sprintf(msg, args...)),
			)
		}),
	})

	return &Consumer{
		reader: reader,
		log:    log,
	}
}

// Fetch retrieves the next message from the Kafka topic.
// It wraps any underlying error with a descriptive prefix for easier debugging.
// The method respects the provided context for cancellation and timeouts.
func (c *Consumer) Fetch(ctx context.Context) (kafka.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("kafkav2.Consumer.Fetch: %w", err)
	}
	return msg, nil
}

// Commit acknowledges the successful processing of a message by committing its offset.
// It wraps any error from the underlying Kafka client with a descriptive prefix.
// Note: Commit should only be called after the message has been fully processed.
func (c *Consumer) Commit(ctx context.Context, msg kafka.Message) error {
	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafkav2.Consumer.Commit: %w", err)
	}
	return nil
}

// Close shuts down the Kafka consumer and releases all associated resources.
// It is safe to call Close multiple times.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
