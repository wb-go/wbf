// Package pgxdriver provides a robust PostgreSQL client built on pgx/v5,
// featuring connection retries with exponential backoff and jitter,
// integrated SQL query building via squirrel, and structured logging.
package pgxdriver

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wb-go/wbf/logger"
)

const (
	_defaultMaxPoolSize    = 100
	_defaultConnAttempts   = 10
	_defaultBaseRetryDelay = 100 * time.Millisecond
	_defaultMaxRetryDelay  = 5 * time.Second

	_backoffMultiplier = 2
)

// Postgres represents a PostgreSQL client with a connection pool and SQL builder.
// It wraps pgxpool.Pool and provides methods for querying, building SQL statements,
// and managing the lifecycle of the database connection.
type Postgres struct {
	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
	logger  logger.Logger

	connAttempts   int
	baseRetryDelay time.Duration
	maxRetryDelay  time.Duration
	maxPoolSize    int32
}

// New creates and initializes a new Postgres client.
// It parses the provided DSN, applies configuration options,
// and attempts to establish a connection pool with exponential backoff and jitter.
// Returns an error if validation fails, the DSN is invalid, or all connection attempts are exhausted.
func New(dsn string, logger logger.Logger, opts ...Option) (*Postgres, error) {
	const op = "dbpg.pgxdriver.New"

	pg := &Postgres{
		logger:         logger,
		connAttempts:   _defaultConnAttempts,
		baseRetryDelay: _defaultBaseRetryDelay,
		maxRetryDelay:  _defaultMaxRetryDelay,
		maxPoolSize:    _defaultMaxPoolSize,
	}

	for _, opt := range opts {
		opt(pg)
	}
	if err := pg.validate(); err != nil {
		return nil, fmt.Errorf("%s: validation: %w", op, err)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: parse pool config: %w", op, err)
	}

	poolConfig.MaxConns = pg.maxPoolSize

	currentBackoff := pg.baseRetryDelay
	for attemptCount := 1; attemptCount <= pg.connAttempts; attemptCount++ {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			return pg, nil
		}
		//nolint:gosec
		jitter := min(time.Duration(
			rand.Int64N(int64(currentBackoff*_backoffMultiplier)),
		), pg.maxRetryDelay)

		pg.logger.Info("postgresql connection attempt failed",
			"operation", op,
			"attempt", attemptCount,
			"retry_after", jitter.String(),
			"error", err,
		)

		time.Sleep(jitter)

		nextBackoff := min(currentBackoff*_backoffMultiplier, pg.maxRetryDelay)
		currentBackoff = nextBackoff
	}
	if err != nil {
		return nil, fmt.Errorf("%s: create new pool: %w", op, err)
	}

	pg.logger.Info("postgresql connection successful")

	return pg, nil
}

// Ping verifies the database connection by sending a lightweight ping request.
// It returns an error if the connection is not alive.
func (p *Postgres) Ping(ctx context.Context) error {
	return p.Pool.Ping(ctx)
}

// Close gracefully shuts down the connection pool and logs the shutdown process.
// It is safe to call Close multiple times.
func (p *Postgres) Close() {
	if p.Pool != nil {
		p.logger.Info("closing postgresql connection pool...")
		p.Pool.Close()
		p.logger.Info("postgresql connection pool closed")
	}
}

// Select starts a new SELECT query using the embedded squirrel builder.
// The returned builder supports chaining methods like From, Where, OrderBy, etc.
func (p *Postgres) Select(columns ...string) squirrel.SelectBuilder {
	return p.Builder.Select(columns...)
}

// Insert starts a new INSERT query using the embedded squirrel builder.
// The `into` parameter specifies the target table name.
func (p *Postgres) Insert(into string) squirrel.InsertBuilder {
	return p.Builder.Insert(into)
}

// Update starts a new UPDATE query using the embedded squirrel builder.
// The `table` parameter specifies the table to update.
func (p *Postgres) Update(table string) squirrel.UpdateBuilder {
	return p.Builder.Update(table)
}

// Delete starts a new DELETE query using the embedded squirrel builder.
// The `from` parameter specifies the table to delete from.
func (p *Postgres) Delete(from string) squirrel.DeleteBuilder {
	return p.Builder.Delete(from)
}
