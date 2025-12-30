package transaction

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

// Option represents a functional configuration option for the transaction manager.
type Option func(*manager)

// MaxAttempts sets the maximum number of transaction execution attempts, including the first try.
// The value must be greater than zero.
func MaxAttempts(attempts int) Option {
	return func(m *manager) {
		m.maxAttempts = attempts
	}
}

// BaseRetryDelay sets the initial delay for the exponential backoff retry logic
// between transaction attempts. The value must be greater than zero.
func BaseRetryDelay(delay time.Duration) Option {
	return func(m *manager) {
		m.baseRetryDelay = delay
	}
}

// MaxRetryDelay sets the upper bound for retry delays between transaction attempts.
// The backoff delay will never exceed this value. The value must be greater than zero
// and greater than or equal to BaseRetryDelay.
func MaxRetryDelay(delay time.Duration) Option {
	return func(m *manager) {
		m.maxRetryDelay = delay
	}
}

// validate checks that all transaction manager configuration parameters are valid.
// It returns an error if any parameter violates its constraints.
func (m *manager) validate() error {
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
