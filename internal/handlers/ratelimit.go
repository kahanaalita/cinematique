package handlers

import (
	"net/http"
	"time"

	"cinematique/internal/ratelimit"

	"github.com/gin-gonic/gin"
)

// RateLimitHandler обработчик для мониторинга rate limiting
type RateLimitHandler struct {
	limiter ratelimit.RateLimiter
	config  ratelimit.Config
}

// NewRateLimitHandler создает новый обработчик для rate limiting
func NewRateLimitHandler(limiter ratelimit.RateLimiter, config ratelimit.Config) *RateLimitHandler {
	return &RateLimitHandler{
		limiter: limiter,
		config:  config,
	}
}

// GetStatus возвращает текущий статус rate limiting для пользователя
func (h *RateLimitHandler) GetStatus(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
			"message": "Rate limiting is disabled",
		})
		return
	}

	// Получаем user_id
	userID := h.config.GetUserID(c)
	if userID == "" {
		userID = "anonymous"
	}

	// Получаем IP адрес
	ip := c.ClientIP()

	// Получаем endpoint из query параметра или используем текущий
	endpoint := c.Query("endpoint")
	if endpoint == "" {
		endpoint = c.Request.URL.Path
	}

	// Получаем текущий счетчик
	ctx := c.Request.Context()
	currentCount, err := h.limiter.GetCurrentCount(ctx, userID, ip, endpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get rate limit status",
		})
		return
	}

	limit := h.limiter.GetLimit()
	remaining := limit - currentCount
	if remaining < 0 {
		remaining = 0
	}

	resetTime := time.Now().Add(h.limiter.GetWindow())

	c.JSON(http.StatusOK, gin.H{
		"enabled":              true,
		"user_id":              userID,
		"ip":                   ip,
		"endpoint":             endpoint,
		"current_count":        currentCount,
		"limit":                limit,
		"remaining":            remaining,
		"window":               h.limiter.GetWindow().String(),
		"reset_time":           resetTime.Unix(),
		"reset_time_human":     resetTime.Format(time.RFC3339),
		"restricted_endpoints": h.config.RestrictedEndpoints,
	})
}
