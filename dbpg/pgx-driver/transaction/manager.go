// Package transaction provides a transaction manager for PostgreSQL with automatic retry
// on transient errors such as serialization failures, deadlocks, and connection issues.
package transaction

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"
	"github.com/wb-go/wbf/logger"
)

const (
	_defaultMaxAttempts    = 3
	_defaultBaseRetryDelay = 10 * time.Millisecond
	_defaultMaxRetryDelay  = 100 * time.Millisecond

	_backoffMultiplier = 2
)

// Manager defines the interface for executing functions within a retriable database transaction.
type Manager interface {
	// ExecuteInTransaction runs the given function inside a PostgreSQL transaction.
	// If the transaction fails due to a retryable error (e.g., serialization failure, deadlock),
	// it will be retried up to maxAttempts times with exponential backoff and jitter.
	// The tsName parameter is used for logging and observability.
	// Returns the last error if all attempts fail, or nil on success.
	ExecuteInTransaction(
		ctx context.Context,
		tsName string,
		fn func(tx pgxdriver.QueryExecuter) error,
	) error
}

// manager is the internal implementation of the Manager interface.
type manager struct {
	pool   *pgxdriver.Postgres
	logger logger.Logger

	maxAttempts    int
	baseRetryDelay time.Duration
	maxRetryDelay  time.Duration
}

// NewManager creates a new transaction manager configured with the given PostgreSQL client and logger.
// It applies optional configuration via functional options.
// Returns an error if validation of options fails.
func NewManager(
	pool *pgxdriver.Postgres,
	logger logger.Logger,
	opts ...Option,
) (Manager, error) {
	tm := &manager{
		pool:   pool,
		logger: logger,

		maxAttempts:    _defaultMaxAttempts,
		baseRetryDelay: _defaultBaseRetryDelay,
		maxRetryDelay:  _defaultMaxRetryDelay,
	}

	for _, opt := range opts {
		opt(tm)
	}
	if err := tm.validate(); err != nil {
		return nil, fmt.Errorf("dbpg.pgx-driver.transaction.NewManager: %w", err)
	}

	return tm, nil
}

// ExecuteInTransaction executes the provided function within a retriable PostgreSQL transaction.
func (tm *manager) ExecuteInTransaction(
	ctx context.Context,
	tsName string,
	fn func(tx pgxdriver.QueryExecuter) error,
) error {
	const op = "dbpg.pgx-driver.transaction.ExecuteInTransaction"
	var lastErr error
	currentBackoff := tm.baseRetryDelay

	for attempt := 1; attempt <= tm.maxAttempts; attempt++ {
		err := tm.doTransaction(ctx, tsName, fn)
		if err == nil {
			return nil
		}

		lastErr = err

		if !isRetryableError(err) || attempt == tm.maxAttempts {
			return err
		}
		//nolint:gosec
		jitter := min(time.Duration(
			rand.Int64N(int64(currentBackoff*_backoffMultiplier)),
		), tm.maxRetryDelay)

		tm.logger.LogAttrs(ctx, logger.WarnLevel, "retrying transaction",
			logger.String("op", op),
			logger.String("transaction", tsName),
			logger.Int("attempt", attempt),
			logger.Int("max_attempts", tm.maxAttempts),
			logger.String("retry_after", jitter.String()),
			logger.Any("error", lastErr),
		)

		select {
		case <-time.After(jitter):
			currentBackoff = min(currentBackoff*_backoffMultiplier, tm.maxRetryDelay)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return fmt.Errorf("%s: %s: %w", op, tsName, lastErr)
}

// doTransaction executes a single transaction attempt: begins, runs the user function, and commits.
// On error, the transaction is rolled back automatically.
func (tm *manager) doTransaction(ctx context.Context, tsName string, fn func(tx pgxdriver.QueryExecuter) error) error {
	tx, err := tm.pool.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer tm.safelyRollback(ctx, tx, tsName)

	if err := fn(&pgxdriver.TxQueryExecuter{Tx: tx}); err != nil {
		return HandleError(tsName, "execute", err)
	}

	return tx.Commit(ctx)
}

// safelyRollback attempts to roll back the transaction and logs only unexpected errors.
// It suppresses pgx.ErrTxClosed, which is normal when the transaction was already committed.
func (tm *manager) safelyRollback(ctx context.Context, tx pgx.Tx, tsName string) {
	const op = "dbpg.pgx-driver.transaction.safelyRollback"

	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		tm.logger.LogAttrs(ctx, logger.ErrorLevel, "rollback failed",
			logger.String("op", op),
			logger.String("transaction", tsName),
			logger.Any("error", err),
		)
	}
}

// isRetryableError determines whether a PostgreSQL error is transient and safe to retry.
// It includes serialization failures (40001), deadlocks (40P01), and various connection errors.
func isRetryableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40P01", "40001", "08000", "08003", "08006", "08001", "08004", "08007", "08P01":
			return true
		}
	}

	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled) {
		return false
	}

	if errors.Is(err, pgx.ErrTxClosed) {
		return true
	}

	return false
}
