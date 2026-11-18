package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	"go.uber.org/zap"
)

const (
	linkCodeLength   = 8
	maxLinkCodeTries = 5
	linkCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

type AccountService struct {
	accounts      ports.AccountRepository
	notifications ports.NotificationRepository
	logger        *logger.Logger
	linkCodeTTL   time.Duration
	now           func() time.Time
}

func NewAccountService(
	accounts ports.AccountRepository,
	notifications ports.NotificationRepository,
	logger *logger.Logger,
	linkCodeTTL time.Duration,
) *AccountService {
	return &AccountService{
		accounts:      accounts,
		notifications: notifications,
		logger:        logger,
		linkCodeTTL:   linkCodeTTL,
		now:           func() time.Time { return time.Now().UTC() },
	}
}

func (s *AccountService) CreateLinkCode(ctx context.Context, userID int) (ports.LinkCodeResult, error) {
	if userID <= 0 {
		return ports.LinkCodeResult{}, fmt.Errorf("user id is required: %w", apperrors.ErrInvalidArgument)
	}
	_, err := s.accounts.GetAccountByUserID(ctx, userID)

	s.logger.Info("checking existing telegram account",
		zap.Int("user_id", userID),
		zap.Bool("account_found", err == nil),
		zap.Error(err),
	)

	if err == nil {
		return ports.LinkCodeResult{}, fmt.Errorf("account already linked: %w", ErrAlreadyLinked)
	} else if !isNotFound(err) {
		return ports.LinkCodeResult{}, err
	}

	code, err := generateLinkCode()
	if err != nil {
		return ports.LinkCodeResult{}, err
	}
	expiresAt := s.now().Add(s.linkCodeTTL)
	linkCode := domain.LinkCode{
		UserID:    userID,
		CodeHash:  hashLinkCode(code),
		ExpiresAt: expiresAt,
	}
	if err := s.accounts.CreateLinkCode(ctx, linkCode); err != nil {
		return ports.LinkCodeResult{}, err
	}
	return ports.LinkCodeResult{Code: code, ExpiresAt: expiresAt}, nil
}

func (s *AccountService) ConsumeLinkCode(
	ctx context.Context,
	code string,
	telegramUserID, chatID int64,
	username *string,
) (domain.TelegramAccount, error) {
	normalized := normalizeLinkCode(code)
	if len(normalized) != linkCodeLength {
		return domain.TelegramAccount{}, ErrCodeInvalid
	}
	if telegramUserID <= 0 || chatID == 0 {
		return domain.TelegramAccount{}, fmt.Errorf("telegram identity is required: %w", apperrors.ErrInvalidArgument)
	}

	linkCode, err := s.accounts.GetLinkCodeByHash(ctx, hashLinkCode(normalized))
	if err != nil {
		if isNotFound(err) {
			return domain.TelegramAccount{}, ErrCodeInvalid
		}
		return domain.TelegramAccount{}, err
	}

	now := s.now()
	if linkCode.Used() {
		return domain.TelegramAccount{}, ErrCodeUsed
	}
	if linkCode.Expired(now) {
		return domain.TelegramAccount{}, ErrCodeInvalid
	}
	if linkCode.Attempts >= maxLinkCodeTries {
		return domain.TelegramAccount{}, ErrCodeInvalid
	}

	if _, err := s.accounts.GetAccountByTelegramID(ctx, telegramUserID); err == nil {
		return domain.TelegramAccount{}, ErrAlreadyLinked
	} else if !isNotFound(err) {
		return domain.TelegramAccount{}, err
	}

	if _, err := s.accounts.GetAccountByUserID(ctx, linkCode.UserID); err == nil {
		return domain.TelegramAccount{}, ErrAlreadyLinked
	} else if !isNotFound(err) {
		return domain.TelegramAccount{}, err
	}

	if err := s.accounts.MarkLinkCodeUsed(ctx, linkCode.ID); err != nil {
		return domain.TelegramAccount{}, ErrCodeUsed
	}

	account := domain.TelegramAccount{
		TelegramUserID: telegramUserID,
		ChatID:         chatID,
		UserID:         linkCode.UserID,
		Username:       username,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.accounts.CreateAccount(ctx, account); err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return domain.TelegramAccount{}, ErrAlreadyLinked
		}
		return domain.TelegramAccount{}, err
	}

	settings := domain.DefaultNotificationSettings(linkCode.UserID, now)
	if err := s.notifications.UpsertSettings(ctx, settings); err != nil {
		s.logger.Warn("failed to create default notification settings", zap.Error(err))
	}

	s.logger.Info("telegram account linked", zap.Int("user_id", linkCode.UserID))
	return account, nil
}

func (s *AccountService) GetLink(ctx context.Context, userID int) (domain.TelegramAccount, error) {
	if userID <= 0 {
		return domain.TelegramAccount{}, fmt.Errorf("user id is required: %w", apperrors.ErrInvalidArgument)
	}
	account, err := s.accounts.GetAccountByUserID(ctx, userID)
	if err != nil {
		if isNotFound(err) {
			return domain.TelegramAccount{}, ErrNotLinked
		}
		return domain.TelegramAccount{}, err
	}
	return account, nil
}

func (s *AccountService) Resolve(ctx context.Context, telegramUserID int64) (domain.TelegramAccount, error) {
	if telegramUserID <= 0 {
		return domain.TelegramAccount{}, fmt.Errorf("telegram user id is required: %w", apperrors.ErrInvalidArgument)
	}
	account, err := s.accounts.GetAccountByTelegramID(ctx, telegramUserID)
	if err != nil {
		if isNotFound(err) {
			return domain.TelegramAccount{}, ErrNotLinked
		}
		return domain.TelegramAccount{}, err
	}
	return account, nil
}

func (s *AccountService) UnlinkByTelegramID(ctx context.Context, telegramUserID int64) error {
	account, err := s.Resolve(ctx, telegramUserID)
	if err != nil {
		return err
	}
	if err := s.accounts.DeleteAccountByTelegramID(ctx, telegramUserID); err != nil {
		if isNotFound(err) {
			return ErrNotLinked
		}
		return err
	}
	s.deleteSettings(ctx, account.UserID)
	s.logger.Info("telegram account unlinked", zap.Int("user_id", account.UserID))
	return nil
}

func (s *AccountService) UnlinkByUserID(ctx context.Context, userID int) error {
	account, err := s.GetLink(ctx, userID)
	if err != nil {
		return err
	}
	if err := s.accounts.DeleteAccountByUserID(ctx, userID); err != nil {
		if isNotFound(err) {
			return ErrNotLinked
		}
		return err
	}
	s.deleteSettings(ctx, userID)
	s.logger.Info("telegram account unlinked", zap.Int("user_id", account.UserID))
	return nil
}

func (s *AccountService) GetSettings(ctx context.Context, userID int) (domain.NotificationSettings, error) {
	if userID <= 0 {
		return domain.NotificationSettings{}, fmt.Errorf("user id is required: %w", apperrors.ErrInvalidArgument)
	}
	settings, err := s.notifications.GetSettings(ctx, userID)
	if err != nil {
		if isNotFound(err) {
			return domain.DefaultNotificationSettings(userID, s.now()), nil
		}
		return domain.NotificationSettings{}, err
	}
	return settings, nil
}

func (s *AccountService) UpdateSettings(ctx context.Context, settings domain.NotificationSettings) error {
	if settings.UserID <= 0 {
		return fmt.Errorf("user id is required: %w", apperrors.ErrInvalidArgument)
	}
	settings.UpdatedAt = s.now()
	if err := s.notifications.UpsertSettings(ctx, settings); err != nil {
		return err
	}
	return nil
}

func (s *AccountService) deleteSettings(ctx context.Context, userID int) {
	if err := s.notifications.DeleteSettings(ctx, userID); err != nil {
		s.logger.Warn("failed to delete notification settings", zap.Int("user_id", userID), zap.Error(err))
	}
}

func generateLinkCode() (string, error) {
	alphabetLength := big.NewInt(int64(len(linkCodeAlphabet)))
	var builder strings.Builder
	builder.Grow(linkCodeLength)
	for i := 0; i < linkCodeLength; i++ {
		index, err := rand.Int(rand.Reader, alphabetLength)
		if err != nil {
			return "", fmt.Errorf("generate link code: %w", err)
		}
		builder.WriteByte(linkCodeAlphabet[index.Int64()])
	}
	return builder.String(), nil
}

func normalizeLinkCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func hashLinkCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func isNotFound(err error) bool {
	return errors.Is(err, apperrors.ErrNotFound) || errors.Is(err, ErrNotLinked) || errors.Is(err, ErrNoActiveDialog)
}
