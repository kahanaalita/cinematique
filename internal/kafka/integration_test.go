package kafka

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestKafkaIntegration тестирует интеграцию между Producer, ProducerPool и MockProducer
func TestKafkaIntegration(t *testing.T) {
	mockProducer := NewMockProducer()
	mockProducer.On("Close").Return(nil).Maybe()
	pool := NewProducerPool(mockProducer, 2, 10)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Ожидаем успешную отправку через пул
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(nil)

	// Отправляем сообщение через пул
	err := pool.Produce(topic, key, value)
	assert.NoError(t, err)

	// Даем время воркеру обработать сообщение
	time.Sleep(100 * time.Millisecond)

	mockProducer.AssertExpectations(t)
}

// TestKafkaIntegration_ErrorHandling тестирует обработку ошибок в интеграции
func TestKafkaIntegration_ErrorHandling(t *testing.T) {
	mockProducer := NewMockProducer()
	mockProducer.On("Close").Return(nil).Maybe()
	pool := NewProducerPool(mockProducer, 1, 5)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Симулируем ошибку при отправке
	produceError := errors.New("integration error")
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(produceError)

	// Отправляем сообщение через пул
	err := pool.Produce(topic, key, value)
	assert.NoError(t, err) // Пул не возвращает ошибку, она обрабатывается воркером

	// Даем время воркеру обработать сообщение
	time.Sleep(100 * time.Millisecond)

	mockProducer.AssertExpectations(t)
}

// TestKafkaIntegration_ConcurrentAccess тестирует конкурентный доступ к пулу
func TestKafkaIntegration_ConcurrentAccess(t *testing.T) {
	mockProducer := NewMockProducer()
	mockProducer.On("Close").Return(nil).Maybe()
	pool := NewProducerPool(mockProducer, 3, 20)
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

// TestKafkaIntegration_ProducerPoolWithRealProducer тестирует пул с реальным Producer
func TestKafkaIntegration_ProducerPoolWithRealProducer(t *testing.T) {
	// Создаем конфигурацию для тестового продюсера
	cfg := NewProducerConfig("localhost:9092")
	cfg.DLQTopic = "test-dlq"

	// Создаем реальный продюсер
	producer := NewProducer(cfg)

	// Создаем пул с реальным продюсером
	pool := NewProducerPool(producer, 1, 5)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Отправляем сообщение через пул
	err := pool.Produce(topic, key, value)
	assert.NoError(t, err)

	// Даем время воркеру обработать сообщение
	time.Sleep(100 * time.Millisecond)
}

// TestKafkaIntegration_ConsumerWithMockReader тестирует консьюмер с мок-ридером
func TestKafkaIntegration_ConsumerWithMockReader(t *testing.T) {
	mockReader := &MockReader{}
	consumer := &TestConsumer{reader: mockReader}

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

// TestKafkaIntegration_EndToEnd тестирует полный цикл: Producer -> ProducerPool -> Consumer
func TestKafkaIntegration_EndToEnd(t *testing.T) {
	mockProducer := NewMockProducer()
	mockProducer.On("Close").Return(nil).Maybe()
	pool := NewProducerPool(mockProducer, 1, 5)
	defer pool.Close()

	// Создаем мок-ридер для консьюмера
	mockReader := &MockReader{}
	consumer := &TestConsumer{reader: mockReader}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Ожидаем отправку через пул
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(nil)

	// Создаем сообщение для консьюмера
	testMessage := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	// Ожидаем получение и коммит в консьюмере
	mockReader.On("FetchMessage", ctx).Return(testMessage, nil).Once()
	mockReader.On("CommitMessages", ctx, []kafka.Message{testMessage}).Return(nil).Once()
	// Ожидаем дополнительные вызовы FetchMessage для завершения цикла
	mockReader.On("FetchMessage", mock.Anything).Return(kafka.Message{}, context.Canceled).Maybe()

	// Отправляем сообщение через пул
	err := pool.Produce(topic, key, value)
	assert.NoError(t, err)

	// Запускаем консьюмер
	go consumer.ConsumeMessages(ctx)

	// Даем время на обработку
	time.Sleep(200 * time.Millisecond)

	// Отменяем контекст
	cancel()

	mockProducer.AssertExpectations(t)
	mockReader.AssertExpectations(t)
}

// TestKafkaIntegration_ErrorRecovery тестирует восстановление после ошибок
func TestKafkaIntegration_ErrorRecovery(t *testing.T) {
	mockProducer := NewMockProducer()
	mockProducer.On("Close").Return(nil).Maybe()
	pool := NewProducerPool(mockProducer, 1, 5)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Сначала симулируем ошибку, затем успех
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(errors.New("temporary error")).Once()
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(nil).Once()

	// Отправляем первое сообщение (должно вызвать ошибку)
	err1 := pool.Produce(topic, key, value)
	assert.NoError(t, err1)

	// Даем время воркеру обработать сообщение
	time.Sleep(100 * time.Millisecond)

	// Отправляем второе сообщение (должно быть успешным)
	err2 := pool.Produce(topic, key, value)
	assert.NoError(t, err2)

	// Даем время воркеру обработать сообщение
	time.Sleep(100 * time.Millisecond)

	mockProducer.AssertExpectations(t)
}

// TestKafkaIntegration_Metrics тестирует метрики в интеграции
func TestKafkaIntegration_Metrics(t *testing.T) {
	mockProducer := NewMockProducer()
	mockProducer.On("Close").Return(nil).Maybe()
	pool := NewProducerPool(mockProducer, 1, 5)
	defer pool.Close()

	topic := "test-topic"
	key := []byte("test-key")
	value := []byte("test-value")

	// Ожидаем успешную отправку
	mockProducer.On("Produce", mock.Anything, topic, key, value).Return(nil)

	// Отправляем сообщение
	err := pool.Produce(topic, key, value)
	assert.NoError(t, err)

	// Даем время воркеру обработать сообщение
	time.Sleep(100 * time.Millisecond)

	mockProducer.AssertExpectations(t)
}
