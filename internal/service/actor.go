package service

import (
	"cinematigue/internal/domain"
)

// Интерфейс для репозитория Actor

type StoreActor interface {
	Create(actor domain.Actor) (int, error)
	GetByID(id int) (domain.Actor, error)
	Update(actor domain.Actor) error
	Delete(id int) error
	GetAll() ([]domain.Actor, error)
	GetMovies(actorID int) ([]domain.Movie, error)
}

// Сервис Actor
type ActorService struct {
	store StoreActor
}

func NewActor(store StoreActor) *ActorService {
	return &ActorService{store: store}
}

func (s *ActorService) Create(actor domain.Actor) (int, error)        { return s.store.Create(actor) }
func (s *ActorService) GetByID(id int) (domain.Actor, error)          { return s.store.GetByID(id) }
func (s *ActorService) Update(actor domain.Actor) error               { return s.store.Update(actor) }
func (s *ActorService) Delete(id int) error                    { return s.store.Delete(id) }
func (s *ActorService) GetAll() ([]domain.Actor, error)               { return s.store.GetAll() }
func (s *ActorService) GetMovies(actorID int) ([]domain.Movie, error) { return s.store.GetMovies(actorID) }
