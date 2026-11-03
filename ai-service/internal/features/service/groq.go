package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/pkg/logger"
	"go.uber.org/zap"
)

const maxGroqResponse = 3 << 20

const groqSystemPrompt = `Ты — ассистент, который помогает планировать покупки.
Твоя задача — извлечь из сообщения пользователя список товаров и предпочтения.

ПРАВИЛА:
1. Верни ТОЛЬКО валидный JSON-объект, без markdown и без пояснений.
2. Формат:
{
  "items": [
    {"name": "молоко", "quantity": 1, "unit": "шт"},
    {"name": "хлеб", "quantity": 2, "unit": "шт"},
    {"name": "яблоки", "quantity": 1, "unit": "кг"}
  ],
  "preferences": "подешевле, рядом с домом"
}
3. Поле "quantity" — целое число. Если не указано — 1.
4. Поле "unit" — "шт", "кг", "л", "г", "мл". Если не указано — "шт".
5. Поле "preferences" — свободный текст, что важно пользователю (цена, близость, доставка). Если не указано — пустая строка.
6. Не выдумывай товары, которых нет в сообщении.
7. Если в сообщении нет товаров — верни {"items": [], "preferences": ""}.`

type GroqClient struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
	retries int
	logger  *logger.Logger
}

func NewGroqClient(
	apiKey, baseURL, model string,
	timeout time.Duration,
	retries int,
	log *logger.Logger,
) *GroqClient {
	return &GroqClient{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		http:    &http.Client{Timeout: timeout},
		retries: retries,
		logger:  log,
	}
}

var _ ports.LLMClient = (*GroqClient)(nil)

func (c *GroqClient) ParseShoppingRequest(ctx context.Context, text string) (*domain.ParsedQuery, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("%w: empty text", domain.ErrInvalidInput)
	}

	body, err := json.Marshal(map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": groqSystemPrompt},
			{"role": "user", "content": text},
		},
		"temperature":     0.0,
		"max_tokens":      600,
		"response_format": map[string]string{"type": "json_object"},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal groq request: %w", err)
	}

	raw, err := c.complete(ctx, body)
	if err != nil {
		return nil, err
	}

	parsed, err := parseParsedQuery(raw)
	if err != nil {
		c.logger.Warn("failed to parse groq response",
			zap.String("raw", raw),
			zap.Error(err))
		return nil, fmt.Errorf("%w: %v", domain.ErrProviderBadResponse, err)
	}
	return parsed, nil
}

func (c *GroqClient) complete(ctx context.Context, body []byte) (string, error) {
	url := c.baseURL + "/chat/completions"

	for attempt := 0; attempt <= c.retries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return "", fmt.Errorf("new request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return "", domain.ErrProviderTimeout
			}
			if attempt < c.retries {
				if !wait(ctx, retryDelay(attempt)) {
					return "", domain.ErrProviderTimeout
				}
				continue
			}
			return "", fmt.Errorf("%w: %v", domain.ErrProviderUnavailable, err)
		}

		data, readErr := io.ReadAll(io.LimitReader(resp.Body, maxGroqResponse+1))
		_ = resp.Body.Close()
		if readErr != nil {
			return "", fmt.Errorf("%w: read body: %v", domain.ErrProviderBadResponse, readErr)
		}
		if len(data) > maxGroqResponse {
			return "", fmt.Errorf("%w: response too large", domain.ErrProviderBadResponse)
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return "", fmt.Errorf("%w: status %d", domain.ErrProviderUnavailable, resp.StatusCode)
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < c.retries {
			if !wait(ctx, retryDelay(attempt)) {
				return "", domain.ErrProviderTimeout
			}
			continue
		}
		if resp.StatusCode >= 500 && attempt < c.retries {
			if !wait(ctx, retryDelay(attempt)) {
				return "", domain.ErrProviderTimeout
			}
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", fmt.Errorf("%w: status %d: %s",
				domain.ErrProviderBadResponse, resp.StatusCode, strings.TrimSpace(string(data)))
		}

		var out struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &out); err != nil {
			return "", fmt.Errorf("%w: unmarshal: %v", domain.ErrProviderBadResponse, err)
		}
		if out.Error != nil {
			return "", fmt.Errorf("%w: %s", domain.ErrProviderBadResponse, out.Error.Message)
		}
		if len(out.Choices) == 0 {
			return "", fmt.Errorf("%w: empty choices", domain.ErrProviderBadResponse)
		}
		return strings.TrimSpace(out.Choices[0].Message.Content), nil
	}

	return "", domain.ErrProviderUnavailable
}

func parseParsedQuery(raw string) (*domain.ParsedQuery, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var payload struct {
		Items []struct {
			Name     string `json:"name"`
			Quantity int    `json:"quantity"`
			Unit     string `json:"unit"`
		} `json:"items"`
		Preferences string `json:"preferences"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	items := make([]domain.Item, 0, len(payload.Items))
	for _, it := range payload.Items {
		name := strings.TrimSpace(it.Name)
		if name == "" {
			continue
		}
		qty := it.Quantity
		if qty <= 0 {
			qty = 1
		}
		unit := strings.TrimSpace(it.Unit)
		if unit == "" {
			unit = "шт"
		}
		items = append(items, domain.Item{
			Name:     name,
			Quantity: qty,
			Unit:     unit,
		})
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no items extracted from text")
	}

	return &domain.ParsedQuery{
		Items:       items,
		Preferences: strings.TrimSpace(payload.Preferences),
	}, nil
}

func retryDelay(n int) time.Duration {
	if n < 0 {
		n = 0
	}
	if n > 4 {
		n = 4
	}
	d := time.Duration(1<<n) * 250 * time.Millisecond
	if d > 3*time.Second {
		return 3 * time.Second
	}
	return d
}

func wait(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
