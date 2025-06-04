package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cinematigue/internal/controller"
	"cinematigue/internal/handlers"
	"cinematigue/internal/postgres"
	"cinematigue/internal/repository"
	"cinematigue/internal/service"

	"github.com/gin-gonic/gin"
)

// Run инициализирует и запускает приложение с поддержкой graceful shutdown
func Run() error {
	// Подключение к базе данных
	db, err := postgres.Connect()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return err
	}

	// Инициализация репозиториев
	movieRepo := repository.NewMovie(db)
	actorRepo := repository.NewActor(db)
	userRepo := repository.NewUserRepository(db)

	// Инициализация сервисов
	movieService := service.NewMovie(movieRepo)
	actorService := service.NewActor(actorRepo)
	authService := service.NewAuthService(userRepo)

	// Инициализация контроллеров
	actorController := controller.NewActorController(actorService)
	movieController := controller.NewMovieController(movieService)

	// Инициализация хендлеров
	actorHandler := handlers.NewActorHandler(actorController)
	movieHandler := handlers.NewMovieHandler(movieController)
	authHandler := handlers.NewAuthHandler(authService)

	// Настройка роутера
	router := gin.Default()

	// --- Регистрация всех маршрутов с авторизацией и ролями ---
	handlers.RegisterAllRoutes(router, actorHandler, movieHandler, authHandler)

	// Создаем HTTP-сервер с настройками
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Канал для graceful shutdown
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

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Пытаемся корректно завершить работу сервера
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	// Закрываем соединение с БД
	if err := db.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}

	log.Println("Server exiting")
	return nil
}
