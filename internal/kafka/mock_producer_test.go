package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewMockProducer(t *testing.T) {
	mockProducer := NewMockProducer()

	assert.NotNil(t, mockProducer)
	assert.Implements(t, (*ProducerInterface)(nil), mockProducer)
}

func TestMockProducer_Produce_Success(t *testing.T) {
	mockProducer := NewMockProducer()

	ctx := context.Background()
	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Ожидаем успешную отправку
	mockProducer.On("Produce", ctx, topic, key, value).Return(nil)

	err := mockProducer.Produce(ctx, topic, key, value)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestMockProducer_Produce_Error(t *testing.T) {
	mockProducer := NewMockProducer()

	ctx := context.Background()
	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Симулируем ошибку при отправке
	produceError := errors.New("mock produce error")
	mockProducer.On("Produce", ctx, topic, key, value).Return(produceError)

	err := mockProducer.Produce(ctx, topic, key, value)

	assert.Error(t, err)
	assert.Equal(t, produceError, err)
	mockProducer.AssertExpectations(t)
}

func TestMockProducer_Close_Success(t *testing.T) {
	mockProducer := NewMockProducer()

	// Ожидаем успешное закрытие
	mockProducer.On("Close").Return(nil)

	err := mockProducer.Close()

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestMockProducer_Close_Error(t *testing.T) {
	mockProducer := NewMockProducer()

	// Симулируем ошибку при закрытии
	closeError := errors.New("mock close error")
	mockProducer.On("Close").Return(closeError)

	err := mockProducer.Close()

	assert.Error(t, err)
	assert.Equal(t, closeError, err)
	mockProducer.AssertExpectations(t)
}

func TestMockProducer_MultipleCalls(t *testing.T) {
	mockProducer := NewMockProducer()

	ctx := context.Background()
	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Ожидаем несколько вызовов Produce
	mockProducer.On("Produce", ctx, topic, key, value).Return(nil).Times(3)

	// Вызываем Produce несколько раз
	for i := 0; i < 3; i++ {
		err := mockProducer.Produce(ctx, topic, key, value)
		assert.NoError(t, err)
	}

	mockProducer.AssertExpectations(t)
}

func TestMockProducer_DifferentParameters(t *testing.T) {
	mockProducer := NewMockProducer()

	ctx := context.Background()

	// Ожидаем вызовы с разными параметрами
	mockProducer.On("Produce", ctx, "topic1", []byte("key1"), []byte("value1")).Return(nil)
	mockProducer.On("Produce", ctx, "topic2", []byte("key2"), []byte("value2")).Return(errors.New("error"))

	// Вызываем с разными параметрами
	err1 := mockProducer.Produce(ctx, "topic1", []byte("key1"), []byte("value1"))
	assert.NoError(t, err1)

	err2 := mockProducer.Produce(ctx, "topic2", []byte("key2"), []byte("value2"))
	assert.Error(t, err2)

	mockProducer.AssertExpectations(t)
}

func TestMockProducer_InterfaceCompliance(t *testing.T) {
	// Проверяем, что MockProducer реализует интерфейс ProducerInterface
	var _ ProducerInterface = (*MockProducer)(nil)

	mockProducer := NewMockProducer()

	// Проверяем, что можно использовать как ProducerInterface
	var producer ProducerInterface = mockProducer

	ctx := context.Background()
	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	mockProducer.On("Produce", ctx, topic, key, value).Return(nil)
	mockProducer.On("Close").Return(nil)

	err := producer.Produce(ctx, topic, key, value)
	assert.NoError(t, err)

	err = producer.Close()
	assert.NoError(t, err)

	mockProducer.AssertExpectations(t)
}

func TestMockProducer_WithProducerPool(t *testing.T) {
	mockProducer := NewMockProducer()
	mockProducer.On("Close").Return(nil).Maybe()
	// Создаем ProducerPool с MockProducer
	pool := NewProducerPool(mockProducer, 1, 5)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Ожидаем успешную отправку через пул
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(nil)

	err := pool.Produce(topic, key, value)
	assert.NoError(t, err)

	// Даем время воркеру обработать сообщение
	time.Sleep(100 * time.Millisecond)

	mockProducer.AssertExpectations(t)
}
