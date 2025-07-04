package service

import (
	"cinematique/internal/domain"
	"errors"
	"fmt"
	"log"
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
	store      StoreMovie
	actorStore StoreActor
}

// NewMovie создаёт сервис фильмов
func NewMovie(store StoreMovie, actorStore StoreActor) *MovieService {
	return &MovieService{store: store, actorStore: actorStore}
}

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
	log.Printf("Starting deletion of movie with ID: %d", id)

	// Проверяем существование фильма
	_, err := s.store.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			log.Printf("Cannot delete: movie with ID %d not found", id)
			return domain.ErrMovieNotFound
		}
		log.Printf("Error getting movie (ID: %d): %v", id, err)
		return fmt.Errorf("checking movie existence: %w", err)
	}

	// Удаляем все связи с актёрами
	log.Printf("Removing all actors from movie (ID: %d)", id)
	if err := s.store.RemoveAllActors(id); err != nil {
		log.Printf("Error removing actors from movie (ID: %d): %v", id, err)
		return fmt.Errorf("removing actors from movie: %w", err)
	}

	// Удаляем фильм
	log.Printf("Deleting movie with ID: %d", id)
	if err := s.store.Delete(id); err != nil {
		log.Printf("Error deleting movie (ID: %d): %v", id, err)
		if errors.Is(err, domain.ErrMovieNotFound) {
			// Это не должно происходить, так как мы уже проверили существование
			log.Printf("Movie (ID: %d) was deleted by another process", id)
			return domain.ErrMovieNotFound
		}
		return fmt.Errorf("deleting movie: %w", err)
	}

	log.Printf("Successfully deleted movie with ID: %d", id)
	return nil
}

// GetAll возвращает все фильмы
func (s *MovieService) GetAll() ([]domain.Movie, error) { return s.store.GetAll() }

// AddActor добавляет актёра к фильму
func (s *MovieService) AddActor(movieID, actorID int) error {
	log.Printf("Adding actor (ID: %d) to movie (ID: %d)", actorID, movieID)

	// Проверяем существование фильма
	_, err := s.store.GetByID(movieID)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			log.Printf("Cannot add actor: movie with ID %d not found", movieID)
			return domain.ErrMovieNotFound
		}
		log.Printf("Error getting movie (ID: %d): %v", movieID, err)
		return fmt.Errorf("checking movie existence: %w", err)
	}

	// TODO: Проверка существования актёра, когда будет доступен сервис актёров

	// Проверяем, не добавлен ли уже актёр
	actors, err := s.store.GetActorsForMovieByID(movieID)
	if err != nil {
		log.Printf("Error getting actors for movie (ID: %d): %v", movieID, err)
		return fmt.Errorf("getting movie actors: %w", err)
	}

	for _, actor := range actors {
		if actor.ID == actorID {
			errMsg := fmt.Sprintf("actor with ID %d is already in the movie", actorID)
			log.Printf("Cannot add actor: %s", errMsg)
			return errors.New(errMsg)
		}
	}

	// Добавляем актёра к фильму
	if err := s.store.AddActor(movieID, actorID); err != nil {
		log.Printf("Error adding actor (ID: %d) to movie (ID: %d): %v", actorID, movieID, err)
		if errors.Is(err, domain.ErrActorNotFound) {
			return fmt.Errorf("actor with ID %d not found: %w", actorID, domain.ErrActorNotFound)
		}
		return fmt.Errorf("adding actor to movie: %w", err)
	}

	log.Printf("Successfully added actor (ID: %d) to movie (ID: %d)", actorID, movieID)
	return nil
}

// RemoveActor удаляет актёра из фильма
func (s *MovieService) RemoveActor(movieID, actorID int) error {
	log.Printf("Removing actor (ID: %d) from movie (ID: %d)", actorID, movieID)

	// Проверяем существование фильма
	_, err := s.store.GetByID(movieID)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			log.Printf("Cannot remove actor: movie with ID %d not found", movieID)
			return domain.ErrMovieNotFound
		}
		log.Printf("Error getting movie (ID: %d): %v", movieID, err)
		return fmt.Errorf("checking movie existence: %w", err)
	}

	// Проверяем, есть ли актёр в фильме
	actors, err := s.store.GetActorsForMovieByID(movieID)
	if err != nil {
		log.Printf("Error getting actors for movie (ID: %d): %v", movieID, err)
		return fmt.Errorf("getting movie actors: %w", err)
	}

	actorFound := false
	for _, actor := range actors {
		if actor.ID == actorID {
			actorFound = true
			break
		}
	}

	if !actorFound {
		errMsg := fmt.Sprintf("actor with ID %d is not in the movie (ID: %d)", actorID, movieID)
		log.Printf("Cannot remove actor: %s", errMsg)
		return errors.New(errMsg)
	}

	// Удаляем актёра из фильма
	if err := s.store.RemoveActor(movieID, actorID); err != nil {
		log.Printf("Error removing actor (ID: %d) from movie (ID: %d): %v", actorID, movieID, err)
		if errors.Is(err, domain.ErrActorNotFound) {
			return fmt.Errorf("actor with ID %d not found: %w", actorID, domain.ErrActorNotFound)
		}
		return fmt.Errorf("removing actor from movie: %w", err)
	}

	log.Printf("Successfully removed actor (ID: %d) from movie (ID: %d)", actorID, movieID)
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
	log.Printf("Updating actors for movie (ID: %d)", movieID)

	// Проверяем существование фильма
	_, err := s.store.GetByID(movieID)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			log.Printf("Cannot update actors: movie with ID %d not found", movieID)
			return domain.ErrMovieNotFound
		}
		log.Printf("Error getting movie (ID: %d): %v", movieID, err)
		return fmt.Errorf("checking movie existence: %w", err)
	}

	// Удаляем всех текущих актёров фильма
	log.Printf("Removing all actors from movie (ID: %d)", movieID)
	if err := s.store.RemoveAllActors(movieID); err != nil {
		log.Printf("Error removing actors from movie (ID: %d): %v", movieID, err)
		return fmt.Errorf("removing actors from movie: %w", err)
	}

	// Добавляем новых актёров
	for _, actorID := range actorIDs {
		log.Printf("Adding actor (ID: %d) to movie (ID: %d)", actorID, movieID)
		if err := s.store.AddActor(movieID, actorID); err != nil {
			log.Printf("Error adding actor (ID: %d) to movie (ID: %d): %v", actorID, movieID, err)
			if errors.Is(err, domain.ErrActorNotFound) {
				return fmt.Errorf("actor with ID %d not found: %w", actorID, domain.ErrActorNotFound)
			}
			return fmt.Errorf("adding actor to movie: %w", err)
		}
	}

	log.Printf("Successfully updated actors for movie (ID: %d)", movieID)
	return nil
}

func (s *MovieService) GetMoviesForActor(actorID int) ([]domain.Movie, error) {
	// Проверяем существование актёра
	_, err := s.actorStore.GetByID(actorID)
	if err != nil {
		if errors.Is(err, domain.ErrActorNotFound) {
			return nil, domain.ErrActorNotFound
		}
		return nil, fmt.Errorf("getting actor: %w", err)
	}

	// Получаем фильмы актёра
	movies, err := s.store.GetMoviesForActor(actorID)
	if err != nil {
		return nil, fmt.Errorf("getting movies for actor: %w", err)
	}

	// Если фильмов нет, возвращаем пустой срез
	if len(movies) == 0 {
		return []domain.Movie{}, nil
	}

	return movies, nil
}

// PartialUpdateMovie частично обновляет фильм
func (s *MovieService) PartialUpdateMovie(id int, update domain.MovieUpdate) error {
	log.Printf("Starting partial update of movie (ID: %d)", id)

	// Проверяем существование фильма
	movie, err := s.store.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			log.Printf("Cannot update: movie with ID %d not found", id)
			return domain.ErrMovieNotFound
		}
		log.Printf("Error getting movie (ID: %d): %v", id, err)
		return fmt.Errorf("getting movie: %w", err)
	}

	// Проверяем, что есть хотя бы одно поле для обновления
	if update.Title == nil && update.Description == nil && update.ReleaseYear == nil && update.Rating == nil {
		errMsg := "no fields to update"
		log.Printf("Cannot update movie (ID: %d): %s", id, errMsg)
		return errors.New(errMsg)
	}

	// Логируем обновляемые поля
	updatedFields := []string{}

	// Обновляем только переданные поля
	if update.Title != nil {
		updatedFields = append(updatedFields, fmt.Sprintf("Title: %s -> %s", movie.Title, *update.Title))
		movie.Title = *update.Title
	}
	if update.Description != nil {
		descPreview := ""
		if movie.Description != "" {
			descPreview = movie.Description
			if len(descPreview) > 20 {
				descPreview = descPreview[:20] + "..."
			}
		}
		updatedFields = append(updatedFields, "Description")
		movie.Description = *update.Description
	}
	if update.ReleaseYear != nil {
		updatedFields = append(updatedFields, fmt.Sprintf("ReleaseYear: %d -> %d", movie.ReleaseYear, *update.ReleaseYear))
		movie.ReleaseYear = *update.ReleaseYear
	}
	if update.Rating != nil {
		updatedFields = append(updatedFields, fmt.Sprintf("Rating: %.1f -> %.1f", movie.Rating, *update.Rating))
		movie.Rating = *update.Rating
	}

	log.Printf("Updating movie (ID: %d) fields: %v", id, updatedFields)

	// Обновляем фильм
	if err := s.store.Update(movie); err != nil {
		log.Printf("Error updating movie (ID: %d): %v", id, err)
		if errors.Is(err, domain.ErrMovieNotFound) {
			// Это не должно происходить, так как мы уже проверили существование
			log.Printf("Movie (ID: %d) was deleted by another process", id)
			return domain.ErrMovieNotFound
		}
		return fmt.Errorf("updating movie: %w", err)
	}

	log.Printf("Successfully updated movie (ID: %d)", id)
	return nil
}
