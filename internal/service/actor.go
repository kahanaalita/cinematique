package service

import (
	"cinematique/internal/domain"
	"errors"
	"fmt"
	"log"
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
	log.Printf("Starting deletion of actor with ID: %d", id)

	// Проверяем существование актёра
	_, err := s.store.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrActorNotFound) {
			log.Printf("Cannot delete: actor with ID %d not found", id)
			return domain.ErrActorNotFound
		}
		// Если произошла другая ошибка, логируем и возвращаем её
		log.Printf("Error getting actor (ID: %d): %v", id, err)
		return fmt.Errorf("getting actor: %w", err)
	}

	// Проверяем, есть ли у актёра связанные фильмы
	movies, err := s.store.GetMovies(id)
	if err != nil {
		log.Printf("Error getting movies for actor (ID: %d): %v", id, err)
		return fmt.Errorf("getting actor movies: %w", err)
	}
	
	log.Printf("Found %d related movies for actor (ID: %d)", len(movies), id)
	if len(movies) > 0 {
		errMsg := fmt.Sprintf("cannot delete actor: has %d related movies. Remove movies first", len(movies))
		log.Printf("Cannot delete actor (ID: %d): %s", id, errMsg)
		return errors.New(errMsg)
	}

	// Удаляем актёра
	log.Printf("Deleting actor with ID: %d from store", id)
	if err := s.store.Delete(id); err != nil {
		log.Printf("Error deleting actor (ID: %d): %v", id, err)
		if errors.Is(err, domain.ErrActorNotFound) {
			// На случай, если актёр был удалён между проверкой и удалением
			log.Printf("Actor (ID: %d) was deleted by another process", id)
			return domain.ErrActorNotFound
		}
		return fmt.Errorf("deleting actor: %w", err)
	}
	
	log.Printf("Successfully deleted actor with ID: %d", id)
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
