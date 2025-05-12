package controller

import "rest api 2/service"

// Интерфейс сервисного слоя для Movie

type ServiceMovie interface {
	Create(movie service.Movie) (int, error)
	GetByID(id int) (service.Movie, error)
	Update(movie service.Movie) error
	Delete(id int) error
	GetAll() ([]service.Movie, error)
}
