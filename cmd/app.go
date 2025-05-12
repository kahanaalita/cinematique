package cmd

import (
	"SQL/controller"
	"log"
	"rest api 2/internal/postgres"
	"rest api 2/internal/repository"
	"rest api 2/service"
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
	cinematique := controller.NewCinematique(movieService, actorService)

}
