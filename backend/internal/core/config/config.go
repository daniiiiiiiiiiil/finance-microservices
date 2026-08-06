package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	TimeZone    *time.Location
	JWTSecret   string
	JWTDuration time.Duration
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

	return &Config{
		TimeZone:    zone,
		JWTSecret:   jwtSecret,
		JWTDuration: duration,
	}, nil
}

func NewConfigMust() *Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to create config: %v", err))
	}
	return config
}
