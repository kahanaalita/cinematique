package cmd

import (
	"log"

	"SQL/internal/postgres"
	"SQL/internal/repository"
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

	_ = movieRepo
	_ = actorRepo
	// Далее можно продолжать выполнение приложения инициализация роутера,сервера и Graceful shutdown

	return nil
}
