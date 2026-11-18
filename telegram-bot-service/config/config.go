package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

const (
	ModeWebhook = "webhook"
	ModePolling = "polling"
)

var webhookSecretPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`)

type Config struct {
	TimeZone    *time.Location
	JWTSecret   string
	JWTDuration time.Duration

	TLSEnabled  bool
	TLSCertFile string
	TLSKeyFile  string
	TLSCAFile   string

	TelegramToken         string
	TelegramMode          string
	TelegramWebhookURL    string
	TelegramWebhookSecret string
	TelegramHTTPAddr      string
	TelegramGRPCAddr      string
	TelegramMetricsAddr   string
	TelegramPublicURL     string
	TelegramWebAppPath    string
	TelegramWebAppMaxAge  time.Duration
	WebAppTokenTTL        time.Duration
	ServiceTokenTTL       time.Duration

	LinkCodeTTL       time.Duration
	DialogTTL         time.Duration
	UpdateDedupTTL    time.Duration
	EventDedupTTL     time.Duration
	GatewayBaseURL    string
	GatewayPublicURL  string
	HTTPTimeout       time.Duration
	NotificationTries int

	RedisAddr string
}

func NewConfig() (*Config, error) {
	tz := envOr("TIME_ZONE", "UTC")
	zone, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("failed to load time zone %q: %w", tz, err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	jwtDuration, err := durationEnv("JWT_DURATION", "24h")
	if err != nil {
		return nil, err
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	mode, err := telegramMode()
	if err != nil {
		return nil, err
	}

	publicURL := trimRightSlash(envOr("TELEGRAM_PUBLIC_URL", "http://localhost:8082"))
	webhookURL := os.Getenv("TELEGRAM_WEBHOOK_URL")
	webhookSecret := os.Getenv("TELEGRAM_WEBHOOK_SECRET")
	if mode == ModeWebhook {
		if webhookURL == "" {
			webhookURL = publicURL + "/telegram/webhook"
		}
		webhookURL = trimRightSlash(webhookURL)
		if webhookSecret == "" {
			return nil, fmt.Errorf("TELEGRAM_WEBHOOK_SECRET is required in webhook mode")
		}
	}
	if webhookSecret != "" && !webhookSecretPattern.MatchString(webhookSecret) {
		return nil, fmt.Errorf("TELEGRAM_WEBHOOK_SECRET must match %s", webhookSecretPattern.String())
	}

	linkCodeTTL, err := durationEnv("TELEGRAM_LINK_CODE_TTL", "10m")
	if err != nil {
		return nil, err
	}
	dialogTTL, err := durationEnv("TELEGRAM_DIALOG_TTL", "30m")
	if err != nil {
		return nil, err
	}
	updateDedupTTL, err := durationEnv("TELEGRAM_UPDATE_DEDUP_TTL", "24h")
	if err != nil {
		return nil, err
	}
	eventDedupTTL, err := durationEnv("TELEGRAM_EVENT_DEDUP_TTL", "168h")
	if err != nil {
		return nil, err
	}
	webAppMaxAge, err := durationEnv("TELEGRAM_WEBAPP_MAX_AGE", "24h")
	if err != nil {
		return nil, err
	}
	webAppTokenTTL, err := durationEnv("TELEGRAM_WEBAPP_TOKEN_TTL", "15m")
	if err != nil {
		return nil, err
	}
	serviceTokenTTL, err := durationEnv("TELEGRAM_SERVICE_TOKEN_TTL", "5m")
	if err != nil {
		return nil, err
	}
	httpTimeout, err := durationEnv("TELEGRAM_HTTP_TIMEOUT", "120s")
	if err != nil {
		return nil, err
	}

	notificationTries, err := intEnv("TELEGRAM_NOTIFICATION_RETRIES", 3)
	if err != nil {
		return nil, err
	}
	if notificationTries < 1 || notificationTries > 10 {
		return nil, fmt.Errorf("TELEGRAM_NOTIFICATION_RETRIES must be between 1 and 10")
	}

	tlsEnabled, _ := strconv.ParseBool(os.Getenv("TLS_ENABLED"))

	return &Config{
		TimeZone:    zone,
		JWTSecret:   jwtSecret,
		JWTDuration: jwtDuration,

		TLSEnabled:  tlsEnabled,
		TLSCertFile: os.Getenv("TLS_CERT_FILE"),
		TLSKeyFile:  os.Getenv("TLS_KEY_FILE"),
		TLSCAFile:   os.Getenv("TLS_CA_FILE"),

		TelegramToken:         token,
		TelegramMode:          mode,
		TelegramWebhookURL:    webhookURL,
		TelegramWebhookSecret: webhookSecret,
		TelegramHTTPAddr:      envOr("TELEGRAM_HTTP_ADDR", ":8082"),
		TelegramGRPCAddr:      envOr("TELEGRAM_GRPC_ADDR", ":50062"),
		TelegramMetricsAddr:   envOr("TELEGRAM_METRICS_ADDR", ":9100"),
		TelegramPublicURL:     publicURL,
		TelegramWebAppPath:    envOr("TELEGRAM_WEBAPP_PATH", "/telegram/webapp"),
		TelegramWebAppMaxAge:  webAppMaxAge,
		WebAppTokenTTL:        webAppTokenTTL,
		ServiceTokenTTL:       serviceTokenTTL,

		LinkCodeTTL:       linkCodeTTL,
		DialogTTL:         dialogTTL,
		UpdateDedupTTL:    updateDedupTTL,
		EventDedupTTL:     eventDedupTTL,
		GatewayBaseURL:    trimRightSlash(envOr("GATEWAY_URL", "http://gateway:8081")),
		GatewayPublicURL:  trimRightSlash(envOr("GATEWAY_PUBLIC_URL", "http://localhost:8081")),
		HTTPTimeout:       httpTimeout,
		NotificationTries: notificationTries,

		RedisAddr: envOr("REDIS_ADDR", "localhost:6379"),
	}, nil
}

func NewConfigMust() *Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to create config: %v", err))
	}
	return config
}

func (c *Config) WebAppURL() string {
	return c.TelegramPublicURL + c.TelegramWebAppPath
}

func (c *Config) WebhookURL() string {
	return trimRightSlash(c.TelegramPublicURL) + "/telegram/webhook"
}

func telegramMode() (string, error) {
	mode := envOr("TELEGRAM_MODE", ModeWebhook)
	switch mode {
	case ModeWebhook, ModePolling:
		return mode, nil
	default:
		return "", fmt.Errorf("TELEGRAM_MODE must be %q or %q", ModeWebhook, ModePolling)
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func durationEnv(name, fallback string) (time.Duration, error) {
	value := envOr(name, fallback)
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}
	return duration, nil
}

func intEnv(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	return parsed, nil
}

func trimRightSlash(value string) string {
	for len(value) > 0 && value[len(value)-1] == '/' {
		value = value[:len(value)-1]
	}
	return value
}
