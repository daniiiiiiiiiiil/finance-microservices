package postgres

import (
	"backend/internal/core/domain"
	errors_core "backend/internal/core/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (r *UserRepository) PatchUser(ctx context.Context, id int, patch domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE users.users SET
		full_name = $1,phone_number = $2,version=version+1
		WHERE id = $3 AND version = $4
		RETURNING id,full_name,phone_number,email,password_hash,version,is_admin`

	row := r.pool.QueryRow(ctx, query,
		patch.FullName,
		patch.PhoneNumber,
		id,
		patch.Version)
	var userModel UserModel
	if err := row.Scan(
		&userModel.ID,
		&userModel.FullName,
		&userModel.PhoneNumber,
		&userModel.Email,
		&userModel.Password,
		&userModel.Version,
		&userModel.IsAdmin,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, fmt.Errorf("user with id='%d' concurrently accesed: %w", id, errors_core.ErrConflict)
		}
		return domain.User{}, fmt.Errorf("err scan user with id='%d': %w", id, err)
	}
	userDomain := domain.NewUser(
		userModel.ID,
		userModel.Version,
		userModel.FullName,
		userModel.Email,
		userModel.Password,
		userModel.PhoneNumber,
		userModel.IsAdmin,
	)
	return userDomain, nil

}
