package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cinematique/internal/ratelimit"
)

func rateLimitMain() {
	// Пример использования Redis Rate Limiter

	// 1. Создаем Redis клиент
	redisClient, err := ratelimit.NewRedisClient("localhost", "6379", "", 0)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// 2. Создаем rate limiter (10 запросов в минуту)
	limiter := ratelimit.NewRedisRateLimiter(redisClient, 10, time.Minute)

	// 3. Тестируем rate limiting
	ctx := context.Background()
	userID := "user123"
	ip := "192.168.1.1"
	endpoint := "/api/movies"

	fmt.Println("Testing Rate Limiter...")
	fmt.Printf("Limit: %d requests per %v\n", limiter.GetLimit(), limiter.GetWindow())
	fmt.Println()

	// Делаем 15 запросов (больше лимита)
	for i := 1; i <= 15; i++ {
		allowed, err := limiter.IsAllowed(ctx, userID, ip, endpoint)
		if err != nil {
			log.Printf("Error checking rate limit: %v", err)
			continue
		}

		currentCount, _ := limiter.GetCurrentCount(ctx, userID, ip, endpoint)

		status := "✅ ALLOWED"
		if !allowed {
			status = "❌ DENIED"
		}

		fmt.Printf("Request %2d: %s (count: %d/%d)\n",
			i, status, currentCount, limiter.GetLimit())

		// Небольшая задержка между запросами
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println()
	fmt.Println("Waiting for window to reset...")
	time.Sleep(61 * time.Second)

	// Проверяем после сброса окна
	allowed, _ := limiter.IsAllowed(ctx, userID, ip, endpoint)
	currentCount, _ := limiter.GetCurrentCount(ctx, userID, ip, endpoint)

	fmt.Printf("After reset: %s (count: %d/%d)\n",
		map[bool]string{true: "✅ ALLOWED", false: "❌ DENIED"}[allowed],
		currentCount, limiter.GetLimit())
}

// Пример интеграции с Gin
func main() {
	rateLimitMain()
}

func ginExample() {
	// Этот код показывает, как интегрировать rate limiter с Gin

	/*
		import (
			"github.com/gin-gonic/gin"
			"cinematique/internal/ratelimit"
		)

		// Создаем Redis клиент
		redisClient, _ := ratelimit.NewRedisClient("localhost", "6379", "", 0)

		// Создаем rate limiter
		limiter := ratelimit.NewRedisRateLimiter(redisClient, 1000, time.Minute)

		// Конфигурация middleware
		config := ratelimit.Config{
			Enabled: true,
			RestrictedEndpoints: []string{"/api/movies", "/api/actors"},
			GetUserID: func(c *gin.Context) string {
				if userID, exists := c.Get("user_id"); exists {
					return userID.(string)
				}
				return "anonymous"
			},
		}

		// Создаем Gin router
		router := gin.Default()

		// Добавляем rate limiting middleware
		router.Use(ratelimit.Middleware(limiter, config))

		// Регистрируем routes
		router.GET("/api/movies", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Movies endpoint"})
		})

		router.Run(":8080")
	*/
}
