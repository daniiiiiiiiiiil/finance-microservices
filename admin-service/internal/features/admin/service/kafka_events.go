package service_admin

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/kafka"
)

func (s *AdminService) sendMetricsEvent(ctx context.Context, metrics Metrics) {
	if s.producer == nil {
		return
	}
	eventData := kafka.MetricsEvent{
		TotalUsers:        metrics.TotalUsers,
		TotalTransactions: metrics.TotalTransactions,
		TotalBalance:      metrics.TotalBalance,
		Timestamp:         time.Now(),
	}
	if err := s.producer.SendEvent(ctx, kafka.EventTypeAdminMetrics, eventData); err != nil {
		fmt.Printf("failed to send kafka event: %v\n", err)
	}
}
