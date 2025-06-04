package handlers

import (
	"net/http"
	"strconv"

	"cinematigue/internal/auth"
	"cinematigue/internal/controller/dto"
	"github.com/gin-gonic/gin"
)

// ActorController описывает методы для работы с актёрами
type ActorController interface {
	CreateActor(c *gin.Context, req dto.CreateActorRequest) (dto.ActorResponse, error)
	GetActorByID(c *gin.Context, id int) (dto.ActorResponse, error)
	UpdateActor(c *gin.Context, id int, req dto.UpdateActorRequest) (dto.ActorResponse, error)
	DeleteActor(c *gin.Context, id int) error
	ListActors(c *gin.Context) (dto.ActorsListResponse, error)
	GetAllActorsWithMovies(c *gin.Context) (dto.ActorsWithFilmsListResponse, error)
	PartialUpdateActor(c *gin.Context, id int, update dto.ActorUpdate) error
}

// MovieController описывает методы для работы с фильмами
type MovieController interface {
	CreateMovie(c *gin.Context, req dto.CreateMovieRequest) (dto.MovieResponse, error)
	GetMovieByID(c *gin.Context, id int) (dto.MovieResponse, error)
	UpdateMovie(c *gin.Context, id int, req dto.UpdateMovieRequest) (dto.MovieResponse, error)
	DeleteMovie(c *gin.Context, id int) error
	ListMovies(c *gin.Context) (dto.MoviesListResponse, error)
	SearchMoviesByTitle(c *gin.Context) (dto.MoviesListResponse, error)
	SearchMoviesByActorName(c *gin.Context) (dto.MoviesListResponse, error)
	GetAllMoviesSorted(c *gin.Context) (dto.MoviesListResponse, error)
	CreateMovieWithActors(c *gin.Context, req dto.MovieWithActorsRequest) (dto.MovieResponse, error)
	UpdateMovieActors(c *gin.Context, movieID int, req dto.UpdateMovieActorsRequest) (dto.MovieActorsResponse, error)
	AddActorToMovie(c *gin.Context, movieID, actorID int) (dto.MovieResponse, error)
	RemoveActorFromMovie(c *gin.Context, movieID, actorID int) (dto.MovieResponse, error)
	GetActorsForMovieByID(c *gin.Context, movieID int) (dto.MovieActorsResponse, error)
	GetMoviesForActor(c *gin.Context, actorID int) (dto.ActorMoviesResponse, error)
	PartialUpdateMovie(c *gin.Context, id int, update dto.MovieUpdate) error
}

// Структуры
type ActorHandler struct {
	controller ActorController
}

type MovieHandler struct {
	controller MovieController
}

// NewActorHandler создаёт хендлер актёров
func NewActorHandler(controller ActorController) *ActorHandler {
	return &ActorHandler{controller: controller}
}

// NewMovieHandler создаёт хендлер фильмов
func NewMovieHandler(controller MovieController) *MovieHandler {
	return &MovieHandler{controller: controller}
}

// Методы ActorHandler ---
// Create создаёт актёра
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

// GetByID возвращает актёра по ID
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

// Update обновляет актёра
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
		if err.Error() == "actor not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// PartialUpdate частично обновляет актёра
func (h *ActorHandler) PartialUpdate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var update dto.ActorUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.controller.PartialUpdateActor(c, id, update); err != nil {
		if err.Error() == "actor not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.Status(http.StatusOK)
}

// Delete удаляет актёра
func (h *ActorHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	err = h.controller.DeleteActor(c, id)
	if err != nil {
		if err.Error() == "actor not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// List возвращает всех актёров
func (h *ActorHandler) List(c *gin.Context) {
	resp, err := h.controller.ListActors(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListWithMovies возвращает актёров с фильмами
func (h *ActorHandler) ListWithMovies(c *gin.Context) {
	resp, err := h.controller.GetAllActorsWithMovies(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Методы MovieHandler ---
// Create создаёт фильм
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

// GetByID возвращает фильм по ID
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

// Update обновляет фильм
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
		if err.Error() == "movie not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// PartialUpdate частично обновляет фильм
func (h *MovieHandler) PartialUpdate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var update dto.MovieUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.controller.PartialUpdateMovie(c, id, update); err != nil {
		if err.Error() == "movie not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.Status(http.StatusOK)
}

// Delete удаляет фильм
func (h *MovieHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	err = h.controller.DeleteMovie(c, id)
	if err != nil {
		if err.Error() == "movie not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// List возвращает все фильмы
func (h *MovieHandler) List(c *gin.Context) {
	resp, err := h.controller.ListMovies(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SearchByTitle ищет фильмы по названию
func (h *MovieHandler) SearchByTitle(c *gin.Context) {
	resp, err := h.controller.SearchMoviesByTitle(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SearchByActorName ищет фильмы по имени актёра
func (h *MovieHandler) SearchByActorName(c *gin.Context) {
	resp, err := h.controller.SearchMoviesByActorName(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListSorted возвращает отсортированные фильмы
func (h *MovieHandler) ListSorted(c *gin.Context) {
	resp, err := h.controller.GetAllMoviesSorted(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateWithActors создаёт фильм с актёрами
func (h *MovieHandler) CreateWithActors(c *gin.Context) {
	var req dto.MovieWithActorsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	resp, err := h.controller.CreateMovieWithActors(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// UpdateMovieActors обновляет актёров фильма
func (h *MovieHandler) UpdateMovieActors(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	var req dto.UpdateMovieActorsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	resp, err := h.controller.UpdateMovieActors(c, movieID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AddActorToMovie добавляет актёра в фильм
func (h *MovieHandler) AddActorToMovie(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	actorID, err := strconv.Atoi(c.Param("actorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid actor id"})
		return
	}

	resp, err := h.controller.AddActorToMovie(c, movieID, actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RemoveActorFromMovie удаляет актёра из фильма
func (h *MovieHandler) RemoveActorFromMovie(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	actorID, err := strconv.Atoi(c.Param("actorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid actor id"})
		return
	}

	resp, err := h.controller.RemoveActorFromMovie(c, movieID, actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetActorsForMovieByID возвращает актёров фильма
func (h *MovieHandler) GetActorsForMovieByID(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	resp, err := h.controller.GetActorsForMovieByID(c, movieID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMoviesForActor возвращает фильмы по актёру
func (h *MovieHandler) GetMoviesForActor(c *gin.Context) {
	actorID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid actor id"})
		return
	}

	resp, err := h.controller.GetMoviesForActor(c, actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// --- Регистрация роутов ---
// RegisterActorRoutes регистрирует маршруты для актёров
func RegisterActorRoutes(router *gin.Engine, handler *ActorHandler) {
	r := router.Group("/actors")
	{
		// Доступны всем авторизованным (чтение и поиск)
		r.GET("", handler.List)
		r.GET(":id", handler.GetByID)
		r.GET("/with-movies", handler.ListWithMovies)

		// Для всех методов, изменяющих данные, требуется роль admin
		r.Use(auth.JWTAuthMiddleware(), auth.OnlyAdminOrReadOnly())
		r.POST("", handler.Create)
		r.PUT(":id", handler.Update)
		r.PATCH(":id", handler.PartialUpdate)
		r.DELETE(":id", handler.Delete)
	}
}

// RegisterMovieRoutes регистрирует маршруты для фильмов
func RegisterMovieRoutes(router *gin.Engine, handler *MovieHandler, requireAuth gin.HandlerFunc) {
	movies := router.Group("/movies")
	{
		// Доступны всем авторизованным (чтение и поиск)
		movies.GET("", handler.List)                            // Список фильмов
		movies.GET("/search", handler.SearchByTitle)            // Поиск фильмов по названию
		movies.GET("/search/actor", handler.SearchByActorName)  // Поиск фильмов по имени актёра
		movies.GET(":id", handler.GetByID)                      // Получить фильм по ID
		movies.GET(":id/actors", handler.GetActorsForMovieByID) // Актёры фильма
		movies.GET("/actor/:id", handler.GetMoviesForActor)     // Фильмы по актёру

		// Для всех методов, изменяющих данные, требуется роль admin
		movies.Use(requireAuth, auth.OnlyAdminOrReadOnly())
		movies.POST("", handler.CreateWithActors)           // Создать фильм
		movies.PUT(":id", handler.Update)                   // Обновить фильм
		movies.PATCH(":id", handler.PartialUpdate)          // Частичное обновление фильма
		movies.DELETE(":id", handler.Delete)                // Удалить фильм
		movies.PUT(":id/actors", handler.UpdateMovieActors) // Обновить актёров фильма
	}
}

// RegisterAuthRoutes регистрирует маршруты для аутентификации
func RegisterAuthRoutes(router *gin.Engine, handler *AuthHandler) {
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
	}
}

// RegisterAllRoutes регистрирует все маршруты
func RegisterAllRoutes(router *gin.Engine, actorHandler *ActorHandler, movieHandler *MovieHandler, authHandler *AuthHandler) {
	RegisterAuthRoutes(router, authHandler)
	RegisterActorRoutes(router, actorHandler)
	RegisterMovieRoutes(router, movieHandler, auth.JWTAuthMiddleware())
}
