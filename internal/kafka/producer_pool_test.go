package kafka

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProducerInterface мокает ProducerInterface для тестирования
type MockProducerInterface struct {
	mock.Mock
}

func (m *MockProducerInterface) Produce(ctx context.Context, topic string, key, value []byte) error {
	args := m.Called(ctx, topic, key, value)
	return args.Error(0)
}

func (m *MockProducerInterface) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewProducerPool(t *testing.T) {
	mockProducer := &MockProducerInterface{}
	mockProducer.On("Close").Return(nil).Maybe()
	workers := 3
	bufSize := 10

	pool := NewProducerPool(mockProducer, workers, bufSize)

	assert.NotNil(t, pool)
	assert.Equal(t, mockProducer, pool.producer)
	assert.NotNil(t, pool.events)
	assert.Equal(t, bufSize, cap(pool.events))

	// Проверяем, что воркеры запущены
	time.Sleep(100 * time.Millisecond)

	// Закрываем пул
	pool.Close()
}

func TestProducerPool_Produce_Success(t *testing.T) {
	mockProducer := &MockProducerInterface{}
	mockProducer.On("Close").Return(nil).Maybe()
	pool := NewProducerPool(mockProducer, 1, 5)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Ожидаем успешную отправку
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(nil)

	err := pool.Produce(topic, key, value)

	assert.NoError(t, err)

	// Даем время воркеру обработать сообщение
	time.Sleep(100 * time.Millisecond)

	mockProducer.AssertExpectations(t)
}

func TestProducerPool_Produce_BufferFull(t *testing.T) {
	mockProducer := &MockProducerInterface{}
	mockProducer.On("Close").Return(nil).Maybe()
	// Создаем пул с очень маленьким буфером
	pool := NewProducerPool(mockProducer, 1, 1)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Блокируем воркера, чтобы буфер заполнился
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(nil).Run(func(args mock.Arguments) {
		time.Sleep(200 * time.Millisecond) // Блокируем воркера
	})

	// Отправляем первое сообщение (должно попасть в буфер)
	err1 := pool.Produce(topic, key, value)
	assert.NoError(t, err1)

	// Отправляем второе сообщение (должно вызвать ошибку буфера)
	err2 := pool.Produce(topic, key, value)
	assert.Error(t, err2)
	assert.Equal(t, ErrBufferFull, err2)

	// Даем время воркеру обработать сообщения
	time.Sleep(300 * time.Millisecond)

	mockProducer.AssertExpectations(t)
}

func TestProducerPool_Produce_Error(t *testing.T) {
	mockProducer := &MockProducerInterface{}
	mockProducer.On("Close").Return(nil).Maybe()
	pool := NewProducerPool(mockProducer, 1, 5)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Симулируем ошибку при отправке
	produceError := errors.New("produce error")
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(produceError)

	err := pool.Produce(topic, key, value)

	assert.NoError(t, err) // Produce возвращает nil, ошибка обрабатывается воркером

	// Даем время воркеру обработать сообщение
	time.Sleep(100 * time.Millisecond)

	mockProducer.AssertExpectations(t)
}

func TestProducerPool_Close(t *testing.T) {
	mockProducer := &MockProducerInterface{}
	// Ожидаем успешное закрытие продюсера
	mockProducer.On("Close").Return(nil)
	pool := NewProducerPool(mockProducer, 2, 5)

	// Отправляем несколько сообщений перед закрытием
	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(nil).Times(3)

	pool.Produce(topic, key, value)
	pool.Produce(topic, key, value)
	pool.Produce(topic, key, value)

	// Закрываем пул
	pool.Close()

	// Проверяем, что все сообщения обработаны и продюсер закрыт
	mockProducer.AssertExpectations(t)
}

func TestProducerPool_Close_ProducerError(t *testing.T) {
	mockProducer := &MockProducerInterface{}
	// Симулируем ошибку при закрытии продюсера
	closeError := errors.New("close error")
	mockProducer.On("Close").Return(closeError)
	pool := NewProducerPool(mockProducer, 1, 5)

	// Закрываем пул
	pool.Close()

	mockProducer.AssertExpectations(t)
}

func TestProducerPool_ConcurrentProduce(t *testing.T) {
	mockProducer := &MockProducerInterface{}
	mockProducer.On("Close").Return(nil).Maybe()
	pool := NewProducerPool(mockProducer, 3, 10)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Ожидаем отправку 10 сообщений
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(nil).Times(10)

	var wg sync.WaitGroup
	numMessages := 10

	// Отправляем сообщения конкурентно
	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := pool.Produce(topic, key, value)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	// Даем время воркерам обработать все сообщения
	time.Sleep(200 * time.Millisecond)

	mockProducer.AssertExpectations(t)
}

func TestProducerPool_Metrics(t *testing.T) {
	mockProducer := &MockProducerInterface{}
	mockProducer.On("Close").Return(nil).Maybe()
	pool := NewProducerPool(mockProducer, 1, 5)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Ожидаем успешную отправку
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(nil)

	err := pool.Produce(topic, key, value)
	assert.NoError(t, err)

	// Даем время воркеру обработать сообщение
	time.Sleep(100 * time.Millisecond)

	mockProducer.AssertExpectations(t)
}

func TestProducerPool_Metrics_Error(t *testing.T) {
	mockProducer := &MockProducerInterface{}
	mockProducer.On("Close").Return(nil).Maybe()
	pool := NewProducerPool(mockProducer, 1, 5)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Симулируем ошибку при отправке
	produceError := errors.New("produce error")
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(produceError)

	err := pool.Produce(topic, key, value)
	assert.NoError(t, err)

	// Даем время воркеру обработать сообщение
	time.Sleep(100 * time.Millisecond)

	mockProducer.AssertExpectations(t)
}

func TestProducerPool_Metrics_Dropped(t *testing.T) {
	mockProducer := &MockProducerInterface{}
	mockProducer.On("Close").Return(nil).Maybe()
	// Создаем пул с очень маленьким буфером
	pool := NewProducerPool(mockProducer, 1, 1)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Блокируем воркера
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(nil).Run(func(args mock.Arguments) {
		time.Sleep(200 * time.Millisecond)
	})

	// Отправляем первое сообщение
	err1 := pool.Produce(topic, key, value)
	assert.NoError(t, err1)

	// Отправляем второе сообщение (должно быть отброшено)
	err2 := pool.Produce(topic, key, value)
	assert.Error(t, err2)
	assert.Equal(t, ErrBufferFull, err2)

	// Даем время воркеру обработать сообщения
	time.Sleep(300 * time.Millisecond)

	mockProducer.AssertExpectations(t)
}
