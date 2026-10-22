package service

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/kafka"
	"go.uber.org/zap"
)

func (s *FinanceService) CreateBudget(ctx context.Context, b domain.Budget) (domain.Budget, error) {
	if b.PeriodStart.IsZero() || b.PeriodEnd.IsZero() {
		from, to, err := domain.PeriodRange(time.Now(), b.Period)
		if err != nil {
			return domain.Budget{}, fmt.Errorf("period range: %w", err)
		}
		b.PeriodStart, b.PeriodEnd = from, to
	}

	if err := b.Validate(); err != nil {
		return domain.Budget{}, fmt.Errorf("validation failed: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Budget{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != context.Canceled {
			s.logger.Error("failed to rollback budget tx", zap.Error(err))
		}
	}()

	created, err := s.repo.CreateBudgetTx(ctx, tx, b)
	if err != nil {
		return domain.Budget{}, fmt.Errorf("create budget: %w", err)
	}

	event, err := domain.NewOutboxEvent(
		int64(created.ID),
		"budget",
		kafka.EventTypeBudgetCreated,
		created,
	)
	if err != nil {
		return domain.Budget{}, fmt.Errorf("create outbox event: %w", err)
	}

	if err := s.outboxRepo.SaveTx(ctx, tx, event); err != nil {
		return domain.Budget{}, fmt.Errorf("save outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Budget{}, fmt.Errorf("commit transaction: %w", err)
	}

	go s.invalidateCache(context.Background(), created.UserID)

	return created, nil
}

func (s *FinanceService) GetBudget(ctx context.Context, id int) (domain.Budget, error) {
	b, err := s.repo.GetBudget(ctx, id)
	if err != nil {
		return domain.Budget{}, fmt.Errorf("get budget: %w", err)
	}
	return b, nil
}

func (s *FinanceService) ListBudgetsWithStatus(ctx context.Context, userID int) ([]domain.BudgetStatus, error) {
	budgets, err := s.repo.ListBudgets(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}

	out := make([]domain.BudgetStatus, 0, len(budgets))
	for _, b := range budgets {
		spent, err := s.repo.GetBudgetSpent(ctx, b)
		if err != nil {
			return nil, fmt.Errorf("get budget spent: %w", err)
		}

		remaining := b.AmountLimit - spent
		pct := 0.0
		if b.AmountLimit > 0 {
			pct = spent / b.AmountLimit * 100
		}
		out = append(out, domain.BudgetStatus{
			Budget:     b,
			Spent:      spent,
			Remaining:  remaining,
			Percentage: pct,
			Exceeded:   spent > b.AmountLimit,
		})
	}
	return out, nil
}

func (s *FinanceService) UpdateBudget(ctx context.Context, b domain.Budget) (domain.Budget, error) {
	existing, err := s.repo.GetBudget(ctx, b.ID)
	if err != nil {
		return domain.Budget{}, fmt.Errorf("get existing budget: %w", err)
	}

	b.Version = existing.Version
	b.UserID = existing.UserID
	if b.PeriodStart.IsZero() || b.PeriodEnd.IsZero() {
		b.PeriodStart = existing.PeriodStart
		b.PeriodEnd = existing.PeriodEnd
	}

	if err := b.Validate(); err != nil {
		return domain.Budget{}, fmt.Errorf("validation failed: %w", err)
	}

	updated, err := s.repo.UpdateBudget(ctx, b)
	if err != nil {
		return domain.Budget{}, fmt.Errorf("update budget: %w", err)
	}

	event, err := domain.NewOutboxEvent(
		int64(updated.ID),
		"budget",
		kafka.EventTypeBudgetUpdated,
		updated,
	)
	if err != nil {
		return domain.Budget{}, fmt.Errorf("create outbox event: %w", err)
	}
	if err := s.outboxRepo.Save(ctx, event); err != nil {
		return domain.Budget{}, fmt.Errorf("save outbox event: %w", err)
	}

	go s.invalidateCache(context.Background(), updated.UserID)

	return updated, nil
}

func (s *FinanceService) DeleteBudget(ctx context.Context, id int) error {
	existing, err := s.repo.GetBudget(ctx, id)
	if err != nil {
		return fmt.Errorf("get budget: %w", err)
	}

	if err := s.repo.DeleteBudget(ctx, id); err != nil {
		return fmt.Errorf("delete budget: %w", err)
	}

	event, err := domain.NewOutboxEvent(
		int64(existing.ID),
		"budget",
		kafka.EventTypeBudgetDeleted,
		existing,
	)
	if err != nil {
		return fmt.Errorf("create outbox event: %w", err)
	}
	if err := s.outboxRepo.Save(ctx, event); err != nil {
		return fmt.Errorf("save outbox event: %w", err)
	}

	go s.invalidateCache(context.Background(), existing.UserID)

	return nil
}

func (s *FinanceService) CheckBudgetExceeded(ctx context.Context, tx domain.Finance) error {
	if tx.TypeTransaction != "expense" {
		return nil
	}

	for _, p := range []domain.BudgetPeriod{
		domain.BudgetPeriodDay,
		domain.BudgetPeriodWeek,
		domain.BudgetPeriodMonth,
		domain.BudgetPeriodYear,
	} {
		b, ok, err := s.repo.FindActiveBudget(ctx, tx.UserID, tx.Category, p, tx.CreatedAt)
		if err != nil {
			return fmt.Errorf("find active budget: %w", err)
		}
		if !ok {
			continue
		}

		spent, err := s.repo.GetBudgetSpent(ctx, b)
		if err != nil {
			return fmt.Errorf("get budget spent: %w", err)
		}
		if spent <= b.AmountLimit {
			continue
		}

		alertKey := fmt.Sprintf("budget_alerted:%d:%d:%d",
			b.UserID, b.ID, b.PeriodStart.Unix())
		exists, err := s.redis.Exists(ctx, alertKey)
		if err != nil {
			s.logger.Warn("failed to check budget alert key", zap.Error(err))
		} else if exists > 0 {
			continue
		}

		event, err := domain.NewOutboxEvent(
			int64(b.ID),
			"budget",
			kafka.EventTypeBudgetExceeded,
			map[string]any{
				"budget_id":    b.ID,
				"user_id":      b.UserID,
				"category":     b.Category,
				"period":       string(b.Period),
				"period_start": b.PeriodStart,
				"period_end":   b.PeriodEnd,
				"amount_limit": b.AmountLimit,
				"spent":        spent,
				"exceeded_by":  spent - b.AmountLimit,
			},
		)
		if err != nil {
			return fmt.Errorf("create budget.exceeded event: %w", err)
		}
		if err := s.outboxRepo.Save(ctx, event); err != nil {
			return fmt.Errorf("save budget.exceeded event: %w", err)
		}

		ttl := time.Until(b.PeriodEnd)
		if ttl <= 0 {
			ttl = time.Hour
		}
		if err := s.redis.Set(ctx, alertKey, true, ttl); err != nil {
			s.logger.Warn("failed to set budget alert key", zap.Error(err))
		}
	}

	return nil
}
