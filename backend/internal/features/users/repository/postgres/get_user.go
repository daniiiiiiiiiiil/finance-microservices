package postgres

import (
	"backend/internal/core/domain"
	errors_core "backend/internal/core/errors"
	"backend/internal/core/repository/postgres/pool"
	"context"
	"errors"
	"fmt"
)

func (r *UserRepository) GetUser(ctx context.Context, id int) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query :=
		`SELECT id,version,full_name,email,password_hash,phone_number,is_admin
		 FROM finance.users 
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
		&userModal.IsAdmin)
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
		userModal.IsAdmin)
	return UserDomain, nil
}
