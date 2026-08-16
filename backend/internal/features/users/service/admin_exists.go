package service_user

import "golang.org/x/net/context"

func (s *UsersService) AdminExists(ctx context.Context) (bool, error) {
	_, total, err := s.userRepository.ListUsers(ctx, 1, 0)
	if err != nil {
		return false, err
	}
	return total > 0, nil
}
