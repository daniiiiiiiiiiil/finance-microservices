package grpc

import (
	"context"
	"errors"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/pkg/grpcutil/interceptors"
	aipb "github.com/daniiiiiiiiiiil/finance-microservices/ai-service/proto/gen/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) PlanShopping(ctx context.Context, req *aipb.PlanShoppingRequest) (*aipb.PlanShoppingResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if req.Text == "" {
		return nil, status.Error(codes.InvalidArgument, "text is required")
	}
	if req.StartAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "start_address is required")
	}

	s.logger.Debug("gRPC PlanShopping",
		zap.Int("user_id", userID),
		zap.String("request_id", req.RequestId),
		zap.String("address", req.StartAddress),
	)

	result, err := s.service.PlanShopping(ctx, ports.PlanRequest{
		RequestID:    req.RequestId,
		UserID:       userID,
		Text:         req.Text,
		StartAddress: req.StartAddress,
	})
	if err != nil {
		s.logger.Error("PlanShopping failed", zap.Error(err))
		return nil, mapError(err)
	}

	return toProtoPlan(result), nil
}

func (s *Server) ListPlans(ctx context.Context, req *aipb.ListPlansRequest) (*aipb.ListPlansResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	limit := int(req.Limit)
	offset := int(req.Offset)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	plans, total, err := s.service.ListPlans(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error("ListPlans failed", zap.Error(err))
		return nil, mapError(err)
	}

	out := make([]*aipb.PlanSummary, 0, len(plans))
	for _, p := range plans {
		out = append(out, &aipb.PlanSummary{
			Id:          int32(p.ID),
			RequestId:   p.RequestID,
			Text:        p.Text,
			Address:     p.Address,
			ItemsCount:  int32(p.ItemsCount),
			StoresCount: int32(p.StoresCount),
			TotalCost:   p.TotalCost,
			Recommended: p.Recommended,
			CreatedAt:   timestamppb.New(p.CreatedAt),
		})
	}

	return &aipb.ListPlansResponse{
		Plans:  out,
		Total:  int32(total),
		Limit:  int32(limit),
		Offset: int32(offset),
	}, nil
}

func (s *Server) GetPlan(ctx context.Context, req *aipb.GetPlanRequest) (*aipb.PlanShoppingResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	plan, err := s.service.GetPlan(ctx, userID, int(req.Id))
	if err != nil {
		s.logger.Error("GetPlan failed", zap.Error(err))
		return nil, mapError(err)
	}

	return toProtoPlan(&plan.Plan), nil
}

func (s *Server) DeletePlan(ctx context.Context, req *aipb.DeletePlanRequest) (*emptypb.Empty, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.service.DeletePlan(ctx, userID, int(req.Id)); err != nil {
		s.logger.Error("DeletePlan failed", zap.Error(err))
		return nil, mapError(err)
	}
	return &emptypb.Empty{}, nil
}

func toProtoPlan(p *domain.PlanResponse) *aipb.PlanShoppingResponse {
	items := make([]*aipb.Item, 0, len(p.Items))
	for _, it := range p.Items {
		items = append(items, &aipb.Item{
			Name:     it.Name,
			Quantity: int32(it.Quantity),
			Unit:     it.Unit,
		})
	}

	stores := make([]*aipb.Store, 0, len(p.Stores))
	for _, st := range p.Stores {
		prices := make([]*aipb.ItemPrice, 0, len(st.Prices))
		for _, pr := range st.Prices {
			prices = append(prices, &aipb.ItemPrice{
				ItemName: pr.ItemName,
				Price:    pr.Price,
				Currency: pr.Currency,
				InStock:  pr.InStock,
			})
		}
		var d *aipb.Delivery
		if st.Delivery != nil {
			d = &aipb.Delivery{
				Available:   st.Delivery.Available,
				Price:       st.Delivery.Price,
				FreeFrom:    st.Delivery.FreeFrom,
				MinOrder:    st.Delivery.MinOrder,
				DurationMin: int32(st.Delivery.DurationMin),
			}
		}
		stores = append(stores, &aipb.Store{
			Id:        st.ID,
			Name:      st.Name,
			Address:   st.Address,
			Latitude:  st.Latitude,
			Longitude: st.Longitude,
			Prices:    prices,
			Delivery:  d,
			SourceUrl: st.SourceURL,
		})
	}

	var route *aipb.Route
	if p.Route != nil {
		order := make([]int32, 0, len(p.Route.Order))
		for _, idx := range p.Route.Order {
			order = append(order, int32(idx))
		}
		legs := make([]*aipb.RouteLeg, 0, len(p.Route.Legs))
		for _, l := range p.Route.Legs {
			legs = append(legs, &aipb.RouteLeg{
				FromIndex:       int32(l.FromIndex),
				ToIndex:         int32(l.ToIndex),
				DistanceMeters:  l.DistanceMeters,
				DurationSeconds: l.DurationSeconds,
			})
		}
		route = &aipb.Route{
			Order:           order,
			DistanceMeters:  p.Route.DistanceMeters,
			DurationSeconds: p.Route.DurationSeconds,
			GeometryGeojson: p.Route.GeometryGeoJSON,
			Legs:            legs,
		}
	}

	return &aipb.PlanShoppingResponse{
		RequestId: p.RequestID,
		UserId:    int32(p.UserID),
		Items:     items,
		Stores:    stores,
		Route:     route,
		Cost: &aipb.CostBreakdown{
			ItemsTotal:   p.Cost.ItemsTotal,
			TravelCost:   p.Cost.TravelCost,
			DeliveryCost: p.Cost.DeliveryCost,
			Total:        p.Cost.Total,
		},
		Recommendation: &aipb.Recommendation{
			Mode:   p.Recommendation.Mode,
			Total:  p.Recommendation.Total,
			Reason: p.Recommendation.Reason,
		},
		CreatedAt: timestamppb.New(p.CreatedAt),
	}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrAddressNotFound):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNoStoresFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrPlanNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrProviderTimeout):
		return status.Error(codes.DeadlineExceeded, "upstream timeout")
	case errors.Is(err, domain.ErrProviderUnavailable):
		return status.Error(codes.Unavailable, "upstream unavailable")
	case errors.Is(err, domain.ErrProviderBadResponse):
		return status.Error(codes.Internal, "upstream bad response")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
