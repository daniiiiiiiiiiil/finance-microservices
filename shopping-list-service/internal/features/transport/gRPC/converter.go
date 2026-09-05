package gRPC

import (
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/proto/shopping/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertShoppingToProto(shopping domain.Shopping) *gen.CreateShoppingResponse {
	return &gen.CreateShoppingResponse{
		Id:             int32(shopping.ID),
		Version:        int32(shopping.Version),
		Title:          shopping.Title,
		Description:    shopping.Description,
		AmountNow:      float32(shopping.AmountNow),
		AmountFinish:   float32(shopping.AmountFinish),
		ImageKey:       shopping.ImageKey,
		Completed:      shopping.Completed,
		CreatedAt:      timestamppb.New(shopping.CreatedAt),
		UpdatedAt:      timestamppb.New(*shopping.UpdatedAt),
		CompletedAt:    timestamppb.New(*shopping.CompletedAt),
		CompletionDate: timestamppb.New(*shopping.CompletionDate),
	}
}

func convertTimestampToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func convertTimeToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}
