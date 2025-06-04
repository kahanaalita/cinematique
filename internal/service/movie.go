package service

import (
	"cinematigue/internal/domain"
)

// StoreMovie определяет интерфейс для работы с хранилищем фильмов
type StoreMovie interface {
	Create(movie domain.Movie) (int, error)                                   // создать фильм
	GetByID(id int) (domain.Movie, error)                                     // получить фильм по ID
	Update(movie domain.Movie) error                                          // обновить фильм
	Delete(id int) error                                                      // удалить фильм
	GetAll() ([]domain.Movie, error)                                          // получить все фильмы
	AddActor(movieID, actorID int) error                                      // добавить актёра к фильму
	RemoveActor(movieID, actorID int) error                                   // удалить актёра из фильма
	GetActorsForMovieByID(movieID int) ([]domain.Actor, error)                // получить актёров фильма
	RemoveAllActors(movieID int) error                                        // удалить всех актёров из фильма
	SearchMoviesByTitle(titleFragment string) ([]domain.Movie, error)         // поиск по названию
	SearchMoviesByActorName(actorNameFragment string) ([]domain.Movie, error) // поиск по актёру
	GetAllMoviesSorted(sortField, sortOrder string) ([]domain.Movie, error)   // сортировка
	CreateMovieWithActors(movie domain.Movie, actorIDs []int) (int, error)    // создать фильм с актёрами
	UpdateMovieActors(movieID int, actorIDs []int) error                      // обновить актёров фильма
	GetMoviesForActor(actorID int) ([]domain.Movie, error)                    // фильмы по актёру
	PartialUpdateMovie(id int, update domain.MovieUpdate) error               // частичное обновление фильма
}

// MovieService реализует бизнес-логику для фильмов
type MovieService struct {
	store StoreMovie
}

// NewMovie создаёт сервис фильмов
func NewMovie(store StoreMovie) *MovieService { return &MovieService{store: store} }

// Create создаёт фильм с актёрами
func (s *MovieService) Create(movie domain.Movie, actorIDs []int) (int, error) {
	id, err := s.store.Create(movie)
	if err != nil {
		return 0, err
	}
	for _, actorID := range actorIDs {
		if err := s.store.AddActor(id, actorID); err != nil {
			_ = s.store.Delete(id)
			return 0, err
		}
	}
	return id, nil
}

// GetByID возвращает фильм с актёрами
func (s *MovieService) GetByID(id int) (domain.Movie, error) {
	movie, err := s.store.GetByID(id)
	if err != nil {
		return domain.Movie{}, err
	}
	actors, err := s.store.GetActorsForMovieByID(id)
	if err != nil {
		return domain.Movie{}, err
	}
	movie.Actors = make([]domain.Actor, len(actors))
	copy(movie.Actors, actors)
	return movie, nil
}

// Update обновляет фильм и связи с актёрами
func (s *MovieService) Update(movie domain.Movie, actorIDs []int) error {
	if err := s.store.Update(movie); err != nil {
		return err
	}
	if err := s.store.RemoveAllActors(movie.ID); err != nil {
		return err
	}
	for _, actorID := range actorIDs {
		if err := s.store.AddActor(movie.ID, actorID); err != nil {
			return err
		}
	}
	return nil
}

// MovieService methods

// Delete удаляет фильм
func (s *MovieService) Delete(id int) error { return s.store.Delete(id) }

// GetAll возвращает все фильмы
func (s *MovieService) GetAll() ([]domain.Movie, error) { return s.store.GetAll() }

// AddActor добавляет актёра к фильму
func (s *MovieService) AddActor(movieID, actorID int) error {
	return s.store.AddActor(movieID, actorID)
}

// RemoveActor удаляет актёра из фильма
func (s *MovieService) RemoveActor(movieID, actorID int) error {
	return s.store.RemoveActor(movieID, actorID)
}

// GetActors возвращает актёров фильма
func (s *MovieService) GetActors(movieID int) ([]domain.Actor, error) {
	return s.store.GetActorsForMovieByID(movieID)
}

// GetActorsForMovieByID возвращает актёров фильма (алиас для GetActors)
func (s *MovieService) GetActorsForMovieByID(movieID int) ([]domain.Actor, error) {
	return s.store.GetActorsForMovieByID(movieID)
}

// RemoveAllActors удаляет всех актёров из фильма
func (s *MovieService) RemoveAllActors(movieID int) error {
    return s.store.RemoveAllActors(movieID)
}

// SearchMoviesByTitle ищет фильмы по названию
func (s *MovieService) SearchMoviesByTitle(titleFragment string) ([]domain.Movie, error) {
	return s.store.SearchMoviesByTitle(titleFragment)
}

// SearchMoviesByActorName ищет фильмы по имени актёра
func (s *MovieService) SearchMoviesByActorName(actorNameFragment string) ([]domain.Movie, error) {
	return s.store.SearchMoviesByActorName(actorNameFragment)
}

// GetAllMoviesSorted возвращает фильмы с сортировкой
func (s *MovieService) GetAllMoviesSorted(sortField, sortOrder string) ([]domain.Movie, error) {
	return s.store.GetAllMoviesSorted(sortField, sortOrder)
}

// CreateMovieWithActors создаёт фильм с актёрами
func (s *MovieService) CreateMovieWithActors(movie domain.Movie, actorIDs []int) (int, error) {
	return s.store.CreateMovieWithActors(movie, actorIDs)
}

// UpdateMovieActors обновляет актёров фильма
func (s *MovieService) UpdateMovieActors(movieID int, actorIDs []int) error {
	return s.store.UpdateMovieActors(movieID, actorIDs)
}

// GetMoviesForActor возвращает фильмы по актёру
func (s *MovieService) GetMoviesForActor(actorID int) ([]domain.Movie, error) {
	return s.store.GetMoviesForActor(actorID)
}

// PartialUpdateMovie частично обновляет фильм
func (s *MovieService) PartialUpdateMovie(id int, update domain.MovieUpdate) error {
	return s.store.PartialUpdateMovie(id, update)
}

