package postgres_auth

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/domain"
)

func (r *AuthRepository) Create(ctx context.Context, email, passwordHash string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
        INSERT INTO auth.credentials (email, password_hash, status)
        VALUES ($1, $2, 'pending')
        RETURNING id`

	var id int
	err := r.pool.QueryRow(ctx, query, email, passwordHash).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert credentials: %w", err)
	}
	return id, nil
}

func (r *AuthRepository) GetByEmail(ctx context.Context, email string) (*domain.Credentials, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
        SELECT id, email, password_hash, status, created_at, updated_at
        FROM auth.credentials
        WHERE email = $1`

	var model AuthModel
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&model.ID,
		&model.Email,
		&model.PasswordHash,
		&model.Status,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}

	cred := model.ToDomain()
	return &cred, nil
}

func (r *AuthRepository) AdminUpdateStatus(ctx context.Context, id int, status string) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `UPDATE auth.credentials SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("could not update status: %w", err)
	}
	return nil
}
