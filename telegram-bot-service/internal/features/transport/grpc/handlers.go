package grpc

import (
	"context"
	"errors"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/service"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/grpcutil/interceptors"
	telegrampb "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/proto/telegram/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateLinkCode(ctx context.Context, _ *telegrampb.CreateLinkCodeRequest) (*telegrampb.CreateLinkCodeResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	result, err := s.accounts.CreateLinkCode(ctx, userID)
	if err != nil {
		return nil, s.mapError("create link code", err)
	}
	return &telegrampb.CreateLinkCodeResponse{
		Code:      result.Code,
		ExpiresAt: timestamppb.New(result.ExpiresAt),
	}, nil
}

func (s *Server) GetLink(ctx context.Context, _ *telegrampb.GetLinkRequest) (*telegrampb.GetLinkResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	account, err := s.accounts.GetLink(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotLinked) {
			return &telegrampb.GetLinkResponse{Linked: false}, nil
		}
		return nil, s.mapError("get link", err)
	}

	response := &telegrampb.GetLinkResponse{
		Linked:         true,
		TelegramUserId: account.TelegramUserID,
		ChatId:         account.ChatID,
		LinkedAt:       timestamppb.New(account.CreatedAt),
	}
	if account.Username != nil {
		response.Username = *account.Username
	}
	return response, nil
}

func (s *Server) Unlink(ctx context.Context, _ *telegrampb.UnlinkRequest) (*telegrampb.UnlinkResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	err := s.accounts.UnlinkByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotLinked) {
			return &telegrampb.UnlinkResponse{Unlinked: false}, nil
		}
		return nil, s.mapError("unlink", err)
	}
	return &telegrampb.UnlinkResponse{Unlinked: true}, nil
}

func (s *Server) GetNotificationSettings(ctx context.Context, _ *telegrampb.GetNotificationSettingsRequest) (*telegrampb.NotificationSettingsResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	settings, err := s.accounts.GetSettings(ctx, userID)
	if err != nil {
		return nil, s.mapError("get notification settings", err)
	}
	return settingsResponse(settings), nil
}

func (s *Server) UpdateNotificationSettings(ctx context.Context, req *telegrampb.UpdateNotificationSettingsRequest) (*telegrampb.NotificationSettingsResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if _, err := s.accounts.GetLink(ctx, userID); err != nil {
		if errors.Is(err, service.ErrNotLinked) {
			return nil, status.Error(codes.FailedPrecondition, "telegram account is not linked")
		}
		return nil, s.mapError("resolve link", err)
	}

	settings, err := s.accounts.GetSettings(ctx, userID)
	if err != nil {
		return nil, s.mapError("get notification settings", err)
	}
	if req.BudgetAlerts != nil {
		settings.BudgetAlerts = *req.BudgetAlerts
	}
	if req.PlanAlerts != nil {
		settings.PlanAlerts = *req.PlanAlerts
	}
	if req.FinanceAlerts != nil {
		settings.FinanceAlerts = *req.FinanceAlerts
	}
	if err := s.accounts.UpdateSettings(ctx, settings); err != nil {
		return nil, s.mapError("update notification settings", err)
	}
	return settingsResponse(settings), nil
}

func settingsResponse(settings domain.NotificationSettings) *telegrampb.NotificationSettingsResponse {
	return &telegrampb.NotificationSettingsResponse{
		BudgetAlerts:  settings.BudgetAlerts,
		PlanAlerts:    settings.PlanAlerts,
		FinanceAlerts: settings.FinanceAlerts,
	}
}

func (s *Server) mapError(operation string, err error) error {
	s.logger.Error("telegram grpc handler failed", zap.String("operation", operation), zap.Error(err))
	switch {
	case errors.Is(err, service.ErrAlreadyLinked):
		return status.Error(codes.AlreadyExists, "telegram account is already linked")
	case errors.Is(err, service.ErrNotLinked):
		return status.Error(codes.NotFound, "telegram account is not linked")
	case errors.Is(err, service.ErrCodeInvalid):
		return status.Error(codes.InvalidArgument, "link code is invalid or expired")
	case errors.Is(err, service.ErrCodeUsed):
		return status.Error(codes.InvalidArgument, "link code has already been used")
	case errors.Is(err, service.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid input")
	case errors.Is(err, apperrors.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, "invalid input")
	case errors.Is(err, apperrors.ErrConflict):
		return status.Error(codes.AlreadyExists, "already exists")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
