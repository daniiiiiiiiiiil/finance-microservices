package service_currency

import (
	"encoding/json"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/kafka"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (s *CurrencyService) sendCurrencyEvent(ctx context.Context, eventType string, data interface{}) {
	if s.producer == nil {
		return
	}
	eventData, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("failed to marshal event", zap.String("event_type", eventType), zap.Error(err))
		return
	}

	if err := s.producer.SendEvent(ctx, eventType, eventData); err != nil {
		s.logger.Error("failed to send kafka event", zap.String("event_type", eventType), zap.Error(err))
	}
}

func (s *CurrencyService) sendConvertedCurrentEvent(ctx context.Context, tx domain.TransactionUSD) {
	if s.producer == nil {
		return
	}
	eventData := kafka.ConvertedEvent{
		TransactionID:    tx.TransactionID,
		AmountUSD:        tx.AmountUSD,
		OriginalAmount:   tx.OriginalAmount,
		OriginalCurrency: tx.OriginalCurrency,
		ConvertedAt:      time.Now(),
	}
	data, err := json.Marshal(eventData)
	if err != nil {
		s.logger.Error("failed to marshal converted event", zap.Error(err))
		return
	}
	if err := s.producer.SendEvent(ctx, "currency.converted", data); err != nil {
		s.logger.Error("failed to send converted event", zap.Error(err))
	}
}
