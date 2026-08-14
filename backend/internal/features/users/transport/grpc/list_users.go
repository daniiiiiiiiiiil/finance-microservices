package grpc

import (
	"backend/internal/features/users/transport/grpc/proto"
	"fmt"

	"golang.org/x/net/context"
)

func (s *UserServer) ListUsers(ctx context.Context, req *proto.ListUsersRequest) (*proto.ListUsersResponse, error) {
	users, total, err := s.service.ListUsers(ctx, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, fmt.Errorf("ListUsers: %v", err)
	}
	s.logger.Debug("Get users")
	return &proto.ListUsersResponse{
		Users:  ConvertUsersToProto(users),
		Total:  int32(total),
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}
