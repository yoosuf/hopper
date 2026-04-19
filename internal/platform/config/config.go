package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Environment string

	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	CORS        CORSConfig
	RateLimit   RateLimitConfig
	Worker      WorkerConfig
	Logging     LoggingConfig
	Metrics     MetricsConfig
	SMTP        SMTPConfig
	Payment     PaymentConfig
	Idempotency IdempotencyConfig
	Redis       RedisConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port int
	Host string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool
	RequestsPerMinute int
	Burst             int
	ByIP              bool
	ByUser            bool
}

// WorkerConfig holds worker configuration
type WorkerConfig struct {
	Enabled                         bool
	PollInterval                    time.Duration
	BatchSize                       int
	MaxRetries                      int
	RetryBaseDelay                  time.Duration
	Concurrency                     int
	JobTimeout                      time.Duration
	OutboxWorkerEnabled             bool
	ScheduledActivatorWorkerEnabled bool
	NotificationWorkerEnabled       bool
	PaymentReconcilerWorkerEnabled  bool
	MaintenanceWorkerEnabled        bool
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool
	Port    int
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host          string
	Port          int
	User          string
	Password      string
	From          string
	SkipTLSVerify bool
}

// PaymentConfig holds payment provider configuration
type PaymentConfig struct {
	Provider            string
	ProviderSecret      string
	StripeAPIKey        string
	StripeWebhookSecret string
}

// IdempotencyConfig holds idempotency configuration
type IdempotencyConfig struct {
	Enabled bool
	TTL     time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Address  string
	Password string
	DB       int
	Enabled  bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		// Only warn if .env file exists but can't be read
		if !os.IsNotExist(err) {
			fmt.Printf("Warning: error loading .env file: %v\n", err)
		}
	}

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "hopper"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "hopper_db"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", ""),
			AccessTokenTTL:  getEnvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: getEnvDuration("JWT_REFRESH_TOKEN_TTL", 168*time.Hour),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			AllowedMethods: getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization", "Idempotency-Key"}),
			MaxAge:         getEnvInt("CORS_MAX_AGE", 3600),
		},
		RateLimit: RateLimitConfig{
			Enabled:           getEnvBool("RATE_LIMIT_ENABLED", true),
			RequestsPerMinute: getEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
			Burst:             getEnvInt("RATE_LIMIT_BURST", 10),
			ByIP:              getEnvBool("RATE_LIMIT_BY_IP", true),
			ByUser:            getEnvBool("RATE_LIMIT_BY_USER", true),
		},
		Worker: WorkerConfig{
			Enabled:                         getEnvBool("WORKER_ENABLED", true),
			PollInterval:                    getEnvDuration("WORKER_POLL_INTERVAL", 5*time.Second),
			BatchSize:                       getEnvInt("WORKER_BATCH_SIZE", 100),
			MaxRetries:                      getEnvInt("WORKER_MAX_RETRIES", 5),
			RetryBaseDelay:                  getEnvDuration("WORKER_RETRY_BASE_DELAY", 1*time.Second),
			Concurrency:                     getEnvInt("WORKER_CONCURRENCY", 10),
			JobTimeout:                      getEnvDuration("WORKER_JOB_TIMEOUT", 30*time.Second),
			OutboxWorkerEnabled:             getEnvBool("OUTBOX_WORKER_ENABLED", true),
			ScheduledActivatorWorkerEnabled: getEnvBool("SCHEDULED_ACTIVATOR_WORKER_ENABLED", true),
			NotificationWorkerEnabled:       getEnvBool("NOTIFICATION_WORKER_ENABLED", true),
			PaymentReconcilerWorkerEnabled:  getEnvBool("PAYMENT_RECONCILER_WORKER_ENABLED", true),
			MaintenanceWorkerEnabled:        getEnvBool("MAINTENANCE_WORKER_ENABLED", true),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvBool("METRICS_ENABLED", true),
			Port:    getEnvInt("METRICS_PORT", 9090),
		},
		SMTP: SMTPConfig{
			Host:          getEnv("SMTP_HOST", ""),
			Port:          getEnvInt("SMTP_PORT", 587),
			User:          getEnv("SMTP_USER", ""),
			Password:      getEnv("SMTP_PASSWORD", ""),
			From:          getEnv("SMTP_FROM", ""),
			SkipTLSVerify: getEnvBool("SMTP_SKIP_TLS_VERIFY", false),
		},
		Payment: PaymentConfig{
			Provider:            getEnv("PAYMENT_PROVIDER", "mock"),
			ProviderSecret:      getEnv("PAYMENT_PROVIDER_SECRET", ""),
			StripeAPIKey:        getEnv("STRIPE_API_KEY", ""),
			StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		},
		Idempotency: IdempotencyConfig{
			Enabled: getEnvBool("IDEMPOTENCY_ENABLED", true),
			TTL:     getEnvDuration("IDEMPOTENCY_TTL", 24*time.Hour),
		},
		Redis: RedisConfig{
			Address:  getEnv("REDIS_ADDRESS", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			Enabled:  getEnvBool("REDIS_ENABLED", false),
		},
	}

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	// Validate SMTP configuration if SMTP is configured
	if c.SMTP.Host != "" && c.SMTP.Password == "" {
		return fmt.Errorf("SMTP_PASSWORD is required when SMTP_HOST is set")
	}

	// Validate payment provider configuration
	if c.Payment.Provider != "mock" {
		if c.Payment.ProviderSecret == "" {
			return fmt.Errorf("PAYMENT_PROVIDER_SECRET is required when PAYMENT_PROVIDER is not 'mock'")
		}
		if c.Payment.Provider == "stripe" && c.Payment.StripeAPIKey == "" {
			return fmt.Errorf("STRIPE_API_KEY is required when PAYMENT_PROVIDER is 'stripe'")
		}
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

// getEnvInt retrieves an environment variable as an integer or returns a default value
func getEnvInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// getEnvBool retrieves an environment variable as a boolean or returns a default value
func getEnvBool(key string, defaultVal bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultVal
}

// getEnvDuration retrieves an environment variable as a duration or returns a default value
func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultVal
}

// getEnvSlice retrieves an environment variable as a string slice or returns a default value
func getEnvSlice(key string, defaultVal []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultVal
}
