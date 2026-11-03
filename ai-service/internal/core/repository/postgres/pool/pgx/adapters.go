package pgx

import (
	"errors"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/repository/postgres/pool"
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
	err := r.Row.Scan(dest...)
	if err != nil {
		return mapErrors(err)
	}
	return nil
}

type pgxCommandTag struct {
	pgconn.CommandTag
}

func mapErrors(err error) error {
	const (
		pgxViolatesForeignKeyErrorCode = "23503"
		pgxUniqueViolationErrorCode    = "23505"
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return pool.ErrNoRows
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgxViolatesForeignKeyErrorCode:
			return pool.ErrViolatesForeignKey
		case pgxUniqueViolationErrorCode:
			return pool.ErrConflict
		}
	}
	return fmt.Errorf("%v: %w", pool.ErrUnknown, err)
}
