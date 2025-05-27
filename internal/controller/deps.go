package controller

import "cinematigue/internal/domain"

// ServiceActor interface moved from actor.go to deps.go
// Интерфейс сервисного слоя для Actor

type ServiceActor interface {
	Create(actor domain.Actor) (int, error)
	GetByID(id int) (domain.Actor, error)
	Update(actor domain.Actor) error
	Delete(id int) error
	GetAll() ([]domain.Actor, error)
	GetMovies(actorID int) ([]domain.Movie, error)
}

// ServiceMovie interface moved from movie.go to deps.go
// Интерфейс сервисного слоя для Movie

type ServiceMovie interface {
	Create(movie domain.Movie) (int, error)
	GetByID(id int) (domain.Movie, error)
	Update(movie domain.Movie) error
	Delete(id int) error
	GetAll() ([]domain.Movie, error)
}
