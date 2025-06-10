package service

import (
	"cinematigue/internal/domain"
	"errors"
	"fmt"
)

// StoreActor определяет интерфейс для работы с хранилищем актёров
type StoreActor interface {
	Create(actor domain.Actor) (int, error)                    // создать актёра
	GetByID(id int) (domain.Actor, error)                      // получить актёра по ID
	Update(actor domain.Actor) error                           // обновить актёра
	Delete(id int) error                                       // удалить актёра
	GetAll() ([]domain.Actor, error)                           // получить всех актёров
	GetMovies(actorID int) ([]domain.Movie, error)             // фильмы по актёру
	PartialUpdateActor(id int, update domain.ActorUpdate) error // частичное обновление
	GetAllActorsWithMovies() ([]domain.Actor, error)           // актёры с фильмами
}

// ActorService реализует бизнес-логику для актёров
type ActorService struct {
	store StoreActor
}

// NewActor создаёт сервис актёров
func NewActor(store StoreActor) *ActorService {
	return &ActorService{store: store}
}

// Create создаёт нового актёра
func (s *ActorService) Create(actor domain.Actor) (int, error) {
	return s.store.Create(actor)
}

// GetByID возвращает актёра по ID
func (s *ActorService) GetByID(id int) (domain.Actor, error) {
	actor, err := s.store.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrActorNotFound) {
			return domain.Actor{}, domain.ErrActorNotFound
		}
		return domain.Actor{}, fmt.Errorf("getting actor: %w", err)
	}
	return actor, nil
}

// Update обновляет данные актёра
func (s *ActorService) Update(actor domain.Actor) error {
	if err := s.store.Update(actor); err != nil {
		if errors.Is(err, domain.ErrActorNotFound) {
			return domain.ErrActorNotFound
		}
		return fmt.Errorf("updating actor: %w", err)
	}
	return nil
}

// Delete удаляет актёра
func (s *ActorService) Delete(id int) error {
	if err := s.store.Delete(id); err != nil {
		if errors.Is(err, domain.ErrActorNotFound) {
			return domain.ErrActorNotFound
		}
		return fmt.Errorf("deleting actor: %w", err)
	}
	return nil
}

// GetAll возвращает всех актёров
func (s *ActorService) GetAll() ([]domain.Actor, error) {
	actors, err := s.store.GetAll()
	if err != nil {
		return nil, fmt.Errorf("getting all actors: %w", err)
	}
	return actors, nil
}

// GetMovies возвращает фильмы актёра
func (s *ActorService) GetMovies(actorID int) ([]domain.Movie, error) {
	movies, err := s.store.GetMovies(actorID)
	if err != nil {
		if errors.Is(err, domain.ErrActorNotFound) {
			return nil, domain.ErrActorNotFound
		}
		return nil, fmt.Errorf("getting actor movies: %w", err)
	}
	return movies, nil
}

// PartialUpdateActor обновляет только переданные поля актёра
func (s *ActorService) PartialUpdateActor(id int, update domain.ActorUpdate) error {
	if err := s.store.PartialUpdateActor(id, update); err != nil {
		if errors.Is(err, domain.ErrActorNotFound) {
			return domain.ErrActorNotFound
		}
		return fmt.Errorf("partially updating actor: %w", err)
	}
	return nil
}

// GetAllActorsWithMovies возвращает актёров с фильмами
func (s *ActorService) GetAllActorsWithMovies() ([]domain.Actor, error) {
	actors, err := s.store.GetAllActorsWithMovies()
	if err != nil {
		return nil, fmt.Errorf("getting all actors with movies: %w", err)
	}
	return actors, nil
}
