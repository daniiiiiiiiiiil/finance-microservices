package gRPC

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/proto/shopping/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *ShoppingListService) GetShoppingImage(ctx context.Context, req *gen.GetShoppingImageRequest) (*gen.GetShoppingImageResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if req.ShoppingId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "shopping id must be positive")
	}

	h.logger.Debug("gRPC GetShoppingImage",
		zap.Int("user_id", userID),
		zap.Int32("shopping_id", req.ShoppingId))

	imageData, imageKey, err := h.service.GetImageByShoppingID(ctx, int(req.ShoppingId), userID)
	if err != nil {
		h.logger.Error("gRPC GetShoppingImage", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.GetShoppingImageResponse{
		ImageData: imageData,
		ImageKey:  &imageKey,
	}, nil
}
