package pgx

import (
	"errors"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/repository/postgres/pool"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgxRows struct {
	pgx.Rows
}

type pgxRow struct {
	pgx.Row
}

func (r pgxRow) Scan(dest ...any) error {
	if err := r.Row.Scan(dest...); err != nil {
		return mapErrors(err)
	}
	return nil
}

type pgxCommandTag struct {
	pgconn.CommandTag
}

func mapErrors(err error) error {
	const (
		uniqueViolation     = "23505"
		foreignKeyViolation = "23503"
	)
	if errors.Is(err, pool.ErrNoRows) {
		return pool.ErrNoRows
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return pool.ErrNoRows
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case uniqueViolation:
			return pool.ErrUniqueViolation
		case foreignKeyViolation:
			return pool.ErrViolatesForeignKey
		}
	}
	return fmt.Errorf("%v: %w", pool.ErrUnknown, err)
}
