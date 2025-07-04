package dto

import "time"

type CreateActorRequest struct {
	Name      string `json:"name"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birth_date"`
}

type UpdateActorRequest struct {
	Name      *string `json:"name,omitempty"`
	Gender    *string `json:"gender,omitempty"`
	BirthDate *string `json:"birth_date,omitempty"`
}

type ActorResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birth_date"`
}

type ActorsListResponse struct {
	Actors []ActorResponse `json:"actors"`
}

type CreateMovieRequest struct {
	Title       string  `json:"title" validate:"required,min=1,max=150"`
	Description string  `json:"description" validate:"max=1000"`
	ReleaseYear int     `json:"release_year" validate:"required"`
	Rating      float64 `json:"rating" validate:"min=0,max=10"`
	ActorIDs    []int   `json:"actor_ids"`
}

type UpdateMovieRequest struct {
	Title       *string  `json:"title,omitempty" validate:"omitempty,min=1,max=150"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	ReleaseYear *int     `json:"release_year,omitempty"`
	Rating      *float64 `json:"rating,omitempty" validate:"omitempty,min=0,max=10"`
	ActorIDs    *[]int   `json:"actor_ids,omitempty"`
}

type MovieResponse struct {
	ID          int            `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	ReleaseYear int            `json:"release_year"`
	Rating      float64        `json:"rating"`
	Actors      []ActorPreview `json:"actors,omitempty"`
}

type ActorPreview struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MoviesListResponse struct {
	Movies []MovieResponse `json:"movies"`
}

// DTO для поиска и фильтрации фильмов

type SearchMoviesRequest struct {
	Query string `json:"query" form:"query"`
}

type MoviesListSortedRequest struct {
	SortField string `json:"sort_field" form:"sort_field"`
	SortOrder string `json:"sort_order" form:"sort_order"`
}

type ActorWithFilms struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	Gender    string          `json:"gender"`
	BirthDate string          `json:"birth_date"`
	Movies    []MovieResponse `json:"movies"`
}

type ActorsWithFilmsListResponse struct {
	Actors []ActorWithFilms `json:"actors"`
}

// MovieWithActorsRequest - запрос на создание фильма с актёрами
type MovieWithActorsRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	ReleaseYear int     `json:"release_year" binding:"required"`
	Rating      float64 `json:"rating" binding:"required"`
	ActorIDs    []int   `json:"actor_ids" binding:"required,min=1"`
}

// UpdateMovieActorsRequest - запрос на обновление списка актёров фильма
type UpdateMovieActorsRequest struct {
	ActorIDs []int `json:"actor_ids" binding:"required,min=1"`
}

// MovieActorsResponse - ответ со списком актёров фильма
type MovieActorsResponse struct {
	Actors []ActorResponse `json:"actors"`
}

// ActorMoviesResponse - ответ со списком фильмов актёра
type ActorMoviesResponse struct {
	Movies []MovieResponse `json:"movies"`
}

// ActorUpdate используется для частичного обновления актёра
type ActorUpdate struct {
	Name      *string    `json:"name,omitempty"`
	Gender    *string    `json:"gender,omitempty"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
}

// MovieUpdate используется для частичного обновления фильма
type MovieUpdate struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	ReleaseYear *int     `json:"release_year,omitempty"`
	Rating      *float64 `json:"rating,omitempty"`
}

// --- AUTH DTOs ---

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Role     string `json:"role,omitempty"` // опционально, по умолчанию user
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"` // in seconds
}
