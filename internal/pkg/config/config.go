package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Telegram TelegramConfig
	App      AppConfig
	Storage  StorageConfig
	SMTP     SMTP
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	URL string
}

type AuthConfig struct {
	YandexClientID     string
	YandexClientSecret string
	SessionSecret      string
	SessionMaxAge      time.Duration
}

type TelegramConfig struct {
	BotToken string
	ChatID   string
}

type AppConfig struct {
	Environment string
	BaseURL     string
}

type StorageConfig struct {
	Driver           string // "local" or "s3"
	LocalUploadDir   string // e.g., "assets/uploads"
	PublicUploadBase string // e.g., "/assets/uploads"
	S3Bucket         string
	S3BaseKey        string
	S3Region         string
	S3Endpoint       string
	S3UsePathStyle   bool
	S3PublicBaseURL  string
}

// SMTP contains settings for sending emails via SMTP.
type SMTP struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
}

func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "3333"),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "file:./data.db"),
		},
		Auth: AuthConfig{
			YandexClientID:     getEnv("YANDEX_CLIENT_ID", ""),
			YandexClientSecret: getEnv("YANDEX_CLIENT_SECRET", ""),
			SessionSecret:      getEnv("SESSION_SECRET", "your-secret-key"),
			SessionMaxAge:      getEnvAsDuration("SESSION_MAX_AGE", 24*time.Hour),
		},
		Telegram: TelegramConfig{
			BotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
			ChatID:   getEnv("TELEGRAM_CHAT_ID", ""),
		},
		App: AppConfig{
			Environment: getEnv("ENVIRONMENT", "development"),
			BaseURL:     getEnv("BASE_URL", "http://localhost:3333"),
		},
		Storage: StorageConfig{
			Driver:           getEnv("STORAGE_DRIVER", "local"),
			LocalUploadDir:   getEnv("LOCAL_UPLOAD_DIR", "assets/uploads"),
			PublicUploadBase: getEnv("PUBLIC_UPLOAD_BASE", "/assets/uploads"),
			S3Bucket:         getEnv("S3_BUCKET", ""),
			S3BaseKey:        getEnv("S3_BASE_KEY", ""),
			S3Region:         getEnv("S3_REGION", "us-east-1"),
			S3Endpoint:       getEnv("S3_ENDPOINT", ""),
			S3UsePathStyle:   getEnvAsBool("S3_USE_PATH_STYLE", false),
			S3PublicBaseURL:  getEnv("S3_PUBLIC_BASE_URL", ""),
		},
		SMTP: SMTP{
			Host:        getEnv("SMTP_HOST", ""),
			Port:        getEnvAsInt("SMTP_PORT", 587),
			Username:    getEnv("SMTP_USERNAME", ""),
			Password:    getEnv("SMTP_PASSWORD", ""),
			FromAddress: getEnv("SMTP_FROM", "noreply@example.com"),
		},
	}

	// Validate required fields
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func (c *Config) validate() error {
	if c.Auth.YandexClientID == "" {
		return fmt.Errorf("YANDEX_CLIENT_ID is required")
	}
	if c.Auth.YandexClientSecret == "" {
		return fmt.Errorf("YANDEX_CLIENT_SECRET is required")
	}
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.Storage.Driver == "s3" && c.Storage.S3Bucket == "" {
		return fmt.Errorf("S3_BUCKET is required when STORAGE_DRIVER=s3")
	}
	return nil
}

func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "1", "true", "yes", "y":
			return true
		case "0", "false", "no", "n":
			return false
		}
	}
	return defaultValue
}
