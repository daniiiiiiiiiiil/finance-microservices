package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/config"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/auth/jwt"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/service"
	telegramtransport "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/transport/telegram"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/webapp"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
)

const httpTestBotToken = "123456:http-test-token"

type fakeSubmitter struct {
	raw    []byte
	called bool
	err    error
}

func (f *fakeSubmitter) SubmitRaw(_ context.Context, raw []byte) error {
	f.called = true
	f.raw = append([]byte(nil), raw...)
	return f.err
}

type fakeHTTPAccounts struct {
	linked  bool
	account domain.TelegramAccount
}

func (f *fakeHTTPAccounts) CreateLinkCode(context.Context, int) (ports.LinkCodeResult, error) {
	return ports.LinkCodeResult{}, nil
}

func (f *fakeHTTPAccounts) ConsumeLinkCode(context.Context, string, int64, int64, *string) (domain.TelegramAccount, error) {
	return domain.TelegramAccount{}, nil
}

func (f *fakeHTTPAccounts) GetLink(context.Context, int) (domain.TelegramAccount, error) {
	return domain.TelegramAccount{}, nil
}

func (f *fakeHTTPAccounts) Resolve(context.Context, int64) (domain.TelegramAccount, error) {
	if !f.linked {
		return domain.TelegramAccount{}, service.ErrNotLinked
	}
	return f.account, nil
}

func (f *fakeHTTPAccounts) UnlinkByTelegramID(context.Context, int64) error { return nil }

func (f *fakeHTTPAccounts) UnlinkByUserID(context.Context, int) error { return nil }

func (f *fakeHTTPAccounts) GetSettings(context.Context, int) (domain.NotificationSettings, error) {
	return domain.DefaultNotificationSettings(7, time.Now()), nil
}

func (f *fakeHTTPAccounts) UpdateSettings(context.Context, domain.NotificationSettings) error {
	return nil
}

func newHTTPTestServer(t *testing.T, accounts ports.AccountServiceInterface, submitter UpdateSubmitter) (*Server, *jwt.Manager) {
	t.Helper()
	log, err := logger.NewLogger(logger.LoggerConfig{Level: "error", Folder: t.TempDir()})
	if err != nil {
		t.Fatalf("create logger: %v", err)
	}
	t.Cleanup(log.Close)

	cfg := &config.Config{
		TelegramWebhookSecret: "webhook-secret",
		TelegramWebAppPath:    "/telegram/webapp",
		TelegramHTTPAddr:      ":0",
		WebAppTokenTTL:        15 * time.Minute,
		GatewayPublicURL:      "http://gateway:8081",
	}
	tokens := jwt.NewManager("jwt-secret", 15*time.Minute)
	server := NewServer(cfg, submitter, accounts, webapp.NewVerifier(httpTestBotToken, time.Hour), tokens, log, nil, nil)
	return server, tokens
}

func signedHTTPInitData(t *testing.T, authDate time.Time, userID int64) string {
	t.Helper()
	values := url.Values{
		"auth_date": {strconv.FormatInt(authDate.Unix(), 10)},
		"user":      {`{"id":` + strconv.FormatInt(userID, 10) + `,"first_name":"Alice"}`},
	}
	dataCheck := ""
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	for index, key := range keys {
		if index > 0 {
			dataCheck += "\n"
		}
		dataCheck += key + "=" + values.Get(key)
	}
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(httpTestBotToken))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	mac.Write([]byte(dataCheck))
	values.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return values.Encode()
}

func TestWebhookRejectsInvalidSecret(t *testing.T) {
	submitter := &fakeSubmitter{}
	server, _ := newHTTPTestServer(t, &fakeHTTPAccounts{}, submitter)
	request := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"update_id":1}`))
	request.Header.Set(webhookSecretHeader, "wrong")
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
	if submitter.called {
		t.Fatal("submitter must not be called")
	}
}

func TestWebhookAcceptsValidUpdate(t *testing.T) {
	submitter := &fakeSubmitter{}
	server, _ := newHTTPTestServer(t, &fakeHTTPAccounts{}, submitter)
	body := `{"update_id":42,"message":{"message_id":1,"date":0,"chat":{"id":100,"type":"private"},"text":"/start"}}`
	request := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(body))
	request.Header.Set(webhookSecretHeader, "webhook-secret")
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !submitter.called || string(submitter.raw) != body {
		t.Fatal("expected raw update to be submitted")
	}
}

func TestWebhookReturnsServiceUnavailableWhenBusy(t *testing.T) {
	submitter := &fakeSubmitter{err: telegramtransport.ErrQueueFull}
	server, _ := newHTTPTestServer(t, &fakeHTTPAccounts{}, submitter)
	request := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"update_id":1,"message":{"message_id":1,"date":0,"chat":{"id":1,"type":"private"},"text":"/start"}}`))
	request.Header.Set(webhookSecretHeader, "webhook-secret")
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
}

func TestWebAppSessionIssuesToken(t *testing.T) {
	accounts := &fakeHTTPAccounts{linked: true, account: domain.TelegramAccount{TelegramUserID: 100, ChatID: 200, UserID: 7}}
	server, tokens := newHTTPTestServer(t, accounts, &fakeSubmitter{})
	body, _ := json.Marshal(sessionRequest{InitData: signedHTTPInitData(t, time.Now(), 100)})
	request := httptest.NewRequest(http.MethodPost, "/telegram/webapp/session", strings.NewReader(string(body)))
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var response sessionResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Token == "" || response.GatewayURL != "http://gateway:8081" || response.User.ID != 100 {
		t.Fatalf("unexpected response: %+v", response)
	}
	claims, err := tokens.Validate(response.Token)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if claims.UserID != 7 {
		t.Fatalf("expected token for user 7, got %d", claims.UserID)
	}
}

func TestWebAppSessionRejectsInvalidInitData(t *testing.T) {
	server, _ := newHTTPTestServer(t, &fakeHTTPAccounts{linked: true}, &fakeSubmitter{})
	body, _ := json.Marshal(sessionRequest{InitData: "auth_date=1&user=%7B%7D&hash=deadbeef"})
	request := httptest.NewRequest(http.MethodPost, "/telegram/webapp/session", strings.NewReader(string(body)))
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestWebAppSessionRejectsUnlinkedUser(t *testing.T) {
	server, _ := newHTTPTestServer(t, &fakeHTTPAccounts{linked: false}, &fakeSubmitter{})
	body, _ := json.Marshal(sessionRequest{InitData: signedHTTPInitData(t, time.Now(), 100)})
	request := httptest.NewRequest(http.MethodPost, "/telegram/webapp/session", strings.NewReader(string(body)))
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestWebAppPageServed(t *testing.T) {
	server, _ := newHTTPTestServer(t, &fakeHTTPAccounts{}, &fakeSubmitter{})
	request := httptest.NewRequest(http.MethodGet, "/telegram/webapp", nil)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "telegram-web-app.js") {
		t.Fatal("expected telegram web app page")
	}
}

func TestHealthWithoutDependencies(t *testing.T) {
	server, _ := newHTTPTestServer(t, &fakeHTTPAccounts{}, &fakeSubmitter{})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestWebhookRejectsWrongMethod(t *testing.T) {
	server, _ := newHTTPTestServer(t, &fakeHTTPAccounts{}, &fakeSubmitter{})
	request := httptest.NewRequest(http.MethodGet, "/telegram/webhook", nil)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", recorder.Code)
	}
}
