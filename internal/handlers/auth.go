package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"cinematique/internal/controller/dto"
	"cinematique/internal/kafka"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	userLoginsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "user_logins_total",
			Help: "Total number of user logins.",
		},
	)
)

func init() {
	prometheus.MustRegister(userLoginsTotal)
}

// AuthHandler отвечает за обработку запросов, связанных с аутентификацией.
type AuthHandler struct {
	service AuthService
	producerPool *kafka.ProducerPool // Используем пул продюсеров
}

// NewAuthHandler создаёт новый обработчик аутентификации.
func NewAuthHandler(service AuthService, producerPool *kafka.ProducerPool) *AuthHandler {
	return &AuthHandler{service: service, producerPool: producerPool}
}

// Register обрабатывает регистрацию пользователя
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный запрос"})
		return
	}
	_, err := h.service.Register(req.Username, req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Отправляем событие регистрации в Kafka
	event := map[string]interface{}{
		"type":      "user_registered",
		"username":  req.Username,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	eventBytes, _ := json.Marshal(event)
	if err := h.producerPool.Produce("user-registration", []byte(req.Username), eventBytes); err != nil {
		// Логируем ошибку, но не блокируем регистрацию пользователя
		// В реальном приложении здесь может быть более сложная логика обработки ошибок
		// например, отправка в Dead Letter Queue или повторная попытка
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send registration event"})
		return
	}

	c.Status(http.StatusCreated)
}

// Login обрабатывает вход пользователя и возвращает JWT токены
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный запрос"})
		return
	}

	tokenPair, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Отправляем событие входа в систему в Kafka
	event := map[string]interface{}{
		"type":      "user_logged_in",
		"username":  req.Username,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	eventBytes, _ := json.Marshal(event)
	if err := h.producerPool.Produce("user_events", []byte(req.Username), eventBytes); err != nil {
		// Логируем ошибку, но не блокируем вход пользователя
		// В реальном приложении здесь может быть более сложная логика обработки ошибок
		// например, отправка в Dead Letter Queue или повторная попытка
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send login event"})
		return
	}

	// Увеличиваем счётчик входов в систему
	userLoginsTotal.Inc()

	// Возвращаем оба токена клиенту
	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	})
}

// Refresh обрабатывает запрос на обновление токена
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный запрос"})
		return
	}

	tokenPair, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	})
}

// Logout обрабатывает выход пользователя из системы
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.service.Logout(req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	c.Status(http.StatusNoContent)
}
