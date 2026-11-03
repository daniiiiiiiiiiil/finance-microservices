package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/pkg/logger"
	"go.uber.org/zap"
)

type CostConfig struct {
	FuelConsumptionLPer100KM float64
	FuelPricePerLiter        float64
	TaxiPricePerKM           float64
	Mode                     string
}

type Service struct {
	llm            ports.LLMClient
	search         ports.SearchClient
	geocoder       ports.Geocoder
	router         ports.Router
	repo           ports.PlanRepository
	eventPublisher ports.EventPublisher
	cost           CostConfig
	logger         *logger.Logger
}

func NewService(
	llm ports.LLMClient,
	search ports.SearchClient,
	geocoder ports.Geocoder,
	router ports.Router,
	repo ports.PlanRepository,
	eventPublisher ports.EventPublisher,
	cost CostConfig,
	log *logger.Logger,
) *Service {
	return &Service{
		llm:            llm,
		search:         search,
		geocoder:       geocoder,
		router:         router,
		repo:           repo,
		eventPublisher: eventPublisher,
		cost:           cost,
		logger:         log,
	}
}

func (s *Service) PlanShopping(ctx context.Context, req ports.PlanRequest) (*domain.PlanResponse, error) {
	if req.Text == "" {
		return nil, fmt.Errorf("%w: empty text", domain.ErrInvalidInput)
	}
	if req.StartAddress == "" {
		return nil, fmt.Errorf("%w: empty start address", domain.ErrInvalidInput)
	}

	if req.RequestID != "" && s.repo != nil {
		if existing, err := s.repo.GetByRequestID(ctx, req.RequestID); err == nil && existing != nil {
			if existing.UserID != req.UserID {
				return nil, fmt.Errorf("%w: request_id already used by another user", domain.ErrInvalidInput)
			}
			return &existing.Plan, nil
		}
	}

	parsed, err := s.llm.ParseShoppingRequest(ctx, req.Text)
	if err != nil {
		return nil, fmt.Errorf("parse request: %w", err)
	}
	if len(parsed.Items) == 0 {
		return nil, fmt.Errorf("%w: no items", domain.ErrInvalidInput)
	}

	startCoords, err := s.geocoder.Geocode(ctx, req.StartAddress)
	if err != nil {
		return nil, fmt.Errorf("geocode start: %w", err)
	}

	stores := make(map[string]*domain.Store)
	for _, item := range parsed.Items {
		found, err := s.search.SearchStores(ctx, item, req.StartAddress)
		if err != nil {
			s.logger.Warn("search stores failed",
				zap.String("item", item.Name),
				zap.Error(err))
			continue
		}
		for _, st := range found {
			if existing, ok := stores[st.ID]; ok {
				existing.Prices = append(existing.Prices, st.Prices...)
				if existing.Delivery == nil {
					existing.Delivery = st.Delivery
				}
			} else {
				stCopy := st
				stores[st.ID] = &stCopy
			}
		}
	}
	if len(stores) == 0 {
		return nil, fmt.Errorf("%w: no stores found for items", domain.ErrNoStoresFound)
	}

	storeList := make([]domain.Store, 0, len(stores))
	for _, st := range stores {
		if st.Latitude == 0 && st.Longitude == 0 && st.Address != "" {
			coords, err := s.geocoder.Geocode(ctx, st.Address)
			if err != nil {
				s.logger.Debug("geocode store failed",
					zap.String("store", st.Name),
					zap.String("address", st.Address),
					zap.Error(err))
				continue
			}
			st.Latitude = coords.Latitude
			st.Longitude = coords.Longitude
		}
		if st.Latitude != 0 || st.Longitude != 0 {
			storeList = append(storeList, *st)
		}
	}
	if len(storeList) == 0 {
		return nil, fmt.Errorf("%w: no stores with coordinates", domain.ErrNoStoresFound)
	}

	storeCoords := make([]domain.Coordinates, len(storeList))
	for i, st := range storeList {
		storeCoords[i] = domain.Coordinates{Latitude: st.Latitude, Longitude: st.Longitude}
	}

	order, route, err := s.router.OptimizeRoute(ctx, startCoords, storeCoords)
	if err != nil {
		s.logger.Warn("optimize route failed", zap.Error(err))
		route = nil
	} else {
		route.Order = order
	}

	cost := s.calculateCost(parsed.Items, storeList, route)

	rec := s.recommend(parsed.Items, storeList, route, cost)

	resp := &domain.PlanResponse{
		RequestID:      req.RequestID,
		UserID:         req.UserID,
		Items:          parsed.Items,
		Stores:         storeList,
		Route:          route,
		Cost:           cost,
		Recommendation: rec,
		CreatedAt:      time.Now().UTC(),
	}

	if s.repo != nil && req.RequestID != "" {
		saved := &domain.SavedPlan{
			UserID:    req.UserID,
			RequestID: req.RequestID,
			Text:      req.Text,
			Address:   req.StartAddress,
			Plan:      *resp,
		}
		if err := s.repo.Save(ctx, saved); err != nil {
			s.logger.Warn("failed to save plan", zap.Error(err))
		}
	}

	return resp, nil
}

func (s *Service) calculateCost(
	items []domain.Item,
	stores []domain.Store,
	route *domain.Route,
) domain.CostBreakdown {
	itemsTotal := 0.0
	for _, item := range items {
		bestPrice := math.MaxFloat64
		for _, st := range stores {
			for _, p := range st.Prices {
				if !sameItem(p.ItemName, item.Name) {
					continue
				}
				if p.Price > 0 && p.Price < bestPrice {
					bestPrice = p.Price
				}
			}
		}
		if bestPrice == math.MaxFloat64 {
			continue
		}
		qty := item.Quantity
		if qty <= 0 {
			qty = 1
		}
		itemsTotal += bestPrice * float64(qty)
	}

	travel := 0.0
	if route != nil && route.DistanceMeters > 0 {
		km := route.DistanceMeters / 1000.0
		switch s.cost.Mode {
		case "taxi":
			travel = km * s.cost.TaxiPricePerKM
		default:
			travel = km * (s.cost.FuelConsumptionLPer100KM / 100.0) * s.cost.FuelPricePerLiter
		}
	}

	delivery := 0.0
	for _, st := range stores {
		if st.Delivery != nil && st.Delivery.Available {
			if st.Delivery.Price > 0 {
				delivery += st.Delivery.Price
				break
			}
		}
	}

	return domain.CostBreakdown{
		ItemsTotal:   round2(itemsTotal),
		TravelCost:   round2(travel),
		DeliveryCost: round2(delivery),
		Total:        round2(itemsTotal + travel + delivery),
	}
}

func (s *Service) recommend(
	items []domain.Item,
	stores []domain.Store,
	route *domain.Route,
	cost domain.CostBreakdown,
) domain.Recommendation {
	selfTotal := cost.ItemsTotal + cost.TravelCost

	deliveryTotal := cost.ItemsTotal + cost.DeliveryCost
	for _, st := range stores {
		if st.Delivery != nil && st.Delivery.Available && st.Delivery.Price > 0 {
			break
		}
	}

	if cost.DeliveryCost == 0 && route == nil {
		return domain.Recommendation{
			Mode:   "self_drive",
			Total:  round2(selfTotal),
			Reason: "доставка недоступна — только самостоятельная поездка",
		}
	}

	if deliveryTotal < selfTotal {
		return domain.Recommendation{
			Mode:  "delivery",
			Total: round2(deliveryTotal),
			Reason: fmt.Sprintf("доставка дешевле поездки на %.0f ₽",
				selfTotal-deliveryTotal),
		}
	}
	return domain.Recommendation{
		Mode:  "self_drive",
		Total: round2(selfTotal),
		Reason: fmt.Sprintf("самостоятельная поездка дешевле доставки на %.0f ₽",
			deliveryTotal-selfTotal),
	}
}

func (s *Service) ListPlans(ctx context.Context, userID, limit, offset int) ([]domain.PlanSummary, int, error) {
	if s.repo == nil {
		return nil, 0, fmt.Errorf("%w: plan history is not configured", domain.ErrInvalidInput)
	}
	plans, err := s.repo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list plans: %w", err)
	}
	total, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("count plans: %w", err)
	}
	return plans, total, nil
}

func (s *Service) GetPlan(ctx context.Context, userID, id int) (*domain.SavedPlan, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("%w: plan history is not configured", domain.ErrInvalidInput)
	}
	plan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get plan: %w", err)
	}
	if plan.UserID != userID {
		return nil, fmt.Errorf("%w: plan does not belong to user", domain.ErrPlanNotFound)
	}
	return plan, nil
}

func (s *Service) DeletePlan(ctx context.Context, userID, id int) error {
	if s.repo == nil {
		return fmt.Errorf("%w: plan history is not configured", domain.ErrInvalidInput)
	}
	plan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get plan: %w", err)
	}
	if plan.UserID != userID {
		return fmt.Errorf("%w: plan does not belong to user", domain.ErrPlanNotFound)
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) DeletePlansByUser(ctx context.Context, userID int) error {
	if s.repo == nil {
		return fmt.Errorf("%w: plan history is not configured", domain.ErrInvalidInput)
	}
	return s.repo.DeleteByUser(ctx, userID)
}

func sameItem(a, b string) bool {
	a = normalizeItem(a)
	b = normalizeItem(b)
	return a == b || strings.Contains(a, b) || strings.Contains(b, a)
}

func normalizeItem(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) > 4 {
		s = s[:len(s)-1]
	}
	return s
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
