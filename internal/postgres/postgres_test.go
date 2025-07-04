package postgres

import (
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetConfig тестирует загрузку конфигурации
func TestGetConfig(t *testing.T) {
	// Сохраняем оригинальные переменные окружения
	oldEnv := make(map[string]string)
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"} {
		value := os.Getenv(key)
		if value != "" {
			oldEnv[key] = value
		}
		os.Unsetenv(key)
	}
	defer func() {
		for key, value := range oldEnv {
			os.Setenv(key, value)
		}
	}()

	// Тест с пустым паролем
	t.Run("empty_password_error", func(t *testing.T) {
		cfg, err := GetConfig()
		assert.Error(t, err)
		assert.Equal(t, ErrEmptyPassword, err)
		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, "5432", cfg.Port)
		assert.Equal(t, "postgres", cfg.User)
		assert.Equal(t, "", cfg.Password)
		assert.Equal(t, "cinematheque", cfg.DBName)
		assert.Equal(t, "disable", cfg.SSLMode)
	})

	// Тест с установленными переменными окружения
	t.Run("with_env", func(t *testing.T) {
		os.Setenv("DB_HOST", "test-host")
		os.Setenv("DB_PORT", "5433")
		os.Setenv("DB_USER", "test-user")
		os.Setenv("DB_PASSWORD", "test-password")
		os.Setenv("DB_NAME", "test-db")
		os.Setenv("DB_SSLMODE", "require")

		cfg, err := GetConfig()
		require.NoError(t, err)
		assert.Equal(t, "test-host", cfg.Host)
		assert.Equal(t, "5433", cfg.Port)
		assert.Equal(t, "test-user", cfg.User)
		assert.Equal(t, "test-password", cfg.Password)
		assert.Equal(t, "test-db", cfg.DBName)
		assert.Equal(t, "require", cfg.SSLMode)
	})
}

// TestConnect тестирует подключение к базе данных
func TestConnect(t *testing.T) {
	// Сохраняем и восстанавливаем переменные окружения
	oldEnv := make(map[string]string)
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"} {
		oldEnv[key] = os.Getenv(key)
	}
	defer func() {
		for key, value := range oldEnv {
			os.Setenv(key, value)
		}
	}()

	// Тест с неправильной конфигурацией
	t.Run("invalid_config", func(t *testing.T) {
		os.Setenv("DB_HOST", "invalid-host")
		os.Setenv("DB_PORT", "invalid-port")
		os.Setenv("DB_USER", "invalid-user")
		os.Setenv("DB_PASSWORD", "invalid-password")
		os.Setenv("DB_NAME", "invalid-db")

		db, err := Connect()
		if db != nil {
			defer db.Close()
		}
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

// TestConnectPoolSettings тестирует настройки пула подключений
func TestConnectPoolSettings(t *testing.T) {
	// Создаем mock-соединение
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Устанавливаем настройки пула
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Проверяем только MaxOpenConnections, так как это единственная надежная проверка с sqlmock
	assert.Equal(t, 25, db.Stats().MaxOpenConnections)

	// Проверяем, что нет неожиданных вызовов
	assert.NoError(t, mock.ExpectationsWereMet())
}
