package pgx

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/repository/postgres/pool"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	*pgxpool.Pool
	opTimeout time.Duration
}

func NewPool(ctx context.Context, config Config) (*Pool, error) {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database)
	return newPool(ctx, connectionString, config.Timeout, config.MaxConns)
}

func NewPoolWithDSN(ctx context.Context, dsn string, timeout time.Duration, maxConns int) (*Pool, error) {
	return newPool(ctx, dsn, timeout, maxConns)
}

func newPool(ctx context.Context, connectionString string, timeout time.Duration, maxConns int) (*Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("parse connection string: %w", err)
	}
	if maxConns > 0 {
		pgxConfig.MaxConns = int32(maxConns)
	}
	connectionPool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}
	if err := connectionPool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping connection pool: %w", err)
	}
	return &Pool{
		Pool:      connectionPool,
		opTimeout: timeout,
	}, nil
}

func (p *Pool) OpTimeout() time.Duration {
	return p.opTimeout
}

func (p *Pool) Query(ctx context.Context, sql string, args ...any) (pool.Rows, error) {
	rows, err := p.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	return pgxRows{rows}, nil
}

func (p *Pool) QueryRow(ctx context.Context, sql string, args ...any) pool.Row {
	return pgxRow{p.Pool.QueryRow(ctx, sql, args...)}
}

func (p *Pool) Exec(ctx context.Context, sql string, args ...any) (pool.CommandTag, error) {
	tag, err := p.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, mapErrors(err)
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
		return nil, mapErrors(err)
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
	return pgxRow{t.Tx.QueryRow(ctx, sql, args...)}
}

func (t *pgxTx) Commit(ctx context.Context) error {
	return t.Tx.Commit(ctx)
}

func (t *pgxTx) Rollback(ctx context.Context) error {
	return t.Tx.Rollback(ctx)
}
