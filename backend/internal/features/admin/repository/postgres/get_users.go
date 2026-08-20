package postgres

import (
	"backend/internal/core/domain"
	"backend/internal/core/pagination"
	"context"
	"fmt"
)

func (r *AdminRepository) GetUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	limit, offset = pagination.LimitOffset(limit, offset)

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM finance.users`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	query := `
    SELECT id, version, full_name, email, password_hash, phone_number, is_admin, status
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
		var model AdminUserModel
		err := rows.Scan(
			&model.ID,
			&model.Version,
			&model.FullName,
			&model.Email,
			&model.Password,
			&model.PhoneNumber,
			&model.IsAdmin,
			&model.Status,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, domain.NewUser(
			model.ID,
			model.Version,
			model.FullName,
			model.Email,
			model.Password,
			model.PhoneNumber,
			model.IsAdmin,
			model.Status,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return users, total, nil
}
