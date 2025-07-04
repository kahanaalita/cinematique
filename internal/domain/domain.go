package domain

import (
	"errors"
	"time"
)

// Actor — доменная модель для таблицы актёров
// Отражает структуру таблицы actors в БД
type Actor struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Gender    string    `json:"gender"`
	BirthDate time.Time `json:"birth_date"`
	Movies    []Movie   `json:"movies,omitempty"`
}

// Movie — доменная модель для таблицы фильмов
// Отражает структуру таблицы movies в БД
type Movie struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ReleaseYear int       `json:"release_year"`
	Rating      float64   `json:"rating"`
	Actors      []Actor   `json:"actors,omitempty"`
}

// ActorUpdate — доменная модель для обновления актёра
type ActorUpdate struct {
	Name      *string    `json:"name,omitempty"`
	Gender    *string    `json:"gender,omitempty"`
	BirthDate *string    `json:"birth_date,omitempty"`
}

// MovieUpdate — доменная модель для обновления фильма
type MovieUpdate struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	ReleaseYear *int     `json:"release_year,omitempty"`
	Rating      *float64 `json:"rating,omitempty"`
}

// ActorWithFilms — актёр с фильмами (для сервисов и DTO)
type ActorWithFilms struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Gender    string    `json:"gender"`
	BirthDate time.Time `json:"birth_date"`
	Movies    []Movie   `json:"movies,omitempty"`
}

// --- USER & AUTH ---

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"` // "user" или "admin"
}

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// Ошибки доменного слоя
var (
	ErrActorNotFound      = errors.New("actor not found")
	ErrMovieNotFound      = errors.New("movie not found")
	ErrEmptyPassword      = errors.New("database password not set")
	ErrEnvNotLoaded       = errors.New("environment variables could not be loaded")
	ErrActorHasMovies     = errors.New("cannot delete actor: has related movies")
)
