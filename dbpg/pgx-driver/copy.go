package pgxdriver

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// ErrInvalidTableName is returned when tableName is not string, []string, or pgx.Identifier.
var ErrInvalidTableName = errors.New("invalid table name type")

// BulkInsert performs a high-performance bulk insert into a PostgreSQL table using the COPY FROM protocol.
// It accepts the table name as a string, []string, or pgx.Identifier, column names, and a 2D slice of row data.
// The function works with any implementation of QueryExecuter (e.g., *Postgres or *TxQueryExecuter),
// making it usable both inside and outside of transactions.
// Returns the number of rows inserted and an error, if any occurred.
func BulkInsert(ctx context.Context, qe QueryExecuter, tableName any, columns []string, data [][]any) (int64, error) {
	const op = "pgxdriver.BulkInsert"

	var ident pgx.Identifier
	switch t := tableName.(type) {
	case string:
		ident = pgx.Identifier{t}
	case []string:
		ident = pgx.Identifier(t)
	case pgx.Identifier:
		ident = t
	default:
		return 0, fmt.Errorf("%w", ErrInvalidTableName)
	}

	count, err := qe.CopyFrom(
		ctx,
		ident,
		columns,
		pgx.CopyFromRows(data),
	)
	if err != nil {
		return 0, fmt.Errorf("%s: copy from: %w", op, err)
	}

	return count, nil
}
