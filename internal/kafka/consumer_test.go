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

// ReaderInterface определяет интерфейс для kafka.Reader
type ReaderInterface interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
	Config() kafka.ReaderConfig
}

// TestConsumer для тестирования с мок-ридером
type TestConsumer struct {
	reader ReaderInterface
}

func (c *TestConsumer) ConsumeMessages(ctx context.Context) {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			// Логируем ошибку коммита
		}
	}
}

func (c *TestConsumer) Close() error {
	return c.reader.Close()
}

// MockReader мокает ReaderInterface для тестирования
type MockReader struct {
	mock.Mock
}

func (m *MockReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return kafka.Message{}, args.Error(1)
	}
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *MockReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockReader) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockReader) Config() kafka.ReaderConfig {
	args := m.Called()
	return args.Get(0).(kafka.ReaderConfig)
}

func TestNewConsumer(t *testing.T) {
	cfg := NewConsumerConfig("localhost:9092", "test-group", "test-topic")

	consumer := NewConsumer(cfg)

	assert.NotNil(t, consumer)
	assert.NotNil(t, consumer.reader)
}

func TestConsumer_ConsumeMessages_Success(t *testing.T) {
	mockReader := &MockReader{}

	consumer := &TestConsumer{
		reader: mockReader,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем тестовое сообщение
	testMessage := kafka.Message{
		Topic:     "test-topic",
		Partition: 0,
		Offset:    123,
		Key:       []byte("test-key"),
		Value:     []byte("test-value"),
	}

	// Ожидаем успешное получение и коммит сообщения
	mockReader.On("FetchMessage", ctx).Return(testMessage, nil).Once()
	mockReader.On("CommitMessages", ctx, []kafka.Message{testMessage}).Return(nil).Once()
	// Ожидаем дополнительные вызовы FetchMessage для завершения цикла
	mockReader.On("FetchMessage", mock.Anything).Return(kafka.Message{}, context.Canceled).Maybe()

	// Запускаем потребление в горутине
	go consumer.ConsumeMessages(ctx)

	// Даем время на обработку
	time.Sleep(100 * time.Millisecond)

	// Отменяем контекст для завершения
	cancel()

	mockReader.AssertExpectations(t)
}

func TestConsumer_ConsumeMessages_FetchError(t *testing.T) {
	mockReader := &MockReader{}

	consumer := &TestConsumer{
		reader: mockReader,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Симулируем ошибку при получении сообщения
	fetchError := errors.New("fetch error")
	mockReader.On("FetchMessage", ctx).Return(kafka.Message{}, fetchError).Once()
	// Ожидаем дополнительные вызовы FetchMessage для завершения цикла
	mockReader.On("FetchMessage", mock.Anything).Return(kafka.Message{}, context.Canceled).Maybe()

	// Запускаем потребление в горутине
	go consumer.ConsumeMessages(ctx)

	// Даем время на обработку
	time.Sleep(100 * time.Millisecond)

	// Отменяем контекст для завершения
	cancel()

	mockReader.AssertExpectations(t)
}

func TestConsumer_ConsumeMessages_CommitError(t *testing.T) {
	mockReader := &MockReader{}

	consumer := &TestConsumer{
		reader: mockReader,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем тестовое сообщение
	testMessage := kafka.Message{
		Topic:     "test-topic",
		Partition: 0,
		Offset:    123,
		Key:       []byte("test-key"),
		Value:     []byte("test-value"),
	}

	// Ожидаем успешное получение, но ошибку при коммите
	mockReader.On("FetchMessage", ctx).Return(testMessage, nil).Once()
	mockReader.On("CommitMessages", ctx, []kafka.Message{testMessage}).Return(errors.New("commit error")).Once()
	// Ожидаем дополнительные вызовы FetchMessage для завершения цикла
	mockReader.On("FetchMessage", mock.Anything).Return(kafka.Message{}, context.Canceled).Maybe()

	// Запускаем потребление в горутине
	go consumer.ConsumeMessages(ctx)

	// Даем время на обработку
	time.Sleep(100 * time.Millisecond)

	// Отменяем контекст для завершения
	cancel()

	mockReader.AssertExpectations(t)
}

func TestConsumer_ConsumeMessages_ContextCancellation(t *testing.T) {
	mockReader := &MockReader{}

	consumer := &TestConsumer{
		reader: mockReader,
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Симулируем отмену контекста
	mockReader.On("FetchMessage", ctx).Return(kafka.Message{}, context.Canceled).Once()
	// Ожидаем дополнительные вызовы FetchMessage для завершения цикла
	mockReader.On("FetchMessage", mock.Anything).Return(kafka.Message{}, context.Canceled).Maybe()

	// Запускаем потребление в горутине
	go consumer.ConsumeMessages(ctx)

	// Даем время на обработку
	time.Sleep(100 * time.Millisecond)

	// Отменяем контекст
	cancel()

	mockReader.AssertExpectations(t)
}

func TestConsumer_Close(t *testing.T) {
	mockReader := &MockReader{}

	consumer := &TestConsumer{
		reader: mockReader,
	}

	// Ожидаем успешное закрытие
	mockReader.On("Close").Return(nil)

	err := consumer.Close()

	assert.NoError(t, err)
	mockReader.AssertExpectations(t)
}

func TestConsumer_Close_Error(t *testing.T) {
	mockReader := &MockReader{}

	consumer := &TestConsumer{
		reader: mockReader,
	}

	// Симулируем ошибку при закрытии
	closeError := errors.New("close error")
	mockReader.On("Close").Return(closeError)

	err := consumer.Close()

	assert.Error(t, err)
	assert.Equal(t, closeError, err)
	mockReader.AssertExpectations(t)
}

func TestNewConsumerConfig(t *testing.T) {
	brokerAddress := "localhost:9092"
	groupID := "test-group"
	topic := "test-topic"

	cfg := NewConsumerConfig(brokerAddress, groupID, topic)

	assert.Equal(t, brokerAddress, cfg.BrokerAddress)
	assert.Equal(t, groupID, cfg.GroupID)
	assert.Equal(t, topic, cfg.Topic)
	assert.Equal(t, int(10e3), cfg.MinBytes) // 10KB
	assert.Equal(t, int(10e6), cfg.MaxBytes) // 10MB
	assert.Equal(t, 1*time.Second, cfg.MaxWait)
}
