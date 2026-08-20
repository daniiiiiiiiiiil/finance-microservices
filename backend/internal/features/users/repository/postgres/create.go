package postgres

import (
	"backend/internal/core/domain"
	"fmt"

	"golang.org/x/net/context"
)

func (r *UserRepository) CreateUser(ctx context.Context, user domain.User) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO users.users (full_name, email, password_hash, phone_number, is_admin,status)
		VALUES ($1, $2, $3, $4, $5,$6)
		RETURNING id`

	var id int
	err := r.pool.QueryRow(ctx, query,
		user.FullName,
		user.Email,
		user.PasswordHash,
		user.PhoneNumber,
		user.IsAdmin,
		user.Status,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	return id, nil
}
