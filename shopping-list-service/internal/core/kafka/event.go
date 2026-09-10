package kafka

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

const (
	EventTypeUserDeleted = "user.deleted"
	//EventTypeUserUpdated         = "user.updated"
	EventTypeShoppingDeleted     = "shopping.deleted"
	EventTypeShoppingCreated     = "shopping.created"
	EventTypeShoppingUpdated     = "shopping.updated"
	EventTypeImageUploaded       = "shopping.image.uploaded"
	EventTypeImageDeleted        = "shopping.image.deleted"
	EventTypeUserDeleteCompleted = "user.delete.completed"
	EventTypeUserDeleteFailed    = "user.delete.failed"
)

type UserDeletedEvent struct {
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	DeletedAt time.Time `json:"deleted_at"`
}

type UserUpdatedEvent struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type ShoppingDeletedEvent struct {
	UserID       int       `json:"user_id"`
	DeletedCount int       `json:"deleted_count"`
	DeletedIDs   []int     `json:"deleted_ids"`
	ImageKeys    []string  `json:"image_keys"`
	DeletedAt    time.Time `json:"deleted_at"`
}

type ShoppingCreatedEvent struct {
	ShoppingID   int       `json:"shopping_id"`
	UserID       int       `json:"user_id"`
	Title        string    `json:"title"`
	AmountNow    float64   `json:"amount_now"`
	AmountFinish float64   `json:"amount_finish"`
	ImageKey     *string   `json:"image_key,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type ShoppingUpdatedEvent struct {
	ShoppingID int                    `json:"shopping_id"`
	UserID     int                    `json:"user_id"`
	OldValues  map[string]interface{} `json:"old_values"`
	NewValues  map[string]interface{} `json:"new_values"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

type ImageUploadedEvent struct {
	ShoppingID int       `json:"shopping_id"`
	UserID     int       `json:"user_id"`
	ImageKey   string    `json:"image_key"`
	Size       int64     `json:"size"`
	Format     string    `json:"format"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type ImageDeletedEvent struct {
	ShoppingID int       `json:"shopping_id"`
	UserID     int       `json:"user_id"`
	ImageKey   string    `json:"image_key"`
	DeletedAt  time.Time `json:"deleted_at"`
}

type UserDeleteCompletedEvent struct {
	UserID               int       `json:"user_id"`
	Status               string    `json:"status"`
	DeletedShoppingCount int       `json:"deleted_shopping_count"`
	DeletedImagesCount   int       `json:"deleted_images_count"`
	CompletedAt          time.Time `json:"completed_at"`
}

type UserDeleteFailedEvent struct {
	UserID     int       `json:"user_id"`
	Reason     string    `json:"reason"`
	Error      string    `json:"error"`
	FailedStep string    `json:"failed_step"`
	RetryCount int       `json:"retry_count"`
	FailedAt   time.Time `json:"failed_at"`
}
