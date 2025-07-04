package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cinematique/internal/auth"
	"cinematique/internal/controller"
	"cinematique/internal/handlers"
	"cinematique/internal/kafka"
	"cinematique/internal/postgres"
	"cinematique/internal/repository"
	"cinematique/internal/service"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	UserRegistrationTopic = "user-registration"
	MovieViewsTopic       = "movie-views"
	MovieSearchesTopic    = "movie-searches"

	UserEventsGroup  = "user-events-group"
	MovieEventsGroup = "movie-events-group"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	// Регистрируем метрики в стандартном реестре Prometheus.
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDurationSeconds)
}

// PrometheusMiddleware собирает метрики HTTP-запросов.
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next() // Обрабатываем запрос

		duration := time.Since(start).Seconds()
		status := http.StatusText(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, status).Inc()
		httpRequestDurationSeconds.WithLabelValues(c.Request.Method, c.Request.URL.Path, status).Observe(duration)
	}
}

// Run инициализирует и запускает приложение с поддержкой корректного завершения (graceful shutdown)
func Run() error {
	// Инициализируем JWT-ключ
	if err := auth.InitJWTKey(); err != nil {
		log.Fatalf("Failed to initialize JWT key: %v", err)
	}

	// Подключаемся к базе данных
	db, err := postgres.Connect()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return err
	}
	defer db.Close()

	// Регистрируем метрики базы данных
	postgres.RegisterDBMetrics(db)

	// Инициализируем Kafka-продюсер и пул
	kafkaBrokerAddress := os.Getenv("KAFKA_BROKER_ADDRESS")
	if kafkaBrokerAddress == "" {
		kafkaBrokerAddress = "kafka:9092" // Адрес по умолчанию для Kafka в docker-compose
	}
	producerCfg := kafka.NewProducerConfig(kafkaBrokerAddress)
	eventProducer := kafka.NewProducer(producerCfg)
	eventProducerPool := kafka.NewProducerPool(eventProducer, 2, 256) // 2 воркера, буфер на 256 сообщений
	defer eventProducerPool.Close()                                  // Корректно закрываем пул при завершении приложения

	// Инициализация Kafka-консьюмеров
	userRegConsumer := kafka.NewConsumer(kafka.NewConsumerConfig(kafkaBrokerAddress, UserEventsGroup, UserRegistrationTopic))
	movieViewsConsumer := kafka.NewConsumer(kafka.NewConsumerConfig(kafkaBrokerAddress, MovieEventsGroup, MovieViewsTopic))
	movieSearchesConsumer := kafka.NewConsumer(kafka.NewConsumerConfig(kafkaBrokerAddress, MovieEventsGroup, MovieSearchesTopic))

	consumers := []*kafka.Consumer{userRegConsumer, movieViewsConsumer, movieSearchesConsumer}

	// Запускаем консьюмеры в отдельных горутинах
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for _, c := range consumers {
		wg.Add(1)
		go func(consumer *kafka.Consumer) {
			defer wg.Done()
			consumer.ConsumeMessages(consumerCtx)
		}(c)
	}

	// Инициализация репозиториев
	movieRepo := repository.NewMovie(db)
	actorRepo := repository.NewActor(db)
	userRepo := repository.NewUserRepository(db)

	// Инициализация сервисов
	movieService := service.NewMovie(movieRepo, actorRepo)
	actorService := service.NewActor(actorRepo)
	authService := service.NewAuthService(userRepo)

	// Инициализация контроллеров
	actorController := controller.NewActorController(actorService)
	movieController := controller.NewMovieController(movieService)

	// Инициализация хендлеров, передавая Kafka продюсер
	actorHandler := handlers.NewActorHandler(actorController)
	movieHandler := handlers.NewMovieHandler(movieController, eventProducerPool)
	authHandler := handlers.NewAuthHandler(authService, eventProducerPool)

	// Настраиваем логирование
	log.SetOutput(os.Stdout)
	log.Println("Logging to stdout is configured")

	// Настраиваем роутер
	router := gin.Default()

	// Добавляем middleware для Prometheus
	router.Use(PrometheusMiddleware())

	// Добавляем endpoint для метрик Prometheus
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Создаём основную группу API с префиксом /api
	api := router.Group("/api")

	// Регистрируем все маршруты (публичные и защищённые)
	handlers.RegisterAllRoutes(api, actorHandler, movieHandler, authHandler)

	// Создаём HTTP-сервер с настройками
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Канал для корректного завершения приложения
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Println("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Ожидаем сигнал завершения
	<-done
	log.Println("Shutting down server...")

	// Создаём контекст с таймаутом для корректного завершения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Пытаемся корректно завершить работу сервера
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	// Останавливаем Kafka-консьюмеры
	log.Println("Stopping Kafka consumers...")
	consumerCancel()
	wg.Wait()
	log.Println("All Kafka consumers have been stopped.")

	for _, c := range consumers {
		if err := c.Close(); err != nil {
			log.Printf("Error closing Kafka consumer: %v", err)
		}
	}

	log.Println("Server exiting")
	return nil
}
