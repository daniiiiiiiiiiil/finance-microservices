package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/metrics"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

const maxResponseBytes = 32 << 20

type Client struct {
	baseURL string
	http    *http.Client
	tokens  ports.TokenIssuer
	logger  *logger.Logger
}

func NewClient(baseURL string, tokens ports.TokenIssuer, timeout time.Duration, logger *logger.Logger) *Client {
	transport := otelhttp.NewTransport(&http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	})
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
		tokens: tokens,
		logger: logger,
	}
}

func (c *Client) GetDashboard(ctx context.Context, userID int) (domain.Dashboard, error) {
	var response dashboardResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "finance/dashboard", "/api/v1/finance/dashboard", nil, nil, &response); err != nil {
		return domain.Dashboard{}, err
	}
	dashboard := domain.Dashboard{
		TotalBalance:    response.TotalBalance,
		MonthlyIncome:   response.MonthlyIncome,
		MonthlyExpenses: response.MonthlyExpenses,
		SavingsRate:     response.SavingsRate,
	}
	for _, stat := range response.DailyStats {
		dashboard.DailyStats = append(dashboard.DailyStats, domain.DailyStat{
			Date:    stat.Date,
			Income:  stat.Income,
			Expense: stat.Expense,
		})
	}
	for _, item := range response.RecentTransactions {
		dashboard.RecentTransactions = append(dashboard.RecentTransactions, domain.Transaction{
			ID:        item.ID,
			Type:      item.Type,
			Amount:    item.Amount,
			Category:  item.Category,
			CreatedAt: item.CreatedAt,
		})
	}
	for _, item := range response.TopCategories {
		dashboard.TopCategories = append(dashboard.TopCategories, domain.CategoryStat{
			Category:   item.Category,
			Total:      item.Total,
			Percentage: item.Percentage,
		})
	}
	return dashboard, nil
}

func (c *Client) GetTransactions(ctx context.Context, userID int, filter domain.TransactionFilter) (domain.TransactionPage, error) {
	query := url.Values{}
	if filter.Type != "" {
		query.Set("type", filter.Type)
	}
	if filter.Category != "" {
		query.Set("category", filter.Category)
	}
	if filter.From != "" {
		query.Set("from", filter.From)
	}
	if filter.To != "" {
		query.Set("to", filter.To)
	}
	query.Set("limit", strconv.Itoa(filter.Limit))
	query.Set("offset", strconv.Itoa(filter.Offset))

	var response transactionsResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "finance/transactions", "/api/v1/finance/transactions", query, nil, &response); err != nil {
		return domain.TransactionPage{}, err
	}
	page := domain.TransactionPage{
		Total:  response.Total,
		Limit:  response.Limit,
		Offset: response.Offset,
	}
	for _, item := range response.Transactions {
		page.Items = append(page.Items, transactionFromDTO(item))
	}
	return page, nil
}

func (c *Client) GetTransaction(ctx context.Context, userID, id int) (domain.Transaction, error) {
	var response transactionResponse
	endpoint := fmt.Sprintf("/api/v1/finance/transaction/%d", id)
	if err := c.doJSON(ctx, userID, http.MethodGet, "finance/transaction", endpoint, nil, nil, &response); err != nil {
		return domain.Transaction{}, err
	}
	return transactionFromDTO(response), nil
}

func (c *Client) CreateTransaction(ctx context.Context, userID int, transactionType string, amount float64, category string, description string) (domain.Transaction, error) {
	body := map[string]any{
		"type_transaction": transactionType,
		"amount":           amount,
		"category":         category,
		"description":      description,
	}
	var response transactionResponse
	if err := c.doJSON(ctx, userID, http.MethodPost, "finance/transaction", "/api/v1/finance/transaction", nil, body, &response); err != nil {
		return domain.Transaction{}, err
	}
	return transactionFromDTO(response), nil
}

func (c *Client) UpdateTransaction(ctx context.Context, userID, id int, transactionType string, amount float64, category string, description string) (domain.Transaction, error) {
	body := map[string]any{
		"type_transaction": transactionType,
		"amount":           amount,
		"category":         category,
		"description":      description,
	}
	endpoint := fmt.Sprintf("/api/v1/finance/transaction/%d", id)
	var response transactionResponse
	if err := c.doJSON(ctx, userID, http.MethodPut, "finance/transaction", endpoint, nil, body, &response); err != nil {
		return domain.Transaction{}, err
	}
	return transactionFromDTO(response), nil
}

func (c *Client) DeleteTransaction(ctx context.Context, userID, id int) error {
	endpoint := fmt.Sprintf("/api/v1/finance/transaction/%d", id)
	return c.doJSON(ctx, userID, http.MethodDelete, "finance/transaction", endpoint, nil, nil, nil)
}

func (c *Client) GetCategories(ctx context.Context, userID int) ([]string, error) {
	var response categoriesResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "finance/categories", "/api/v1/finance/categories", nil, nil, &response); err != nil {
		return nil, err
	}
	return response.Categories, nil
}

func (c *Client) GetAnalytics(ctx context.Context, userID int, period string) (domain.Analytics, error) {
	query := url.Values{}
	if period != "" {
		query.Set("period", period)
	}
	var response analyticsResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "finance/analytics", "/api/v1/finance/analytics", query, nil, &response); err != nil {
		return domain.Analytics{}, err
	}
	analytics := domain.Analytics{
		Period:      response.Period,
		From:        response.From,
		To:          response.To,
		Income:      response.Income,
		Expenses:    response.Expenses,
		Net:         response.Net,
		SavingsRate: response.SavingsRate,
	}
	for _, point := range response.Series {
		analytics.Series = append(analytics.Series, domain.AnalyticsPoint{
			Period:   point.Period,
			Income:   point.Income,
			Expenses: point.Expenses,
		})
	}
	for _, item := range response.TopCategories {
		analytics.TopCategories = append(analytics.TopCategories, domain.CategoryStat{
			Category:   item.Category,
			Total:      item.Total,
			Percentage: item.Percentage,
		})
	}
	return analytics, nil
}

func (c *Client) ListBudgets(ctx context.Context, userID int) ([]domain.BudgetStatus, error) {
	var response budgetsResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "finance/budgets", "/api/v1/finance/budgets", nil, nil, &response); err != nil {
		return nil, err
	}
	statuses := make([]domain.BudgetStatus, 0, len(response.Budgets))
	for _, item := range response.Budgets {
		statuses = append(statuses, domain.BudgetStatus{
			Budget:     budgetFromDTO(item.Budget),
			Spent:      item.Spent,
			Remaining:  item.Remaining,
			Percentage: item.Percentage,
			Exceeded:   item.Exceeded,
		})
	}
	return statuses, nil
}

func (c *Client) CreateBudget(ctx context.Context, userID int, category string, amountLimit float64, period string) (domain.Budget, error) {
	body := map[string]any{
		"category":     category,
		"amount_limit": amountLimit,
		"period":       period,
	}
	var response createBudgetResponse
	if err := c.doJSON(ctx, userID, http.MethodPost, "finance/budget", "/api/v1/finance/budget", nil, body, &response); err != nil {
		return domain.Budget{}, err
	}
	return budgetFromDTO(response.Budget), nil
}

func (c *Client) UpdateBudget(ctx context.Context, userID, id int, category string, amountLimit float64, period string) (domain.Budget, error) {
	body := map[string]any{
		"category":     category,
		"amount_limit": amountLimit,
		"period":       period,
	}
	endpoint := fmt.Sprintf("/api/v1/finance/budget/%d", id)
	var response updateBudgetResponse
	if err := c.doJSON(ctx, userID, http.MethodPut, "finance/budget", endpoint, nil, body, &response); err != nil {
		return domain.Budget{}, err
	}
	return budgetFromDTO(response.Budget), nil
}

func (c *Client) DeleteBudget(ctx context.Context, userID, id int) error {
	endpoint := fmt.Sprintf("/api/v1/finance/budget/%d", id)
	return c.doJSON(ctx, userID, http.MethodDelete, "finance/budget", endpoint, nil, nil, nil)
}

func (c *Client) Export(ctx context.Context, userID int, format string) (domain.ExportFile, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case "csv", "pdf", "json", "txt":
	default:
		return domain.ExportFile{}, fmt.Errorf("unsupported export format %q: %w", format, apperrors.ErrInvalidArgument)
	}
	var response exportResponse
	endpoint := "/api/v1/finance/export/" + format
	if err := c.doJSON(ctx, userID, http.MethodGet, "finance/export/"+format, endpoint, nil, nil, &response); err != nil {
		return domain.ExportFile{}, err
	}
	return domain.ExportFile{Key: response.Key, URL: response.URL}, nil
}

func (c *Client) DownloadExport(ctx context.Context, userID int, key string) (domain.DownloadedFile, error) {
	if key == "" {
		return domain.DownloadedFile{}, fmt.Errorf("export key is required: %w", apperrors.ErrInvalidArgument)
	}
	query := url.Values{}
	query.Set("key", key)
	data, headers, err := c.doRaw(ctx, userID, http.MethodGet, "finance/export/file", "/api/v1/finance/export/file", query, nil)
	if err != nil {
		return domain.DownloadedFile{}, err
	}
	filename := filenameFromHeaders(headers, key)
	return domain.DownloadedFile{Data: data, Filename: filename}, nil
}

func (c *Client) ListShopping(ctx context.Context, userID, limit, offset int) (domain.ShoppingPage, error) {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))
	var response shoppingListResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "shopping/list", "/api/v1/shopping", query, nil, &response); err != nil {
		return domain.ShoppingPage{}, err
	}
	page := domain.ShoppingPage{Total: response.Total}
	for _, item := range response.List {
		page.Items = append(page.Items, shoppingFromDTO(item))
	}
	return page, nil
}

func (c *Client) GetShopping(ctx context.Context, userID, id int) (domain.ShoppingItem, error) {
	endpoint := fmt.Sprintf("/api/v1/shopping/%d", id)
	var response shoppingResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "shopping/get", endpoint, nil, nil, &response); err != nil {
		return domain.ShoppingItem{}, err
	}
	return shoppingFromDTO(response), nil
}

func (c *Client) CreateShopping(ctx context.Context, userID int, input domain.ShoppingInput) (domain.ShoppingItem, error) {
	body := map[string]any{
		"title":         input.Title,
		"amount_now":    input.AmountNow,
		"amount_finish": input.AmountFinish,
		"completed":     input.Completed,
	}
	if input.Description != nil {
		body["description"] = *input.Description
	}
	if len(input.ImageData) > 0 && input.Filename != "" {
		body["image_data"] = input.ImageData
		body["filename"] = input.Filename
	}
	var response shoppingResponse
	if err := c.doJSON(ctx, userID, http.MethodPost, "shopping/create", "/api/v1/shopping", nil, body, &response); err != nil {
		return domain.ShoppingItem{}, err
	}
	return shoppingFromDTO(response), nil
}

func (c *Client) UpdateShopping(ctx context.Context, userID int, item domain.ShoppingItem, input domain.ShoppingInput) (domain.ShoppingItem, error) {
	body := map[string]any{
		"version":       item.Version,
		"title":         input.Title,
		"amount_now":    input.AmountNow,
		"amount_finish": input.AmountFinish,
		"completed":     input.Completed,
	}
	if input.Description != nil {
		body["description"] = *input.Description
	}
	if len(input.ImageData) > 0 && input.Filename != "" {
		body["image_data"] = input.ImageData
		body["filename"] = input.Filename
	}
	endpoint := fmt.Sprintf("/api/v1/shopping/%d", item.ID)
	var response shoppingResponse
	if err := c.doJSON(ctx, userID, http.MethodPut, "shopping/update", endpoint, nil, body, &response); err != nil {
		return domain.ShoppingItem{}, err
	}
	return shoppingFromDTO(response), nil
}

func (c *Client) DeleteShopping(ctx context.Context, userID, id int) error {
	endpoint := fmt.Sprintf("/api/v1/shopping/%d", id)
	return c.doJSON(ctx, userID, http.MethodDelete, "shopping/delete", endpoint, nil, nil, nil)
}

func (c *Client) CompleteShopping(ctx context.Context, userID, id int, completed bool) error {
	endpoint := fmt.Sprintf("/api/v1/shopping/%d", id)
	body := map[string]any{"completed": completed}
	return c.doJSON(ctx, userID, http.MethodPatch, "shopping/complete", endpoint, nil, body, nil)
}

func (c *Client) ShoppingTotal(ctx context.Context, userID int) (int, error) {
	var response shoppingTotalResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "shopping/total", "/api/v1/shopping/total", nil, nil, &response); err != nil {
		return 0, err
	}
	return response.Total, nil
}

func (c *Client) ShoppingImage(ctx context.Context, userID, id int) (domain.ShoppingImage, error) {
	endpoint := fmt.Sprintf("/api/v1/shopping/%d/image", id)
	var response shoppingImageResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "shopping/image", endpoint, nil, nil, &response); err != nil {
		return domain.ShoppingImage{}, err
	}
	filename := fmt.Sprintf("shopping-%d.jpg", id)
	if response.ImageKey != nil && *response.ImageKey != "" {
		filename = path.Base(*response.ImageKey)
	}
	return domain.ShoppingImage{Data: response.ImageData, Filename: filename}, nil
}

func (c *Client) PlanShopping(ctx context.Context, userID int, request domain.AIPlanRequest) (domain.AIPlan, error) {
	requestID := request.RequestID
	if requestID == "" {
		requestID = uuid.NewString()
	}
	body := aiPlanRequestDTO{
		RequestID:    requestID,
		Text:         request.Text,
		StartAddress: request.StartAddress,
		Latitude:     request.Latitude,
		Longitude:    request.Longitude,
	}
	var response aiPlanResponse
	if err := c.doJSON(ctx, userID, http.MethodPost, "ai/plan", "/api/v1/ai/plan", nil, body, &response); err != nil {
		return domain.AIPlan{}, err
	}
	return aiPlanFromDTO(response), nil
}

func (c *Client) ListPlans(ctx context.Context, userID, limit, offset int) (domain.AIPlanPage, error) {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))
	var response aiPlanListResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "ai/plans", "/api/v1/ai/plans", query, nil, &response); err != nil {
		return domain.AIPlanPage{}, err
	}
	page := domain.AIPlanPage{
		Total:  response.Total,
		Limit:  response.Limit,
		Offset: response.Offset,
	}
	for _, item := range response.Plans {
		page.Items = append(page.Items, domain.AIPlanSummary{
			ID:          item.ID,
			RequestID:   item.RequestID,
			Text:        item.Text,
			Address:     item.Address,
			ItemsCount:  item.ItemsCount,
			StoresCount: item.StoresCount,
			TotalCost:   item.TotalCost,
			Recommended: item.Recommended,
			CreatedAt:   item.CreatedAt,
		})
	}
	return page, nil
}

func (c *Client) GetPlan(ctx context.Context, userID, id int) (domain.AIPlan, error) {
	endpoint := fmt.Sprintf("/api/v1/ai/plans/%d", id)
	var response aiPlanResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "ai/plan", endpoint, nil, nil, &response); err != nil {
		return domain.AIPlan{}, err
	}
	return aiPlanFromDTO(response), nil
}

func (c *Client) DeletePlan(ctx context.Context, userID, id int) error {
	endpoint := fmt.Sprintf("/api/v1/ai/plans/%d", id)
	return c.doJSON(ctx, userID, http.MethodDelete, "ai/plan", endpoint, nil, nil, nil)
}

func (c *Client) GetRates(ctx context.Context, userID int, base string) (domain.CurrencyRates, error) {
	base = strings.ToUpper(strings.TrimSpace(base))
	if len(base) != 3 {
		return domain.CurrencyRates{}, fmt.Errorf("invalid base currency %q: %w", base, apperrors.ErrInvalidArgument)
	}
	endpoint := "/api/v1/currency/" + url.PathEscape(base)
	var response currencyRatesResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "currency/rates", endpoint, nil, nil, &response); err != nil {
		return domain.CurrencyRates{}, err
	}
	return domain.CurrencyRates{Base: response.Base, Rates: response.Rates, Timestamp: response.Timestamp}, nil
}

func (c *Client) Convert(ctx context.Context, userID int, from, to string, amount float64) (domain.Conversion, error) {
	body := map[string]any{
		"from":   strings.ToUpper(strings.TrimSpace(from)),
		"to":     strings.ToUpper(strings.TrimSpace(to)),
		"amount": amount,
	}
	var response currencyConvertResponse
	if err := c.doJSON(ctx, userID, http.MethodPost, "currency/convert", "/api/v1/currency/convert", nil, body, &response); err != nil {
		return domain.Conversion{}, err
	}
	return domain.Conversion{
		From:      response.From,
		To:        response.To,
		Amount:    response.Amount,
		Result:    response.Result,
		Rate:      response.Rate,
		Timestamp: response.Timestamp,
	}, nil
}

func (c *Client) GetProfile(ctx context.Context, userID int) (domain.UserProfile, error) {
	endpoint := fmt.Sprintf("/api/v1/user/%d", userID)
	var response userResponse
	if err := c.doJSON(ctx, userID, http.MethodGet, "user/profile", endpoint, nil, nil, &response); err != nil {
		return domain.UserProfile{}, err
	}
	return domain.UserProfile{
		ID:       response.ID,
		FullName: response.FullName,
		Email:    response.Email,
		IsAdmin:  response.IsAdmin,
		Status:   response.Status,
	}, nil
}

func (c *Client) doJSON(ctx context.Context, userID int, method, route, endpoint string, query url.Values, body, out any) error {
	var payload []byte
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		payload = encoded
	}
	data, _, err := c.doRaw(ctx, userID, method, route, endpoint, query, payload)
	if err != nil {
		return err
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode gateway response: %w", err)
	}
	return nil
}

func (c *Client) doRaw(ctx context.Context, userID int, method, route, endpoint string, query url.Values, body []byte) ([]byte, http.Header, error) {
	if userID <= 0 {
		return nil, nil, fmt.Errorf("user id is required: %w", apperrors.ErrUnauthorized)
	}
	token, err := c.tokens.Generate(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("issue service token: %w", err)
	}

	requestURL := c.baseURL + endpoint
	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}

	ctx, span := otel.Tracer("telegram-bot-service").Start(ctx, method+" "+route)
	defer span.End()
	span.SetAttributes(attribute.String("http.request.method", method), attribute.String("url.path", endpoint))

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	request, err := http.NewRequestWithContext(ctx, method, requestURL, reader)
	if err != nil {
		return nil, nil, fmt.Errorf("create gateway request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	response, err := c.http.Do(request)
	elapsed := time.Since(start)
	if err != nil {
		metrics.RecordGateway(method, route, "transport_error", elapsed)
		c.logger.Error("gateway request failed", zap.String("method", method), zap.String("route", route), zap.Error(err))
		return nil, nil, fmt.Errorf("gateway request: %w", apperrors.ErrUnavailable)
	}
	defer response.Body.Close()

	metrics.RecordGateway(method, route, strconv.Itoa(response.StatusCode), elapsed)

	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("read gateway response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		appErr := mapStatus(response.StatusCode, data)
		c.logger.Warn("gateway returned error",
			zap.String("method", method),
			zap.String("route", route),
			zap.Int("status", response.StatusCode),
			zap.String("error", appErr.Error()),
		)
		return nil, nil, appErr
	}

	return data, response.Header, nil
}

func mapStatus(statusCode int, body []byte) error {
	var payload errorResponse
	_ = json.Unmarshal(body, &payload)
	message := payload.Message
	if message == "" {
		message = payload.Error
	}
	if message == "" {
		message = http.StatusText(statusCode)
	}

	var sentinel error
	switch statusCode {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		sentinel = apperrors.ErrInvalidArgument
	case http.StatusUnauthorized:
		sentinel = apperrors.ErrUnauthorized
	case http.StatusForbidden:
		sentinel = apperrors.ErrForbidden
	case http.StatusNotFound:
		sentinel = apperrors.ErrNotFound
	case http.StatusConflict:
		sentinel = apperrors.ErrConflict
	default:
		if statusCode >= http.StatusInternalServerError {
			sentinel = apperrors.ErrUnavailable
		} else {
			sentinel = apperrors.ErrUnexpected
		}
	}
	return fmt.Errorf("%s: %w", message, sentinel)
}

func transactionFromDTO(dto transactionResponse) domain.Transaction {
	return domain.Transaction{
		ID:          dto.ID,
		Version:     dto.Version,
		Type:        dto.TypeTransaction,
		Amount:      dto.Amount,
		Category:    dto.Category,
		Description: dto.Description,
		CreatedAt:   dto.CreatedAt,
		UserID:      dto.UserID,
	}
}

func budgetFromDTO(dto budgetResponse) domain.Budget {
	return domain.Budget{
		ID:          dto.ID,
		Version:     dto.Version,
		UserID:      dto.UserID,
		Category:    dto.Category,
		AmountLimit: dto.AmountLimit,
		Period:      dto.Period,
		PeriodStart: dto.PeriodStart,
		PeriodEnd:   dto.PeriodEnd,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}
}

func shoppingFromDTO(dto shoppingResponse) domain.ShoppingItem {
	return domain.ShoppingItem{
		ID:             dto.ID,
		Version:        dto.Version,
		Title:          dto.Title,
		Description:    dto.Description,
		AmountNow:      dto.AmountNow,
		AmountFinish:   dto.AmountFinish,
		ImageKey:       dto.ImageKey,
		Completed:      dto.Completed,
		CreatedAt:      dto.CreatedAt,
		UpdatedAt:      dto.UpdatedAt,
		CompletedAt:    dto.CompletedAt,
		CompletionDate: dto.CompletionDate,
	}
}

func aiPlanFromDTO(dto aiPlanResponse) domain.AIPlan {
	plan := domain.AIPlan{
		RequestID: dto.RequestID,
		UserID:    dto.UserID,
		Cost: domain.AICost{
			ItemsTotal:   dto.Cost.ItemsTotal,
			TravelCost:   dto.Cost.TravelCost,
			DeliveryCost: dto.Cost.DeliveryCost,
			Total:        dto.Cost.Total,
		},
		Recommendation: domain.AIRecommendation{
			Mode:   dto.Recommend.Mode,
			Total:  dto.Recommend.Total,
			Reason: dto.Recommend.Reason,
		},
		CreatedAt: dto.CreatedAt,
	}
	for _, item := range dto.Items {
		plan.Items = append(plan.Items, domain.AIPlanItem{
			Name:     item.Name,
			Quantity: item.Quantity,
			Unit:     item.Unit,
		})
	}
	for _, store := range dto.Stores {
		storeDomain := domain.AIStore{
			ID:        store.ID,
			Name:      store.Name,
			Address:   store.Address,
			Latitude:  store.Latitude,
			Longitude: store.Longitude,
			SourceURL: store.SourceURL,
		}
		for _, price := range store.Prices {
			storeDomain.Prices = append(storeDomain.Prices, domain.AIItemPrice{
				ItemName: price.ItemName,
				Price:    price.Price,
				Currency: price.Currency,
				InStock:  price.InStock,
			})
		}
		if store.Delivery != nil {
			storeDomain.Delivery = &domain.AIDelivery{
				Available:   store.Delivery.Available,
				Price:       store.Delivery.Price,
				FreeFrom:    store.Delivery.FreeFrom,
				MinOrder:    store.Delivery.MinOrder,
				DurationMin: store.Delivery.DurationMin,
			}
		}
		plan.Stores = append(plan.Stores, storeDomain)
	}
	if dto.Route != nil {
		route := &domain.AIRoute{
			Order:           dto.Route.Order,
			DistanceMeters:  dto.Route.DistanceMeters,
			DurationSeconds: dto.Route.DurationSeconds,
			GeometryGeoJSON: dto.Route.GeometryGeoJSON,
		}
		for _, leg := range dto.Route.Legs {
			route.Legs = append(route.Legs, domain.AIRouteLeg{
				FromIndex:       leg.FromIndex,
				ToIndex:         leg.ToIndex,
				DistanceMeters:  leg.DistanceMeters,
				DurationSeconds: leg.DurationSeconds,
			})
		}
		plan.Route = route
	}
	return plan
}

func filenameFromHeaders(headers http.Header, key string) string {
	disposition := headers.Get("Content-Disposition")
	if disposition != "" {
		for _, part := range strings.Split(disposition, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "filename=") {
				name := strings.Trim(strings.TrimPrefix(part, "filename="), `"`)
				if name != "" {
					return path.Base(name)
				}
			}
		}
	}
	return path.Base(key)
}
