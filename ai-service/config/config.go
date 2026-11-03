package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TimeZone *time.Location

	JWTSecret   string
	JWTDuration time.Duration

	GroqAPIKey     string
	GroqBaseURL    string
	GroqModel      string
	GroqTimeout    time.Duration
	GroqMaxRetries int

	TavilyAPIKey  string
	TavilyBaseURL string
	TavilyTimeout time.Duration

	NominatimURL       string
	NominatimUserAgent string
	NominatimTimeout   time.Duration

	OSRMURL     string
	OSRMTimeout time.Duration

	FuelConsumptionLPer100KM float64
	FuelPricePerLiter        float64
	TaxiPricePerKM           float64
	TravelCostMode           string

	GRPCPort    string
	MetricsPort string

	TLSEnabled  bool
	TLSCertFile string
	TLSKeyFile  string
	TLSCAFile   string

	RedisAddr string
	CacheTTL  time.Duration

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDatabase string
	PostgresTimeout  time.Duration
}

func NewConfig() (*Config, error) {
	tz := getEnv("TIME_ZONE", "UTC")
	zone, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("failed to load time zone %q: %w", tz, err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	jwtDuration, err := getDuration("JWT_DURATION", 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_DURATION: %w", err)
	}

	groqTimeout, err := getDuration("GROQ_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid GROQ_TIMEOUT: %w", err)
	}

	tavilyTimeout, err := getDuration("TAVILY_TIMEOUT", 20*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid TAVILY_TIMEOUT: %w", err)
	}

	nominatimTimeout, err := getDuration("NOMINATIM_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid NOMINATIM_TIMEOUT: %w", err)
	}

	osrmTimeout, err := getDuration("OSRM_TIMEOUT", 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid OSRM_TIMEOUT: %w", err)
	}

	postgresTimeout, err := getDuration("POSTGRES_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid POSTGRES_TIMEOUT: %w", err)
	}

	cacheTTL, err := getDuration("CACHE_TTL", 10*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid CACHE_TTL: %w", err)
	}

	fuelConsumption, err := getFloat("FUEL_CONSUMPTION_L_PER_100KM", 8.0)
	if err != nil {
		return nil, fmt.Errorf("invalid FUEL_CONSUMPTION_L_PER_100KM: %w", err)
	}

	fuelPrice, err := getFloat("FUEL_PRICE_PER_LITER", 55.0)
	if err != nil {
		return nil, fmt.Errorf("invalid FUEL_PRICE_PER_LITER: %w", err)
	}

	taxiPrice, err := getFloat("TAXI_PRICE_PER_KM", 30.0)
	if err != nil {
		return nil, fmt.Errorf("invalid TAXI_PRICE_PER_KM: %w", err)
	}

	travelMode := getEnv("TRAVEL_COST_MODE", "fuel")
	if travelMode != "fuel" && travelMode != "taxi" {
		return nil, fmt.Errorf("invalid TRAVEL_COST_MODE %q: must be 'fuel' or 'taxi'", travelMode)
	}

	groqRetries, err := strconv.Atoi(getEnv("GROQ_MAX_RETRIES", "2"))
	if err != nil {
		return nil, fmt.Errorf("invalid GROQ_MAX_RETRIES: %w", err)
	}

	tlsEnabled, _ := strconv.ParseBool(getEnv("TLS_ENABLED", "false"))

	return &Config{
		TimeZone:    zone,
		JWTSecret:   jwtSecret,
		JWTDuration: jwtDuration,

		GroqAPIKey:     os.Getenv("GROQ_API_KEY"),
		GroqBaseURL:    getEnv("GROQ_BASE_URL", "https://api.groq.com/openai/v1"),
		GroqModel:      getEnv("GROQ_MODEL", "llama-3.3-70b-versatile"),
		GroqTimeout:    groqTimeout,
		GroqMaxRetries: groqRetries,

		TavilyAPIKey:  os.Getenv("TAVILY_API_KEY"),
		TavilyBaseURL: getEnv("TAVILY_BASE_URL", "https://api.tavily.com"),
		TavilyTimeout: tavilyTimeout,

		NominatimURL:       getEnv("NOMINATIM_URL", "https://nominatim.openstreetmap.org"),
		NominatimUserAgent: getEnv("NOMINATIM_USER_AGENT", "finance-ai-service/1.0"),
		NominatimTimeout:   nominatimTimeout,

		OSRMURL:     getEnv("OSRM_URL", "http://router.project-osrm.org"),
		OSRMTimeout: osrmTimeout,

		FuelConsumptionLPer100KM: fuelConsumption,
		FuelPricePerLiter:        fuelPrice,
		TaxiPricePerKM:           taxiPrice,
		TravelCostMode:           travelMode,

		GRPCPort:    getEnv("AI_GRPC_PORT", "50056"),
		MetricsPort: getEnv("AI_METRICS_PORT", "9099"),

		TLSEnabled:  tlsEnabled,
		TLSCertFile: os.Getenv("TLS_CERT_FILE"),
		TLSKeyFile:  os.Getenv("TLS_KEY_FILE"),
		TLSCAFile:   os.Getenv("TLS_CA_FILE"),

		RedisAddr: os.Getenv("REDIS_ADDR"),
		CacheTTL:  cacheTTL,

		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresDatabase: os.Getenv("POSTGRES_DB"),
		PostgresTimeout:  postgresTimeout,
	}, nil
}

func NewConfigMust() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to create config: %v", err))
	}
	return cfg
}

func (c *Config) Validate() error {
	if c.GroqAPIKey == "" {
		return fmt.Errorf("GROQ_API_KEY is required")
	}
	if c.TavilyAPIKey == "" {
		return fmt.Errorf("TAVILY_API_KEY is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	return time.ParseDuration(v)
}

func getFloat(key string, fallback float64) (float64, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	return strconv.ParseFloat(v, 64)
}

type LoggerConfig struct {
	Level  string `envconfig:"LEVEL" required:"true"`
	Folder string `envconfig:"FOLDER" required:"true"`
}

func NewLoggerConfig() (LoggerConfig, error) {
	var cfg LoggerConfig
	if err := envconfig.Process("LOGGER", &cfg); err != nil {
		return LoggerConfig{}, fmt.Errorf("process envconfig: %w", err)
	}
	return cfg, nil
}
