package pgx

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/repository/postgres/pool"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	*pgxpool.Pool
	opTimeout time.Duration
}

func NewPool(ctx context.Context, config Config) (*Pool, error) {
	conntectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database)
	pgxconfig, err := pgxpool.ParseConfig(conntectionString)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string '%s': %w", conntectionString, err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, pgxconfig)
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging connection pool: %w", err)
	}
	return &Pool{
		Pool:      pool,
		opTimeout: config.Timeout,
	}, nil
}

func (p *Pool) OpTimeout() time.Duration {
	return p.opTimeout
}

func (p *Pool) Query(ctx context.Context, sql string, args ...any) (pool.Rows, error) {
	rows, err := p.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query '%s': %w", sql, err)
	}
	return pgxRows{rows}, nil
}

func (p *Pool) QueryRow(ctx context.Context, sql string, args ...any) pool.Row {
	row := p.Pool.QueryRow(ctx, sql, args...)
	return pgxRow{row}
}

func (p *Pool) Exec(ctx context.Context, sql string, arguments ...any) (pool.CommandTag, error) {
	tag, err := p.Pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return nil, fmt.Errorf("error executing query '%s': %w", sql, err)
	}
	return pgxCommandTag{tag}, nil
}

func (p *Pool) Begin(ctx context.Context) (pool.Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &pgxTx{tx}, nil
}

type pgxTx struct {
	pgx.Tx
}

func (t *pgxTx) Exec(ctx context.Context, sql string, args ...any) (pool.CommandTag, error) {
	tag, err := t.Tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgxCommandTag{tag}, nil
}

func (t *pgxTx) Query(ctx context.Context, sql string, args ...any) (pool.Rows, error) {
	rows, err := t.Tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgxRows{rows}, nil
}

func (t *pgxTx) QueryRow(ctx context.Context, sql string, args ...any) pool.Row {
	row := t.Tx.QueryRow(ctx, sql, args...)
	return pgxRow{row}
}

func (t *pgxTx) Commit(ctx context.Context) error {
	return t.Tx.Commit(ctx)
}

func (t *pgxTx) Rollback(ctx context.Context) error {
	return t.Tx.Rollback(ctx)
}
