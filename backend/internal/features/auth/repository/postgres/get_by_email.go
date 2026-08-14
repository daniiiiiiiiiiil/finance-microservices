package postgres_auth

import (
	"backend/internal/core/domain"
	"context"
	"fmt"
)

func (r *AuthRepository) GetByEmail(ctx context.Context, email string) (*domain.Credentials, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
        SELECT id, email, password_hash, status, created_at, updated_at
        FROM auth.credentials
        WHERE email = $1`

	var cred domain.Credentials
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&cred.ID,
		&cred.Email,
		&cred.PasswordHash,
		&cred.Status,
		&cred.CreatedAt,
		&cred.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &cred, nil
}
