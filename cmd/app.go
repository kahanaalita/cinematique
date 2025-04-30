package cmd

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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

	// Инициализация handler  и HTTP-сервера
	type Handler struct {
		MovieRepo repository.Movie
		ActorRepo repository.Actor
	}

	func (h Handler) GetMovie(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		movie, err := h.MovieRepo.GetByID(id)
		if err != nil {
			http.Error(w, "movie not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(movie)
	}

	handler := Handler{MovieRepo: movieRepo, ActorRepo: actorRepo}
	http.HandleFunc("/movies/get", handler.GetMovie)

	log.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	return nil
}
