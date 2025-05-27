package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"cinematigue/internal/config"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Конфигурация подключения к БД теперь находится в пакете config (internal/config/config.go)

var (
	ErrEmptyPassword = errors.New("database password not set")
	ErrEnvNotLoaded  = errors.New("environment variables could not be loaded")
)

// GetConfig возвращает конфигурацию на основе переменных окружения
func GetConfig() (config.Config, error) {
	// Загрузка .env файла с игнорированием ошибки если файл не найден
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: error loading .env file: %v", err)
	}

	cfg := config.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "cinematheque"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	if cfg.Password == "" {
		return cfg, ErrEmptyPassword
	}

	return cfg, nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Connect устанавливает соединение с базой данных PostgreSQL
func Connect() (*sql.DB, error) {
	cfg, err := GetConfig()
	if err != nil {
		return nil, fmt.Errorf("database configuration failed: %w", err)
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Настройки пула подключений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Проверка подключения
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}
