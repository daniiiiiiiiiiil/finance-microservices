package gRPC

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/proto/shopping/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *ShoppingListService) CompletedShopping(ctx context.Context, req *gen.CompletedShoppingRequest) (*emptypb.Empty, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	h.logger.Debug("gRPC Completed shopping service", zap.Int("id", userID))
	if err := h.service.CompletedShopping(ctx, int(req.Id), req.Completed, userID); err != nil {
		return nil, fmt.Errorf("completed shopping service: %w", err)
	}
	return nil, nil
}
