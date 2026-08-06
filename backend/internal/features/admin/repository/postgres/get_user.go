package postgres

import (
	"backend/internal/core/domain"
	errors_core "backend/internal/core/errors"
	"backend/internal/core/repository/postgres/pool"
	"context"
	"errors"
	"fmt"
)

func (r *AdminRepository) GetUser(ctx context.Context, id int) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, full_name, email, password_hash, phone_number, is_admin
		FROM finance.users
		WHERE id = $1
	`

	var model AdminUserModel
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&model.ID,
		&model.Version,
		&model.FullName,
		&model.Email,
		&model.Password,
		&model.PhoneNumber,
		&model.IsAdmin,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.User{}, fmt.Errorf("user with id %d: %w", id, errors_core.ErrNotFound)
		}
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}

	return domain.NewUser(
		model.ID,
		model.Version,
		model.FullName,
		model.Email,
		model.Password,
		model.PhoneNumber,
		model.IsAdmin,
	), nil
}
