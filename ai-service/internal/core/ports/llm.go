package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/domain"
)

type LLMClient interface {
	ParseShoppingRequest(ctx context.Context, text string) (*domain.ParsedQuery, error)
}
