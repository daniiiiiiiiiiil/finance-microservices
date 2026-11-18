package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
)

func linkedAccount() domain.TelegramAccount {
	return domain.TelegramAccount{TelegramUserID: 100, ChatID: 200, UserID: 7, CreatedAt: time.Now()}
}

func linkedAccounts() *fakeAccountService {
	return &fakeAccountService{linked: true, account: linkedAccount()}
}

func TestRouterLinksAccountFromStartCode(t *testing.T) {
	accounts := &fakeAccountService{}
	dialogs := newFakeDialogService()
	platform := &fakePlatform{}
	telegram := newFakeTelegram()
	router := newTestRouter(t, accounts, dialogs, platform, telegram)

	if err := router.Process(context.Background(), textUpdate(1, 100, 200, "/start ABCDEFGH")); err != nil {
		t.Fatalf("process start: %v", err)
	}
	if accounts.consumedCode != "ABCDEFGH" {
		t.Fatalf("expected code to be consumed, got %q", accounts.consumedCode)
	}
	if !containsText(telegram.allTexts(), "Аккаунт привязан") {
		t.Fatalf("unexpected messages: %v", telegram.allTexts())
	}
}

func TestRouterShowsMenuForLinkedUser(t *testing.T) {
	accounts := linkedAccounts()
	telegram := newFakeTelegram()
	router := newTestRouter(t, accounts, newFakeDialogService(), &fakePlatform{}, telegram)

	if err := router.Process(context.Background(), textUpdate(2, 100, 200, "привет")); err != nil {
		t.Fatalf("process message: %v", err)
	}
	if !containsText(telegram.allTexts(), "FinGoals") {
		t.Fatalf("expected main menu, got %v", telegram.allTexts())
	}
	last, ok := telegram.lastMessage()
	if !ok || last.Keyboard == nil || len(last.Keyboard.Rows) == 0 {
		t.Fatal("expected inline keyboard")
	}
}

func TestRouterRejectsCallbackForUnlinkedUser(t *testing.T) {
	accounts := &fakeAccountService{linked: false}
	telegram := newFakeTelegram()
	router := newTestRouter(t, accounts, newFakeDialogService(), &fakePlatform{}, telegram)

	if err := router.Process(context.Background(), callbackUpdate(3, 100, 200, "fin")); err != nil {
		t.Fatalf("process callback: %v", err)
	}
	if !containsText(telegram.allTexts(), "не привязан") {
		t.Fatalf("expected unlinked message, got %v", telegram.allTexts())
	}
	if telegram.callbacks != 1 {
		t.Fatalf("expected callback acknowledgement, got %d", telegram.callbacks)
	}
}

func TestTransactionFlowCreatesTransactionWithDescription(t *testing.T) {
	accounts := linkedAccounts()
	dialogs := newFakeDialogService()
	platform := &fakePlatform{categories: []string{"Продукты", "Кафе"}}
	telegram := newFakeTelegram()
	router := newTestRouter(t, accounts, dialogs, platform, telegram)
	ctx := context.Background()

	if err := router.Process(ctx, callbackUpdate(10, 100, 200, "fin:tx:new:expense")); err != nil {
		t.Fatalf("start transaction: %v", err)
	}
	state, ok := dialogs.Get(ctx, 200)
	if !ok || state.Step != domain.StepTransactionAmount || state.Get("type") != "expense" {
		t.Fatalf("unexpected state: %+v", state)
	}

	if err := router.Process(ctx, textUpdate(11, 100, 200, "1 250,50")); err != nil {
		t.Fatalf("amount: %v", err)
	}
	if !containsText(telegram.allTexts(), "Выберите категорию") {
		t.Fatalf("expected category prompt, got %v", telegram.allTexts())
	}

	if err := router.Process(ctx, callbackUpdate(12, 100, 200, "fin:tx:cat:0")); err != nil {
		t.Fatalf("category: %v", err)
	}
	state, _ = dialogs.Get(ctx, 200)
	if state.Step != domain.StepTransactionNote || state.Get("category") != "Продукты" {
		t.Fatalf("unexpected state after category: %+v", state)
	}

	if err := router.Process(ctx, textUpdate(13, 100, 200, "обед с коллегами")); err != nil {
		t.Fatalf("description: %v", err)
	}
	if !containsText(telegram.allTexts(), "Проверьте операцию") {
		t.Fatalf("expected confirmation preview, got %v", telegram.allTexts())
	}

	if err := router.Process(ctx, callbackUpdate(14, 100, 200, "fin:tx:confirm")); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if platform.created == nil {
		t.Fatal("expected transaction to be created")
	}
	if platform.created.Type != "expense" || platform.created.Amount != 1250.5 || platform.created.Category != "Продукты" {
		t.Fatalf("unexpected transaction: %+v", platform.created)
	}
	if platform.created.Description != "обед с коллегами" {
		t.Fatalf("unexpected description: %q", platform.created.Description)
	}
	if !containsText(telegram.allTexts(), "Операция создана") {
		t.Fatalf("expected success message, got %v", telegram.allTexts())
	}
}

func TestTransactionFlowValidatesAmount(t *testing.T) {
	accounts := linkedAccounts()
	dialogs := newFakeDialogService()
	telegram := newFakeTelegram()
	router := newTestRouter(t, accounts, dialogs, &fakePlatform{}, telegram)
	ctx := context.Background()

	if err := router.Process(ctx, callbackUpdate(20, 100, 200, "fin:tx:new:income")); err != nil {
		t.Fatalf("start transaction: %v", err)
	}
	if err := router.Process(ctx, textUpdate(21, 100, 200, "-100")); err != nil {
		t.Fatalf("invalid amount: %v", err)
	}
	if !containsText(telegram.allTexts(), "Сумма должна быть положительным числом") {
		t.Fatalf("expected validation error, got %v", telegram.allTexts())
	}
	state, _ := dialogs.Get(ctx, 200)
	if state.Step != domain.StepTransactionAmount {
		t.Fatalf("state must remain on amount step, got %s", state.Step)
	}
}

func TestTransactionDeleteRequiresConfirmation(t *testing.T) {
	accounts := linkedAccounts()
	platform := &fakePlatform{}
	telegram := newFakeTelegram()
	router := newTestRouter(t, accounts, newFakeDialogService(), platform, telegram)

	if err := router.Process(context.Background(), callbackUpdate(30, 100, 200, "fin:tx:del:5")); err != nil {
		t.Fatalf("delete callback: %v", err)
	}
	last, _ := telegram.lastMessage()
	if last.Keyboard == nil || len(last.Keyboard.Rows) != 1 {
		t.Fatalf("expected confirmation keyboard, got %+v", last)
	}
	found := false
	for _, button := range last.Keyboard.Rows[0] {
		if button.CallbackData == "fin:tx:del:5:yes" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected explicit delete confirmation button")
	}
}

func TestRunnerDeduplicatesUpdates(t *testing.T) {
	accounts := linkedAccounts()
	telegram := newFakeTelegram()
	router := newTestRouter(t, accounts, newFakeDialogService(), &fakePlatform{}, telegram)
	dedup := newFakeDedup()
	runner := NewRunner(router, dedup, newRouterTestLogger(t), 1, 8, time.Hour, 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner.Start(ctx)

	raw, err := json.Marshal(textUpdate(45, 100, 200, "привет"))
	if err != nil {
		t.Fatalf("marshal update: %v", err)
	}
	if err := runner.SubmitRaw(ctx, raw); err != nil {
		t.Fatalf("submit first update: %v", err)
	}
	if err := runner.SubmitRaw(ctx, raw); err != nil {
		t.Fatalf("submit duplicate update: %v", err)
	}

	select {
	case <-telegram.messageCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for update processing")
	}
	select {
	case <-telegram.messageCh:
		t.Fatal("duplicate update was processed twice")
	case <-time.After(150 * time.Millisecond):
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
	defer stopCancel()
	runner.Stop(stopCtx)
}

func TestRunnerRejectsMalformedUpdate(t *testing.T) {
	router := newTestRouter(t, linkedAccounts(), newFakeDialogService(), &fakePlatform{}, newFakeTelegram())
	runner := NewRunner(router, newFakeDedup(), newRouterTestLogger(t), 1, 8, time.Hour, time.Second)
	if err := runner.SubmitRaw(context.Background(), []byte("{")); err == nil {
		t.Fatal("expected malformed update error")
	}
}

func TestShoppingImageUploadUpdatesList(t *testing.T) {
	accounts := linkedAccounts()
	dialogs := newFakeDialogService()
	platform := &fakePlatform{}
	telegram := newFakeTelegram()
	router := newTestRouter(t, accounts, dialogs, platform, telegram)
	ctx := context.Background()

	if err := router.Process(ctx, callbackUpdate(80, 100, 200, "sh:up:5")); err != nil {
		t.Fatalf("start upload: %v", err)
	}
	state, ok := dialogs.Get(ctx, 200)
	if !ok || state.Flow != domain.FlowShopping || state.Step != domain.StepShoppingImage {
		t.Fatalf("unexpected state: %+v", state)
	}

	if err := router.Process(ctx, photoUpdate(81, 100, 200, "large-file")); err != nil {
		t.Fatalf("upload photo: %v", err)
	}
	if platform.updatedShopping == nil {
		t.Fatal("expected shopping update")
	}
	if string(platform.updatedShopping.ImageData) != "image" || platform.updatedShopping.Filename != "photo.jpg" {
		t.Fatalf("unexpected image payload: %+v", platform.updatedShopping)
	}
	if _, ok := dialogs.Get(ctx, 200); ok {
		t.Fatal("expected dialog to be cleared")
	}
}

func TestParseAmount(t *testing.T) {
	cases := []struct {
		input string
		want  float64
		ok    bool
	}{
		{"100", 100, true},
		{"1 250,50", 1250.5, true},
		{"12.346", 12.35, true},
		{"0", 0, false},
		{"-10", 0, false},
		{"abc", 0, false},
		{"", 0, false},
	}
	for _, testCase := range cases {
		got, err := parseAmount(testCase.input)
		if testCase.ok && err != nil {
			t.Fatalf("parse %q: %v", testCase.input, err)
		}
		if !testCase.ok && err == nil {
			t.Fatalf("expected error for %q", testCase.input)
		}
		if testCase.ok && got != testCase.want {
			t.Fatalf("parse %q: expected %v, got %v", testCase.input, testCase.want, got)
		}
	}
}

func TestParseCommandStripsBotName(t *testing.T) {
	command, args := parseCommand("/start@fin_bot ABCDEFGH")
	if command != "/start" || len(args) != 1 || args[0] != "ABCDEFGH" {
		t.Fatalf("unexpected command: %q %v", command, args)
	}
}

func TestRunnerSubmitRawWithoutUpdateIDOrContentIsIgnored(t *testing.T) {
	router := newTestRouter(t, linkedAccounts(), newFakeDialogService(), &fakePlatform{}, newFakeTelegram())
	runner := NewRunner(router, newFakeDedup(), newRouterTestLogger(t), 1, 8, time.Hour, time.Second)
	if err := runner.SubmitRaw(context.Background(), []byte(`{"update_id":1}`)); err != nil {
		t.Fatalf("submit empty update: %v", err)
	}
}

func TestRouterAnswersCallbackEvenWithoutData(t *testing.T) {
	telegram := newFakeTelegram()
	router := newTestRouter(t, linkedAccounts(), newFakeDialogService(), &fakePlatform{}, telegram)
	update := callbackUpdate(70, 100, 200, "")
	if err := router.Process(context.Background(), update); err != nil {
		t.Fatalf("process: %v", err)
	}
	if telegram.callbacks != 1 {
		t.Fatalf("expected callback acknowledgement, got %d", telegram.callbacks)
	}
}

func TestRunnerQueueFullReleasesDedupClaim(t *testing.T) {
	router := newTestRouter(t, linkedAccounts(), newFakeDialogService(), &fakePlatform{}, newFakeTelegram())
	dedup := newFakeDedup()
	runner := NewRunner(router, dedup, newRouterTestLogger(t), 1, 1, time.Hour, time.Second)
	first, _ := json.Marshal(textUpdate(100, 100, 200, "привет"))
	if err := runner.SubmitRaw(context.Background(), first); err != nil {
		t.Fatalf("submit first: %v", err)
	}
	second, _ := json.Marshal(textUpdate(999, 100, 200, "привет"))
	err := runner.SubmitRaw(context.Background(), second)
	if !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}
	claimed, err := dedup.Claim(context.Background(), "telegram:update:999", 60)
	if err != nil || !claimed {
		t.Fatalf("expected dedup claim to be released, claimed=%v err=%v", claimed, err)
	}
}

func TestRunnerRejectsSubmitAfterStop(t *testing.T) {
	router := newTestRouter(t, linkedAccounts(), newFakeDialogService(), &fakePlatform{}, newFakeTelegram())
	runner := NewRunner(router, newFakeDedup(), newRouterTestLogger(t), 1, 1, time.Hour, time.Second)
	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	runner.Stop(stopCtx)

	raw, _ := json.Marshal(textUpdate(500, 100, 200, "привет"))
	if err := runner.SubmitRaw(context.Background(), raw); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull after stop, got %v", err)
	}
}

func TestKeyboardHelpers(t *testing.T) {
	keyboard := ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("A", "a")))
	if keyboard == nil || len(keyboard.Rows) != 1 || keyboard.Rows[0][0].CallbackData != "a" {
		t.Fatalf("unexpected keyboard: %+v", keyboard)
	}
}
