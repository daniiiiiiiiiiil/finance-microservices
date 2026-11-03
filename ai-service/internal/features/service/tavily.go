package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/pkg/logger"
	"go.uber.org/zap"
)

const maxTavilyResponse = 5 << 20

type TavilyClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
	logger  *logger.Logger
}

func NewTavilyClient(apiKey, baseURL string, timeout time.Duration, log *logger.Logger) *TavilyClient {
	return &TavilyClient{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
		logger:  log,
	}
}

var _ ports.SearchClient = (*TavilyClient)(nil)

func (c *TavilyClient) SearchStores(
	ctx context.Context,
	item domain.Item,
	nearAddress string,
) ([]domain.Store, error) {
	query := fmt.Sprintf("супермаркет продуктовый магазин %s цена адрес рядом с %s", item.Name, nearAddress)

	payload, err := json.Marshal(map[string]any{
		"query":               query,
		"search_depth":        "advanced",
		"max_results":         5,
		"include_answer":      false,
		"include_raw_content": false,
		"include_domains":     []string{},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal tavily request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/search", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, domain.ErrProviderTimeout
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxTavilyResponse+1))
	if err != nil {
		return nil, fmt.Errorf("%w: read body: %v", domain.ErrProviderBadResponse, err)
	}
	if len(data) > maxTavilyResponse {
		return nil, fmt.Errorf("%w: response too large", domain.ErrProviderBadResponse)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("%w: status %d", domain.ErrProviderUnavailable, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d: %s",
			domain.ErrProviderBadResponse, resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var out struct {
		Results []struct {
			Title   string  `json:"title"`
			URL     string  `json:"url"`
			Content string  `json:"content"`
			Score   float64 `json:"score"`
		} `json:"results"`
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("%w: unmarshal: %v", domain.ErrProviderBadResponse, err)
	}
	if out.Error != "" {
		return nil, fmt.Errorf("%w: %s", domain.ErrProviderBadResponse, out.Error)
	}
	if len(out.Results) == 0 {
		c.logger.Debug("tavily returned no results", zap.String("item", item.Name))
		return nil, nil
	}

	stores := make([]domain.Store, 0, len(out.Results))
	for _, r := range out.Results {
		if r.Score < 0.3 {
			continue
		}
		store := c.parseResult(r.Title, r.URL, r.Content, item)
		if store != nil {
			stores = append(stores, *store)
		}
	}

	if len(stores) == 0 {
		c.logger.Warn("tavily: no parsable stores",
			zap.String("item", item.Name),
			zap.Int("raw_results", len(out.Results)))
		return nil, nil
	}

	return stores, nil
}

var (
	// Цена вида "123,45 руб", "1 234 ₽", "999 RUB".
	priceRe = regexp.MustCompile(`(?i)(\d{1,3}(?:[\s\x{00A0}]\d{3})*(?:[.,]\d{1,2})?)\s*(?:руб|₽|rub|р\.)`)
	// Доставка: "доставка 200 руб", "бесплатная доставка", "от 1000 руб".
	deliveryRe = regexp.MustCompile(`(?i)доставк[а-я]*[^\d]*(\d{1,5})\s*(?:руб|₽)`)
	// Адрес — эвристика: "ул. Ленина, 1", "г. Москва, ул. Тверская, 1".
	addressRe = regexp.MustCompile(`(?i)(?:г\.\s*)?[А-ЯЁ][а-яё-]+(?:[,\s]+(?:ул\.|улица|пр-т|проспект|пер\.|переулок|д\.|дом)?[^|.\n]{3,80})`)
)

func (c *TavilyClient) parseResult(title, url, content string, item domain.Item) *domain.Store {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	storeName := extractStoreName(title, url)
	if storeName == "" {
		return nil
	}

	price, ok := extractPrice(content)
	if !ok {
		return nil
	}

	address := ""
	if m := addressRe.FindString(content); m != "" {
		address = strings.TrimSpace(m)
	}

	delivery := extractDelivery(content)

	return &domain.Store{
		ID:      storeName + ":" + url,
		Name:    storeName,
		Address: address,
		Prices: []domain.ItemPrice{
			{
				ItemName: item.Name,
				Price:    price,
				Currency: "RUB",
				InStock:  true,
			},
		},
		Delivery:  delivery,
		SourceURL: url,
	}
}

func extractStoreName(title, sourceURL string) string {
	for _, sep := range []string{" — ", " – ", " - ", ": ", " | ", " · "} {
		if idx := strings.Index(title, sep); idx > 0 {
			name := strings.TrimSpace(title[:idx])
			if len(name) >= 2 && len(name) <= 60 {
				return name
			}
		}
	}
	u, err := url.Parse(sourceURL)
	if err != nil {
		return ""
	}
	host := strings.TrimPrefix(u.Host, "www.")
	if idx := strings.Index(host, "."); idx > 0 {
		return host[:idx]
	}
	return host
}

func extractPrice(content string) (float64, bool) {
	m := priceRe.FindStringSubmatch(content)
	if len(m) < 2 {
		return 0, false
	}
	raw := strings.ReplaceAll(m[1], "\u00a0", "")
	raw = strings.ReplaceAll(raw, " ", "")
	raw = strings.ReplaceAll(raw, ",", ".")
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v <= 0 || v > 1_000_000 {
		return 0, false
	}
	return v, true
}

func extractDelivery(content string) *domain.Delivery {
	lower := strings.ToLower(content)
	if strings.Contains(lower, "бесплатная доставка") || strings.Contains(lower, "доставка бесплатно") {
		return &domain.Delivery{Available: true, Price: 0}
	}
	if m := deliveryRe.FindStringSubmatch(content); len(m) >= 2 {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			return &domain.Delivery{Available: true, Price: v}
		}
	}
	if strings.Contains(lower, "доставка") {
		return &domain.Delivery{Available: true, Price: 0}
	}
	return nil
}

func parseURL(raw string) (*urlParts, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &urlParts{Host: u.Host}, nil
}

type urlParts struct {
	Host string
}
