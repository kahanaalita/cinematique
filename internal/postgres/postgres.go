package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"cinematique/internal/config"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ErrEmptyPassword = errors.New("database password not set")
	ErrEnvNotLoaded  = errors.New("environment variables could not be loaded")

	dbConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Current number of active database connections.",
		},
	)
)

// GetConfig возвращает конфигурацию на основе переменных окружения
func GetConfig() (config.Config, error) {
	// Загрузка .env файла с игнорированием ошибки если файл не найден
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: error loading .env file: %v", err)
	}

	// Логируем загруженные переменные окружения для отладки
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "cinematheque")
	sslMode := getEnv("DB_SSLMODE", "disable")

	log.Printf("DB Config - Host: %s, Port: %s, User: %s, DBName: %s, SSLMode: %s", 
		host, port, user, dbName, sslMode)

	cfg := config.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
		SSLMode:  sslMode,
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

	// Логируем параметры подключения (без пароля)
	log.Printf("Connecting to database: host=%s, port=%s, user=%s, dbname=%s, sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.DBName, cfg.SSLMode)

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

// DBStatsCollector реализует интерфейс prometheus.Collector.
type DBStatsCollector struct {
	db *sql.DB
}

// NewDBStatsCollector создает новый экземпляр DBStatsCollector.
func NewDBStatsCollector(db *sql.DB) *DBStatsCollector {
	return &DBStatsCollector{db: db}
}

// Describe реализует метод интерфейса prometheus.Collector.
func (collector *DBStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	dbConnectionsActive.Describe(ch)
}

// Collect реализует метод интерфейса prometheus.Collector.
func (collector *DBStatsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := collector.db.Stats()
	dbConnectionsActive.Set(float64(stats.OpenConnections))
	ch <- dbConnectionsActive
}

// RegisterDBMetrics регистрирует метрики подключений к базе данных.
func RegisterDBMetrics(db *sql.DB) {
	prometheus.MustRegister(NewDBStatsCollector(db))
	log.Println("Database metrics registered.")
}