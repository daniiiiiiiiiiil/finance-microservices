package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/repository/postgres/pool"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
)

func (r *Repository) CreateAccount(ctx context.Context, account domain.TelegramAccount) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO telegram.accounts (telegram_user_id, chat_id, user_id, username)
		OVERRIDING SYSTEM VALUE
		VALUES ($1, $2, $3, $4)`

	_, err := r.pool.Exec(ctx, query,
		account.TelegramUserID,
		account.ChatID,
		account.UserID,
		account.Username,
	)
	if err != nil {
		if errors.Is(err, pool.ErrUniqueViolation) {
			return fmt.Errorf("telegram account already linked: %w", apperrors.ErrConflict)
		}
		return fmt.Errorf("create telegram account: %w", err)
	}
	return nil
}

func (r *Repository) GetAccountByTelegramID(ctx context.Context, telegramUserID int64) (domain.TelegramAccount, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	return r.getAccount(ctx, "telegram_user_id = $1", telegramUserID)
}

func (r *Repository) GetAccountByUserID(ctx context.Context, userID int) (domain.TelegramAccount, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	return r.getAccount(ctx, "user_id = $1", userID)
}

func (r *Repository) getAccount(ctx context.Context, condition string, value any) (domain.TelegramAccount, error) {
	query := `
		SELECT telegram_user_id, chat_id, user_id, username, created_at, updated_at
		FROM telegram.accounts
		WHERE ` + condition

	var model accountModel
	err := r.pool.QueryRow(ctx, query, value).Scan(
		&model.TelegramUserID,
		&model.ChatID,
		&model.UserID,
		&model.Username,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.TelegramAccount{}, fmt.Errorf("telegram account: %w", apperrors.ErrNotFound)
		}
		return domain.TelegramAccount{}, fmt.Errorf("get telegram account: %w", err)
	}
	return accountDomainFromModel(model), nil
}

func (r *Repository) DeleteAccountByTelegramID(ctx context.Context, telegramUserID int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tag, err := r.pool.Exec(ctx, `DELETE FROM telegram.accounts WHERE telegram_user_id = $1`, telegramUserID)
	if err != nil {
		return fmt.Errorf("delete telegram account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("telegram account: %w", apperrors.ErrNotFound)
	}
	return nil
}

func (r *Repository) DeleteAccountByUserID(ctx context.Context, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tag, err := r.pool.Exec(ctx, `DELETE FROM telegram.accounts WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete telegram account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("telegram account: %w", apperrors.ErrNotFound)
	}
	return nil
}
