package controller

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"cinematigue/internal/controller/dto"
	"cinematigue/internal/domain"
)

// actorController контроллер актёров.
type actorController struct {
	actorService ServiceActor
}

// PartialUpdateActor частично обновляет данные актёра
func (c *actorController) PartialUpdateActor(ctx *gin.Context, id int, update dto.ActorUpdate) error {
	// Получаем текущие данные актёра
	actor, err := c.actorService.GetByID(id)
	if err != nil {
		return fmt.Errorf("getting actor: %w", err)
	}

	// Обновляем только переданные поля
	if update.Name != nil {
		actor.Name = *update.Name
	}
	if update.Gender != nil {
		actor.Gender = *update.Gender
	}
	if update.BirthDate != nil {
		actor.BirthDate = *update.BirthDate
	}

	// Валидация обновленных данных
	if err := validateActorInput(
		actor.Name,
		actor.Gender,
		actor.BirthDate.Format("2006-01-02"),
	); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Сохраняем изменения
	if err := c.actorService.Update(actor); err != nil {
		return fmt.Errorf("updating actor: %w", err)
	}

	return nil
}

// NewActorController создаёт новый контроллер актёров.
func NewActorController(actorService ServiceActor) *actorController {
	return &actorController{
		actorService: actorService,
	}
}

// validateActorInput проверяет корректность входных данных актёра.
func validateActorInput(name, gender, birthDate string) error {
	name = strings.TrimSpace(name)
	if len(name) == 0 || len(name) > 100 {
		return fmt.Errorf("name: must be 1-100 characters")
	}

	gender = strings.ToLower(strings.TrimSpace(gender))
	if gender != "male" && gender != "female" && gender != "other" {
		return fmt.Errorf("gender: must be 'male', 'female' or 'other'")
	}

	birth, err := time.Parse("2006-01-02", birthDate)
	if err != nil {
		return fmt.Errorf("birth_date: must be in format YYYY-MM-DD")
	}

	if birth.After(time.Now()) {
		return fmt.Errorf("birth_date: cannot be in the future")
	}

	minDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	if birth.Before(minDate) {
		return fmt.Errorf("birth_date: cannot be before 1900-01-01")
	}

	return nil
}

// CreateActor создаёт нового актёра.
func (c *actorController) CreateActor(ctx *gin.Context, req dto.CreateActorRequest) (dto.ActorResponse, error) {
	if err := validateActorInput(req.Name, req.Gender, req.BirthDate); err != nil {
		return dto.ActorResponse{}, err
	}
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return dto.ActorResponse{}, err
	}
	actor := domain.Actor{
		Name:      req.Name,
		Gender:    req.Gender,
		BirthDate: birthDate,
	}
	id, err := c.actorService.Create(actor)
	if err != nil {
		return dto.ActorResponse{}, err
	}
	return dto.ActorResponse{
		ID:        id,
		Name:      actor.Name,
		Gender:    actor.Gender,
		BirthDate: req.BirthDate,
	}, nil
}

// GetActorByID возвращает актёра по ID.
func (c *actorController) GetActorByID(ctx *gin.Context, id int) (dto.ActorResponse, error) {
	actor, err := c.actorService.GetByID(id)
	if err != nil {
		return dto.ActorResponse{}, err
	}
	return dto.ActorResponse{
		ID:        actor.ID,
		Name:      actor.Name,
		Gender:    actor.Gender,
		BirthDate: actor.BirthDate.Format("2006-01-02"),
	}, nil
}

// UpdateActor обновляет данные актёра.
func (c *actorController) UpdateActor(ctx *gin.Context, id int, req dto.UpdateActorRequest) (dto.ActorResponse, error) {
	actor, err := c.actorService.GetByID(id)
	if err != nil {
		return dto.ActorResponse{}, fmt.Errorf("getting actor: %w", err)
	}

	// Подготавливаем обновленные значения
	updatedName := actor.Name
	updatedGender := actor.Gender
	updatedBirthDate := actor.BirthDate

	// Обновляем только переданные поля
	if req.Name != nil {
		updatedName = *req.Name
	}
	if req.Gender != nil {
		updatedGender = *req.Gender
	}
	if req.BirthDate != nil {
		birthDate, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			return dto.ActorResponse{}, fmt.Errorf("invalid birth date format: %w", err)
		}
		updatedBirthDate = birthDate
	}

	// Валидируем все поля разом
	if err := validateActorInput(
		updatedName,
		updatedGender,
		updatedBirthDate.Format("2006-01-02"),
	); err != nil {
		return dto.ActorResponse{}, fmt.Errorf("validation error: %w", err)
	}

	// Применяем обновления
	actor.Name = updatedName
	actor.Gender = updatedGender
	actor.BirthDate = updatedBirthDate
	err = c.actorService.Update(actor)
	if err != nil {
		return dto.ActorResponse{}, err
	}
	return dto.ActorResponse{
		ID:        actor.ID,
		Name:      actor.Name,
		Gender:    actor.Gender,
		BirthDate: actor.BirthDate.Format("2006-01-02"),
	}, nil
}

// DeleteActor удаляет актёра.
func (c *actorController) DeleteActor(ctx *gin.Context, id int) error {
	// Проверяем, есть ли у актёра связанные фильмы
	movies, err := c.actorService.GetMovies(id)
	if err != nil {
		return fmt.Errorf("getting actor movies: %w", err)
	}
	if len(movies) > 0 {
		return fmt.Errorf("cannot delete actor: has %d related movies. Remove movies first", len(movies))
	}
	if err := c.actorService.Delete(id); err != nil {
		return fmt.Errorf("deleting actor: %w", err)
	}
	return nil
}

// ListActors возвращает всех актёров.
func (c *actorController) ListActors(ctx *gin.Context) (dto.ActorsListResponse, error) {
	actors, err := c.actorService.GetAll()
	if err != nil {
		return dto.ActorsListResponse{}, err
	}

	response := dto.ActorsListResponse{
		Actors: make([]dto.ActorResponse, 0, len(actors)),
	}

	for _, actor := range actors {
		response.Actors = append(response.Actors, dto.ActorResponse{
			ID:        actor.ID,
			Name:      actor.Name,
			Gender:    actor.Gender,
			BirthDate: actor.BirthDate.Format("2006-01-02"),
		})
	}

	return response, nil
}

// GetAllActorsWithMovies возвращает актёров с фильмами.
func (c *actorController) GetAllActorsWithMovies(ctx *gin.Context) (dto.ActorsWithFilmsListResponse, error) {
	actors, err := c.actorService.GetAllActorsWithMovies()
	if err != nil {
		return dto.ActorsWithFilmsListResponse{}, fmt.Errorf("getting actors with movies: %w", err)
	}

	// Преобразуем доменные структуры в DTO
	result := make([]dto.ActorWithFilms, 0, len(actors))
	for _, actor := range actors {
		// Преобразуем фильмы актёра
		movies := make([]dto.MovieResponse, 0, len(actor.Movies))
		for _, movie := range actor.Movies {
			// Преобразуем актёров фильма (если есть)
			actors := make([]dto.ActorPreview, 0, len(movie.Actors))
			for _, actor := range movie.Actors {
				actors = append(actors, dto.ActorPreview{
					ID:   actor.ID,
					Name: actor.Name,
				})
			}

			movies = append(movies, dto.MovieResponse{
				ID:          movie.ID,
				Title:       movie.Title,
				Description: movie.Description,
				ReleaseYear: movie.ReleaseYear,
				Rating:      movie.Rating,
				Actors:      actors,
			})
		}

		// Создаём DTO актёра с фильмами
		actorDTO := dto.ActorWithFilms{
			ID:        actor.ID,
			Name:      actor.Name,
			Gender:    actor.Gender,
			BirthDate: actor.BirthDate.Format("2006-01-02"),
			Movies:    movies,
		}

		result = append(result, actorDTO)
	}

	return dto.ActorsWithFilmsListResponse{Actors: result}, nil
}
