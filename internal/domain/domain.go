package domain

import (
	"errors"
	"time"
)

// Actor — доменная модель для таблицы актёров
// Отражает структуру таблицы actors в БД

type Actor struct {
	ID        int
	Name      string
	Gender    string
	BirthDate time.Time
}

// Movie — доменная модель для таблицы фильмов
// Отражает структуру таблицы movies в БД

type Movie struct {
	ID          int
	Title       string
	Description string
	ReleaseYear int
	Rating      float64
}

// Ошибки доменного слоя
var (
	ErrActorNotFound = errors.New("actor not found")
	ErrMovieNotFound = errors.New("movie not found")
	ErrEmptyPassword = errors.New("database password not set")
	ErrEnvNotLoaded  = errors.New("environment variables could not be loaded")
)
