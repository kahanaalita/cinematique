package service

import (
	"cinematigue/internal/domain"
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

// Методы реализующие интерфейс StoreActor
func (s *ActorService) Create(actor domain.Actor) (int, error)        { return s.store.Create(actor) }
func (s *ActorService) GetByID(id int) (domain.Actor, error)          { return s.store.GetByID(id) }
func (s *ActorService) Update(actor domain.Actor) error               { return s.store.Update(actor) }
func (s *ActorService) Delete(id int) error                          { return s.store.Delete(id) }
func (s *ActorService) GetAll() ([]domain.Actor, error)               { return s.store.GetAll() }
func (s *ActorService) GetMovies(actorID int) ([]domain.Movie, error) { return s.store.GetMovies(actorID) }

// PartialUpdateActor обновляет только переданные поля актёра
func (s *ActorService) PartialUpdateActor(id int, update domain.ActorUpdate) error {
	return s.store.PartialUpdateActor(id, update)
}

// GetAllActorsWithMovies возвращает актёров с фильмами
func (s *ActorService) GetAllActorsWithMovies() ([]domain.Actor, error) {
	return s.store.GetAllActorsWithMovies()
}
