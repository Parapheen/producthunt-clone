package env

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func IsProduction() bool {
	return GetEnv("APP_ENV", "development") == "production"
}

func Load() {
	_ = godotenv.Load()
}

func GetEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func GetEnvUint32(key string, defaultVal uint32) uint32 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseUint(val, 10, 32); err == nil {
			return uint32(parsed)
		}
	}
	return defaultVal
}

func GetEnvUint8(key string, defaultVal uint8) uint8 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseUint(val, 10, 8); err == nil {
			return uint8(parsed)
		}
	}
	return defaultVal
}

func GetEnvInt64(key string, defaultVal int64) int64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func GetEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			return parsed
		}
	}
	return defaultVal
}
