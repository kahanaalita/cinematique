package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"

	"cinematique/internal/controller/dto"
	"cinematique/internal/domain"
)


// movieController обрабатывает запросы, связанные с фильмами
type movieController struct {
	movieService ServiceMovie
}

// NewMovieController создаёт контроллер фильмов
func NewMovieController(movieService ServiceMovie) *movieController {
	return &movieController{
		movieService: movieService,
	}
}

// validateMovie проверяет валидность данных фильма
func validateMovie(title, description string, rating float64) error {
	title = strings.TrimSpace(title)
	if len(title) < 1 || len(title) > 150 {
		return fmt.Errorf("title: must be 1-150 characters")
	}

	if len(description) > 1000 {
		return fmt.Errorf("description: too long (max 1000 characters)")
	}

	if rating < 0 || rating > 10 {
		return fmt.Errorf("rating: must be between 0 and 10")
	}

	return nil
}

// CreateMovie создаёт фильм
func (c *movieController) CreateMovie(ctx *gin.Context, req dto.CreateMovieRequest) (dto.MovieResponse, error) {
	// Валидация входных данных
	if err := validateMovie(req.Title, req.Description, req.Rating); err != nil {
		return dto.MovieResponse{}, fmt.Errorf("validation error: %w", err)
	}

	movie := domain.Movie{
		Title:       req.Title,
		Description: req.Description,
		ReleaseYear: req.ReleaseYear,
		Rating:      req.Rating,
	}

	// Создаем фильм и добавляем связи с актерами
	id, err := c.movieService.Create(movie, req.ActorIDs)
	if err != nil {
		return dto.MovieResponse{}, err
	}

	// Получаем созданный фильм с актерами
	createdMovie, err := c.movieService.GetByID(id)
	if err != nil {
		return dto.MovieResponse{}, err
	}

	// Конвертируем в DTO
	return c.toMovieResponse(createdMovie), nil
}

// GetMovieByID возвращает фильм по ID
func (c *movieController) GetMovieByID(ctx *gin.Context, id int) (dto.MovieResponse, error) {
	movie, err := c.movieService.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return dto.MovieResponse{}, domain.ErrMovieNotFound
		}
		return dto.MovieResponse{}, fmt.Errorf("getting movie: %w", err)
	}

	return c.toMovieResponse(movie), nil
}

// UpdateMovie обновляет фильм
func (c *movieController) UpdateMovie(ctx *gin.Context, id int, req dto.UpdateMovieRequest) (dto.MovieResponse, error) {
	movie, err := c.movieService.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return dto.MovieResponse{}, domain.ErrMovieNotFound
		}
		return dto.MovieResponse{}, fmt.Errorf("getting movie: %w", err)
	}

	// Валидация обновляемых полей
	title := movie.Title
	description := movie.Description
	rating := movie.Rating

	if req.Title != nil {
		title = *req.Title
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.Rating != nil {
		rating = *req.Rating
	}

	if err := validateMovie(title, description, rating); err != nil {
		return dto.MovieResponse{}, fmt.Errorf("validation error: %w", err)
	}

	// Обновляем только переданные поля
	if req.Title != nil {
		movie.Title = *req.Title
	}
	if req.Description != nil {
		movie.Description = *req.Description
	}
	if req.ReleaseYear != nil {
		movie.ReleaseYear = *req.ReleaseYear
	}
	if req.Rating != nil {
		movie.Rating = *req.Rating
	}

	// Обновляем фильм и связи с актерами, если они были переданы
	var actorIDs []int
	if req.ActorIDs != nil {
		actorIDs = *req.ActorIDs
	}

	err = c.movieService.Update(movie, actorIDs)
	if err != nil {
		return dto.MovieResponse{}, err
	}

	// Получаем обновленный фильм с актерами
	updatedMovie, err := c.movieService.GetByID(id)
	if err != nil {
		return dto.MovieResponse{}, err
	}

	return c.toMovieResponse(updatedMovie), nil
}

// DeleteMovie удаляет фильм
func (c *movieController) DeleteMovie(ctx *gin.Context, id int) error {
	if err := c.movieService.Delete(id); err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return domain.ErrMovieNotFound
		}
		return fmt.Errorf("deleting movie: %w", err)
	}
	return nil
}

// ListMovies возвращает все фильмы
func (c *movieController) ListMovies(ctx *gin.Context) (dto.MoviesListResponse, error) {
	movies, err := c.movieService.GetAll()
	if err != nil {
		return dto.MoviesListResponse{}, err
	}

	response := dto.MoviesListResponse{
		Movies: make([]dto.MovieResponse, 0, len(movies)),
	}

	for _, movie := range movies {
		response.Movies = append(response.Movies, c.toMovieResponse(movie))
	}

	return response, nil
}

// SearchMoviesByTitle ищет фильмы по названию
func (c *movieController) SearchMoviesByTitle(ctx *gin.Context) (dto.MoviesListResponse, error) {
	query := ctx.Query("title")
	if query == "" {
		return dto.MoviesListResponse{}, errors.New("title parameter is required")
	}
	movies, err := c.movieService.SearchMoviesByTitle(query)
	if err != nil {
		return dto.MoviesListResponse{}, err
	}
	return dto.MoviesListResponse{Movies: c.toMovieResponses(movies)}, nil
}

// SearchMoviesByActorName ищет фильмы по имени актёра
func (c *movieController) SearchMoviesByActorName(ctx *gin.Context) (dto.MoviesListResponse, error) {
	query := ctx.Query("actorName")
	if query == "" {
		return dto.MoviesListResponse{}, errors.New("actorName parameter is required")
	}
	movies, err := c.movieService.SearchMoviesByActorName(query)
	if err != nil {
		return dto.MoviesListResponse{}, err
	}
	return dto.MoviesListResponse{Movies: c.toMovieResponses(movies)}, nil
}

// GetAllMoviesSorted возвращает фильмы с сортировкой
func (c *movieController) GetAllMoviesSorted(ctx *gin.Context) (dto.MoviesListResponse, error) {
	sortField := ctx.DefaultQuery("sort_field", "rating")
	sortOrder := ctx.DefaultQuery("sort_order", "DESC")
	movies, err := c.movieService.GetAllMoviesSorted(sortField, sortOrder)
	if err != nil {
		return dto.MoviesListResponse{}, err
	}
	return dto.MoviesListResponse{Movies: c.toMovieResponses(movies)}, nil
}

// toMovieResponse конвертирует Movie в DTO
func (c *movieController) toMovieResponse(movie domain.Movie) dto.MovieResponse {
	// Конвертируем актеров в формат DTO
	var actorPreviews []dto.ActorPreview
	if len(movie.Actors) > 0 {
		actorPreviews = make([]dto.ActorPreview, 0, len(movie.Actors))
		for _, actor := range movie.Actors {
			actorPreviews = append(actorPreviews, dto.ActorPreview{
				ID:   actor.ID,
				Name: actor.Name,
			})
		}
	} else {
		actorPreviews = nil
	}

	return dto.MovieResponse{
		ID:          movie.ID,
		Title:       movie.Title,
		Description: movie.Description,
		ReleaseYear: movie.ReleaseYear,
		Rating:      movie.Rating,
		Actors:      actorPreviews,
	}
}

// toMovieResponses конвертирует []Movie в []DTO
func (c *movieController) toMovieResponses(movies []domain.Movie) []dto.MovieResponse {
	responses := make([]dto.MovieResponse, 0, len(movies))
	for _, m := range movies {
		responses = append(responses, c.toMovieResponse(m))
	}
	return responses
}

// CreateMovieWithActors создаёт фильм с актёрами
func (c *movieController) CreateMovieWithActors(ctx *gin.Context, req dto.MovieWithActorsRequest) (dto.MovieResponse, error) {
	// Валидация входных данных
	if err := validateMovie(req.Title, req.Description, req.Rating); err != nil {
		return dto.MovieResponse{}, fmt.Errorf("validation error: %w", err)
	}

	movie := domain.Movie{
		Title:       req.Title,
		Description: req.Description,
		ReleaseYear: req.ReleaseYear,
		Rating:      req.Rating,
	}

	// Создаем фильм с актёрами
	id, err := c.movieService.CreateMovieWithActors(movie, req.ActorIDs)
	if err != nil {
		return dto.MovieResponse{}, err
	}

	// Получаем созданный фильм с актёрами
	createdMovie, err := c.movieService.GetByID(id)
	if err != nil {
		return dto.MovieResponse{}, err
	}

	return c.toMovieResponse(createdMovie), nil
}

// UpdateMovieActors обновляет актёров фильма
func (c *movieController) UpdateMovieActors(ctx *gin.Context, movieID int, req dto.UpdateMovieActorsRequest) (dto.MovieActorsResponse, error) {
	// Обновляем связи фильма с актёрами
	err := c.movieService.UpdateMovieActors(movieID, req.ActorIDs)
	if err != nil {
		return dto.MovieActorsResponse{}, err
	}

	// Получаем обновлённый список актёров фильма
	actors, err := c.movieService.GetActors(movieID)
	if err != nil {
		return dto.MovieActorsResponse{}, err
	}

	// Конвертируем актёров в DTO
	actorResponses := make([]dto.ActorResponse, len(actors))
	for i, actor := range actors {
		actorResponses[i] = dto.ActorResponse{
			ID:        actor.ID,
			Name:      actor.Name,
			Gender:    actor.Gender,
			BirthDate: actor.BirthDate.Format("2006-01-02"),
		}
	}

	return dto.MovieActorsResponse{Actors: actorResponses}, nil
}

// AddActorToMovie добавляет актёра в фильм
func (c *movieController) AddActorToMovie(ctx *gin.Context, movieID, actorID int) (dto.MovieResponse, error) {
	// Добавляем актёра в фильм
	err := c.movieService.AddActor(movieID, actorID)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) || errors.Is(err, domain.ErrActorNotFound) {
			return dto.MovieResponse{}, err
		}
		return dto.MovieResponse{}, fmt.Errorf("adding actor to movie: %w", err)
	}

	// Получаем обновлённый фильм
	updatedMovie, err := c.movieService.GetByID(movieID)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return dto.MovieResponse{}, domain.ErrMovieNotFound
		}
		return dto.MovieResponse{}, fmt.Errorf("getting updated movie: %w", err)
	}

	return c.toMovieResponse(updatedMovie), nil
}

// RemoveActorFromMovie удаляет актёра из фильма
func (c *movieController) RemoveActorFromMovie(ctx *gin.Context, movieID, actorID int) (dto.MovieResponse, error) {
	// Удаляем актёра из фильма
	err := c.movieService.RemoveActor(movieID, actorID)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) || errors.Is(err, domain.ErrActorNotFound) {
			return dto.MovieResponse{}, err
		}
		return dto.MovieResponse{}, fmt.Errorf("removing actor from movie: %w", err)
	}

	// Получаем обновлённый фильм
	updatedMovie, err := c.movieService.GetByID(movieID)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return dto.MovieResponse{}, domain.ErrMovieNotFound
		}
		return dto.MovieResponse{}, fmt.Errorf("getting updated movie: %w", err)
	}

	return c.toMovieResponse(updatedMovie), nil
}

// GetActorsForMovieByID возвращает актёров фильма
func (c *movieController) GetActorsForMovieByID(ctx *gin.Context, movieID int) (dto.MovieActorsResponse, error) {
	// Проверяем существование фильма
	_, err := c.movieService.GetByID(movieID)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return dto.MovieActorsResponse{}, domain.ErrMovieNotFound
		}
		return dto.MovieActorsResponse{}, domain.ErrMovieNotFound
	}

	actors, err := c.movieService.GetActorsForMovieByID(movieID)
	if err != nil {
		return dto.MovieActorsResponse{}, fmt.Errorf("getting actors for movie: %w", err)
	}

	// Конвертируем актёров в DTO
	actorResponses := make([]dto.ActorResponse, len(actors))
	for i, actor := range actors {
		actorResponses[i] = dto.ActorResponse{
			ID:        actor.ID,
			Name:      actor.Name,
			Gender:    actor.Gender,
			BirthDate: actor.BirthDate.Format("2006-01-02"),
		}
	}

	return dto.MovieActorsResponse{Actors: actorResponses}, nil
}

// GetMoviesForActor возвращает фильмы по актёру
func (c *movieController) GetMoviesForActor(ctx *gin.Context, actorID int) (dto.ActorMoviesResponse, error) {
	// TODO: Добавить проверку существования актёра, когда будет доступен сервис актёров

	movies, err := c.movieService.GetMoviesForActor(actorID)
	if err != nil {
		if errors.Is(err, domain.ErrActorNotFound) {
			return dto.ActorMoviesResponse{}, domain.ErrActorNotFound
		}
		return dto.ActorMoviesResponse{}, fmt.Errorf("getting movies for actor: %w", err)
	}

	return dto.ActorMoviesResponse{
		Movies: c.toMovieResponses(movies),
	}, nil
}

// PartialUpdateMovie частично обновляет фильм
func (c *movieController) PartialUpdateMovie(ctx *gin.Context, id int, update dto.MovieUpdate) error {
	// Получаем текущий фильм
	movie, err := c.movieService.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return domain.ErrMovieNotFound
		}
		return fmt.Errorf("getting movie: %w", err)
	}

	// Обновляем только переданные поля
	if update.Title != nil {
		movie.Title = *update.Title
	}
	if update.Description != nil {
		movie.Description = *update.Description
	}
	if update.ReleaseYear != nil {
		movie.ReleaseYear = *update.ReleaseYear
	}
	if update.Rating != nil {
		movie.Rating = *update.Rating
	}

	// Валидация обновленных данных
	if err := validateMovie(movie.Title, movie.Description, movie.Rating); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Сохраняем изменения (передаем пустой слайс actorIDs, так как мы не обновляем актеров)
	if err := c.movieService.Update(movie, []int{}); err != nil {
		return fmt.Errorf("updating movie: %w", err)
	}

	return nil
}
