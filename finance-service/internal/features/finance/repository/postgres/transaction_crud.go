package postgres

import (
	"errors"
	"fmt"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/repository/postgres/pool"
	errors_my "github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/errors"
)

func (r *FinanceRepository) CreateTransactionTx(ctx context.Context, tx pool.Tx, transaction domain.Finance) (domain.Finance, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO finance.transactions (type_transaction, amount, category, created_at, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, version, type_transaction, amount, category, created_at, user_id
	`

	var model FinanceModel
	err := tx.QueryRow(ctx, query,
		transaction.TypeTransaction,
		transaction.Amount,
		transaction.Category,
		transaction.CreatedAt,
		transaction.UserID,
	).Scan(
		&model.ID,
		&model.Version,
		&model.TypeTransaction,
		&model.Amount,
		&model.Category,
		&model.CreatedAt,
		&model.UserID,
	)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("insert transaction: %w", err)
	}

	return financeDomainFromModel(model), nil
}

func (r *FinanceRepository) GetTransaction(ctx context.Context, id int) (domain.Finance, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, type_transaction, amount, category, created_at, user_id
		FROM finance.transactions
		WHERE id = $1
	`

	var model FinanceModel
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&model.ID,
		&model.Version,
		&model.TypeTransaction,
		&model.Amount,
		&model.Category,
		&model.CreatedAt,
		&model.UserID,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.Finance{}, fmt.Errorf("transaction with id %d: %w", id, errors_my.ErrNotFound)
		}
		return domain.Finance{}, fmt.Errorf("get transaction: %w", err)
	}

	return financeDomainFromModel(model), nil
}

func (r *FinanceRepository) UpdateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE finance.transactions SET
			type_transaction = $1,
			amount = $2,
			category = $3,
				version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING id, version, type_transaction, amount, category, created_at, user_id
	`

	var model FinanceModel
	err := r.pool.QueryRow(ctx, query,
		transaction.TypeTransaction,
		transaction.Amount,
		transaction.Category,
		transaction.ID,
		transaction.Version,
	).Scan(
		&model.ID,
		&model.Version,
		&model.TypeTransaction,
		&model.Amount,
		&model.Category,
		&model.CreatedAt,
		&model.UserID,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.Finance{}, fmt.Errorf("transaction with id='%d' concurrently accessed: %w", transaction.ID, errors_my.ErrConflict)
		}
		return domain.Finance{}, fmt.Errorf("update transaction: %w", err)
	}

	return financeDomainFromModel(model), nil
}

func (r *FinanceRepository) DeleteTransaction(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `DELETE FROM finance.transactions WHERE id = $1`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("transaction with id %d: %w", id, errors_my.ErrNotFound)
	}

	return nil
}
