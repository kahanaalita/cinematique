package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient мок для Redis клиента
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRedisRateLimiter_IsAllowed(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockRedisClient)
	limiter := NewRedisRateLimiter(mockClient, 10, time.Minute)

	t.Run("first request should be allowed", func(t *testing.T) {
		// Настраиваем мок для первого запроса
		incrCmd := redis.NewIntCmd(ctx)
		incrCmd.SetVal(1)
		mockClient.On("Incr", ctx, mock.AnythingOfType("string")).Return(incrCmd).Once()

		expireCmd := redis.NewBoolCmd(ctx)
		expireCmd.SetVal(true)
		mockClient.On("Expire", ctx, mock.AnythingOfType("string"), time.Minute).Return(expireCmd).Once()

		allowed, err := limiter.IsAllowed(ctx, "user123", "192.168.1.1", "/api/movies")

		assert.NoError(t, err)
		assert.True(t, allowed)
		mockClient.AssertExpectations(t)
	})

	t.Run("request within limit should be allowed", func(t *testing.T) {
		mockClient = new(MockRedisClient) // Создаем новый мок
		limiter = NewRedisRateLimiter(mockClient, 10, time.Minute)

		incrCmd := redis.NewIntCmd(ctx)
		incrCmd.SetVal(5) // 5-й запрос из 10
		mockClient.On("Incr", ctx, mock.AnythingOfType("string")).Return(incrCmd).Once()

		allowed, err := limiter.IsAllowed(ctx, "user123", "192.168.1.1", "/api/movies")

		assert.NoError(t, err)
		assert.True(t, allowed)
		mockClient.AssertExpectations(t)
	})

	t.Run("request exceeding limit should be denied", func(t *testing.T) {
		mockClient = new(MockRedisClient)
		limiter = NewRedisRateLimiter(mockClient, 10, time.Minute)

		incrCmd := redis.NewIntCmd(ctx)
		incrCmd.SetVal(11) // 11-й запрос из 10
		mockClient.On("Incr", ctx, mock.AnythingOfType("string")).Return(incrCmd).Once()

		allowed, err := limiter.IsAllowed(ctx, "user123", "192.168.1.1", "/api/movies")

		assert.NoError(t, err)
		assert.False(t, allowed)
		mockClient.AssertExpectations(t)
	})
}

func TestRedisRateLimiter_GetCurrentCount(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockRedisClient)
	limiter := NewRedisRateLimiter(mockClient, 10, time.Minute)

	t.Run("should return current count", func(t *testing.T) {
		getCmd := redis.NewStringCmd(ctx)
		getCmd.SetVal("5")
		mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(getCmd).Once()

		count, err := limiter.GetCurrentCount(ctx, "user123", "192.168.1.1", "/api/movies")

		assert.NoError(t, err)
		assert.Equal(t, 5, count)
		mockClient.AssertExpectations(t)
	})

	t.Run("should return 0 when key doesn't exist", func(t *testing.T) {
		mockClient = new(MockRedisClient)
		limiter = NewRedisRateLimiter(mockClient, 10, time.Minute)

		getCmd := redis.NewStringCmd(ctx)
		getCmd.SetErr(redis.Nil)
		mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(getCmd).Once()

		count, err := limiter.GetCurrentCount(ctx, "user123", "192.168.1.1", "/api/movies")

		assert.NoError(t, err)
		assert.Equal(t, 0, count)
		mockClient.AssertExpectations(t)
	})
}
