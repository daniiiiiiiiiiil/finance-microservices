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

func (h *ShoppingListService) GetTotalShopping(ctx context.Context, req *emptypb.Empty) (*gen.GetTotalResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	h.logger.Debug("gRPC Completed shopping service", zap.Int("id", userID))
	shopping, err := h.service.GetTotal(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("gRPC GetShopping failed: %w", err)
	}
	return convertTotal(shopping), nil
}
