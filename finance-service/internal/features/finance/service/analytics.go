package service

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
)

func (s *FinanceService) GetAnalytics(ctx context.Context, userID int, period string) (domain.Analytics, error) {
	if userID <= 0 {
		return domain.Analytics{}, fmt.Errorf("invalid user id")
	}

	p := domain.AnalyticsPeriod(period)
	switch p {
	case domain.AnalyticsPeriodDay, domain.AnalyticsPeriodWeek, domain.AnalyticsPeriodMonth, domain.AnalyticsPeriodYear:
	default:
		return domain.Analytics{}, fmt.Errorf("invalid period %q: expected day, week, month, or year", period)
	}

	result, err := s.analyticsRepo.GetAnalytics(ctx, userID, p)
	if err != nil {
		return domain.Analytics{}, fmt.Errorf("get financial analytics: %w", err)
	}
	return result, nil
}
