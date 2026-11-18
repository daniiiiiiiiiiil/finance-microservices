package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	telegrampb "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/proto/telegram/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeGRPCAccounts struct {
	linked          bool
	account         domain.TelegramAccount
	settings        domain.NotificationSettings
	created         bool
	unlinked        bool
	settingsUpdated bool
}

func (f *fakeGRPCAccounts) CreateLinkCode(context.Context, int) (ports.LinkCodeResult, error) {
	f.created = true
	return ports.LinkCodeResult{Code: "ABCDEFGH", ExpiresAt: time.Now().Add(10 * time.Minute)}, nil
}

func (f *fakeGRPCAccounts) ConsumeLinkCode(context.Context, string, int64, int64, *string) (domain.TelegramAccount, error) {
	return f.account, nil
}

func (f *fakeGRPCAccounts) GetLink(context.Context, int) (domain.TelegramAccount, error) {
	if !f.linked {
		return domain.TelegramAccount{}, service.ErrNotLinked
	}
	return f.account, nil
}

func (f *fakeGRPCAccounts) Resolve(context.Context, int64) (domain.TelegramAccount, error) {
	if !f.linked {
		return domain.TelegramAccount{}, service.ErrNotLinked
	}
	return f.account, nil
}

func (f *fakeGRPCAccounts) UnlinkByTelegramID(context.Context, int64) error {
	f.unlinked = true
	return nil
}

func (f *fakeGRPCAccounts) UnlinkByUserID(context.Context, int) error {
	if !f.linked {
		return service.ErrNotLinked
	}
	f.unlinked = true
	f.linked = false
	return nil
}

func (f *fakeGRPCAccounts) GetSettings(context.Context, int) (domain.NotificationSettings, error) {
	if f.settings.UserID == 0 {
		return domain.DefaultNotificationSettings(7, time.Now()), nil
	}
	return f.settings, nil
}

func (f *fakeGRPCAccounts) UpdateSettings(_ context.Context, settings domain.NotificationSettings) error {
	f.settings = settings
	f.settingsUpdated = true
	return nil
}

func newGRPCTestServer(t *testing.T, accounts *fakeGRPCAccounts) *Server {
	t.Helper()
	log, err := logger.NewLogger(logger.LoggerConfig{Level: "error", Folder: t.TempDir()})
	if err != nil {
		t.Fatalf("create logger: %v", err)
	}
	t.Cleanup(log.Close)
	return NewServer(accounts, log)
}

func authenticatedContext(userID int) context.Context {
	return context.WithValue(context.Background(), interceptors.UserIDKey, userID)
}

func TestCreateLinkCodeRequiresAuthentication(t *testing.T) {
	server := newGRPCTestServer(t, &fakeGRPCAccounts{})
	_, err := server.CreateLinkCode(context.Background(), &telegrampb.CreateLinkCodeRequest{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestCreateLinkCodeReturnsCode(t *testing.T) {
	accounts := &fakeGRPCAccounts{}
	server := newGRPCTestServer(t, accounts)
	response, err := server.CreateLinkCode(authenticatedContext(7), &telegrampb.CreateLinkCodeRequest{})
	if err != nil {
		t.Fatalf("create link code: %v", err)
	}
	if response.Code != "ABCDEFGH" || response.ExpiresAt == nil {
		t.Fatalf("unexpected response: %+v", response)
	}
	if !accounts.created {
		t.Fatal("expected account service call")
	}
}

func TestGetLinkReturnsUnlinkedState(t *testing.T) {
	server := newGRPCTestServer(t, &fakeGRPCAccounts{linked: false})
	response, err := server.GetLink(authenticatedContext(7), &telegrampb.GetLinkRequest{})
	if err != nil {
		t.Fatalf("get link: %v", err)
	}
	if response.Linked {
		t.Fatal("expected linked=false")
	}
}

func TestGetLinkReturnsAccountDetails(t *testing.T) {
	username := "alice"
	accounts := &fakeGRPCAccounts{
		linked:  true,
		account: domain.TelegramAccount{TelegramUserID: 100, ChatID: 200, UserID: 7, Username: &username, CreatedAt: time.Now()},
	}
	server := newGRPCTestServer(t, accounts)
	response, err := server.GetLink(authenticatedContext(7), &telegrampb.GetLinkRequest{})
	if err != nil {
		t.Fatalf("get link: %v", err)
	}
	if !response.Linked || response.TelegramUserId != 100 || response.Username != username {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestUnlinkReturnsFalseWhenNotLinked(t *testing.T) {
	server := newGRPCTestServer(t, &fakeGRPCAccounts{linked: false})
	response, err := server.Unlink(authenticatedContext(7), &telegrampb.UnlinkRequest{})
	if err != nil {
		t.Fatalf("unlink: %v", err)
	}
	if response.Unlinked {
		t.Fatal("expected unlinked=false")
	}
}

func TestUnlinkReturnsTrue(t *testing.T) {
	server := newGRPCTestServer(t, &fakeGRPCAccounts{linked: true})
	response, err := server.Unlink(authenticatedContext(7), &telegrampb.UnlinkRequest{})
	if err != nil {
		t.Fatalf("unlink: %v", err)
	}
	if !response.Unlinked {
		t.Fatal("expected unlinked=true")
	}
}

func TestGetNotificationSettingsReturnsDefaults(t *testing.T) {
	server := newGRPCTestServer(t, &fakeGRPCAccounts{})
	response, err := server.GetNotificationSettings(authenticatedContext(7), &telegrampb.GetNotificationSettingsRequest{})
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if !response.BudgetAlerts || !response.PlanAlerts || response.FinanceAlerts {
		t.Fatalf("unexpected defaults: %+v", response)
	}
}

func TestUpdateNotificationSettingsRequiresLink(t *testing.T) {
	server := newGRPCTestServer(t, &fakeGRPCAccounts{linked: false})
	enabled := true
	_, err := server.UpdateNotificationSettings(authenticatedContext(7), &telegrampb.UpdateNotificationSettingsRequest{FinanceAlerts: &enabled})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", err)
	}
}

func TestUpdateNotificationSettingsAppliesPatch(t *testing.T) {
	accounts := &fakeGRPCAccounts{linked: true, settings: domain.DefaultNotificationSettings(7, time.Now())}
	server := newGRPCTestServer(t, accounts)
	enabled := true
	disabled := false
	response, err := server.UpdateNotificationSettings(authenticatedContext(7), &telegrampb.UpdateNotificationSettingsRequest{
		FinanceAlerts: &enabled,
		PlanAlerts:    &disabled,
	})
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if !response.FinanceAlerts || response.PlanAlerts {
		t.Fatalf("unexpected response: %+v", response)
	}
	if !accounts.settingsUpdated {
		t.Fatal("expected settings update")
	}
}
