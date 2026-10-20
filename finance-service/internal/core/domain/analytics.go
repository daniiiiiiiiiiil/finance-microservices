package domain

import "time"

type AnalyticsPeriod string

const (
	AnalyticsPeriodDay   AnalyticsPeriod = "day"
	AnalyticsPeriodWeek  AnalyticsPeriod = "week"
	AnalyticsPeriodMonth AnalyticsPeriod = "month"
	AnalyticsPeriodYear  AnalyticsPeriod = "year"
)

type Analytics struct {
	Period        string           `json:"period"`
	From          time.Time        `json:"from"`
	To            time.Time        `json:"to"`
	Income        float64          `json:"income"`
	Expenses      float64          `json:"expenses"`
	Net           float64          `json:"net"`
	SavingsRate   float64          `json:"savings_rate"`
	Series        []AnalyticsPoint `json:"series"`
	TopCategories []CategoryStat   `json:"top_categories"`
}

type AnalyticsPoint struct {
	Period   time.Time `json:"period"`
	Income   float64   `json:"income"`
	Expenses float64   `json:"expenses"`
}
