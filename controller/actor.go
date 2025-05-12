package controller

import (
	"rest api 2/service"
)

// Интерфейс сервисного слоя для Actor

type ServiceActor interface {
	Create(actor service.Actor) (int, error)
	GetByID(id int) (service.Actor, error)
	Update(actor service.Actor) error
	Delete(id int) error
	GetAll() ([]service.Actor, error)
	GetMovies(actorID int) ([]service.Movie, error)
}
