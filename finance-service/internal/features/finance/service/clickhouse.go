package service

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (s *FinanceService) SaveTransactionToAnalytics(ctx context.Context, tx domain.Finance) error {
	if err := s.analyticsRepo.SaveTransaction(ctx, tx); err != nil {
		s.logger.Error("failed to save transaction to clickhouse",
			zap.Int("transaction_id", tx.ID),
			zap.Error(err))
		return err
	}
	s.logger.Info("saved transaction to clickhouse",
		zap.Int("transaction_id", tx.ID))
	return nil
}

func (s *FinanceService) DeleteTransactionFromAnalytics(ctx context.Context, transactionID int) error {
	if err := s.analyticsRepo.DeleteTransaction(ctx, transactionID); err != nil {
		s.logger.Error("failed to delete transaction from clickhouse",
			zap.Int("transaction_id", transactionID),
			zap.Error(err))
		return err
	}
	s.logger.Debug("transaction deleted from clickhouse",
		zap.Int("transaction_id", transactionID))
	return nil
}
