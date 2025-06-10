package service

import (
	"cinematigue/internal/domain"
	"errors"
	"fmt"
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
		if errors.Is(err, domain.ErrMovieNotFound) {
			return domain.Movie{}, domain.ErrMovieNotFound
		}
		return domain.Movie{}, fmt.Errorf("getting movie by ID: %w", err)
	}

	actors, err := s.store.GetActorsForMovieByID(id)
	if err != nil {
		// Don't fail if we can't get actors, just log and continue with empty actors list
		// This is a design decision - we might want to handle this differently
		// For now, we'll just log the error and continue with empty actors
		// return domain.Movie{}, fmt.Errorf("getting actors for movie: %w", err)
	}

	movie.Actors = make([]domain.Actor, len(actors))
	copy(movie.Actors, actors)
	return movie, nil
}

// Update обновляет фильм и связи с актёрами
func (s *MovieService) Update(movie domain.Movie, actorIDs []int) error {
	// Проверяем существование фильма
	_, err := s.store.GetByID(movie.ID)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return domain.ErrMovieNotFound
		}
		return fmt.Errorf("checking movie existence: %w", err)
	}

	if err := s.store.Update(movie); err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return domain.ErrMovieNotFound
		}
		return fmt.Errorf("updating movie: %w", err)
	}

	if err := s.store.RemoveAllActors(movie.ID); err != nil {
		return fmt.Errorf("removing actors from movie: %w", err)
	}

	for _, actorID := range actorIDs {
		if err := s.store.AddActor(movie.ID, actorID); err != nil {
			if errors.Is(err, domain.ErrActorNotFound) {
				return domain.ErrActorNotFound
			}
			return fmt.Errorf("adding actor to movie: %w", err)
		}
	}

	return nil
}

// MovieService methods

// Delete удаляет фильм
func (s *MovieService) Delete(id int) error {
	// Проверяем существование фильма
	_, err := s.store.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return domain.ErrMovieNotFound
		}
		return fmt.Errorf("checking movie existence: %w", err)
	}

	// Удаляем все связи с актёрами
	if err := s.store.RemoveAllActors(id); err != nil {
		return fmt.Errorf("removing actors from movie: %w", err)
	}

	// Удаляем фильм
	if err := s.store.Delete(id); err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return domain.ErrMovieNotFound
		}
		return fmt.Errorf("deleting movie: %w", err)
	}

	return nil
}

// GetAll возвращает все фильмы
func (s *MovieService) GetAll() ([]domain.Movie, error) { return s.store.GetAll() }

// AddActor добавляет актёра к фильму
func (s *MovieService) AddActor(movieID, actorID int) error {
	// Проверяем существование фильма
	_, err := s.store.GetByID(movieID)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return domain.ErrMovieNotFound
		}
		return fmt.Errorf("checking movie existence: %w", err)
	}

	// TODO: Проверка существования актёра, когда будет доступен сервис актёров

	// Добавляем актёра к фильму
	if err := s.store.AddActor(movieID, actorID); err != nil {
		if errors.Is(err, domain.ErrActorNotFound) || errors.Is(err, domain.ErrMovieNotFound) {
			return err
		}
		return fmt.Errorf("adding actor to movie: %w", err)
	}

	return nil
}

// RemoveActor удаляет актёра из фильма
func (s *MovieService) RemoveActor(movieID, actorID int) error {
	// Проверяем существование фильма
	_, err := s.store.GetByID(movieID)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return domain.ErrMovieNotFound
		}
		return fmt.Errorf("checking movie existence: %w", err)
	}

	// Удаляем актёра из фильма
	if err := s.store.RemoveActor(movieID, actorID); err != nil {
		if errors.Is(err, domain.ErrActorNotFound) || errors.Is(err, domain.ErrMovieNotFound) {
			return err
		}
		return fmt.Errorf("removing actor from movie: %w", err)
	}

	return nil
}

// GetActors возвращает актёров фильма
func (s *MovieService) GetActors(movieID int) ([]domain.Actor, error) {
	// Проверяем существование фильма
	_, err := s.store.GetByID(movieID)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return nil, domain.ErrMovieNotFound
		}
		return nil, fmt.Errorf("checking movie existence: %w", err)
	}

	actors, err := s.store.GetActorsForMovieByID(movieID)
	if err != nil {
		// Возвращаем пустой список, если не удалось получить актёров
		return []domain.Actor{}, nil
	}

	return actors, nil
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
	// TODO: Проверка существования актёра, когда будет доступен сервис актёров

	movies, err := s.store.GetMoviesForActor(actorID)
	if err != nil {
		if errors.Is(err, domain.ErrActorNotFound) {
			return nil, domain.ErrActorNotFound
		}
		return nil, fmt.Errorf("getting movies for actor: %w", err)
	}

	return movies, nil
}

// PartialUpdateMovie частично обновляет фильм
func (s *MovieService) PartialUpdateMovie(id int, update domain.MovieUpdate) error {
	// Проверяем существование фильма
	_, err := s.store.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return domain.ErrMovieNotFound
		}
		return fmt.Errorf("checking movie existence: %w", err)
	}

	// Выполняем частичное обновление
	if err := s.store.PartialUpdateMovie(id, update); err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return domain.ErrMovieNotFound
		}
		return fmt.Errorf("partially updating movie: %w", err)
	}

	return nil
}
