package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/repository/postgres/pool"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
)

func (r *Repository) CreateLinkCode(ctx context.Context, code domain.LinkCode) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	if _, err := r.pool.Exec(ctx, `DELETE FROM telegram.link_codes WHERE expires_at < NOW()`); err != nil {
		return fmt.Errorf("cleanup expired link codes: %w", err)
	}

	query := `
		INSERT INTO telegram.link_codes (user_id, code_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query, code.UserID, code.CodeHash, code.ExpiresAt).Scan(&code.ID, &code.CreatedAt)
	if err != nil {
		if errors.Is(err, pool.ErrUniqueViolation) {
			return fmt.Errorf("link code collision: %w", apperrors.ErrConflict)
		}
		return fmt.Errorf("create link code: %w", err)
	}
	return nil
}

func (r *Repository) GetLinkCodeByHash(ctx context.Context, codeHash string) (domain.LinkCode, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, user_id, code_hash, expires_at, used_at, attempts, created_at
		FROM telegram.link_codes
		WHERE code_hash = $1`

	var model linkCodeModel
	err := r.pool.QueryRow(ctx, query, codeHash).Scan(
		&model.ID,
		&model.UserID,
		&model.CodeHash,
		&model.ExpiresAt,
		&model.UsedAt,
		&model.Attempts,
		&model.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.LinkCode{}, fmt.Errorf("link code: %w", apperrors.ErrNotFound)
		}
		return domain.LinkCode{}, fmt.Errorf("get link code: %w", err)
	}
	return linkCodeDomainFromModel(model), nil
}

func (r *Repository) MarkLinkCodeUsed(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tag, err := r.pool.Exec(ctx, `UPDATE telegram.link_codes SET used_at = NOW() WHERE id = $1 AND used_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("mark link code used: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("link code already used: %w", apperrors.ErrConflict)
	}
	return nil
}

func (r *Repository) IncrementLinkCodeAttempts(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	if _, err := r.pool.Exec(ctx, `UPDATE telegram.link_codes SET attempts = attempts + 1 WHERE id = $1`, id); err != nil {
		return fmt.Errorf("increment link code attempts: %w", err)
	}
	return nil
}
