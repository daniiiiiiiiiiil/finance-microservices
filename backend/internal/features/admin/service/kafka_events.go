package service_admin

import (
	"backend/internal/core/kafka"
	"context"
	"fmt"
	"time"
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
