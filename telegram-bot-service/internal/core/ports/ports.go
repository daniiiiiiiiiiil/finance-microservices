package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
)

type AccountRepository interface {
	CreateLinkCode(ctx context.Context, code domain.LinkCode) error
	GetLinkCodeByHash(ctx context.Context, codeHash string) (domain.LinkCode, error)
	MarkLinkCodeUsed(ctx context.Context, id int64) error
	IncrementLinkCodeAttempts(ctx context.Context, id int64) error
	CreateAccount(ctx context.Context, account domain.TelegramAccount) error
	GetAccountByTelegramID(ctx context.Context, telegramUserID int64) (domain.TelegramAccount, error)
	GetAccountByUserID(ctx context.Context, userID int) (domain.TelegramAccount, error)
	DeleteAccountByTelegramID(ctx context.Context, telegramUserID int64) error
	DeleteAccountByUserID(ctx context.Context, userID int) error
}

type NotificationRepository interface {
	GetSettings(ctx context.Context, userID int) (domain.NotificationSettings, error)
	UpsertSettings(ctx context.Context, settings domain.NotificationSettings) error
	DeleteSettings(ctx context.Context, userID int) error
}

type DialogStore interface {
	Get(ctx context.Context, chatID int64) (domain.DialogState, error)
	Set(ctx context.Context, state domain.DialogState) error
	Delete(ctx context.Context, chatID int64) error
}

type DedupStore interface {
	Claim(ctx context.Context, key string, ttlSeconds int) (bool, error)
	Release(ctx context.Context, key string) error
}

type TelegramAPI interface {
	SendMessage(ctx context.Context, chatID int64, text string, keyboard *Keyboard) (int, error)
	EditMessage(ctx context.Context, chatID int64, messageID int, text string, keyboard *Keyboard) error
	EditReplyMarkup(ctx context.Context, chatID int64, messageID int, keyboard *Keyboard) error
	DeleteMessage(ctx context.Context, chatID int64, messageID int) error
	AnswerCallback(ctx context.Context, callbackID, text string, alert bool) error
	SendPhoto(ctx context.Context, chatID int64, data []byte, filename, caption string) error
	SendDocument(ctx context.Context, chatID int64, data []byte, filename, caption string) error
	SendLocation(ctx context.Context, chatID int64, latitude, longitude float64) error
	DownloadFile(ctx context.Context, fileID string) ([]byte, error)
	SendChatAction(ctx context.Context, chatID int64, action string) error
	SetWebhook(ctx context.Context, url, secretToken string) error
	GetMe(ctx context.Context) (BotIdentity, error)
	StartPolling(ctx context.Context, handler func(ctx context.Context, update []byte) error) error
}

type BotIdentity struct {
	ID       int64
	Username string
}

type PlatformClient interface {
	GetDashboard(ctx context.Context, userID int) (domain.Dashboard, error)
	GetTransactions(ctx context.Context, userID int, filter domain.TransactionFilter) (domain.TransactionPage, error)
	GetTransaction(ctx context.Context, userID, id int) (domain.Transaction, error)
	CreateTransaction(ctx context.Context, userID int, transactionType string, amount float64, category string, description string) (domain.Transaction, error)
	UpdateTransaction(ctx context.Context, userID, id int, transactionType string, amount float64, category string, description string) (domain.Transaction, error)
	DeleteTransaction(ctx context.Context, userID, id int) error
	GetCategories(ctx context.Context, userID int) ([]string, error)
	GetAnalytics(ctx context.Context, userID int, period string) (domain.Analytics, error)
	ListBudgets(ctx context.Context, userID int) ([]domain.BudgetStatus, error)
	CreateBudget(ctx context.Context, userID int, category string, amountLimit float64, period string) (domain.Budget, error)
	UpdateBudget(ctx context.Context, userID, id int, category string, amountLimit float64, period string) (domain.Budget, error)
	DeleteBudget(ctx context.Context, userID, id int) error
	Export(ctx context.Context, userID int, format string) (domain.ExportFile, error)
	DownloadExport(ctx context.Context, userID int, key string) (domain.DownloadedFile, error)

	ListShopping(ctx context.Context, userID, limit, offset int) (domain.ShoppingPage, error)
	GetShopping(ctx context.Context, userID, id int) (domain.ShoppingItem, error)
	CreateShopping(ctx context.Context, userID int, input domain.ShoppingInput) (domain.ShoppingItem, error)
	UpdateShopping(ctx context.Context, userID int, item domain.ShoppingItem, input domain.ShoppingInput) (domain.ShoppingItem, error)
	DeleteShopping(ctx context.Context, userID, id int) error
	CompleteShopping(ctx context.Context, userID, id int, completed bool) error
	ShoppingTotal(ctx context.Context, userID int) (int, error)
	ShoppingImage(ctx context.Context, userID, id int) (domain.ShoppingImage, error)

	PlanShopping(ctx context.Context, userID int, request domain.AIPlanRequest) (domain.AIPlan, error)
	ListPlans(ctx context.Context, userID, limit, offset int) (domain.AIPlanPage, error)
	GetPlan(ctx context.Context, userID, id int) (domain.AIPlan, error)
	DeletePlan(ctx context.Context, userID, id int) error

	GetRates(ctx context.Context, userID int, base string) (domain.CurrencyRates, error)
	Convert(ctx context.Context, userID int, from, to string, amount float64) (domain.Conversion, error)

	GetProfile(ctx context.Context, userID int) (domain.UserProfile, error)
}

type TokenIssuer interface {
	Generate(userID int) (string, error)
}
