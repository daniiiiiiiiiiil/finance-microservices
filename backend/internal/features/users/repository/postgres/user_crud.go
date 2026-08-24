package postgres

import (
	"backend/internal/core/domain"
	"backend/internal/core/repository/postgres/pool"
	errors_core "backend/pkg/errors"
	"database/sql"
	"errors"
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

func (r *UserRepository) GetUser(ctx context.Context, id int) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query :=
		`SELECT id,version,full_name,email,password_hash,phone_number,is_admin,status
		 FROM users.users 
         WHERE id = $1;`
	row := r.pool.QueryRow(ctx, query, id)
	var userModal UserModel
	err := row.Scan(
		&userModal.ID,
		&userModal.Version,
		&userModal.FullName,
		&userModal.Email,
		&userModal.Password,
		&userModal.PhoneNumber,
		&userModal.IsAdmin,
		&userModal.Status,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.User{}, fmt.Errorf("user with id %d : %w", id, errors_core.ErrNotFound)
		}
		return domain.User{}, fmt.Errorf("scan error: %w", err)
	}
	UserDomain := domain.NewUser(
		userModal.ID,
		userModal.Version,
		userModal.FullName,
		userModal.Email,
		userModal.Password,
		userModal.PhoneNumber,
		userModal.IsAdmin,
		userModal.Status)
	return UserDomain, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, full_name, email, password_hash, phone_number, is_admin,status
		FROM users.users
		WHERE email = $1`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
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
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) PatchUser(ctx context.Context, id int, patch domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE users.users SET
		full_name = $1,phone_number = $2,version=version+1
		WHERE id = $3 AND version = $4
		RETURNING id,full_name,phone_number,email,password_hash,version,is_admin,status`

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
		&userModel.Status,
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
		userModel.Status,
	)
	return userDomain, nil

}

func (r *UserRepository) DeleteUserTx(ctx context.Context, tx pool.Tx, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	cmdTag, err := tx.Exec(ctx, `DELETE FROM users.users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("user with id %d: %w", id, errors_core.ErrNotFound)
	}

	return nil
}
