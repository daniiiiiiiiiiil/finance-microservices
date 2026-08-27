package service_currency

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/kafka"
	"go.uber.org/zap"
)

func (s *CurrencyService) ConvertTransactionToUSD(ctx context.Context, tx kafka.TransactionCurrencyEvent) error {
	currency := tx.Currency
	if currency == "" {
		currency = "RUB"
	}

	conversion, err := s.Convert(ctx, &domain.Conversion{
		From:   currency,
		To:     "USD",
		Amount: tx.Amount,
	})
	if err != nil {
		s.logger.Error("failed to convert to USD",
			zap.Int("txID", tx.TransactionID),
			zap.Error(err),
		)
		return fmt.Errorf("convert to usd: %w", err)
	}

	txUSD := domain.TransactionUSD{
		TransactionID:    tx.TransactionID,
		AmountUSD:        conversion.Result,
		OriginalAmount:   tx.Amount,
		OriginalCurrency: currency,
		ConvertedAt:      time.Now(),
	}

	if err := s.rateCache.SetTransactionUSD(ctx, txUSD, 24*time.Hour); err != nil {
		s.logger.Error("failed to save transaction USD", zap.Error(err))
		return fmt.Errorf("save transaction usd: %w", err)
	}

	s.logger.Info("transaction converted to USD",
		zap.Int("txID", tx.TransactionID),
		zap.Float64("original_amount", tx.Amount),
		zap.String("currency", currency),
		zap.Float64("usd", conversion.Result),
	)

	return nil
}

func (s *CurrencyService) DeleteConvertedUSD(ctx context.Context, txID int) error {
	if err := s.rateCache.DeleteConvertedUSD(ctx, txID); err != nil {
		s.logger.Warn("failed to delete converted USD", zap.Error(err))
		return fmt.Errorf("delete converted usd: %w", err)
	}

	s.logger.Debug("deleted converted USD",
		zap.Int("txID", txID),
	)

	return nil
}

func (s *CurrencyService) StartConsumer(ctx context.Context) {
	if s.eventConsumer == nil {
		return
	}

	s.eventConsumer.RegisterHandler(kafka.EventTypeTransactionCreated, func(ctx context.Context, event kafka.Event) error {
		var currencyEvent kafka.TransactionCurrencyEvent
		if err := json.Unmarshal(event.Data, &currencyEvent); err != nil {
			return fmt.Errorf("failed to unmarshal transaction event: %w", err)
		}
		return s.ConvertTransactionToUSD(ctx, currencyEvent)
	})

	s.eventConsumer.RegisterHandler(kafka.EventTypeTransactionDeleted, func(ctx context.Context, event kafka.Event) error {
		var currencyEvent kafka.TransactionCurrencyEvent
		if err := json.Unmarshal(event.Data, &currencyEvent); err != nil {
			return fmt.Errorf("failed to unmarshal transaction event: %w", err)
		}
		return s.DeleteConvertedUSD(ctx, currencyEvent.TransactionID)
	})

	go func() {
		s.logger.Info("start consumer currency", zap.String("group", "currency-group"))
		if err := s.eventConsumer.Start(ctx); err != nil {
			s.logger.Error("consumer error", zap.Error(err))
		}
	}()
}
