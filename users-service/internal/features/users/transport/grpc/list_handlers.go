package grpc

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/proto/users/gen"
)

func (s *UserServer) ListUsers(ctx context.Context, req *gen.ListUsersRequest) (*gen.ListUsersResponse, error) {
	users, total, err := s.service.ListUsers(ctx, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, fmt.Errorf("ListUsers: %v", err)
	}
	s.logger.Debug("Get users")
	return &gen.ListUsersResponse{
		Users:  ConvertUsersToProto(users),
		Total:  int32(total),
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}
