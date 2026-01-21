package kafkav2

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/logger"
)

// Producer wraps kafka.Writer to provide structured logging and consistent error handling.
// It is configured with strong durability guarantees (RequireAll acks) and a 10-second write timeout.
type Producer struct {
	writer *kafka.Writer
	log    logger.Logger
}

// NewProducer creates a new Kafka producer configured for the given brokers and topic.
// It uses LeastBytes balancer, requires acknowledgments from all in-sync replicas,
// and has a 10-second write timeout. All internal logs are routed through the provided logger
// with structured attributes.
func NewProducer(brokers []string, topic string, log logger.Logger) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
			Logger: kafka.LoggerFunc(func(msg string, args ...any) {
				log.LogAttrs(context.Background(), logger.InfoLevel, "producer info",
					logger.String("message", fmt.Sprintf(msg, args...)),
				)
			}),
			ErrorLogger: kafka.LoggerFunc(func(msg string, args ...any) {
				log.LogAttrs(context.Background(), logger.ErrorLevel, "producer error",
					logger.String("error", fmt.Sprintf(msg, args...)),
				)
			}),
		},
		log: log,
	}
}

// Send publishes a single message to the Kafka topic.
// It wraps any underlying error with a descriptive prefix for easier debugging.
// The operation respects the provided context for cancellation and timeouts.
func (p *Producer) Send(ctx context.Context, key, value []byte, headers ...kafka.Header) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:     key,
		Value:   value,
		Headers: headers,
	})
	if err != nil {
		return fmt.Errorf("kafkav2.Producer.Send: %w", err)
	}
	return nil
}

// Close gracefully shuts down the producer and flushes any pending messages.
// It is safe to call Close multiple times.
func (p *Producer) Close() error {
	return p.writer.Close()
}
