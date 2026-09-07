package gRPC

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/proto/shopping/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *ShoppingListService) GetShopping(ctx context.Context, req *gen.GetShoppingRequest) (*gen.GetShoppingResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	h.logger.Debug("gRPC get shopping", zap.Int("user_id", userID))
	shopping, err := h.service.GetShopping(ctx, int(req.Id), userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return convertGetShoppingToProto(shopping), nil
}
