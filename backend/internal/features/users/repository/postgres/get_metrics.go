package postgres

import (
	"context"
	"fmt"
)

func (r *UserRepository) GetTotalUsers(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM finance.users`).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return total, nil
}
