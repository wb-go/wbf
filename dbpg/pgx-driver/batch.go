package pgxdriver

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// BatchInsert executes the same SQL statement multiple times with different parameters
// using pgx's batch protocol for improved performance over individual Exec calls.
// It accepts a slice of parameter rows and queues each as a separate batch operation.
// The function works with any QueryExecuter (e.g., *Postgres or *TxQueryExecuter),
// enabling use both inside and outside of transactions.
// Returns an error if any statement in the batch fails.
func BatchInsert(ctx context.Context, qe QueryExecuter, sql string, rows [][]any) error {
	const op = "dbpg.pgx-driver.BatchInsert"

	batch := &pgx.Batch{}
	for _, row := range rows {
		batch.Queue(sql, row...)
	}

	results := qe.SendBatch(ctx, batch)
	defer func() {
		_ = results.Close()
	}()

	for i := range rows {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("%s: executing statement at index %d: %w", op, i, err)
		}
	}

	return nil
}
