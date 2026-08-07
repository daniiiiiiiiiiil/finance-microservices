package postgres_auth

import (
	"context"
	"fmt"
)

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
