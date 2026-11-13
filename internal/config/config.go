package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName    string
	AppEnv     string
	AppPort    string
	AppBaseURL string

	DBHost                         string
	DBPort                         string
	DBUser                         string
	DBPassword                     string
	DBName                         string
	DBSSLMode                      string
	DBTimezone                     string
	ActivityCommentWhatsAppEnabled bool
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		AppName:                        getEnv("APP_NAME", "backend-presensi"),
		AppEnv:                         getEnv("APP_ENV", "development"),
		AppPort:                        getEnv("APP_PORT", "8080"),
		AppBaseURL:                     getEnv("APP_BASE_URL", "http://localhost:8080"),
		ActivityCommentWhatsAppEnabled: getEnvBool("ACTIVITY_COMMENT_WHATSAPP_ENABLED", false),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     getEnv("DB_NAME", "presensi"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		DBTimezone: getEnv("DB_TIMEZONE", "Asia/Jakarta"),
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}
