package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
)

type AnalyticsRepository interface {
	GetDashboard(ctx context.Context, userID int) (domain.Dashboard, error)
	GetAnalytics(ctx context.Context, userID int, period domain.AnalyticsPeriod) (domain.Analytics, error)
	SaveTransaction(ctx context.Context, tx domain.Finance) error
	DeleteTransaction(ctx context.Context, transactionID int) error
}
