package controller

import (
	"time"

	"github.com/gin-gonic/gin"

	"cinematigue/internal/controller/dto"
	"cinematigue/internal/domain"
)

type actorController struct {
	actorService ServiceActor
}

func NewActorController(actorService ServiceActor) *actorController {
	return &actorController{
		actorService: actorService,
	}
}

func (c *actorController) CreateActor(ctx *gin.Context, req dto.CreateActorRequest) (dto.ActorResponse, error) {
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

func (c *actorController) UpdateActor(ctx *gin.Context, id int, req dto.UpdateActorRequest) (dto.ActorResponse, error) {
	actor, err := c.actorService.GetByID(id)
	if err != nil {
		return dto.ActorResponse{}, err
	}

	// Обновляем только переданные поля
	if req.Name != nil {
		actor.Name = *req.Name
	}
	if req.Gender != nil {
		actor.Gender = *req.Gender
	}
	if req.BirthDate != nil {
		birthDate, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			return dto.ActorResponse{}, err
		}
		actor.BirthDate = birthDate
	}

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

func (c *actorController) DeleteActor(ctx *gin.Context, id int) error {
	return c.actorService.Delete(id)
}

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
