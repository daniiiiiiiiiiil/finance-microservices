package service_user

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (s *UsersService) AdminExists(ctx context.Context) (bool, error) {
	key := "admin:exists"
	var exists bool
	err := s.redis.Get(ctx, key, &exists)
	if err == nil {
		return exists, nil
	}

	exists, err = s.userRepository.AdminExists(ctx)
	if err != nil {
		return false, err
	}

	if err := s.redis.Set(ctx, key, exists, 5*time.Minute); err != nil {
		s.logger.Warn("failed to cache admin exists", zap.Error(err))
	}

	return exists, nil
}

func (s *UsersService) GetMetrics(ctx context.Context) (int, error) {
	total, err := s.userRepository.GetTotalUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("get total users: %w", err)
	}
	return total, nil
}
