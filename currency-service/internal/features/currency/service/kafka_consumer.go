package service_currency

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/kafka"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (s *CurrencyService) StartConsumer(ctx context.Context) {
	config := kafka.NewConfig()
	config.ConsumerGroup = "currency-group"

	consumer := kafka.NewConsumer(config, *s.logger)

	consumer.RegisterHandler(kafka.EventTypeTransactionCreated, s.handleTransactionCreated)
	consumer.RegisterHandler(kafka.EventTypeTransactionDeleted, s.handleTransactionDeleted)

	go func() {
		s.logger.Info("start consumer currency", zap.String("group", "currency-group"))
		if err := consumer.Start(ctx); err != nil {
			s.logger.Error("consumer error", zap.Error(err))
		}
	}()

}

func (s *CurrencyService) handleTransactionCreated(ctx context.Context, event kafka.Event) error {
	if s.isDuplicate(ctx, event.ID) {
		s.logger.Debug("duplicate event ignored", zap.String("event_id", event.ID))
		return nil
	}
	var currencyEvent kafka.TransactionCurrencyEvent
	if err := json.Unmarshal(event.Data, &currencyEvent); err != nil {
		return fmt.Errorf("failed to unmarshal transaction event: %w", err)
	}
	s.logger.Debug("handling currency created", zap.String("event", "currency-create"), zap.Int("TransactionID", currencyEvent.TransactionID))

	currency := currencyEvent.Currency
	if currency == "" {
		currency = "RUB"
	}

	conversion, err := s.Convert(ctx, &domain.Conversion{
		From:   currency,
		To:     "USD",
		Amount: currencyEvent.Amount,
	})
	if err != nil {
		s.logger.Error("failed to convert to USD",
			zap.Int("txID", currencyEvent.TransactionID),
			zap.Error(err),
		)
		return nil
	}

	txUSD := domain.TransactionUSD{
		TransactionID:    currencyEvent.TransactionID,
		AmountUSD:        conversion.Result,
		OriginalAmount:   currencyEvent.Amount,
		OriginalCurrency: currency,
		ConvertedAt:      time.Now(),
	}
	if err := s.rateCache.SetTransactionUSD(ctx, txUSD, 24*time.Hour); err != nil {
		s.logger.Error("failed to save transaction USD", zap.Error(err))
	}

	s.logger.Info("transaction converted to USD",
		zap.Int("txID", currencyEvent.TransactionID),
		zap.Float64("original_amount", currencyEvent.Amount),
		zap.String("currency", currency),
		zap.Float64("usd", conversion.Result),
	)

	if s.producer != nil {
		convertedEvent := kafka.ConvertedEvent{
			TransactionID:    currencyEvent.TransactionID,
			AmountUSD:        conversion.Result,
			OriginalAmount:   currencyEvent.Amount,
			OriginalCurrency: currency,
			ConvertedAt:      time.Now(),
		}
		s.sendConvertedEvent(ctx, convertedEvent)
	}

	return nil
}

func (s *CurrencyService) handleTransactionDeleted(ctx context.Context, event kafka.Event) error {
	if s.isDuplicate(ctx, event.ID) {
		s.logger.Debug("duplicate event ignored", zap.String("event_id", event.ID))
		return nil
	}

	var currencyEvent kafka.TransactionCurrencyEvent
	if err := json.Unmarshal(event.Data, &currencyEvent); err != nil {
		return fmt.Errorf("failed to unmarshal transaction event: %w", err)
	}
	s.logger.Debug("handling currency deleted", zap.String("event", "currency-delete"), zap.Int("TransactionID", currencyEvent.TransactionID))
	if err := s.rateCache.DeleteConvertedUSD(ctx, currencyEvent.TransactionID); err != nil {
		s.logger.Warn("failed to delete converted USD", zap.Error(err))
	}
	return nil
}

func (s *CurrencyService) isDuplicate(ctx context.Context, eventID string) bool {
	key := fmt.Sprintf("currency:processed:%s", eventID)
	exists, err := s.rateCache.Exists(ctx, key)
	if err != nil {
		s.logger.Warn("failed to check duplicate", zap.Error(err))
		return false
	}
	if exists > 0 {
		return true
	}
	if err := s.rateCache.Set(ctx, key, "processed", 24*time.Hour); err != nil {
		s.logger.Warn("failed to save processed marker", zap.Error(err))
	}
	return false
}

func (s *CurrencyService) sendConvertedEvent(ctx context.Context, event kafka.ConvertedEvent) {
	if s.producer == nil {
		return
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("failed to marshal converted event", zap.Error(err))
		return
	}

	if err := s.producer.SendEvent(ctx, "currency.converted", eventData); err != nil {
		s.logger.Error("failed to send converted event", zap.Error(err))
	}
}
