package cmd

import (
	"log"
	"cinematigue/internal/controller"
	"cinematigue/internal/handlers"
	"cinematigue/internal/postgres"
	"cinematigue/internal/repository"
	"cinematigue/internal/service"

	"github.com/gin-gonic/gin"
)

// Run инициализирует и запускает приложение
func Run() error {
	// Подключение к базе данных
	db, err := postgres.Connect()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return err
	}
	defer db.Close()

	// Инициализация репозиториев
	movieRepo := repository.NewMovie(db)
	actorRepo := repository.NewActor(db)

	// Инициализация сервисов
	movieService := service.NewMovie(movieRepo)
	actorService := service.NewActor(actorRepo)

	// Инициализация контроллеров
	actorController := controller.NewActorController(actorService)
	movieController := controller.NewMovieController(movieService)

	// Инициализация хендлеров
	actorHandler := handlers.NewActorHandler(actorController)
	movieHandler := handlers.NewMovieHandler(movieController)

	// Настройка роутера
	router := gin.Default()

	// Регистрация маршрутов
	handlers.RegisterActorRoutes(router, actorHandler)
	handlers.RegisterMovieRoutes(router, movieHandler)

	// Запуск сервера
	log.Println("Starting server on :8080")
	return router.Run(":8080")
}
