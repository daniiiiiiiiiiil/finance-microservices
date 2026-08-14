package postgres_auth

import (
	"fmt"

	"golang.org/x/net/context"
)

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
