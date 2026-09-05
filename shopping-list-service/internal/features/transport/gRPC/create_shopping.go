package gRPC

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/proto/shopping/gen"
	"go.uber.org/zap"
)

func (h *ShoppingListService) CreateShopping(ctx context.Context, req *gen.CreateShoppingRequest) (*gen.CreateShoppingResponse, error) {
	h.logger.Debug("gRPC CreateShopping", zap.String("Title", req.Title))

	shopping, err := h.service.CreateShopping(ctx, domain.Shopping{
		Title:          req.Title,
		Description:    req.Description,
		AmountNow:      float64(req.AmountNow),
		AmountFinish:   float64(req.AmountFinish),
		ImageKey:       req.ImageKey,
		Completed:      false,
		CreatedAt:      time.Now(),
		CompletionDate: convertTimestampToTime(req.CompletionDate),
	})
	if err != nil {
		h.logger.Error("gRPC CreateShopping error", zap.Error(err))
		return nil, fmt.Errorf("CreateShopping handler error: %w", err)
	}
	return convertShoppingToProto(shopping), nil
}
