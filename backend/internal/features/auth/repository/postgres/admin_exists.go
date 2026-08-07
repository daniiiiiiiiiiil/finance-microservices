package postgres_auth

import (
	"context"
	"fmt"
)

func (r *AuthRepository) AdminExists(ctx context.Context) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM finance.users WHERE is_admin = true`
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check admin exists: %w", err)
	}
	return count > 0, nil
}
