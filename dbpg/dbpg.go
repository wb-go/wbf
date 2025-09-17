// Package dbpg предоставляет управление подключениями к PostgreSQL с поддержкой master-slave.
package dbpg

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq" // Драйвер PostgreSQL.
	"github.com/wb-go/wbf/retry"
)

// DB представляет подключение к базе данных с master и slave узлами.
type DB struct {
	balancer *balancer

	Master *sql.DB
	Slaves []*sql.DB
}

// Options содержит опции конфигурации подключения к базе данных.
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

// New создает новый экземпляр DB с master и slave подключениями.
func New(masterDSN string, slaveDSNs []string, opts *Options) (*DB, error) {
	master, err := sql.Open("postgres", masterDSN)
	if err != nil {
		return nil, err
	}
	applyOptions(master, opts)

	// Предварительно выделяем память для лучшей производительности.
	slaves := make([]*sql.DB, 0, len(slaveDSNs))
	for _, dsn := range slaveDSNs {
		slave, err := sql.Open("postgres", dsn)
		if err != nil {
			return nil, err
		}
		applyOptions(slave, opts)
		slaves = append(slaves, slave)
	}

	// Создаем balancer для использования slaves
	balancer := newBalancer(len(slaveDSNs))

	return &DB{Master: master, Slaves: slaves, balancer: balancer}, nil
}

func (db *DB) selectDB() *sql.DB {
	if len(db.Slaves) > 0 {
		return db.Slaves[db.balancer.index()]
	}

	return db.Master
}

// QueryContext выполняет запрос на slave если доступен, иначе на master.
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.selectDB().QueryContext(ctx, query, args...)
}

// QueryRowContext выполняет запрос на slave если доступен, иначе на master.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.selectDB().QueryRowContext(ctx, query, args...)
}

// ExecContext выполняет команду на master базе данных.
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.Master.ExecContext(ctx, query, args...)
}

// ExecWithRetry выполняет команду с стратегией повторных попыток.
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

// QueryWithRetry выполняет запрос с стратегией повторных попыток.
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

// QueryRowWithRetry выполняет запрос с стратегией повторных попыток.
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

// BatchExec выполняет несколько запросов пакетно асинхронно.
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
