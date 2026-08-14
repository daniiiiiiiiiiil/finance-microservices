package service_admin

import (
	"context"
	"fmt"
)

func (s *AdminService) DeleteUser(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("Id must be greater than 0")
	}
	if err := s.userClient.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("DeleteUser id %v: %s", id, err)
	}
	return nil
}
