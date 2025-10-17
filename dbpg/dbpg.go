// Package dbpg provides PostgreSQL connection management with master-slave support.
package dbpg

import (
	"context"
	"database/sql"
	"time"

	// Register PostgreSQL driver for database/sql.
	_ "github.com/lib/pq"

	"github.com/wb-go/wbf/retry"
)

// DB represents a database connection with master and slave nodes.
type DB struct {
	balancer *balancer

	Master *sql.DB
	Slaves []*sql.DB
}

// Options defines database connection configuration options.
type Options struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func applyOptions(db *sql.DB, opts *Options) {
	if opts == nil {
		return
	}
	if opts.MaxOpenConns > 0 {
		db.SetMaxOpenConns(opts.MaxOpenConns)
	}
	if opts.MaxIdleConns > 0 {
		db.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(opts.ConnMaxLifetime)
	}
}

// New creates a new DB instance with master and slave connections.
func New(masterDSN string, slaveDSNs []string, opts *Options) (*DB, error) {
	master, err := sql.Open("postgres", masterDSN)
	if err != nil {
		return nil, err
	}
	applyOptions(master, opts)

	// Preallocate memory for better performance.
	slaves := make([]*sql.DB, 0, len(slaveDSNs))
	for _, dsn := range slaveDSNs {
		slave, err := sql.Open("postgres", dsn)
		if err != nil {
			return nil, err
		}
		applyOptions(slave, opts)
		slaves = append(slaves, slave)
	}

	// Create balancer.
	balancer := newBalancer(len(slaveDSNs))

	return &DB{Master: master, Slaves: slaves, balancer: balancer}, nil
}

// QueryContext executes a query on a slave if available, otherwise on the master.
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.selectDB().QueryContext(ctx, query, args...)
}

// QueryRowContext executes a single-row query on a slave if available, otherwise on the master.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.selectDB().QueryRowContext(ctx, query, args...)
}

// ExecContext executes a command on the master database.
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.Master.ExecContext(ctx, query, args...)
}

// ExecWithRetry executes a command with a retry strategy.
func (db *DB) ExecWithRetry(
	ctx context.Context,
	strategy retry.Strategy,
	query string,
	args ...interface{},
) (sql.Result, error) {
	var res sql.Result
	err := retry.Do(func() error {
		r, e := db.ExecContext(ctx, query, args...)
		if e == nil {
			res = r
		}
		return e
	}, strategy)
	return res, err
}

// QueryWithRetry executes a query with a retry strategy.
func (db *DB) QueryWithRetry(
	ctx context.Context,
	strategy retry.Strategy,
	query string,
	args ...interface{},
) (*sql.Rows, error) {
	var rows *sql.Rows
	err := retry.Do(func() error {
		r, e := db.QueryContext(ctx, query, args...)
		if e == nil {
			if rowsErr := r.Err(); rowsErr != nil {
				defer func() {
					_ = r.Close()
				}()
				return rowsErr
			}
			rows = r
		}
		return e
	}, strategy)

	return rows, err
}

// QueryRowWithRetry executes a single-row query with a retry strategy.
func (db *DB) QueryRowWithRetry(
	ctx context.Context,
	strategy retry.Strategy,
	query string,
	args ...interface{},
) (*sql.Row, error) {
	var row *sql.Row
	err := retry.Do(func() error {
		r := db.QueryRowContext(ctx, query, args...)
		row = r
		return r.Err()
	}, strategy)

	return row, err
}

// BatchExec executes multiple queries asynchronously in a batch.
func (db *DB) BatchExec(ctx context.Context, in <-chan string) {
	go func() {
		for query := range in {
			_, _ = db.ExecContext(ctx, query) // Errors can be logged if needed.
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
}

// selectDB returns a database for query execution: slave (round-robin) or master.
func (db *DB) selectDB() *sql.DB {
	if len(db.Slaves) > 0 {
		// Select a slave using balancer.
		return db.Slaves[db.balancer.index()]
	}

	return db.Master
}
