package telegram

import (
	"strings"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
)

func TestFormatAmount(t *testing.T) {
	if got := formatAmount(1234567.5); got != "1 234 567.50" {
		t.Fatalf("unexpected amount: %s", got)
	}
	if got := formatAmount(-1000); got != "-1 000.00" {
		t.Fatalf("unexpected negative amount: %s", got)
	}
}

func TestProgressBar(t *testing.T) {
	if got := progressBar(50, 100); !strings.Contains(got, "50%") || !strings.Contains(got, "▰") {
		t.Fatalf("unexpected progress bar: %s", got)
	}
	if got := progressBar(200, 100); !strings.Contains(got, "200%") {
		t.Fatalf("unexpected clamped progress bar: %s", got)
	}
}

func TestFormatAIPlanContainsCostsAndStores(t *testing.T) {
	plan := domain.AIPlan{
		RequestID: "req",
		Items:     []domain.AIPlanItem{{Name: "Молоко", Quantity: 2, Unit: "л"}},
		Stores: []domain.AIStore{{
			Name:      "Магазин у дома",
			Address:   "ул. Ленина, 1",
			Latitude:  55.75,
			Longitude: 37.61,
			Prices:    []domain.AIItemPrice{{ItemName: "Молоко", Price: 89.9, Currency: "RUB", InStock: true}},
			Delivery:  &domain.AIDelivery{Available: true, Price: 199, DurationMin: 40},
		}},
		Cost:           domain.AICost{ItemsTotal: 179.8, TravelCost: 50, DeliveryCost: 199, Total: 428.8},
		Recommendation: domain.AIRecommendation{Mode: "delivery", Reason: "доставка дешевле"},
		Route:          &domain.AIRoute{DistanceMeters: 5400, DurationSeconds: 900},
	}
	text := formatAIPlan(plan)
	for _, fragment := range []string{"Молоко", "Магазин у дома", "428.80", "доставка дешевле", "5.4 км", "40 мин"} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("expected %q in plan text: %s", fragment, text)
		}
	}
}

func TestParseDate(t *testing.T) {
	date, err := parseDate("2026-01-31")
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}
	if date.Format("2006-01-02") != "2026-01-31" {
		t.Fatalf("unexpected date: %s", date)
	}
	if _, err := parseDate("31.01.2026"); err == nil {
		t.Fatal("expected invalid date error")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("abcdef", 4); got != "abc…" {
		t.Fatalf("unexpected truncation: %s", got)
	}
	if got := truncate("abc", 10); got != "abc" {
		t.Fatalf("unexpected unchanged value: %s", got)
	}
}

func TestFilterDescription(t *testing.T) {
	text := filterDescription(domain.TransactionFilter{Type: "expense", Category: "Кафе", From: "2026-01-01", To: "2026-01-31"})
	for _, fragment := range []string{"расход", "Кафе", "2026-01-01", "2026-01-31"} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("expected %q in filter description: %s", fragment, text)
		}
	}
	if text := filterDescription(domain.TransactionFilter{}); !strings.Contains(text, "не заданы") {
		t.Fatalf("unexpected empty filter description: %s", text)
	}
}

func TestBudgetPeriodLabel(t *testing.T) {
	if budgetPeriodLabel("month") != "месяц" {
		t.Fatal("unexpected month label")
	}
}

func TestFormatDateUsesLocalTime(t *testing.T) {
	value := time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC)
	if formatDate(value) == "" {
		t.Fatal("expected formatted date")
	}
}
