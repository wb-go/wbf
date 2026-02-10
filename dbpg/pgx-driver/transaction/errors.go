package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
)

var (
	// ErrMaxRetriesExceeded is returned when a transaction fails after exhausting all retry attempts.
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
	// ErrTransactionTimeout is returned when a transaction exceeds its context deadline.
	ErrTransactionTimeout = errors.New("transaction timeout")
	// ErrConflictingData indicates a unique constraint violation (PostgreSQL error code 23505).
	// It is typically returned when inserting or updating data that conflicts with an existing unique row.
	ErrConflictingData = errors.New("data conflicts with existing data in unique column")
	// ErrInvalidData indicates a foreign key violation or other referential integrity error (PostgreSQL error code 23503).
	ErrInvalidData = errors.New("invalid data")
)

// HandleError wraps a raw error with contextual information and maps PostgreSQL error codes
// to semantic, user-friendly error types.
// It takes the operation name (e.g., "transfer_funds"), a step description (e.g., "execute"),
// and the original error, then returns a wrapped error suitable for logging or upstream handling.
// If the input error is nil, it returns nil.
func HandleError(operation, step string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s: %s: timeout: %w", operation, step, ErrTransactionTimeout)
	}

	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%s: %s: canceled: %w", operation, step, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40P01":
			return fmt.Errorf("%s: %s: deadlock: %w", operation, step, err)
		case "40001":
			return fmt.Errorf("%s: %s: serialization failure: %w", operation, step, err)
		case "57014":
			return fmt.Errorf("%s: %s: statement timeout: %w", operation, step, err)
		case "55P03":
			return fmt.Errorf("%s: %s: lock timeout: %w", operation, step, err)
		case "23505":
			return fmt.Errorf(
				"%s: %s: unique constraint violation: %w",
				operation,
				step,
				ErrConflictingData,
			)
		case "23503":
			return fmt.Errorf(
				"%s: %s: foreign key violation: %w",
				operation,
				step,
				ErrInvalidData,
			)
		}
	}

	if errors.Is(err, ErrMaxRetriesExceeded) {
		return fmt.Errorf("%s: %s: max retries exceeded: %w", operation, step, err)
	}

	if errors.Is(err, ErrTransactionTimeout) {
		return fmt.Errorf("%s: %s: transaction timeout: %w", operation, step, err)
	}

	return fmt.Errorf("%s: %s: %w", operation, step, err)
}
