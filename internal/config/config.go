package config

import (
	"cinematique/internal/keycloak"
	"os"
	"strconv"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// KeycloakConfig содержит настройки Keycloak
type KeycloakConfig struct {
	Enabled   bool   `json:"enabled"`
	ServerURL string `json:"server_url"`
	Realm     string `json:"realm"`
	ClientID  string `json:"client_id"`
}

// RedisConfig содержит настройки Redis
type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// RateLimitConfig содержит настройки rate limiting
type RateLimitConfig struct {
	Enabled             bool     `json:"enabled"`
	RequestsPerMinute   int      `json:"requests_per_minute"`
	WindowSeconds       int      `json:"window_seconds"`
	RestrictedEndpoints []string `json:"restricted_endpoints"`
}

// AppConfig содержит всю конфигурацию приложения
type AppConfig struct {
	Database  Config          `json:"database"`
	Keycloak  KeycloakConfig  `json:"keycloak"`
	Redis     RedisConfig     `json:"redis"`
	RateLimit RateLimitConfig `json:"rate_limit"`
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *AppConfig {
	return &AppConfig{
		Database: Config{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "cinematique"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Keycloak: KeycloakConfig{
			Enabled:   getEnvBool("KEYCLOAK_ENABLED", false),
			ServerURL: getEnv("KEYCLOAK_SERVER_URL", ""),
			Realm:     getEnv("KEYCLOAK_REALM", ""),
			ClientID:  getEnv("KEYCLOAK_CLIENT_ID", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		RateLimit: RateLimitConfig{
			Enabled:           getEnvBool("RATE_LIMIT_ENABLED", true),
			RequestsPerMinute: getEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 1000),
			WindowSeconds:     getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60),
			RestrictedEndpoints: []string{
				"/api/movies",
				"/api/actors",
			},
		},
	}
}

// ToKeycloakConfig преобразует в конфигурацию клиента Keycloak
func (kc *KeycloakConfig) ToKeycloakConfig() keycloak.Config {
	return keycloak.Config{
		ServerURL: kc.ServerURL,
		Realm:     kc.Realm,
		ClientID:  kc.ClientID,
	}
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool получает булеву переменную окружения
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvInt получает целочисленную переменную окружения
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
