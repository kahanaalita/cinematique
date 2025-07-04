package controller

import "cinematique/internal/domain"

// ServiceActor интерфейс сервисного слоя для Actor
type ServiceActor interface {
	Create(actor domain.Actor) (int, error)
	GetByID(id int) (domain.Actor, error)
	Update(actor domain.Actor) error
	Delete(id int) error
	GetAll() ([]domain.Actor, error)
	GetMovies(actorID int) ([]domain.Movie, error)
	GetAllActorsWithMovies() ([]domain.Actor, error)
}

// ServiceMovie интерфейс сервисного слоя для Movie
type ServiceMovie interface {
	Create(movie domain.Movie, actorIDs []int) (int, error)
	GetByID(id int) (domain.Movie, error)
	Update(movie domain.Movie, actorIDs []int) error
	Delete(id int) error
	GetAll() ([]domain.Movie, error)
	AddActor(movieID, actorID int) error
	RemoveActor(movieID, actorID int) error
	GetActors(movieID int) ([]domain.Actor, error)
	GetActorsForMovieByID(movieID int) ([]domain.Actor, error)
	GetMoviesForActor(actorID int) ([]domain.Movie, error)
	SearchMoviesByTitle(titleFragment string) ([]domain.Movie, error)
	SearchMoviesByActorName(actorNameFragment string) ([]domain.Movie, error)
	GetAllMoviesSorted(sortField, sortOrder string) ([]domain.Movie, error)
	CreateMovieWithActors(movie domain.Movie, actorIDs []int) (int, error)
	UpdateMovieActors(movieID int, actorIDs []int) error
	PartialUpdateMovie(id int, update domain.MovieUpdate) error
}
