package gRPC

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/proto/shopping/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *ShoppingListService) DeleteShopping(ctx context.Context, req *gen.DeleteShoppingRequest) (*emptypb.Empty, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	h.logger.Debug("gRPC DeleteShopping", zap.Int("user_id", userID), zap.Int32("id", req.Id))

	if err := h.service.DeleteShoppingList(ctx, int(req.Id), userID); err != nil {
		h.logger.Error("gRPC DeleteShopping error", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}
