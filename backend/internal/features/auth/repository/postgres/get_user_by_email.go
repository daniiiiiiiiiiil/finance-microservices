package postgres_auth

import (
	service_auth "backend/internal/features/auth/service"
	"context"
	"fmt"
)

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (service_auth.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, full_name, email, password_hash, phone_number, is_admin
		FROM finance.users
		WHERE email = $1`

	var user service_auth.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.PhoneNumber,
		&user.IsAdmin,
	)
	if err != nil {
		return service_auth.User{}, fmt.Errorf("scan user: %w", err)
	}
	return user, nil
}
