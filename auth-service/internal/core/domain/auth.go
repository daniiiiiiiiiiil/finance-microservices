package domain

import "time"

type Credentials struct {
	ID           int
	Email        string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
