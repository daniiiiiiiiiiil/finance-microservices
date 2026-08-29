package service_currency

import (
	"fmt"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/domain"
	"go.uber.org/zap"
)

func (s *CurrencyService) GetTransactionUSD(ctx context.Context, txID int) (domain.TransactionUSD, error) {
	if txID <= 0 {
		return domain.TransactionUSD{}, fmt.Errorf("txID must be greater than zero")
	}

	tx, err := s.rateCache.GetTransactionUSD(ctx, txID)
	if err != nil {
		s.logger.Warn("converted USD not found in cache", zap.Int("txID", txID), zap.Error(err))
		return domain.TransactionUSD{}, err
	}

	s.logger.Debug("converted USD found in cache",
		zap.Int("txID", txID),
		zap.Float64("usd", tx.AmountUSD))
	return tx, nil
}
