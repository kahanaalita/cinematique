package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// WriterInterface определяет интерфейс для kafka.Writer
type WriterInterface interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
	Stats() kafka.WriterStats
}

// TestProducer для тестирования с мок-писателями
type TestProducer struct {
	writer    WriterInterface
	dlqWriter WriterInterface
}

func (p *TestProducer) Produce(ctx context.Context, topic string, key, value []byte) error {
	message := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	err := p.writer.WriteMessages(ctx, message)
	if err != nil {
		// Попытка отправить в DLQ
		if p.dlqWriter != nil {
			dlqMessage := kafka.Message{
				Key:   key,
				Value: []byte("original_topic: " + topic + ", error: " + err.Error() + ", message: " + string(value)),
			}
			if dlqErr := p.dlqWriter.WriteMessages(context.Background(), dlqMessage); dlqErr != nil {
				return errors.New("failed to write to topic " + topic + " and DLQ: " + err.Error())
			}
		}
		return err
	}
	return nil
}

func (p *TestProducer) Close() error {
	err := p.writer.Close()
	if err != nil {
		return err
	}

	if p.dlqWriter != nil {
		if dlqErr := p.dlqWriter.Close(); dlqErr != nil {
			return dlqErr
		}
	}
	return nil
}

// MockWriter мокает WriterInterface для тестирования
type MockWriter struct {
	mock.Mock
}

func (m *MockWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockWriter) Stats() kafka.WriterStats {
	args := m.Called()
	return args.Get(0).(kafka.WriterStats)
}

func TestNewProducer(t *testing.T) {
	cfg := NewProducerConfig("localhost:9092")
	cfg.DLQTopic = "test-dlq"

	producer := NewProducer(cfg)

	assert.NotNil(t, producer)
	assert.NotNil(t, producer.writer)
	assert.NotNil(t, producer.dlqWriter)
}

func TestProducer_Produce_Success(t *testing.T) {
	mockWriter := &MockWriter{}
	mockDLQWriter := &MockWriter{}

	producer := &TestProducer{
		writer:    mockWriter,
		dlqWriter: mockDLQWriter,
	}

	ctx := context.Background()
	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Ожидаем успешную отправку
	mockWriter.On("WriteMessages", ctx, mock.AnythingOfType("[]kafka.Message")).Return(nil)

	err := producer.Produce(ctx, topic, key, value)

	assert.NoError(t, err)
	mockWriter.AssertExpectations(t)
}

func TestProducer_Produce_Error_WithDLQ(t *testing.T) {
	mockWriter := &MockWriter{}
	mockDLQWriter := &MockWriter{}

	producer := &TestProducer{
		writer:    mockWriter,
		dlqWriter: mockDLQWriter,
	}

	ctx := context.Background()
	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Симулируем ошибку при отправке в основную тему
	produceError := errors.New("kafka write error")
	mockWriter.On("WriteMessages", ctx, mock.AnythingOfType("[]kafka.Message")).Return(produceError)

	// Ожидаем успешную отправку в DLQ
	mockDLQWriter.On("WriteMessages", mock.Anything, mock.AnythingOfType("[]kafka.Message")).Return(nil)

	err := producer.Produce(ctx, topic, key, value)

	// Должна вернуться исходная ошибка
	assert.Error(t, err)
	assert.Equal(t, produceError, err)

	mockWriter.AssertExpectations(t)
	mockDLQWriter.AssertExpectations(t)
}

func TestProducer_Produce_Error_WithoutDLQ(t *testing.T) {
	mockWriter := &MockWriter{}

	producer := &TestProducer{
		writer:    mockWriter,
		dlqWriter: nil, // Отключаем DLQ
	}

	ctx := context.Background()
	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Симулируем ошибку при отправке
	produceError := errors.New("kafka write error")
	mockWriter.On("WriteMessages", ctx, mock.AnythingOfType("[]kafka.Message")).Return(produceError)

	err := producer.Produce(ctx, topic, key, value)

	assert.Error(t, err)
	assert.Equal(t, produceError, err)

	mockWriter.AssertExpectations(t)
}

func TestProducer_Produce_Error_DLQFailure(t *testing.T) {
	mockWriter := &MockWriter{}
	mockDLQWriter := &MockWriter{}

	producer := &TestProducer{
		writer:    mockWriter,
		dlqWriter: mockDLQWriter,
	}

	ctx := context.Background()
	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Симулируем ошибку при отправке в основную тему
	produceError := errors.New("kafka write error")
	mockWriter.On("WriteMessages", ctx, mock.AnythingOfType("[]kafka.Message")).Return(produceError)

	// Симулируем ошибку при отправке в DLQ
	dlqError := errors.New("dlq write error")
	mockDLQWriter.On("WriteMessages", mock.Anything, mock.AnythingOfType("[]kafka.Message")).Return(dlqError)

	err := producer.Produce(ctx, topic, key, value)

	// Должна вернуться ошибка с информацией о обеих неудачах
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write to topic test-topic and DLQ")

	mockWriter.AssertExpectations(t)
	mockDLQWriter.AssertExpectations(t)
}

func TestProducer_Close_Success(t *testing.T) {
	mockWriter := &MockWriter{}
	mockDLQWriter := &MockWriter{}

	producer := &TestProducer{
		writer:    mockWriter,
		dlqWriter: mockDLQWriter,
	}

	// Ожидаем успешное закрытие обоих писателей
	mockWriter.On("Close").Return(nil)
	mockDLQWriter.On("Close").Return(nil)

	err := producer.Close()

	assert.NoError(t, err)
	mockWriter.AssertExpectations(t)
	mockDLQWriter.AssertExpectations(t)
}

func TestProducer_Close_MainWriterError(t *testing.T) {
	mockWriter := &MockWriter{}
	mockDLQWriter := &MockWriter{}

	producer := &TestProducer{
		writer:    mockWriter,
		dlqWriter: mockDLQWriter,
	}

	// Симулируем ошибку при закрытии основного писателя
	closeError := errors.New("close error")
	mockWriter.On("Close").Return(closeError)

	err := producer.Close()

	assert.Error(t, err)
	assert.Equal(t, closeError, err)

	mockWriter.AssertExpectations(t)
	mockDLQWriter.AssertExpectations(t)
}

func TestProducer_Close_DLQWriterError(t *testing.T) {
	mockWriter := &MockWriter{}
	mockDLQWriter := &MockWriter{}

	producer := &TestProducer{
		writer:    mockWriter,
		dlqWriter: mockDLQWriter,
	}

	// Симулируем ошибку при закрытии DLQ писателя
	mockWriter.On("Close").Return(nil)
	dlqError := errors.New("dlq close error")
	mockDLQWriter.On("Close").Return(dlqError)

	err := producer.Close()

	assert.Error(t, err)
	assert.Equal(t, dlqError, err)

	mockWriter.AssertExpectations(t)
	mockDLQWriter.AssertExpectations(t)
}

func TestNewProducerConfig(t *testing.T) {
	brokerAddress := "localhost:9092"
	cfg := NewProducerConfig(brokerAddress)

	assert.Equal(t, brokerAddress, cfg.BrokerAddress)
	assert.NotNil(t, cfg.Balancer)
	assert.Equal(t, 10*time.Millisecond, cfg.BatchTimeout)
	assert.Equal(t, 100, cfg.BatchSize)
	assert.Equal(t, kafka.RequireAll, cfg.RequiredAcks)
	assert.Equal(t, 5, cfg.MaxAttempts)
	assert.Equal(t, kafka.Snappy, cfg.Compression)
	assert.False(t, cfg.Async)
	assert.Equal(t, "dead-letter-queue", cfg.DLQTopic)
}
