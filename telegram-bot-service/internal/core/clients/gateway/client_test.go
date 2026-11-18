package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
)

type staticTokenIssuer struct{}

func (staticTokenIssuer) Generate(userID int) (string, error) {
	return fmt.Sprintf("token-%d", userID), nil
}

func newGatewayTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	log, err := logger.NewLogger(logger.LoggerConfig{Level: "error", Folder: t.TempDir()})
	if err != nil {
		t.Fatalf("create logger: %v", err)
	}
	t.Cleanup(log.Close)
	return log
}

func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewClient(server.URL, staticTokenIssuer{}, 5*time.Second, newGatewayTestLogger(t)), server
}

func TestGetDashboardSendsTokenAndParsesResponse(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token-42" {
			t.Errorf("unexpected authorization header: %s", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/api/v1/finance/dashboard" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"totalBalance": 1500.5,
			"monthlyIncome": 5000,
			"monthlyExpenses": 3499.5,
			"savingsRate": 30,
			"dailyStats": [{"date":"2026-01-01","income":100,"expense":50}],
			"recentTransactions": [{"id":1,"type":"expense","amount":100,"category":"Продукты","createdAt":"2026-01-01T10:00:00Z"}],
			"topCategories": [{"category":"Продукты","total":100,"percentage":50}]
		}`)
	}))

	dashboard, err := client.GetDashboard(context.Background(), 42)
	if err != nil {
		t.Fatalf("get dashboard: %v", err)
	}
	if dashboard.TotalBalance != 1500.5 || len(dashboard.RecentTransactions) != 1 || dashboard.RecentTransactions[0].Category != "Продукты" {
		t.Fatalf("unexpected dashboard: %+v", dashboard)
	}
	if len(dashboard.TopCategories) != 1 || dashboard.TopCategories[0].Percentage != 50 {
		t.Fatalf("unexpected categories: %+v", dashboard.TopCategories)
	}
}

func TestCreateTransactionSendsDescription(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body["description"] != "обед" {
			t.Errorf("unexpected description: %v", body["description"])
		}
		if body["type_transaction"] != "expense" {
			t.Errorf("unexpected type: %v", body["type_transaction"])
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":10,"version":1,"typeTransaction":"expense","amount":250,"category":"Кафе","description":"обед","createdAt":"2026-01-01T12:00:00Z","userId":42}`)
	}))

	transaction, err := client.CreateTransaction(context.Background(), 42, "expense", 250, "Кафе", "обед")
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	if transaction.ID != 10 || transaction.Description != "обед" {
		t.Fatalf("unexpected transaction: %+v", transaction)
	}
}

func TestGetTransactionsSendsFilters(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("type") != "expense" || query.Get("category") != "Кафе" || query.Get("from") != "2026-01-01" || query.Get("to") != "2026-01-31" {
			t.Errorf("unexpected filters: %s", r.URL.RawQuery)
		}
		if query.Get("limit") != "10" || query.Get("offset") != "20" {
			t.Errorf("unexpected pagination: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"transactions":[],"total":0,"limit":10,"offset":20}`)
	}))

	page, err := client.GetTransactions(context.Background(), 42, domain.TransactionFilter{
		Type: "expense", Category: "Кафе", From: "2026-01-01", To: "2026-01-31", Limit: 10, Offset: 20,
	})
	if err != nil {
		t.Fatalf("get transactions: %v", err)
	}
	if page.Limit != 10 || page.Offset != 20 {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func TestGatewayErrorMapping(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"code":5,"message":"transaction not found"}`)
	}))

	_, err := client.GetTransaction(context.Background(), 42, 999)
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "transaction not found") {
		t.Fatalf("expected gateway message, got %v", err)
	}
}

func TestDownloadExportReturnsRawFile(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/finance/export/file" || r.URL.Query().Get("key") != "exports/user_42/report.csv" {
			t.Errorf("unexpected request: %s", r.URL.String())
		}
		w.Header().Set("Content-Disposition", `attachment; filename="report.csv"`)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "id,amount\n1,100")
	}))

	file, err := client.DownloadExport(context.Background(), 42, "exports/user_42/report.csv")
	if err != nil {
		t.Fatalf("download export: %v", err)
	}
	if string(file.Data) != "id,amount\n1,100" || file.Filename != "report.csv" {
		t.Fatalf("unexpected file: %+v", file)
	}
}

func TestPlanShoppingSendsCoordinates(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body aiPlanRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body.Latitude != 55.75 || body.Longitude != 37.61 {
			t.Errorf("unexpected coordinates: %+v", body)
		}
		if body.RequestID == "" {
			t.Error("expected request id")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"requestId":"req","userId":42,"items":[],"stores":[],"cost":{"total":0},"recommendation":{"mode":"self_drive"},"createdAt":"2026-01-01T00:00:00Z"}`)
	}))

	plan, err := client.PlanShopping(context.Background(), 42, domain.AIPlanRequest{
		Text: "молоко", StartAddress: "55.75, 37.61", Latitude: 55.75, Longitude: 37.61,
	})
	if err != nil {
		t.Fatalf("plan shopping: %v", err)
	}
	if plan.RequestID != "req" || plan.Recommendation.Mode != "self_drive" {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestGetRatesUppercasesBase(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/currency/USD" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"base":"USD","rates":{"RUB":91.5},"timestamp":"2026-01-01T00:00:00Z"}`)
	}))

	rates, err := client.GetRates(context.Background(), 42, "usd")
	if err != nil {
		t.Fatalf("get rates: %v", err)
	}
	if rates.Base != "USD" || rates.Rates["RUB"] != 91.5 {
		t.Fatalf("unexpected rates: %+v", rates)
	}
}

func TestClientRejectsInvalidUserID(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	if _, err := client.GetDashboard(context.Background(), 0); err == nil {
		t.Fatal("expected error for invalid user id")
	}
}

func TestDoRawReadsBoundedErrorBody(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"message":"boom"}`)
	}))
	_, err := client.GetProfile(context.Background(), 1)
	if !errors.Is(err, apperrors.ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
}
