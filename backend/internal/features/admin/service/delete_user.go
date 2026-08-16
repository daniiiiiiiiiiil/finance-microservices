package service_admin

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
)

func (s *AdminService) DeleteUser(ctx context.Context, id int) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		auth := md.Get("authorization")
		if len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}
	if id <= 0 {
		return fmt.Errorf("Id must be greater than 0")
	}
	if err := s.userClient.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("DeleteUser id %v: %s", id, err)
	}
	return nil
}
