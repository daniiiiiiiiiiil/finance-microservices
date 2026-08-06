package postgres_auth

import (
	"backend/internal/core/repository/postgres/pool"
	service_auth "backend/internal/features/auth/service"
	"context"
	"fmt"
)

type AuthRepository struct {
	pool pool.Pool
}

func NewAuthRepository(pool pool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

func (r *AuthRepository) CreateUserWithAuth(ctx context.Context, fullName, email, passwordHash string, phoneNumber *string, isAdmin bool) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO finance.users (full_name, email, password_hash, phone_number, is_admin)
		VALUES ($1, $2, $3, $4,$5)
		RETURNING id`

	var userID int
	err := r.pool.QueryRow(ctx, query, fullName, email, passwordHash, phoneNumber, isAdmin).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return userID, nil
}

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

func (r *AuthRepository) AdminExists(ctx context.Context) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM finance.users WHERE is_admin = true`
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check admin exists: %w", err)
	}
	return count > 0, nil
}
