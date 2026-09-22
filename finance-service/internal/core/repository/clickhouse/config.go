package clickhouse

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string        `envconfig:"HOST" required:"true"`
	Port     string        `envconfig:"PORT" default:"9000"`
	Database string        `envconfig:"DB" required:"true"`
	User     string        `envconfig:"USER" required:"true"`
	Password string        `envconfig:"PASSWORD" required:"true"`
	Timeout  time.Duration `envconfig:"TIMEOUT" default:"5s"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("CLICKHOUSE", &config); err != nil {
		return Config{}, fmt.Errorf("error parsing config: %w", err)
	}
	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(err)
	}
	return config
}
