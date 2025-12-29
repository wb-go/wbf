package kafkav2

import (
	"errors"
	"time"
)

var (
	// ErrInvalidMaxAttempts is returned when MaxPoolAttempts <= 0.
	ErrInvalidMaxAttempts = errors.New("invalid maxPoolAttempts: must be > 0")
	// ErrInvalidBaseRetryDelay is returned when BaseRetryDelay <= 0.
	ErrInvalidBaseRetryDelay = errors.New("invalid base retry delay: must be > 0")
	// ErrInvalidMaxRetryDelay is returned when MaxRetryDelay <= 0.
	ErrInvalidMaxRetryDelay = errors.New("invalid max retry delay: must be > 0")
	// ErrBaseExceedsMaxDelay is returned when BaseRetryDelay > MaxRetryDelay.
	ErrBaseExceedsMaxDelay = errors.New("baseRetryDelay cannot exceed maxRetryDelay")
)

// ProcessorOption represents a functional configuration option for the message processor.
// Use these options to customize retry behavior when creating a Processor instance.
type ProcessorOption func(*Processor)

// MaxAttempts sets the maximum number of processing attempts for a single Kafka message,
// including the initial attempt. The value must be greater than zero.
func MaxAttempts(attempts int) ProcessorOption {
	return func(m *Processor) {
		m.maxAttempts = attempts
	}
}

// BaseRetryDelay sets the initial delay for the exponential backoff retry logic
// between message processing attempts. The value must be greater than zero.
func BaseRetryDelay(delay time.Duration) ProcessorOption {
	return func(m *Processor) {
		m.baseRetryDelay = delay
	}
}

// MaxRetryDelay sets the upper bound for retry delays between message processing attempts.
// The backoff delay will never exceed this value. The value must be greater than zero
// and greater than or equal to BaseRetryDelay.
func MaxRetryDelay(delay time.Duration) ProcessorOption {
	return func(m *Processor) {
		m.maxRetryDelay = delay
	}
}

// validate checks that all Processor configuration parameters are valid.
// It returns an error if any parameter violates its constraints.
func (m *Processor) validate() error {
	if m.maxAttempts <= 0 {
		return ErrInvalidMaxAttempts
	}

	if m.baseRetryDelay <= 0 {
		return ErrInvalidBaseRetryDelay
	}

	if m.maxRetryDelay <= 0 {
		return ErrInvalidMaxRetryDelay
	}

	if m.baseRetryDelay > m.maxRetryDelay {
		return ErrBaseExceedsMaxDelay
	}
	return nil
}
