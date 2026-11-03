package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/domain"
)

type Geocoder interface {
	Geocode(ctx context.Context, address string) (domain.Coordinates, error)
}

type Router interface {
	Route(ctx context.Context, start domain.Coordinates, points []domain.Coordinates) (*domain.Route, error)
	OptimizeRoute(ctx context.Context, start domain.Coordinates, points []domain.Coordinates) ([]int, *domain.Route, error)
}
