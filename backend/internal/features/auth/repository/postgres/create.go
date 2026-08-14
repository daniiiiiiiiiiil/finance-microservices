package postgres_auth

import (
	"context"
	"fmt"
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
