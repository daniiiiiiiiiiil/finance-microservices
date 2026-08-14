package postgres_auth

import (
	"fmt"

	"golang.org/x/net/context"
)

func (r *AuthRepository) DeleteByEmail(ctx context.Context, email string) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `DELETE FROM auth.credentials WHERE email = $1;`
	_, err := r.pool.Exec(ctx, query, email)
	if err != nil {
		return fmt.Errorf("could not delete user by email: %w", err)
	}
	return err
}
