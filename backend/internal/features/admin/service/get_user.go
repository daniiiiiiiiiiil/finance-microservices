package service

import (
	"backend/internal/core/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *AdminService) GetUser(ctx context.Context, id int) (domain.User, error) {
	key := fmt.Sprintf("user:%d", id)
	var user domain.User

	err := s.redis.Get(ctx, key, &user)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			user, err = s.repo.GetUser(ctx, id)
			if err != nil {
				return domain.User{}, fmt.Errorf("get user from postgres: %w", err)
			}
			if err := s.redis.Set(ctx, key, user, 10*time.Minute); err != nil {
				return domain.User{}, fmt.Errorf("set user to redis: %w", err)
			}
			return user, nil
		}
		return domain.User{}, fmt.Errorf("get user from redis/postgres: %w", err)
	}
	return user, nil
}
