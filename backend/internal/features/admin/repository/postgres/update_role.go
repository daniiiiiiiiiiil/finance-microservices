package postgres

import (
	"backend/internal/core/domain"
	errors_core "backend/internal/core/errors"
	"backend/internal/core/repository/postgres/pool"
	"context"
	"errors"
	"fmt"
)

func (r *AdminRepository) UpdateUserRoleTx(ctx context.Context, tx pool.Tx, id int, isAdmin bool) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE finance.users SET
		is_admin = $1,
		version = version + 1
		WHERE id = $2
		RETURNING id, version, full_name, email, password_hash, phone_number, is_admin
	`

	var model AdminUserModel
	err := tx.QueryRow(ctx, query, isAdmin, id).Scan(
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
		return domain.User{}, fmt.Errorf("update user role: %w", err)
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
