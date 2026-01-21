package pgxdriver

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// QueryExecuter defines a unified interface for executing SQL queries and commands.
// It is implemented by both the main Postgres client and transaction wrappers,
// enabling seamless use of the same logic in and outside of transactions.
type QueryExecuter interface {
	// Query executes a query that returns rows, such as a SELECT.
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)

	// QueryRow executes a query that is expected to return at most one row.
	// It is safe to call Scan on the returned Row even if no row is found.
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

	// Exec executes a query that does not return rows, such as INSERT, UPDATE, or DELETE.
	// Returns the command tag and any error.
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)

	// SendBatch sends a batch of queries to the server in a single round trip.
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults

	// CopyFrom performs a PostgreSQL COPY FROM operation for high-performance bulk inserts.
	// Returns the number of rows copied.
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

// Query executes a query that returns rows, such as a SELECT.
// Delegates to the underlying pgxpool.Pool.
func (p *Postgres) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return p.Pool.Query(ctx, sql, args...)
}

// QueryRow executes a query expected to return at most one row.
// Delegates to the underlying pgxpool.Pool.
func (p *Postgres) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.Pool.QueryRow(ctx, sql, args...)
}

// Exec executes a non-query SQL statement (e.g., INSERT, UPDATE, DELETE).
// Delegates to the underlying pgxpool.Pool.
func (p *Postgres) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return p.Pool.Exec(ctx, sql, args...)
}

// SendBatch sends a batch of queries to the server using pgx's batch protocol.
// Delegates to the underlying pgxpool.Pool.
func (p *Postgres) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return p.Pool.SendBatch(ctx, b)
}

// CopyFrom performs a high-performance bulk insert using PostgreSQL's COPY FROM protocol.
// Delegates to the underlying pgxpool.Pool.
func (p *Postgres) CopyFrom(
	ctx context.Context,
	tableName pgx.Identifier,
	columnNames []string,
	rowSrc pgx.CopyFromSource,
) (int64, error) {
	return p.Pool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

// TxQueryExecuter wraps a pgx.Tx to satisfy the QueryExecuter interface,
// allowing transactional and non-transactional code to share the same execution path.
type TxQueryExecuter struct {
	Tx pgx.Tx
}

// Query executes a query within a transaction.
func (t *TxQueryExecuter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return t.Tx.Query(ctx, sql, args...)
}

// QueryRow executes a single-row query within a transaction.
func (t *TxQueryExecuter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return t.Tx.QueryRow(ctx, sql, args...)
}

// Exec executes a non-query statement within a transaction.
func (t *TxQueryExecuter) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return t.Tx.Exec(ctx, sql, args...)
}

// SendBatch sends a batch of queries within a transaction.
func (t *TxQueryExecuter) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return t.Tx.SendBatch(ctx, b)
}

// CopyFrom performs a COPY FROM operation within a transaction.
func (t *TxQueryExecuter) CopyFrom(
	ctx context.Context,
	tableName pgx.Identifier,
	columnNames []string,
	rowSrc pgx.CopyFromSource,
) (int64, error) {
	return t.Tx.CopyFrom(ctx, tableName, columnNames, rowSrc)
}
