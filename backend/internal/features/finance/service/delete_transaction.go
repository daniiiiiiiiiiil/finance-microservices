package service

import (
	"context"
	"fmt"
)

func (s *FinanceService) DeleteTransaction(ctx context.Context, id int) error {
	if err := s.repo.DeleteTransaction(ctx, id); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	return nil
}
