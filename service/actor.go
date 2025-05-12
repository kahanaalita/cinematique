package service

import "time"

// Интерфейс для репозитория Actor

type StoreActor interface {
	Create(actor Actor) (int, error)
	GetByID(id int) (Actor, error)
	Update(actor Actor) error
	Delete(id int) error
	GetAll() ([]Actor, error)
	GetMovies(actorID int) ([]Movie, error)
}

// Структура Actor
type Actor struct {
	ID        int
	Name      string
	Gender    string
	BirthDate time.Time
}

// Сервис Actor
type ActorService struct {
	store StoreActor
}

func NewActor(store StoreActor) *ActorService {
	return &ActorService{store: store}
}

func (s *ActorService) Create(actor Actor) (int, error)        { return s.store.Create(actor) }
func (s *ActorService) GetByID(id int) (Actor, error)          { return s.store.GetByID(id) }
func (s *ActorService) Update(actor Actor) error               { return s.store.Update(actor) }
func (s *ActorService) Delete(id int) error                    { return s.store.Delete(id) }
func (s *ActorService) GetAll() ([]Actor, error)               { return s.store.GetAll() }
func (s *ActorService) GetMovies(actorID int) ([]Movie, error) { return s.store.GetMovies(actorID) }
