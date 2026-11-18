package webapp

import (
	"encoding/hex"
	"errors"
	"net/url"
	"strconv"
	"testing"
	"time"
)

const testBotToken = "123456:test-token"

func signedInitData(t *testing.T, values url.Values) string {
	t.Helper()
	dataCheck := buildDataCheckString(values)
	secret := hmacSHA256([]byte("WebAppData"), []byte(testBotToken))
	values.Set("hash", hex.EncodeToString(hmacSHA256(secret, []byte(dataCheck))))
	return values.Encode()
}

func baseInitValues(authDate time.Time) url.Values {
	return url.Values{
		"auth_date": {strconv.FormatInt(authDate.Unix(), 10)},
		"query_id":  {"AAH-test"},
		"user":      {`{"id":777,"first_name":"Alice","username":"alice"}`},
	}
}

func TestVerifyValidInitData(t *testing.T) {
	now := time.Now().UTC()
	verifier := NewVerifier(testBotToken, 24*time.Hour)
	verifier.now = func() time.Time { return now }

	initData, err := verifier.Verify(signedInitData(t, baseInitValues(now.Add(-time.Minute))))
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if initData.User.ID != 777 || initData.User.Username != "alice" {
		t.Fatalf("unexpected user: %+v", initData.User)
	}
	if initData.QueryID != "AAH-test" {
		t.Fatalf("unexpected query id: %s", initData.QueryID)
	}
}

func TestVerifyRejectsTamperedData(t *testing.T) {
	now := time.Now().UTC()
	verifier := NewVerifier(testBotToken, 24*time.Hour)
	verifier.now = func() time.Time { return now }
	values := baseInitValues(now)
	signedInitData(t, values)

	tampered := values
	tampered.Set("user", `{"id":888,"first_name":"Bob"}`)
	if _, err := verifier.Verify(tampered.Encode()); !errors.Is(err, ErrInvalidInitData) {
		t.Fatalf("expected invalid init data, got %v", err)
	}
}

func TestVerifyRejectsExpiredData(t *testing.T) {
	now := time.Now().UTC()
	verifier := NewVerifier(testBotToken, 10*time.Minute)
	verifier.now = func() time.Time { return now }

	raw := signedInitData(t, baseInitValues(now.Add(-time.Hour)))
	if _, err := verifier.Verify(raw); !errors.Is(err, ErrExpiredInitData) {
		t.Fatalf("expected expired init data, got %v", err)
	}
}

func TestVerifyRejectsFutureAuthDate(t *testing.T) {
	now := time.Now().UTC()
	verifier := NewVerifier(testBotToken, 24*time.Hour)
	verifier.now = func() time.Time { return now }

	raw := signedInitData(t, baseInitValues(now.Add(10*time.Minute)))
	if _, err := verifier.Verify(raw); !errors.Is(err, ErrInvalidInitData) {
		t.Fatalf("expected invalid init data, got %v", err)
	}
}

func TestVerifyRejectsMissingHash(t *testing.T) {
	verifier := NewVerifier(testBotToken, 24*time.Hour)
	if _, err := verifier.Verify(baseInitValues(time.Now()).Encode()); !errors.Is(err, ErrInvalidInitData) {
		t.Fatalf("expected invalid init data, got %v", err)
	}
}

func TestVerifyRejectsWrongBotToken(t *testing.T) {
	now := time.Now().UTC()
	verifier := NewVerifier("another-token", 24*time.Hour)
	verifier.now = func() time.Time { return now }
	raw := signedInitData(t, baseInitValues(now))
	if _, err := verifier.Verify(raw); !errors.Is(err, ErrInvalidInitData) {
		t.Fatalf("expected invalid init data, got %v", err)
	}
}

func TestVerifyRejectsMissingUser(t *testing.T) {
	now := time.Now().UTC()
	verifier := NewVerifier(testBotToken, 24*time.Hour)
	verifier.now = func() time.Time { return now }
	values := url.Values{"auth_date": {strconv.FormatInt(now.Unix(), 10)}}
	raw := signedInitData(t, values)
	if _, err := verifier.Verify(raw); !errors.Is(err, ErrInvalidInitData) {
		t.Fatalf("expected invalid init data, got %v", err)
	}
}

func TestVerifyRejectsEmptyInput(t *testing.T) {
	verifier := NewVerifier(testBotToken, 24*time.Hour)
	if _, err := verifier.Verify(""); !errors.Is(err, ErrInvalidInitData) {
		t.Fatalf("expected invalid init data, got %v", err)
	}
}
