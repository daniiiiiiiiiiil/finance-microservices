package webapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidInitData = fmt.Errorf("invalid telegram init data")
	ErrExpiredInitData = fmt.Errorf("telegram init data expired")
)

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type InitData struct {
	User       User
	AuthDate   time.Time
	QueryID    string
	ChatID     int64
	ChatType   string
	StartParam string
}

type Verifier struct {
	botToken string
	maxAge   time.Duration
	now      func() time.Time
}

func NewVerifier(botToken string, maxAge time.Duration) *Verifier {
	return &Verifier{
		botToken: botToken,
		maxAge:   maxAge,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (v *Verifier) Verify(raw string) (InitData, error) {
	if strings.TrimSpace(raw) == "" {
		return InitData{}, ErrInvalidInitData
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return InitData{}, ErrInvalidInitData
	}

	providedHash := values.Get("hash")
	if providedHash == "" {
		return InitData{}, ErrInvalidInitData
	}
	values.Del("hash")

	dataCheckString := buildDataCheckString(values)
	secret := hmacSHA256([]byte("WebAppData"), []byte(v.botToken))
	expected := hmacSHA256(secret, []byte(dataCheckString))

	provided, err := hex.DecodeString(strings.ToLower(providedHash))
	if err != nil {
		return InitData{}, ErrInvalidInitData
	}
	if !hmac.Equal(provided, expected) {
		return InitData{}, ErrInvalidInitData
	}

	authDateUnix, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil {
		return InitData{}, ErrInvalidInitData
	}
	authDate := time.Unix(authDateUnix, 0).UTC()
	now := v.now()
	if authDate.After(now.Add(5 * time.Minute)) {
		return InitData{}, ErrInvalidInitData
	}
	if v.maxAge > 0 && now.Sub(authDate) > v.maxAge {
		return InitData{}, ErrExpiredInitData
	}

	var user User
	if err := json.Unmarshal([]byte(values.Get("user")), &user); err != nil {
		return InitData{}, ErrInvalidInitData
	}
	if user.ID <= 0 {
		return InitData{}, ErrInvalidInitData
	}

	initData := InitData{
		User:       user,
		AuthDate:   authDate,
		QueryID:    values.Get("query_id"),
		ChatType:   values.Get("chat_type"),
		StartParam: values.Get("start_param"),
	}
	if chatID := values.Get("chat_instance"); chatID != "" {
		if parsed, err := strconv.ParseInt(chatID, 10, 64); err == nil {
			initData.ChatID = parsed
		}
	}
	return initData, nil
}

func buildDataCheckString(values url.Values) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for index, key := range keys {
		if index > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(values.Get(key))
	}
	return builder.String()
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
