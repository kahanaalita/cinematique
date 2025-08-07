package ratelimit

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter интерфейс для rate limiting
type RateLimiter interface {
	IsAllowed(ctx context.Context, userID, ip, endpoint string) (bool, error)
	GetCurrentCount(ctx context.Context, userID, ip, endpoint string) (int, error)
	GetLimit() int
	GetWindow() time.Duration
}

// Config конфигурация для rate limiter middleware
type Config struct {
	Enabled bool
	// Endpoints которые нужно ограничивать (если пусто - все endpoints)
	RestrictedEndpoints []string
	// Функция для извлечения user_id из контекста
	GetUserID func(c *gin.Context) string
}

// Middleware создает middleware для rate limiting
func Middleware(limiter RateLimiter, config Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("RateLimit middleware triggered for path:", c.Request.URL.Path)
		// Если rate limiting отключен, пропускаем
		if !config.Enabled {
			log.Println("Rate limiting disabled")
			c.Next()
			return
		}

		// Проверяем, нужно ли ограничивать данный endpoint
		if len(config.RestrictedEndpoints) > 0 {
			currentPath := c.Request.URL.Path
			shouldLimit := false
			for _, endpoint := range config.RestrictedEndpoints {
				if strings.HasPrefix(currentPath, endpoint) {
					shouldLimit = true
					break
				}
			}
			if !shouldLimit {
				c.Next()
				return
			}
		}

		// Получаем user_id
		userID := config.GetUserID(c)
		if userID == "" {
			userID = "anonymous"
		}

		// Получаем IP адрес
		ip := getClientIP(c)

		// Получаем endpoint
		endpoint := c.Request.URL.Path

		// Проверяем лимит
		ctx := c.Request.Context()
		allowed, err := limiter.IsAllowed(ctx, userID, ip, endpoint)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limiter error",
			})
			c.Abort()
			return
		}

		if !allowed {
			// Получаем текущий счетчик для информативности
			currentCount, _ := limiter.GetCurrentCount(ctx, userID, ip, endpoint)

			c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.GetLimit()))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(limiter.GetWindow()).Unix(), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":         "Too many requests",
				"message":       fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %v", limiter.GetLimit(), limiter.GetWindow()),
				"current_count": currentCount,
				"limit":         limiter.GetLimit(),
				"window":        limiter.GetWindow().String(),
				"retry_after":   int(limiter.GetWindow().Seconds()),
			})
			c.Abort()
			return
		}

		// Добавляем заголовки с информацией о лимитах
		currentCount, _ := limiter.GetCurrentCount(ctx, userID, ip, endpoint)
		remaining := limiter.GetLimit() - currentCount
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.GetLimit()))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(limiter.GetWindow()).Unix(), 10))

		c.Next()
	}
}

// getClientIP извлекает IP адрес клиента
func getClientIP(c *gin.Context) string {
	// Проверяем заголовки прокси
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For может содержать несколько IP через запятую
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	if ip := c.GetHeader("X-Client-IP"); ip != "" {
		return ip
	}

	// Используем RemoteAddr как fallback
	ip := c.ClientIP()
	return ip
}
