package postgres_auth

import (
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/domain"
)

type AuthModel struct {
	ID           int
	Email        string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (m *AuthModel) ToDomain() domain.Credentials {
	return domain.Credentials{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Status:       m.Status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func FromDomain(cred domain.Credentials) AuthModel {
	return AuthModel{
		ID:           cred.ID,
		Email:        cred.Email,
		PasswordHash: cred.PasswordHash,
		Status:       cred.Status,
		CreatedAt:    cred.CreatedAt,
		UpdatedAt:    cred.UpdatedAt,
	}
}
