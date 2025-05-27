package service

import (
	"cinematigue/internal/domain"
)

// Интерфейс для репозитория Movie
type StoreMovie interface {
	Create(movie domain.Movie) (int, error)
	GetByID(id int) (domain.Movie, error)
	Update(movie domain.Movie) error
	Delete(id int) error
	GetAll() ([]domain.Movie, error)
}

// Сервис Movie
type MovieService struct {
	store StoreMovie
}

func NewMovie(store StoreMovie) *MovieService {
	return &MovieService{store: store}
}

func (s *MovieService) Create(movie domain.Movie) (int, error) { return s.store.Create(movie) }
func (s *MovieService) GetByID(id int) (domain.Movie, error)   { return s.store.GetByID(id) }
func (s *MovieService) Update(movie domain.Movie) error        { return s.store.Update(movie) }
func (s *MovieService) Delete(id int) error             { return s.store.Delete(id) }
func (s *MovieService) GetAll() ([]domain.Movie, error)        { return s.store.GetAll() }
