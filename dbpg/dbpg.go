package dbpg

import (
	"context"
	"database/sql"
	"time"

	"github.com/pozedorum/wbf/retry"

	_ "github.com/lib/pq"
)

type DB struct {
	Master *sql.DB
	Slaves []*sql.DB
}

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

func New(masterDSN string, slaveDSNs []string, opts *Options) (*DB, error) {
	master, err := sql.Open("postgres", masterDSN)
	if err != nil {
		return nil, err
	}
	applyOptions(master, opts)
	var slaves []*sql.DB
	for _, dsn := range slaveDSNs {
		slave, err := sql.Open("postgres", dsn)
		if err != nil {
			return nil, err
		}
		applyOptions(slave, opts)
		slaves = append(slaves, slave)
	}
	return &DB{Master: master, Slaves: slaves}, nil
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if len(db.Slaves) > 0 {
		// Простейший round-robin
		return db.Slaves[0].QueryContext(ctx, query, args...)
	}
	return db.Master.QueryContext(ctx, query, args...)
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.Master.ExecContext(ctx, query, args...)
}

func (db *DB) ExecWithRetry(ctx context.Context, strat retry.Strategy, query string, args ...interface{}) (sql.Result, error) {
	var res sql.Result
	err := retry.Do(func() error {
		r, e := db.ExecContext(ctx, query, args...)
		if e == nil {
			res = r
		}
		return e
	}, strat)
	return res, err
}

func (db *DB) QueryWithRetry(ctx context.Context, strat retry.Strategy, query string, args ...interface{}) (*sql.Rows, error) {
	var rows *sql.Rows
	err := retry.Do(func() error {
		r, e := db.QueryContext(ctx, query, args...)
		if e == nil {
			rows = r
		}
		return e
	}, strat)
	return rows, err
}

func (db *DB) BatchExec(ctx context.Context, in <-chan string) {
	go func() {
		for query := range in {
			_, _ = db.ExecContext(ctx, query) // Ошибки можно логировать
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
}
