package service_user

import (
	"context"
	"fmt"
)

func (s *UsersService) GetMetrics(ctx context.Context) (int, error) {
	total, err := s.userRepository.GetTotalUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("get total users: %w", err)
	}
	return total, nil
}
