package gateway

import "time"

type dashboardResponse struct {
	TotalBalance       float64                `json:"totalBalance"`
	MonthlyIncome      float64                `json:"monthlyIncome"`
	MonthlyExpenses    float64                `json:"monthlyExpenses"`
	SavingsRate        float64                `json:"savingsRate"`
	DailyStats         []dailyStatResponse    `json:"dailyStats"`
	RecentTransactions []recentTransactionDTO `json:"recentTransactions"`
	TopCategories      []categoryStatResponse `json:"topCategories"`
}

type dailyStatResponse struct {
	Date    string  `json:"date"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

type recentTransactionDTO struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"createdAt"`
}

type categoryStatResponse struct {
	Category   string  `json:"category"`
	Total      float64 `json:"total"`
	Percentage float64 `json:"percentage"`
}

type transactionResponse struct {
	ID              int       `json:"id"`
	Version         int       `json:"version"`
	TypeTransaction string    `json:"typeTransaction"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"createdAt"`
	UserID          int       `json:"userId"`
}

type transactionsResponse struct {
	Transactions []transactionResponse `json:"transactions"`
	Total        int                   `json:"total"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}

type categoriesResponse struct {
	Categories []string `json:"categories"`
}

type analyticsResponse struct {
	Period        string                 `json:"period"`
	From          time.Time              `json:"from"`
	To            time.Time              `json:"to"`
	Income        float64                `json:"income"`
	Expenses      float64                `json:"expenses"`
	Net           float64                `json:"net"`
	SavingsRate   float64                `json:"savingsRate"`
	Series        []analyticsPointDTO    `json:"series"`
	TopCategories []categoryStatResponse `json:"topCategories"`
}

type analyticsPointDTO struct {
	Period   time.Time `json:"period"`
	Income   float64   `json:"income"`
	Expenses float64   `json:"expenses"`
}

type budgetResponse struct {
	ID          int       `json:"id"`
	Version     int       `json:"version"`
	UserID      int       `json:"userId"`
	Category    string    `json:"category"`
	AmountLimit float64   `json:"amountLimit"`
	Period      string    `json:"period"`
	PeriodStart time.Time `json:"periodStart"`
	PeriodEnd   time.Time `json:"periodEnd"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type budgetStatusResponse struct {
	Budget     budgetResponse `json:"budget"`
	Spent      float64        `json:"spent"`
	Remaining  float64        `json:"remaining"`
	Percentage float64        `json:"percentage"`
	Exceeded   bool           `json:"exceeded"`
}

type budgetsResponse struct {
	Budgets []budgetStatusResponse `json:"budgets"`
}

type createBudgetResponse struct {
	Budget budgetResponse `json:"budget"`
}

type updateBudgetResponse struct {
	Budget budgetResponse `json:"budget"`
}

type exportResponse struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

type shoppingResponse struct {
	ID             int        `json:"id"`
	Version        int        `json:"version"`
	Title          string     `json:"title"`
	Description    *string    `json:"description"`
	AmountNow      float64    `json:"amountNow"`
	AmountFinish   float64    `json:"amountFinish"`
	ImageKey       *string    `json:"imageKey"`
	Completed      bool       `json:"completed"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt"`
	CompletedAt    *time.Time `json:"completedAt"`
	CompletionDate *time.Time `json:"completionDate"`
}

type shoppingListResponse struct {
	List  []shoppingResponse `json:"list"`
	Total int                `json:"total"`
}

type shoppingTotalResponse struct {
	Total int `json:"total"`
}

type shoppingImageResponse struct {
	ImageData []byte  `json:"imageData"`
	ImageKey  *string `json:"imageKey"`
}

type aiPlanRequestDTO struct {
	RequestID    string  `json:"request_id"`
	Text         string  `json:"text"`
	StartAddress string  `json:"start_address"`
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
}

type aiPlanResponse struct {
	RequestID string              `json:"requestId"`
	UserID    int                 `json:"userId"`
	Items     []aiPlanItemDTO     `json:"items"`
	Stores    []aiStoreDTO        `json:"stores"`
	Route     *aiRouteDTO         `json:"route"`
	Cost      aiCostDTO           `json:"cost"`
	Recommend aiRecommendationDTO `json:"recommendation"`
	CreatedAt time.Time           `json:"createdAt"`
}

type aiPlanItemDTO struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Unit     string `json:"unit"`
}

type aiItemPriceDTO struct {
	ItemName string  `json:"itemName"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
	InStock  bool    `json:"inStock"`
}

type aiDeliveryDTO struct {
	Available   bool    `json:"available"`
	Price       float64 `json:"price"`
	FreeFrom    float64 `json:"freeFrom"`
	MinOrder    float64 `json:"minOrder"`
	DurationMin int     `json:"durationMin"`
}

type aiStoreDTO struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Address   string           `json:"address"`
	Latitude  float64          `json:"latitude"`
	Longitude float64          `json:"longitude"`
	Prices    []aiItemPriceDTO `json:"prices"`
	Delivery  *aiDeliveryDTO   `json:"delivery"`
	SourceURL string           `json:"sourceUrl"`
}

type aiRouteLegDTO struct {
	FromIndex       int     `json:"fromIndex"`
	ToIndex         int     `json:"toIndex"`
	DistanceMeters  float64 `json:"distanceMeters"`
	DurationSeconds float64 `json:"durationSeconds"`
}

type aiRouteDTO struct {
	Order           []int           `json:"order"`
	DistanceMeters  float64         `json:"distanceMeters"`
	DurationSeconds float64         `json:"durationSeconds"`
	GeometryGeoJSON string          `json:"geometryGeojson"`
	Legs            []aiRouteLegDTO `json:"legs"`
}

type aiCostDTO struct {
	ItemsTotal   float64 `json:"itemsTotal"`
	TravelCost   float64 `json:"travelCost"`
	DeliveryCost float64 `json:"deliveryCost"`
	Total        float64 `json:"total"`
}

type aiRecommendationDTO struct {
	Mode   string  `json:"mode"`
	Total  float64 `json:"total"`
	Reason string  `json:"reason"`
}

type aiPlanSummaryDTO struct {
	ID          int       `json:"id"`
	RequestID   string    `json:"requestId"`
	Text        string    `json:"text"`
	Address     string    `json:"address"`
	ItemsCount  int       `json:"itemsCount"`
	StoresCount int       `json:"storesCount"`
	TotalCost   float64   `json:"totalCost"`
	Recommended string    `json:"recommended"`
	CreatedAt   time.Time `json:"createdAt"`
}

type aiPlanListResponse struct {
	Plans  []aiPlanSummaryDTO `json:"plans"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

type currencyRatesResponse struct {
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
	Timestamp time.Time          `json:"timestamp"`
}

type currencyConvertResponse struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    float64   `json:"amount"`
	Result    float64   `json:"result"`
	Rate      float64   `json:"rate"`
	Timestamp time.Time `json:"timestamp"`
}

type userResponse struct {
	ID       int    `json:"id"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"isAdmin"`
	Status   string `json:"status"`
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
