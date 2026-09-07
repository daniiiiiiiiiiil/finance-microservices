package gRPC

import (
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/proto/shopping/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertShoppingToProto(shopping domain.Shopping) *gen.CreateShoppingResponse {
	return &gen.CreateShoppingResponse{
		Id:           int32(shopping.ID),
		Version:      int32(shopping.Version),
		Title:        shopping.Title,
		Description:  shopping.Description,
		AmountNow:    float32(shopping.AmountNow),
		AmountFinish: float32(shopping.AmountFinish),
		ImageKey:     shopping.ImageKey,
		Completed:    shopping.Completed,
		CreatedAt:    timestamppb.New(shopping.CreatedAt),
		UpdatedAt: func() *timestamppb.Timestamp {
			if shopping.UpdatedAt != nil {
				return timestamppb.New(*shopping.UpdatedAt)
			}
			return nil
		}(),
		CompletedAt: func() *timestamppb.Timestamp {
			if shopping.CompletedAt != nil {
				return timestamppb.New(*shopping.CompletedAt)
			}
			return nil
		}(),
		CompletionDate: func() *timestamppb.Timestamp {
			if shopping.CompletionDate != nil {
				return timestamppb.New(*shopping.CompletionDate)
			}
			return nil
		}(),
	}
}

func convertGetShoppingToProto(shopping domain.Shopping) *gen.GetShoppingResponse {
	return &gen.GetShoppingResponse{
		Id:           int32(shopping.ID),
		Version:      int32(shopping.Version),
		Title:        shopping.Title,
		Description:  shopping.Description,
		AmountNow:    float32(shopping.AmountNow),
		AmountFinish: float32(shopping.AmountFinish),
		ImageKey:     shopping.ImageKey,
		Completed:    shopping.Completed,
		CreatedAt:    timestamppb.New(shopping.CreatedAt),
		UpdatedAt: func() *timestamppb.Timestamp {
			if shopping.UpdatedAt != nil {
				return timestamppb.New(*shopping.UpdatedAt)
			}
			return nil
		}(),
		CompletedAt: func() *timestamppb.Timestamp {
			if shopping.CompletedAt != nil {
				return timestamppb.New(*shopping.CompletedAt)
			}
			return nil
		}(),
		CompletionDate: func() *timestamppb.Timestamp {
			if shopping.CompletionDate != nil {
				return timestamppb.New(*shopping.CompletionDate)
			}
			return nil
		}(),
	}
}

func convertUpdatedShoppingToProto(shopping domain.Shopping) *gen.UpdateShoppingResponse {
	return &gen.UpdateShoppingResponse{
		Id:           int32(shopping.ID),
		Version:      int32(shopping.Version),
		Title:        shopping.Title,
		Description:  shopping.Description,
		AmountNow:    float32(shopping.AmountNow),
		AmountFinish: float32(shopping.AmountFinish),
		ImageKey:     shopping.ImageKey,
		Completed:    shopping.Completed,
		CreatedAt:    timestamppb.New(shopping.CreatedAt),
		UpdatedAt: func() *timestamppb.Timestamp {
			if shopping.UpdatedAt != nil {
				return timestamppb.New(*shopping.UpdatedAt)
			}
			return nil
		}(),
		CompletedAt: func() *timestamppb.Timestamp {
			if shopping.CompletedAt != nil {
				return timestamppb.New(*shopping.CompletedAt)
			}
			return nil
		}(),
		CompletionDate: func() *timestamppb.Timestamp {
			if shopping.CompletionDate != nil {
				return timestamppb.New(*shopping.CompletionDate)
			}
			return nil
		}(),
	}
}

func ConvertShoppingListToProto(users []domain.Shopping) []*gen.GetShoppingResponse {
	result := make([]*gen.GetShoppingResponse, len(users))
	for i, user := range users {
		result[i] = convertGetShoppingToProto(user)
	}
	return result
}

func convertTotal(count int) *gen.GetTotalResponse {
	return &gen.GetTotalResponse{
		Total: int32(count),
	}
}

func convertTimestampToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}
func convertTimestampToTimes(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
func convertTimestampToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func convertTimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func convertTimePtrToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
