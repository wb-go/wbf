// Package dlq provides a Dead Letter Queue (DLQ) client for publishing failed Kafka messages
// to a dedicated error topic with structured metadata and safe serialization.
package dlq

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/logger"
)

// Publisher defines the minimal interface required to send messages to Kafka.
// It is used to decouple the DLQ from concrete Kafka producer implementations.
type Publisher interface {
	// Send publishes a message to Kafka with optional headers.
	Send(ctx context.Context, key, value []byte, headers ...kafka.Header) error
}

// DLQ represents a Dead Letter Queue client that captures failed messages
// and publishes them to a dedicated error topic for later analysis.
type DLQ struct {
	producer Publisher
	logger   logger.Logger
}

// New creates a new DLQ instance with the given publisher and logger.
// The publisher must be configured to send messages to the DLQ topic.
func New(producer Publisher, logger logger.Logger) *DLQ {
	return &DLQ{producer: producer, logger: logger}
}

// PublishError serializes the original Kafka message, error, and metadata into a structured JSON payload,
// then sends it to the DLQ topic. The message value is safely encoded in base64 to support binary data.
// If JSON marshaling fails, a fallback plain-text payload is used to prevent total data loss.
// Returns an error if sending to Kafka fails.
func (d *DLQ) PublishError(ctx context.Context, msg kafka.Message, err error, attempt int) error {
	const op = "dlq.PublishError"

	payload := map[string]any{
		"original_topic": msg.Topic,
		"error":          err.Error(),
		"attempt":        attempt,
		"timestamp":      time.Now().UTC(),
		"data_base64":    base64.StdEncoding.EncodeToString(msg.Value),
	}

	val, errMarshal := json.Marshal(payload)
	if errMarshal != nil {
		d.logger.LogAttrs(ctx, logger.ErrorLevel, "failed to marshal dlq payload",
			logger.String("op", op),
			logger.Any("err", errMarshal),
		)

		fallbackData := fmt.Sprintf(`{"status":"marshal_error","raw_data":"%s","error":"%s"}`,
			string(msg.Value), err.Error())
		val = []byte(fallbackData)
	}

	if errSend := d.producer.Send(ctx, msg.Key, val); errSend != nil {
		return fmt.Errorf("%s: send to kafka: %w", op, errSend)
	}

	return nil
}
