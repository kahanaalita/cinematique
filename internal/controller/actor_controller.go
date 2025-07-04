package controller

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"cinematique/internal/controller/dto"
	"cinematique/internal/domain"
)

// actorController контроллер актёров.
type actorController struct {
	actorService ServiceActor
}

// PartialUpdateActor частично обновляет данные актёра
func (c *actorController) PartialUpdateActor(ctx *gin.Context, id int, update dto.ActorUpdate) (dto.ActorResponse, error) {
	// Логируем входные данные
	log.Printf("PartialUpdateActor вызван с id=%d, update=%+v", id, update)
	
	// Получаем текущие данные актёра
	actor, err := c.actorService.GetByID(id)
	if err != nil {
		log.Printf("Ошибка получения актёра с id=%d: %v", id, err)
		if errors.Is(err, domain.ErrActorNotFound) {
			return dto.ActorResponse{}, domain.ErrActorNotFound
		}
		return dto.ActorResponse{}, fmt.Errorf("получение актёра: %w", err)
	}
	log.Printf("Текущие данные актёра: %+v", actor)

	// Логируем обновляемые поля
	log.Printf("Обновляем актёра с полями: Name=%v, Gender=%v, BirthDate=%v", 
		update.Name, update.Gender, update.BirthDate)

	// Создаем обновленную структуру актёра
	updatedActor := domain.Actor{
		ID:        id,
		Name:      actor.Name,
		Gender:    actor.Gender,
		BirthDate: actor.BirthDate,
	}
	
	// Обновляем только переданные поля
	if update.Name != nil {
		updatedActor.Name = *update.Name
	}
	if update.Gender != nil {
		updatedActor.Gender = *update.Gender
	}
	if update.BirthDate != nil {
		updatedActor.BirthDate = *update.BirthDate
	}

	// Валидируем обновленные данные
	if err := validateActorInput(updatedActor.Name, updatedActor.Gender, updatedActor.BirthDate.Format("2006-01-02")); err != nil {
		log.Printf("Ошибка валидации для актёра (ID: %d): %v", id, err)
		return dto.ActorResponse{}, fmt.Errorf("ошибка валидации: %w", err)
	}

	// Обновляем актёра в хранилище
	if err := c.actorService.Update(updatedActor); err != nil {
		log.Printf("Ошибка обновления актёра (ID: %d): %v", id, err)
		return dto.ActorResponse{}, fmt.Errorf("обновление актёра: %w", err)
	}

	// Получаем обновленные данные актёра
	updated, err := c.actorService.GetByID(id)
	if err != nil {
		log.Printf("Ошибка получения обновлённых данных актёра (ID: %d): %v", id, err)
		return dto.ActorResponse{}, fmt.Errorf("получение обновлённых данных актёра: %w", err)
	}

	// Преобразуем в DTO и возвращаем
	return dto.ActorResponse{
		ID:        updated.ID,
		Name:      updated.Name,
		Gender:    updated.Gender,
		BirthDate: updated.BirthDate.Format("2006-01-02"),
	}, nil
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
		return fmt.Errorf("имя: должно быть от 1 до 100 символов")
	}

	gender = strings.ToLower(strings.TrimSpace(gender))
	if gender != "male" && gender != "female" && gender != "other" {
		return fmt.Errorf("пол: должно быть 'male', 'female' или 'other'")
	}

	birth, err := time.Parse("2006-01-02", birthDate)
	if err != nil {
		return fmt.Errorf("дата рождения: должна быть в формате YYYY-MM-DD")
	}

	if birth.After(time.Now()) {
		return fmt.Errorf("дата рождения: не может быть в будущем")
	}

	minDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	if birth.Before(minDate) {
		return fmt.Errorf("дата рождения: не может быть раньше 1900-01-01")
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
		if errors.Is(err, domain.ErrActorNotFound) {
			return dto.ActorResponse{}, domain.ErrActorNotFound
		}
		return dto.ActorResponse{}, fmt.Errorf("получение актёра: %w", err)
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
		if errors.Is(err, domain.ErrActorNotFound) {
			return dto.ActorResponse{}, domain.ErrActorNotFound
		}
		return dto.ActorResponse{}, fmt.Errorf("получение актёра: %w", err)
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
			return dto.ActorResponse{}, fmt.Errorf("неверный формат даты рождения: %w", err)
		}
		updatedBirthDate = birthDate
	}

	// Валидируем все поля разом
	if err := validateActorInput(
		updatedName,
		updatedGender,
		updatedBirthDate.Format("2006-01-02"),
	); err != nil {
		return dto.ActorResponse{}, fmt.Errorf("ошибка валидации: %w", err)
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

// DeleteActor удаляет актёра
func (c *actorController) DeleteActor(ctx *gin.Context, id int) error {
	log.Printf("Попытка удаления актёра с ID: %d", id)

	// Проверяем существование актёра
	_, err := c.actorService.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrActorNotFound) {
			return domain.ErrActorNotFound
		}
		return fmt.Errorf("получение актёра для удаления: %w", err)
	}

	// Проверяем, есть ли у актёра связанные фильмы
	movies, err := c.actorService.GetMovies(id)
	if err != nil {
		return fmt.Errorf("получение фильмов актёра для удаления: %w", err)
	}
	if len(movies) > 0 {
		log.Printf("Невозможно удалить актёра (ID: %d): актёр имеет связанные фильмы", id)
		return domain.ErrActorHasMovies
	}

	err = c.actorService.Delete(id)
	if err != nil {
		return fmt.Errorf("ошибка удаления актёра (ID: %d): %w", id, err)
	}

	log.Printf("Актёр с ID: %d успешно удалён", id)
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
		return dto.ActorsWithFilmsListResponse{}, fmt.Errorf("получение актёров с фильмами: %w", err)
	}

	// Преобразуем доменные структуры в DTO
	result := make([]dto.ActorWithFilms, 0, len(actors))
	for _, actor := range actors {
		// Преобразуем фильмы актёра
		movies := make([]dto.MovieResponse, 0, len(actor.Movies))
		for _, movie := range actor.Movies {
			// Преобразуем актёров фильма (если есть)
			var actorsList []dto.ActorPreview
			if len(movie.Actors) > 0 {
				actorsList = make([]dto.ActorPreview, 0, len(movie.Actors))
				for _, actor := range movie.Actors {
					actorsList = append(actorsList, dto.ActorPreview{
						ID:   actor.ID,
						Name: actor.Name,
					})
				}
			} else {
				actorsList = nil
			}

			movies = append(movies, dto.MovieResponse{
				ID:          movie.ID,
				Title:       movie.Title,
				Description: movie.Description,
				ReleaseYear: movie.ReleaseYear,
				Rating:      movie.Rating,
				Actors:      actorsList,
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
