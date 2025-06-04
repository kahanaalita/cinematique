package handlers

import (
	"cinematigue/internal/controller/dto"
	"cinematigue/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AuthHandler обрабатывает HTTP-запросы, связанные с аутентификацией
type AuthHandler struct {
	service *service.AuthService
}

// NewAuthHandler создаёт обработчик аутентификации
func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register обрабатывает регистрацию пользователя
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	_, err := h.service.Register(req.Username, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

// Login обрабатывает вход пользователя и возвращает JWT-токен
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	token, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.AuthResponse{Token: token})
}
