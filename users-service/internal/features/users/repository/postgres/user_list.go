package postgres

import (
	"fmt"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"
)

func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users.users`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	query := `
		SELECT id, version, full_name, email, password_hash, phone_number, is_admin,status
		FROM users.users
		ORDER BY id
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.Version,
			&user.FullName,
			&user.Email,
			&user.PasswordHash,
			&user.PhoneNumber,
			&user.IsAdmin,
			&user.Status,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return users, total, nil
}

func (r *UserRepository) GetTotalUsers(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users.users`).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return total, nil
}
