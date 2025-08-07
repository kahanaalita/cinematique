package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient интерфейс для работы с Redis
type RedisClient interface {
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Close() error
}

// RedisRateLimiter реализует rate limiting через Redis
type RedisRateLimiter struct {
	client RedisClient
	limit  int
	window time.Duration
}

// NewRedisRateLimiter создает новый rate limiter
func NewRedisRateLimiter(client RedisClient, limit int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

// IsAllowed проверяет, разрешен ли запрос
func (r *RedisRateLimiter) IsAllowed(ctx context.Context, userID, ip, endpoint string) (bool, error) {
	// Формируем ключ: ratelimit:{user_id}:{ip}:{endpoint}:{timestamp_minute}
	now := time.Now()
	timestampMinute := now.Truncate(time.Minute).Unix()
	key := fmt.Sprintf("ratelimit:%s:%s:%s:%d", userID, ip, endpoint, timestampMinute)

	// Увеличиваем счетчик
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to increment counter: %w", err)
	}

	// Если это первый запрос для данного ключа, устанавливаем TTL
	if count == 1 {
		if err := r.client.Expire(ctx, key, r.window).Err(); err != nil {
			return false, fmt.Errorf("failed to set TTL: %w", err)
		}
	}

	// Проверяем лимит
	return count <= int64(r.limit), nil
}

// GetCurrentCount возвращает текущее количество запросов
func (r *RedisRateLimiter) GetCurrentCount(ctx context.Context, userID, ip, endpoint string) (int, error) {
	now := time.Now()
	timestampMinute := now.Truncate(time.Minute).Unix()
	key := fmt.Sprintf("ratelimit:%s:%s:%s:%d", userID, ip, endpoint, timestampMinute)

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get current count: %w", err)
	}

	count, err := strconv.Atoi(result)
	if err != nil {
		return 0, fmt.Errorf("failed to parse count: %w", err)
	}

	return count, nil
}

// GetLimit возвращает установленный лимит
func (r *RedisRateLimiter) GetLimit() int {
	return r.limit
}

// GetWindow возвращает временное окно
func (r *RedisRateLimiter) GetWindow() time.Duration {
	return r.window
}
