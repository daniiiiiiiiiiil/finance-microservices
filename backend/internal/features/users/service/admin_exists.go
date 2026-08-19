package service_user

import "golang.org/x/net/context"

func (s *UsersService) AdminExists(ctx context.Context) (bool, error) {
	exists, err := s.userRepository.AdminExists(ctx)
	if err != nil {
		return false, err
	}
	return exists, nil
}
