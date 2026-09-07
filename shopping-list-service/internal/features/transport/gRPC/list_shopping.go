// internal/features/transport/gRPC/list_shopping.go
package gRPC

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/proto/shopping/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *ShoppingListService) ListShopping(ctx context.Context, req *gen.ListShoppingRequest) (*gen.ListShoppingResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	h.logger.Debug("gRPC ListShopping request received",
		zap.Int("user_id", userID),
		zap.Int32("limit", req.Limit),
		zap.Int32("offset", req.Offset),
	)

	shopping, total, err := h.service.ListShopping(ctx, int(req.Limit), int(req.Offset), userID)
	if err != nil {
		h.logger.Error("gRPC ListShopping failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &gen.ListShoppingResponse{
		List:  ConvertShoppingListToProto(shopping),
		Total: int32(total),
	}, nil
}
