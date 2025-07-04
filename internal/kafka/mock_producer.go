package kafka

import (
	"context"
	"github.com/stretchr/testify/mock"
)

// MockProducer реализует интерфейс ProducerInterface для тестирования.
type MockProducer struct {
	mock.Mock
}

// NewMockProducer создаёт новый экземпляр MockProducer
func NewMockProducer() *MockProducer {
	return &MockProducer{}
}

// Produce реализует мок-метод отправки сообщения.
func (m *MockProducer) Produce(ctx context.Context, topic string, key, value []byte) error {
	args := m.Called(ctx, topic, key, value)
	return args.Error(0)
}

// Close реализует мок-метод закрытия продюсера.
func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}
