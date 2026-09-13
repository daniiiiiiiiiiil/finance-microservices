package kafka

import (
	"time"
)

const (
	EventTypeUserDeleted           = "user.deleted"
	EventTypeUserRegisterRequested = "user.register.requested"
	EventTypeUserCreated           = "user.created"
	EventTypeUserUpdated           = "user.updated"

	EventTypeUserMarkDeleting    = "user.mark_deleting"
	EventTypeUserRestore         = "user.restore"
	EventTypeUserFinalizeDelete  = "user.finalize_delete"
	EventTypeUserDeleteCompleted = "user.delete.completed"
	EventTypeUserDeleteFailed    = "user.delete.failed"

	EventTypeAuthCreateCredentials = "auth.create_credentials"
	EventTypeAuthDeleteCredentials = "auth.delete_credentials"
	EventTypeAuthActivateAccount   = "auth.activate_account"
	EventTypeUserCreateProfile     = "user.create_profile"
	EventTypeUserDeleteProfile     = "user.delete_profile"

	EventTypeShoppingDeleteUserData = "shopping.delete_user_data"
	EventTypeShoppingDeleted        = "shopping.deleted"

	EventTypeFinanceDeleteUserTransactions = "finance.delete_user_transactions"
	EventTypeFinanceDeleted                = "finance.deleted"
)

type UserDeletedEvent struct {
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	DeletedAt time.Time `json:"deleted_at"`
}

type UserRegisterRequestedEvent struct {
	Email        string  `json:"email"`
	FullName     string  `json:"full_name"`
	PasswordHash string  `json:"password_hash"`
	PhoneNumber  *string `json:"phone_number,omitempty"`
	IsAdmin      bool    `json:"is_admin"`
}

type UserMarkDeletingEvent struct {
	UserID    int       `json:"user_id"`
	SagaID    string    `json:"saga_id"`
	Timestamp time.Time `json:"timestamp"`
}

type UserRestoreEvent struct {
	UserID    int       `json:"user_id"`
	SagaID    string    `json:"saga_id"`
	Timestamp time.Time `json:"timestamp"`
}

type UserFinalizeDeleteEvent struct {
	UserID    int       `json:"user_id"`
	SagaID    string    `json:"saga_id"`
	Timestamp time.Time `json:"timestamp"`
}

type UserDeleteCompletedEvent struct {
	UserID      int       `json:"user_id"`
	SagaID      string    `json:"saga_id"`
	CompletedAt time.Time `json:"completed_at"`
}

type UserDeleteFailedEvent struct {
	UserID   int       `json:"user_id"`
	SagaID   string    `json:"saga_id"`
	Reason   string    `json:"reason"`
	Error    string    `json:"error"`
	FailedAt time.Time `json:"failed_at"`
}

type ShoppingDeleteUserDataEvent struct {
	UserID    int       `json:"user_id"`
	SagaID    string    `json:"saga_id"`
	Timestamp time.Time `json:"timestamp"`
}

type FinanceDeleteUserTransactionsEvent struct {
	UserID    int       `json:"user_id"`
	SagaID    string    `json:"saga_id"`
	Timestamp time.Time `json:"timestamp"`
}
