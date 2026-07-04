// Package postgresrepo
package postgresrepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTransactionsUnavailable = errors.New("postgres transactions unavailable")

type queryer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repo struct {
	db   queryer
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{db: pool, pool: pool}
}

func (r *Repo) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return r.db.Exec(ctx, sql, arguments...)
}

func (r *Repo) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return r.db.Query(ctx, sql, args...)
}

func (r *Repo) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return r.db.QueryRow(ctx, sql, args...)
}

func (r *Repo) WithTx(ctx context.Context, fn func(*Repo) error) error {
	if r.pool == nil {
		return ErrTransactionsUnavailable
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	txRepo := &Repo{db: tx, pool: r.pool}
	if err := fn(txRepo); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
