package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/ports"
)

func (p *AdminEventPublisher) SendMetricsEvent(ctx context.Context, metrics ports.Metrics) {
	eventData := kafka.MetricsEvent{
		TotalUsers:        metrics.TotalUsers,
		TotalTransactions: metrics.TotalTransactions,
		TotalBalance:      metrics.TotalBalance,
		Timestamp:         time.Now(),
	}
	if err := p.Publish(ctx, kafka.EventTypeAdminMetrics, eventData); err != nil {
		fmt.Printf("failed to send kafka event: %v\n", err)
	}
}
