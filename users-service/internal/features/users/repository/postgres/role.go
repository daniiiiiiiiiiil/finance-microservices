package postgres

import (
	"errors"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/repository/postgres/pool"
	errors_core "github.com/daniiiiiiiiiiil/finance-microservices/users-service/pkg/errors"
	"golang.org/x/net/context"
)

func (r *UserRepository) UpdateRole(ctx context.Context, id int, isAdmin bool) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE users.users SET
		is_admin = $1,
		version = version + 1
		WHERE id = $2
		RETURNING id, version, full_name, email, password_hash, phone_number, is_admin,status`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, isAdmin, id).Scan(
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
		if errors.Is(err, pool.ErrNoRows) {
			return domain.User{}, fmt.Errorf("user with id %d: %w", id, errors_core.ErrNotFound)
		}
		return domain.User{}, fmt.Errorf("update role: %w", err)
	}

	return user, nil
}
