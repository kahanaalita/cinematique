package handlers

import (
	"encoding/json" // Добавляем импорт encoding/json
	"errors"
	"fmt" // Добавляем импорт fmt
	"log" // Добавляем импорт log
	"net/http"
	"strconv"
	"strings" // Добавляем импорт strings
	"time"    // Добавляем импорт time

	"cinematique/internal/auth"
	"cinematique/internal/controller/dto"
	"cinematique/internal/domain"
	"cinematique/internal/kafka" // Добавляем импорт kafka
	"cinematique/internal/keycloak"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	moviesSearchedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "movies_searched_total",
			Help: "Общее количество поисковых запросов фильмов.",
		},
	)
	moviesViewedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "movies_viewed_total",
			Help: "Общее количество просмотров страниц фильмов.",
		},
	)
)

func init() {
	prometheus.MustRegister(moviesSearchedTotal)
	prometheus.MustRegister(moviesViewedTotal)
}

// ActorController описывает методы для работы с актёрами
type ActorController interface {
	CreateActor(c *gin.Context, req dto.CreateActorRequest) (dto.ActorResponse, error)
	GetActorByID(c *gin.Context, id int) (dto.ActorResponse, error)
	UpdateActor(c *gin.Context, id int, req dto.UpdateActorRequest) (dto.ActorResponse, error)
	DeleteActor(c *gin.Context, id int) error
	ListActors(c *gin.Context) (dto.ActorsListResponse, error)
	GetAllActorsWithMovies(c *gin.Context) (dto.ActorsWithFilmsListResponse, error)
	PartialUpdateActor(c *gin.Context, id int, update dto.ActorUpdate) (dto.ActorResponse, error)
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
	controller   MovieController
	producerPool *kafka.ProducerPool // Используем пул продюсеров
}

// NewActorHandler создаёт обработчик (handler) для актёров
func NewActorHandler(controller ActorController) *ActorHandler {
	return &ActorHandler{controller: controller}
}

// NewMovieHandler создаёт обработчик (handler) для фильмов
func NewMovieHandler(controller MovieController, producerPool *kafka.ProducerPool) *MovieHandler {
	return &MovieHandler{controller: controller, producerPool: producerPool}
}

// Методы ActorHandler ---
// Create создаёт актёра
func (h *ActorHandler) Create(c *gin.Context) {
	var req dto.CreateActorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Validate required fields
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}
	if req.Gender == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gender is required"})
		return
	}
	if req.BirthDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "BirthDate is required"})
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
		if errors.Is(err, domain.ErrActorNotFound) {
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
	log.Println("Handling PATCH /api/actors/:id request")

	// Получаем ID актера из URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errMsg := "invalid id"
		log.Printf("Error: %s", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}
	log.Printf("Updating actor with ID: %d", id)

	// Парсим тело запроса
	var update dto.ActorUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		errMsg := "invalid request body: " + err.Error()
		log.Printf("Error: %s", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}
	log.Printf("Update data: %+v", update)

	// Проверяем, что хотя бы одно поле для обновления указано
	if update.Name == nil && update.Gender == nil && update.BirthDate == nil {
		errMsg := "no fields to update"
		log.Printf("Error: %s", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	// Вызываем метод контроллера для обновления актера
	updatedActor, err := h.controller.PartialUpdateActor(c, id, update)
	if err != nil {
		errMsg := fmt.Sprintf("Error updating actor: %v", err)
		log.Printf("Error: %s", errMsg)

		switch {
		case errors.Is(err, domain.ErrActorNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "actor not found"})
		case strings.Contains(err.Error(), "validation error"):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	log.Printf("Successfully updated actor with ID: %d", id)

	// Возвращаем обновленные данные актера
	c.JSON(http.StatusOK, updatedActor)
}

// Delete удаляет актёра
func (h *ActorHandler) Delete(c *gin.Context) {
	log.Printf("ActorHandler.Delete called with id: %s", c.Param("id"))

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Error parsing id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	fmt.Printf("=== Starting Delete handler for actor ID: %d ===\n", id)
	err = h.controller.DeleteActor(c, id)
	if err != nil {
		errMsg := err.Error()
		fmt.Printf("=== Error in Delete handler: %v, type: %T ===\n", errMsg, err)

		switch {
		case errors.Is(err, domain.ErrActorNotFound):
			fmt.Println("=== Returning 404 Not Found ===")
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		case strings.Contains(errMsg, "cannot delete actor: has") && strings.Contains(errMsg, "related movies"):
			fmt.Println("=== Returning 409 Conflict ===")
			c.JSON(http.StatusConflict, gin.H{"error": errMsg})
		default:
			fmt.Println("=== Returning 500 Internal Server Error ===")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	fmt.Println("=== Actor deleted successfully, returning 204 No Content ===")
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

	// Валидация обязательных полей
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}
	if req.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}
	if req.Rating < 0 || req.Rating > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating must be between 0 and 10"})
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
	moviesViewedTotal.Inc() // Увеличиваем счетчик при просмотре фильма

	// Отправляем событие просмотра фильма в Kafka
	event := map[string]interface{}{
		"type":      "movie_viewed",
		"movie_id":  id,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	eventBytes, _ := json.Marshal(event)
	h.producerPool.Produce("movie-views", []byte(strconv.Itoa(id)), eventBytes)

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

// Search ищет фильмы по названию или имени актёра
func (h *MovieHandler) Search(c *gin.Context) {
	title := c.Query("title")
	actorName := c.Query("actorName")

	var resp dto.MoviesListResponse
	var err error

	if title != "" {
		resp, err = h.controller.SearchMoviesByTitle(c)
	} else if actorName != "" {
		resp, err = h.controller.SearchMoviesByActorName(c)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one search parameter (title or actorName) is required"})
		return
	}

	if err != nil {
		// Check for specific errors from the controller indicating missing parameters
		if strings.Contains(err.Error(), "parameter is required") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Отправляем событие поиска фильма в Kafka
	event := map[string]interface{}{
		"type":      "movie_searched",
		"query":     c.Request.URL.Query(),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	eventBytes, _ := json.Marshal(event)
	h.producerPool.Produce("movie-searches", []byte(c.Request.URL.RawQuery), eventBytes)

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
	movieID, err := strconv.Atoi(c.Param("movieId"))
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
	movieID, err := strconv.Atoi(c.Param("movieId"))
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
		if errors.Is(err, domain.ErrMovieNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
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
		if errors.Is(err, domain.ErrActorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Если актёр существует, но фильмов нет — возвращаем пустой массив
	if resp.Movies == nil {
		resp.Movies = []dto.MovieResponse{}
	}

	c.JSON(http.StatusOK, resp)
}

// --- Регистрация роутов ---
// RegisterActorRoutes регистрирует маршруты для актёров
func RegisterActorRoutes(router *gin.RouterGroup, handler *ActorHandler, _ gin.HandlerFunc) {
	r := router.Group("/actors")

	// Группа для методов чтения (доступны всем аутентифицированным)
	r.GET("", handler.List)
	r.GET(":id", handler.GetByID)
	r.GET("/with-movies", handler.ListWithMovies)

	// Группа для методов записи (требуются права администратора)
	// JWTAuthMiddleware уже применен, поэтому проверяем только роль
	r.Use(auth.OnlyAdminOrReadOnly())

	r.POST("", handler.Create)
	r.PUT(":id", handler.Update)
	r.PATCH(":id", handler.PartialUpdate)
	r.DELETE(":id", handler.Delete)
}

// RegisterMovieRoutes регистрирует маршруты для фильмов
func RegisterMovieRoutes(router *gin.RouterGroup, handler *MovieHandler) {
	movies := router.Group("/movies")

	// Конкретные маршруты идут первыми
	movies.GET("", handler.List)
	movies.GET("/search", handler.Search)
	movies.GET("/sorted", handler.ListSorted)

	// Маршрут для получения фильмов актёра
	movies.GET("/actor/:id", handler.GetMoviesForActor)

	// Параметризованные маршруты идут после конкретных
	movies.GET(":id", handler.GetByID)
	movies.GET(":id/actors", handler.GetActorsForMovieByID)

	// Группа для методов записи (требуются права администратора)
	movies.Use(auth.OnlyAdminOrReadOnly())
	movies.POST("", handler.Create)
	movies.POST("/with-actors", handler.CreateWithActors)
	movies.PUT(":id", handler.Update)
	movies.PATCH(":id", handler.PartialUpdate)
	movies.DELETE(":id", handler.Delete)
	movies.POST(":id/actors", handler.UpdateMovieActors)
	movies.POST("add-actor/:movieId/:actorId", handler.AddActorToMovie)
	movies.DELETE("remove-actor/:movieId/:actorId", handler.RemoveActorFromMovie)
}

// RegisterAuthRoutes регистрирует маршруты для аутентификации
func RegisterAuthRoutes(router *gin.RouterGroup, handler *AuthHandler) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
		authGroup.POST("/refresh", handler.Refresh) // Добавляем эндпоинт для обновления токена
		authGroup.POST("/logout", handler.Logout)   // Добавляем эндпоинт для выхода
	}
}

// RegisterRateLimitRoutes регистрирует маршруты для мониторинга rate limiting
func RegisterRateLimitRoutes(router *gin.RouterGroup, handler *RateLimitHandler) {
	if handler != nil {
		rateLimitGroup := router.Group("/rate-limit")
		{
			rateLimitGroup.GET("/status", handler.GetStatus)
		}
	}
}

// RegisterAllRoutes регистрирует все маршруты
func RegisterAllRoutes(router *gin.RouterGroup, actorHandler *ActorHandler, movieHandler *MovieHandler, authHandler *AuthHandler, rateLimitHandler *RateLimitHandler) {
	// 1. Регистрируем публичные маршруты (без аутентификации)
	RegisterAuthRoutes(router, authHandler)

	// 2. Создаем группу для защищенных маршрутов
	protected := router.Group("/")
	// 3. Применяем гибридный middleware (поддерживает JWT и Keycloak)
	keycloakManager := keycloak.GetGlobalManager()
	var keycloakClient keycloak.KeycloakClient
	if keycloakManager.IsEnabled() {
		keycloakClient = keycloakManager.GetDefaultClient()
	}
	protected.Use(auth.HybridAuthMiddleware(keycloakClient))

	// 4. Регистрируем защищенные маршруты
	RegisterActorRoutes(protected, actorHandler, func(c *gin.Context) {})
	RegisterMovieRoutes(protected, movieHandler)
	RegisterRateLimitRoutes(protected, rateLimitHandler)
}
