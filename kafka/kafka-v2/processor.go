// Package kafkav2 provides a robust Kafka client implementation with structured logging,
// retry-capable message processing, and Dead Letter Queue (DLQ) integration.
package kafkav2

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/kafka/dlq"
	"github.com/wb-go/wbf/logger"
)

const (
	_defaultMaxAttempts    = 3
	_defaultBaseRetryDelay = 10 * time.Millisecond
	_defaultMaxRetryDelay  = 100 * time.Millisecond

	_backoffMultiplier = 2
)

// Handler is a function type that processes a single Kafka message.
// It receives the message and a context, and returns an error if processing fails.
// Returning nil signals successful processing and triggers offset commit.
type Handler func(ctx context.Context, msg kafka.Message) error

// Processor manages the lifecycle of Kafka message processing,
// including retry attempts, backoff delays, and DLQ fallback.
type Processor struct {
	consumer *Consumer
	dlq      *dlq.DLQ
	logger   logger.Logger

	maxAttempts    int
	baseRetryDelay time.Duration
	maxRetryDelay  time.Duration
}

// NewProcessor creates a new message processor with the given consumer, DLQ client, and logger.
// It applies optional configuration via functional options and validates the resulting settings.
// Returns an error if validation fails.
func NewProcessor(c *Consumer, d *dlq.DLQ, logger logger.Logger, opts ...ProcessorOption) (*Processor, error) {
	p := &Processor{
		consumer:       c,
		dlq:            d,
		logger:         logger,
		maxAttempts:    _defaultMaxAttempts,
		baseRetryDelay: _defaultBaseRetryDelay,
		maxRetryDelay:  _defaultMaxRetryDelay,
	}

	for _, opt := range opts {
		opt(p)
	}

	if err := p.validate(); err != nil {
		return nil, fmt.Errorf("kafka.kafka-v2.NewProcessor: validation: %w", err)
	}

	return p, nil
}

// Start launches a background goroutine that continuously fetches and processes Kafka messages.
// Processing respects the provided context for graceful shutdown.
// The method returns immediately; errors during message processing are logged but not returned.
func (p *Processor) Start(ctx context.Context, handler Handler) {
	go func() {
		for {
			msg, err := p.consumer.Fetch(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				p.logger.LogAttrs(ctx, logger.ErrorLevel, "fetch error",
					logger.Any("error", err),
				)
				continue
			}

			p.processWithRetry(ctx, msg, handler)
		}
	}()
}

// processWithRetry executes the handler up to maxAttempts times with exponential backoff and jitter.
// If all retries fail and a DLQ is configured, the message is published to the DLQ.
// Regardless of DLQ outcome, the message offset is committed to prevent infinite reprocessing.
func (p *Processor) processWithRetry(ctx context.Context, msg kafka.Message, handler Handler) {
	var lastErr error

	currentBackoff := p.baseRetryDelay

	for attempt := 1; attempt <= p.maxAttempts; attempt++ {
		lastErr = handler(ctx, msg)
		if lastErr == nil {
			if err := p.consumer.Commit(ctx, msg); err != nil {
				p.logger.LogAttrs(ctx, logger.ErrorLevel, "failed to commit message offset",
					logger.Int64("offset", msg.Offset),
					logger.String("topic", msg.Topic),
					logger.Any("error", err),
				)
			}
			return
		}

		p.logger.LogAttrs(ctx, logger.WarnLevel, "retryable error",
			logger.Int("attempt", attempt),
			logger.Any("err", lastErr),
		)

		if attempt >= p.maxAttempts {
			break
		}
		//nolint:gosec
		jitter := min(time.Duration(
			rand.Int64N(int64(currentBackoff*_backoffMultiplier)),
		), p.maxRetryDelay)

		select {
		case <-time.After(jitter):
		case <-ctx.Done():
			return
		}

		nextBackoff := min(currentBackoff*_backoffMultiplier, p.maxRetryDelay)
		currentBackoff = nextBackoff
	}

	if p.dlq != nil {
		if err := p.dlq.PublishError(ctx, msg, lastErr, p.maxAttempts); err != nil {
			p.logger.LogAttrs(ctx, logger.ErrorLevel, "DLQ unavailable, skipping commit to prevent data loss",
				logger.Any("err", err),
			)
			return
		}
	}

	if err := p.consumer.Commit(ctx, msg); err != nil {
		p.logger.LogAttrs(ctx, logger.ErrorLevel, "final commit error",
			logger.Any("err", err),
		)
	}
}
