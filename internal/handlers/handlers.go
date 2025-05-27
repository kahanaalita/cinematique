package handlers

import (
	"net/http"
	"strconv"

	"cinematigue/internal/controller/dto"
	"github.com/gin-gonic/gin"
)

// --- Интерфейсы ActorController ---
type ActorController interface {
	CreateActor(c *gin.Context, req dto.CreateActorRequest) (dto.ActorResponse, error)
	GetActorByID(c *gin.Context, id int) (dto.ActorResponse, error)
	UpdateActor(c *gin.Context, id int, req dto.UpdateActorRequest) (dto.ActorResponse, error)
	DeleteActor(c *gin.Context, id int) error
	ListActors(c *gin.Context) (dto.ActorsListResponse, error)
}

// --- Интерфейсы MovieController ---
type MovieController interface {
	CreateMovie(c *gin.Context, req dto.CreateMovieRequest) (dto.MovieResponse, error)
	GetMovieByID(c *gin.Context, id int) (dto.MovieResponse, error)
	UpdateMovie(c *gin.Context, id int, req dto.UpdateMovieRequest) (dto.MovieResponse, error)
	DeleteMovie(c *gin.Context, id int) error
	ListMovies(c *gin.Context) (dto.MoviesListResponse, error)
}

// --- Handler структуры ---
type ActorHandler struct {
	controller ActorController
}

type MovieHandler struct {
	controller MovieController
}

func NewActorHandler(controller ActorController) *ActorHandler {
	return &ActorHandler{controller: controller}
}

func NewMovieHandler(controller MovieController) *MovieHandler {
	return &MovieHandler{controller: controller}
}

// --- Actor endpoints ---
func (h *ActorHandler) Create(c *gin.Context) {
	var req dto.CreateActorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := h.controller.CreateActor(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *ActorHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.controller.GetActorByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ActorHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req dto.UpdateActorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := h.controller.UpdateActor(c, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ActorHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	err = h.controller.DeleteActor(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ActorHandler) List(c *gin.Context) {
	resp, err := h.controller.ListActors(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Movie endpoints ---
func (h *MovieHandler) Create(c *gin.Context) {
	var req dto.CreateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := h.controller.CreateMovie(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *MovieHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.controller.GetMovieByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *MovieHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req dto.UpdateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := h.controller.UpdateMovie(c, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *MovieHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	err = h.controller.DeleteMovie(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MovieHandler) List(c *gin.Context) {
	resp, err := h.controller.ListMovies(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RegisterActorRoutes registers all actor-related routes
func RegisterActorRoutes(router *gin.Engine, handler *ActorHandler) {
	router.POST("/actors", handler.Create)
	router.GET("/actors/:id", handler.GetByID)
	router.PUT("/actors/:id", handler.Update)
	router.DELETE("/actors/:id", handler.Delete)
	router.GET("/actors", handler.List)
}

// RegisterMovieRoutes registers all movie-related routes
func RegisterMovieRoutes(router *gin.Engine, handler *MovieHandler) {
	router.POST("/movies", handler.Create)
	router.GET("/movies/:id", handler.GetByID)
	router.PUT("/movies/:id", handler.Update)
	router.DELETE("/movies/:id", handler.Delete)
	router.GET("/movies", handler.List)
}
