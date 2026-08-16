package postgres_auth

import (
	"context"
	"fmt"
)

func (r *AuthRepository) AdminExists(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var count int
	err := r.pool.QueryRow(ctx, `
        SELECT COUNT(*) 
        FROM users.users 
        WHERE is_admin = true
    `).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check admin exists: %w", err)
	}
	return count > 0, nil
}
