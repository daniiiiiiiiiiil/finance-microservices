package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	TimeZone    *time.Location
	JWTSecret   string
	JWTDuration time.Duration
	TLSEnabled  bool
	TLSCertFile string
	TLSKeyFile  string
	TLSCAFile   string

	KafkaBrokers       []string
	KafkaTopic         string
	KafkaConsumerGroup string
	KafkaMaxRetries    int
	KafkaRetryBackoff  time.Duration

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresTimeout  time.Duration

	RedisAddr string
}

func NewConfig() (*Config, error) {
	tz := os.Getenv("TIME_ZONE")
	if tz == "" {
		tz = "UTC"
	}
	zone, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("failed to load time zone %q: %v", tz, err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	jwtDuration := os.Getenv("JWT_DURATION")
	if jwtDuration == "" {
		jwtDuration = "24h"
	}
	duration, err := time.ParseDuration(jwtDuration)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_DURATION: %w", err)
	}

	tlsEnabled, _ := strconv.ParseBool(os.Getenv("TLS_ENABLED"))
	tlsCertFile := os.Getenv("TLS_CERT_FILE")
	tlsKeyFile := os.Getenv("TLS_KEY_FILE")
	tlsCAFile := os.Getenv("TLS_CA_FILE")

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "finance-events"
	}
	kafkaConsumerGroup := os.Getenv("KAFKA_CONSUMER_GROUP")
	if kafkaConsumerGroup == "" {
		kafkaConsumerGroup = "users-service"
	}

	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDB := os.Getenv("POSTGRES_DB")
	postgresTimeout := 5 * time.Second

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	return &Config{
		TimeZone:    zone,
		JWTSecret:   jwtSecret,
		JWTDuration: duration,
		TLSEnabled:  tlsEnabled,
		TLSCertFile: tlsCertFile,
		TLSKeyFile:  tlsKeyFile,
		TLSCAFile:   tlsCAFile,

		KafkaBrokers:       []string{kafkaBrokers},
		KafkaTopic:         kafkaTopic,
		KafkaConsumerGroup: kafkaConsumerGroup,
		KafkaMaxRetries:    3,
		KafkaRetryBackoff:  time.Second,

		PostgresHost:     postgresHost,
		PostgresPort:     postgresPort,
		PostgresUser:     postgresUser,
		PostgresPassword: postgresPassword,
		PostgresDB:       postgresDB,
		PostgresTimeout:  postgresTimeout,

		RedisAddr: redisAddr,
	}, nil
}

func NewConfigMust() *Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to create config: %v", err))
	}
	return config
}
