package domain

import "time"

type ShoppingRequest struct {
	Text         string
	StartAddress string
}

type ParsedQuery struct {
	Items       []Item
	Preferences string
}

type Item struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Unit     string `json:"unit"`
}

type Store struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Address   string      `json:"address"`
	Latitude  float64     `json:"latitude"`
	Longitude float64     `json:"longitude"`
	Prices    []ItemPrice `json:"prices"`
	Delivery  *Delivery   `json:"delivery,omitempty"`
	SourceURL string      `json:"source_url,omitempty"`
}

type ItemPrice struct {
	ItemName string  `json:"item_name"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
	InStock  bool    `json:"in_stock"`
}

type Delivery struct {
	Available   bool    `json:"available"`
	Price       float64 `json:"price"`
	FreeFrom    float64 `json:"free_from,omitempty"`
	MinOrder    float64 `json:"min_order,omitempty"`
	DurationMin int     `json:"duration_min,omitempty"`
}

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Route struct {
	Order           []int      `json:"order"`
	DistanceMeters  float64    `json:"distance_meters"`
	DurationSeconds float64    `json:"duration_seconds"`
	GeometryGeoJSON string     `json:"geometry_geojson"`
	Legs            []RouteLeg `json:"legs,omitempty"`
}

type RouteLeg struct {
	FromIndex       int     `json:"from_index"`
	ToIndex         int     `json:"to_index"`
	DistanceMeters  float64 `json:"distance_meters"`
	DurationSeconds float64 `json:"duration_seconds"`
}

type CostBreakdown struct {
	ItemsTotal   float64 `json:"items_total"`
	TravelCost   float64 `json:"travel_cost"`
	DeliveryCost float64 `json:"delivery_cost"`
	Total        float64 `json:"total"`
}

type Recommendation struct {
	Mode   string  `json:"mode"`
	Total  float64 `json:"total"`
	Reason string  `json:"reason"`
}

type PlanResponse struct {
	RequestID      string         `json:"request_id"`
	UserID         int            `json:"user_id"`
	Items          []Item         `json:"items"`
	Stores         []Store        `json:"stores"`
	Route          *Route         `json:"route,omitempty"`
	Cost           CostBreakdown  `json:"cost"`
	Recommendation Recommendation `json:"recommendation"`
	CreatedAt      time.Time      `json:"created_at"`
}
