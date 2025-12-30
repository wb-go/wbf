package pgxdriver

import (
	"errors"
	"time"
)

var (
	// ErrInvalidMaxPoolSize is returned when MaxPoolSize <= 0.
	ErrInvalidMaxPoolSize = errors.New("invalid maxPoolSize: must be > 0")
	// ErrInvalidConnAttempts is returned when MaxConnAttempts <= 0.
	ErrInvalidConnAttempts = errors.New("invalid connAttempts: must be > 0")
	// ErrInvalidBaseRetryDelay is returned when BaseRetryDelay <= 0.
	ErrInvalidBaseRetryDelay = errors.New("invalid base retry delay: must be > 0")
	// ErrInvalidMaxRetryDelay is returned when MaxRetryDelay <= 0.
	ErrInvalidMaxRetryDelay = errors.New("invalid max retry delay: must be > 0")
	// ErrBaseExceedsMaxDelay is returned when BaseRetryDelay > MaxRetryDelay.
	ErrBaseExceedsMaxDelay = errors.New("baseRetryDelay cannot exceed maxRetryDelay")
)

// Option represents a functional configuration option for the Postgres client.
// Use Option functions like MaxPoolSize or BaseRetryDelay to customize client behavior.
type Option func(*Postgres)

// MaxPoolSize sets the maximum number of concurrent connections in the connection pool.
// The size must be greater than zero.
func MaxPoolSize(size int32) Option {
	return func(p *Postgres) {
		p.maxPoolSize = size
	}
}

// MaxConnAttempts sets the maximum number of attempts to establish a database connection
// during client initialization. The value must be greater than zero.
func MaxConnAttempts(attempts int) Option {
	return func(p *Postgres) {
		p.connAttempts = attempts
	}
}

// BaseRetryDelay sets the initial delay for the exponential backoff retry logic
// when connecting to the database. The value must be greater than zero.
func BaseRetryDelay(delay time.Duration) Option {
	return func(p *Postgres) {
		p.baseRetryDelay = delay
	}
}

// MaxRetryDelay sets the upper bound for retry delays during connection attempts.
// The backoff delay will never exceed this value. The value must be greater than zero
// and greater than or equal to BaseRetryDelay.
func MaxRetryDelay(delay time.Duration) Option {
	return func(p *Postgres) {
		p.maxRetryDelay = delay
	}
}

// validate checks that all Postgres client configuration parameters are valid.
// It returns an error if any parameter violates its constraints.
func (p *Postgres) validate() error {
	if p.maxPoolSize <= 0 {
		return ErrInvalidMaxPoolSize
	}

	if p.connAttempts <= 0 {
		return ErrInvalidConnAttempts
	}

	if p.baseRetryDelay <= 0 {
		return ErrInvalidBaseRetryDelay
	}

	if p.maxRetryDelay <= 0 {
		return ErrInvalidMaxRetryDelay
	}

	if p.baseRetryDelay > p.maxRetryDelay {
		return ErrBaseExceedsMaxDelay
	}
	return nil
}
