package service_auth

import (
	"time"

	"golang.org/x/net/context"
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
