package grpc

import (
	"backend/internal/features/users/transport/grpc/proto"

	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *UserServer) AdminExists(ctx context.Context, req *emptypb.Empty) (*proto.AdminExistsResponse, error) {
	exists, err := s.service.AdminExists(ctx)
	if err != nil {
		return nil, err
	}
	return &proto.AdminExistsResponse{Exists: exists}, nil
}
