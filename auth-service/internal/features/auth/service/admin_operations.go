package service_auth

import (
	"fmt"
	"time"

	"context"
)

func (s *AuthService) GenerateToken(userID int, email string, isAdmin bool) (string, error) {
	return s.jwtManager.Generate(userID, email, isAdmin)
}

func (s *AuthService) AdminExists(ctx context.Context) (bool, error) {
	exists, err := s.usersClient.AdminExists(ctx)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *AuthService) RateLimitCheck(ctx context.Context, key string, limit int64, ttl time.Duration) (bool, error) {
	return s.rateLimit.Check(ctx, key, limit, ttl)
}

func (s *AuthService) AddToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	return s.blacklist.Add(ctx, token, ttl)
}

func (s *AuthService) DeleteCredentials(ctx context.Context, email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if err := s.credRepo.DeleteByEmail(ctx, email); err != nil {
		return fmt.Errorf("delete credentials: %w", err)
	}
	return nil
}
