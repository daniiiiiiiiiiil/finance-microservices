package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/pkg/logger"
)

const maxOSMResponse = 5 << 20

type OSMClient struct {
	nominatimURL string
	osrmURL      string
	userAgent    string
	http         *http.Client
	logger       *logger.Logger
}

func NewOSMClient(
	nominatimURL, osrmURL, userAgent string,
	timeout time.Duration,
	log *logger.Logger,
) *OSMClient {
	return &OSMClient{
		nominatimURL: strings.TrimRight(nominatimURL, "/"),
		osrmURL:      strings.TrimRight(osrmURL, "/"),
		userAgent:    userAgent,
		http:         &http.Client{Timeout: timeout},
		logger:       log,
	}
}

var _ ports.Geocoder = (*OSMClient)(nil)
var _ ports.Router = (*OSMClient)(nil)

func (c *OSMClient) Geocode(ctx context.Context, address string) (domain.Coordinates, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return domain.Coordinates{}, fmt.Errorf("%w: empty address", domain.ErrInvalidInput)
	}

	q := url.Values{}
	q.Set("q", address)
	q.Set("format", "json")
	q.Set("limit", "1")
	q.Set("addressdetails", "0")

	endpoint := c.nominatimURL + "/search?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.Coordinates{}, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return domain.Coordinates{}, domain.ErrProviderTimeout
		}
		return domain.Coordinates{}, fmt.Errorf("%w: %v", domain.ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxOSMResponse+1))
	if err != nil {
		return domain.Coordinates{}, fmt.Errorf("%w: read body: %v", domain.ErrProviderBadResponse, err)
	}
	if resp.StatusCode != http.StatusOK {
		return domain.Coordinates{}, fmt.Errorf("%w: status %d: %s",
			domain.ErrProviderBadResponse, resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.Unmarshal(data, &results); err != nil {
		return domain.Coordinates{}, fmt.Errorf("%w: unmarshal: %v", domain.ErrProviderBadResponse, err)
	}
	if len(results) == 0 {
		return domain.Coordinates{}, fmt.Errorf("%w: %q", domain.ErrAddressNotFound, address)
	}

	lat, err := parseFloat(results[0].Lat)
	if err != nil {
		return domain.Coordinates{}, fmt.Errorf("%w: lat: %v", domain.ErrProviderBadResponse, err)
	}
	lon, err := parseFloat(results[0].Lon)
	if err != nil {
		return domain.Coordinates{}, fmt.Errorf("%w: lon: %v", domain.ErrProviderBadResponse, err)
	}
	return domain.Coordinates{Latitude: lat, Longitude: lon}, nil
}

func (c *OSMClient) Route(ctx context.Context, start domain.Coordinates, points []domain.Coordinates) (*domain.Route, error) {
	if len(points) == 0 {
		return nil, fmt.Errorf("%w: empty points", domain.ErrInvalidInput)
	}

	coords := make([]string, 0, len(points)+1)
	coords = append(coords, formatLonLat(start))
	for _, p := range points {
		coords = append(coords, formatLonLat(p))
	}

	q := url.Values{}
	q.Set("overview", "full")
	q.Set("geometries", "geojson")
	q.Set("steps", "false")

	endpoint := c.osrmURL + "/route/v1/driving/" + strings.Join(coords, ";") + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, domain.ErrProviderTimeout
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxOSMResponse+1))
	if err != nil {
		return nil, fmt.Errorf("%w: read body: %v", domain.ErrProviderBadResponse, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d: %s",
			domain.ErrProviderBadResponse, resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var out struct {
		Code   string `json:"code"`
		Routes []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
			Geometry struct {
				Type        string          `json:"type"`
				Coordinates json.RawMessage `json:"coordinates"`
			} `json:"geometry"`
			Legs []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
			} `json:"legs"`
		} `json:"routes"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("%w: unmarshal: %v", domain.ErrProviderBadResponse, err)
	}
	if out.Code != "" && out.Code != "Ok" {
		return nil, fmt.Errorf("%w: osrm code %q", domain.ErrProviderBadResponse, out.Code)
	}
	if len(out.Routes) == 0 {
		return nil, fmt.Errorf("%w: no routes", domain.ErrProviderBadResponse)
	}

	r := out.Routes[0]
	geometry, _ := json.Marshal(map[string]any{
		"type":        r.Geometry.Type,
		"coordinates": json.RawMessage(r.Geometry.Coordinates),
	})

	legs := make([]domain.RouteLeg, 0, len(r.Legs))
	for i, l := range r.Legs {
		fromIdx := i - 1
		legs = append(legs, domain.RouteLeg{
			FromIndex:       fromIdx,
			ToIndex:         i,
			DistanceMeters:  l.Distance,
			DurationSeconds: l.Duration,
		})
	}

	return &domain.Route{
		DistanceMeters:  r.Distance,
		DurationSeconds: r.Duration,
		GeometryGeoJSON: string(geometry),
		Legs:            legs,
	}, nil
}

func (c *OSMClient) OptimizeRoute(
	ctx context.Context,
	start domain.Coordinates,
	points []domain.Coordinates,
) ([]int, *domain.Route, error) {
	if len(points) == 0 {
		return nil, nil, fmt.Errorf("%w: empty points", domain.ErrInvalidInput)
	}
	if len(points) == 1 {
		route, err := c.Route(ctx, start, points)
		if err != nil {
			return nil, nil, err
		}
		return []int{0}, route, nil
	}
	if len(points) > 5 {
		return c.greedyOptimize(ctx, start, points)
	}

	bestOrder := make([]int, len(points))
	bestRoute := (*domain.Route)(nil)
	bestDistance := 0.0

	forEachPermutation(len(points), func(order []int) {
		ordered := make([]domain.Coordinates, len(order))
		for i, idx := range order {
			ordered[i] = points[idx]
		}
		route, err := c.Route(ctx, start, ordered)
		if err != nil {
			return
		}
		if bestRoute == nil || route.DistanceMeters < bestDistance {
			bestOrder = append([]int(nil), order...)
			bestRoute = route
			bestDistance = route.DistanceMeters
		}
	})

	if bestRoute == nil {
		return nil, nil, fmt.Errorf("%w: no route found", domain.ErrProviderBadResponse)
	}

	bestRoute.Order = bestOrder

	for i := range bestRoute.Legs {
		if i < len(bestOrder) {
			bestRoute.Legs[i].ToIndex = bestOrder[i]
		}
		if i-1 >= 0 && i-1 < len(bestOrder) {
			bestRoute.Legs[i].FromIndex = bestOrder[i-1]
		} else {
			bestRoute.Legs[i].FromIndex = -1
		}
	}

	return bestOrder, bestRoute, nil
}

func (c *OSMClient) greedyOptimize(
	ctx context.Context,
	start domain.Coordinates,
	points []domain.Coordinates,
) ([]int, *domain.Route, error) {
	order := make([]int, len(points))
	for i := range order {
		order[i] = i
	}

	bestRoute, err := c.routeByOrder(ctx, start, points, order)
	if err != nil {
		return nil, nil, err
	}

	improved := true
	for improved {
		improved = false
		for i := 0; i < len(order)-1; i++ {
			for j := i + 1; j < len(order); j++ {
				candidate := append([]int(nil), order...)
				candidate[i], candidate[j] = candidate[j], candidate[i]
				r, err := c.routeByOrder(ctx, start, points, candidate)
				if err != nil {
					continue
				}
				if r.DistanceMeters < bestRoute.DistanceMeters {
					order = candidate
					bestRoute = r
					improved = true
				}
			}
		}
	}

	bestRoute.Order = order
	for i := range bestRoute.Legs {
		if i < len(order) {
			bestRoute.Legs[i].ToIndex = order[i]
		}
		if i-1 >= 0 && i-1 < len(order) {
			bestRoute.Legs[i].FromIndex = order[i-1]
		} else {
			bestRoute.Legs[i].FromIndex = -1
		}
	}
	return order, bestRoute, nil
}

func (c *OSMClient) routeByOrder(
	ctx context.Context,
	start domain.Coordinates,
	points []domain.Coordinates,
	order []int,
) (*domain.Route, error) {
	ordered := make([]domain.Coordinates, len(order))
	for i, idx := range order {
		ordered[i] = points[idx]
	}
	return c.Route(ctx, start, ordered)
}

func forEachPermutation(n int, fn func(order []int)) {
	perm := make([]int, n)
	for i := 0; i < n; i++ {
		perm[i] = i
	}
	var generate func(k int)
	generate = func(k int) {
		if k == n {
			fn(append([]int(nil), perm...))
			return
		}
		for i := k; i < n; i++ {
			perm[k], perm[i] = perm[i], perm[k]
			generate(k + 1)
			perm[k], perm[i] = perm[i], perm[k]
		}
	}
	generate(0)
}

func formatLonLat(c domain.Coordinates) string {
	return fmt.Sprintf("%.6f,%.6f", c.Longitude, c.Latitude)
}

func parseFloat(s string) (float64, error) {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return 0, err
	}
	return f, nil
}
