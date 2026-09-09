package gRPC

import (
	"context"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/proto/shopping/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *ShoppingListService) UpdateShopping(ctx context.Context, req *gen.UpdateShoppingRequest) (*gen.UpdateShoppingResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	h.logger.Debug("gRPC UpdateShopping", zap.Int("ID", userID), zap.Any("req", req))
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.AmountNow <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount now is required")
	}
	if req.AmountFinish <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount finish is required")
	}
	existing, err := h.service.GetShopping(ctx, int(req.Id), userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "shopping not found")
	}

	now := time.Now()

	var completedAt *time.Time

	if req.Completed {
		if existing.Completed {
			completedAt = existing.CompletedAt
		} else {
			if req.CompletionDate != nil {
				t := req.CompletionDate.AsTime()
				completedAt = &t
			} else {
				completedAt = &now
			}
		}
	} else {
		completedAt = nil
	}
	shopping := domain.Shopping{
		ID:             int(req.Id),
		Title:          req.Title,
		Description:    req.Description,
		AmountNow:      float64(req.AmountNow),
		AmountFinish:   float64(req.AmountFinish),
		ImageKey:       existing.ImageKey,
		Completed:      req.Completed,
		CreatedAt:      existing.CreatedAt,
		UpdatedAt:      &now,
		CompletedAt:    completedAt,
		CompletionDate: convertTimestampToTimePtr(req.CompletionDate),
	}
	updated, err := h.service.UpdateShopping(ctx, &shopping, userID, req.ImageData, req.Filename)
	if err != nil {
		h.logger.Error("gRPC UpdateShopping error", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return convertUpdatedShoppingToProto(updated), nil
}
