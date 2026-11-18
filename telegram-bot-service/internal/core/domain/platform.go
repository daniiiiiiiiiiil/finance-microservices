package domain

import "time"

type Transaction struct {
	ID          int
	Version     int
	Type        string
	Amount      float64
	Category    string
	Description string
	CreatedAt   time.Time
	UserID      int
}

type TransactionFilter struct {
	Type     string
	Category string
	From     string
	To       string
	Limit    int
	Offset   int
}

type TransactionPage struct {
	Items  []Transaction
	Total  int
	Limit  int
	Offset int
}

type CategoryStat struct {
	Category   string
	Total      float64
	Percentage float64
}

type DailyStat struct {
	Date    string
	Income  float64
	Expense float64
}

type Dashboard struct {
	TotalBalance       float64
	MonthlyIncome      float64
	MonthlyExpenses    float64
	SavingsRate        float64
	DailyStats         []DailyStat
	RecentTransactions []Transaction
	TopCategories      []CategoryStat
}

type Analytics struct {
	Period        string
	From          time.Time
	To            time.Time
	Income        float64
	Expenses      float64
	Net           float64
	SavingsRate   float64
	Series        []AnalyticsPoint
	TopCategories []CategoryStat
}

type AnalyticsPoint struct {
	Period   time.Time
	Income   float64
	Expenses float64
}

type Budget struct {
	ID          int
	Version     int
	UserID      int
	Category    string
	AmountLimit float64
	Period      string
	PeriodStart time.Time
	PeriodEnd   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BudgetStatus struct {
	Budget     Budget
	Spent      float64
	Remaining  float64
	Percentage float64
	Exceeded   bool
}

type ShoppingItem struct {
	ID             int
	Version        int
	Title          string
	Description    *string
	AmountNow      float64
	AmountFinish   float64
	ImageKey       *string
	Completed      bool
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	CompletedAt    *time.Time
	CompletionDate *time.Time
}

type ShoppingPage struct {
	Items []ShoppingItem
	Total int
}

type ShoppingInput struct {
	Title          string
	Description    *string
	AmountNow      float64
	AmountFinish   float64
	Completed      bool
	ImageData      []byte
	Filename       string
	CompletionDate *time.Time
}

type ShoppingImage struct {
	Data     []byte
	Filename string
}

type AIPlanRequest struct {
	RequestID    string
	Text         string
	StartAddress string
	Latitude     float64
	Longitude    float64
}

type AIPlanItem struct {
	Name     string
	Quantity int
	Unit     string
}

type AIItemPrice struct {
	ItemName string
	Price    float64
	Currency string
	InStock  bool
}

type AIDelivery struct {
	Available   bool
	Price       float64
	FreeFrom    float64
	MinOrder    float64
	DurationMin int
}

type AIStore struct {
	ID        string
	Name      string
	Address   string
	Latitude  float64
	Longitude float64
	Prices    []AIItemPrice
	Delivery  *AIDelivery
	SourceURL string
}

type AIRouteLeg struct {
	FromIndex       int
	ToIndex         int
	DistanceMeters  float64
	DurationSeconds float64
}

type AIRoute struct {
	Order           []int
	DistanceMeters  float64
	DurationSeconds float64
	GeometryGeoJSON string
	Legs            []AIRouteLeg
}

type AICost struct {
	ItemsTotal   float64
	TravelCost   float64
	DeliveryCost float64
	Total        float64
}

type AIRecommendation struct {
	Mode   string
	Total  float64
	Reason string
}

type AIPlan struct {
	RequestID      string
	UserID         int
	Items          []AIPlanItem
	Stores         []AIStore
	Route          *AIRoute
	Cost           AICost
	Recommendation AIRecommendation
	CreatedAt      time.Time
}

type AIPlanSummary struct {
	ID          int
	RequestID   string
	Text        string
	Address     string
	ItemsCount  int
	StoresCount int
	TotalCost   float64
	Recommended string
	CreatedAt   time.Time
}

type AIPlanPage struct {
	Items  []AIPlanSummary
	Total  int
	Limit  int
	Offset int
}

type CurrencyRates struct {
	Base      string
	Rates     map[string]float64
	Timestamp time.Time
}

type Conversion struct {
	From      string
	To        string
	Amount    float64
	Result    float64
	Rate      float64
	Timestamp time.Time
}

type UserProfile struct {
	ID       int
	FullName string
	Email    string
	IsAdmin  bool
	Status   string
}

type ExportFile struct {
	Key string
	URL string
}

type DownloadedFile struct {
	Data     []byte
	Filename string
}

type Notification struct {
	Kind  string
	Title string
	Text  string
}
