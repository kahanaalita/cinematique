# Тестирование Kafka компонентов

Этот документ описывает подход к тестированию Kafka функциональности в проекте Cinematique.

## Структура тестов

### Unit-тесты

1. **`producer_test.go`** - тесты для Kafka Producer
   - Тестирование успешной отправки сообщений
   - Тестирование обработки ошибок
   - Тестирование DLQ (Dead Letter Queue)
   - Тестирование закрытия соединений

2. **`consumer_test.go`** - тесты для Kafka Consumer
   - Тестирование потребления сообщений
   - Тестирование обработки ошибок получения
   - Тестирование ошибок коммита
   - Тестирование отмены контекста

3. **`producer_pool_test.go`** - тесты для ProducerPool
   - Тестирование создания пула
   - Тестирование отправки сообщений через пул
   - Тестирование переполнения буфера
   - Тестирование конкурентного доступа
   - Тестирование метрик

4. **`mock_producer_test.go`** - тесты для MockProducer
   - Тестирование мок-реализации
   - Тестирование интерфейса
   - Тестирование интеграции с пулом

### Интеграционные тесты

5. **`integration_test.go`** - интеграционные тесты
   - Тестирование взаимодействия между компонентами
   - End-to-end тестирование
   - Тестирование восстановления после ошибок
   - Тестирование метрик в интеграции

## Запуск тестов

### Запуск всех тестов Kafka
```bash
go test ./internal/kafka/...
```

### Запуск конкретного файла тестов
```bash
go test ./internal/kafka/producer_test.go
go test ./internal/kafka/consumer_test.go
go test ./internal/kafka/producer_pool_test.go
go test ./internal/kafka/mock_producer_test.go
go test ./internal/kafka/integration_test.go
```

### Запуск с подробным выводом
```bash
go test -v ./internal/kafka/...
```

### Запуск с покрытием кода
```bash
go test -cover ./internal/kafka/...
go test -coverprofile=coverage.out ./internal/kafka/...
go tool cover -html=coverage.out
```

### Запуск конкретного теста
```bash
go test -run TestProducer_Produce_Success ./internal/kafka/
go test -run TestProducerPool_ConcurrentProduce ./internal/kafka/
```

## Подходы к тестированию

### 1. Моки и интерфейсы

Мы используем интерфейсы для абстракции зависимостей:

```go
type WriterInterface interface {
    WriteMessages(ctx context.Context, msgs ...kafka.Message) error
    Close() error
    Stats() kafka.WriterStats
}

type ProducerInterface interface {
    Produce(ctx context.Context, topic string, key, value []byte) error
    Close() error
}
```

### 2. TestProducer и TestConsumer

Созданы тестовые версии Producer и Consumer, которые принимают интерфейсы вместо конкретных реализаций:

```go
type TestProducer struct {
    writer    WriterInterface
    dlqWriter WriterInterface
}

type TestConsumer struct {
    reader ReaderInterface
}
```

### 3. MockWriter и MockReader

Моки для симуляции поведения Kafka Writer и Reader:

```go
type MockWriter struct {
    mock.Mock
}

type MockReader struct {
    mock.Mock
}
```

## Сценарии тестирования

### Producer тесты

1. **Успешная отправка** - проверяет корректную отправку сообщения
2. **Ошибка отправки с DLQ** - проверяет отправку в Dead Letter Queue при ошибке
3. **Ошибка отправки без DLQ** - проверяет поведение без настроенного DLQ
4. **Ошибка DLQ** - проверяет обработку ошибки при отправке в DLQ
5. **Закрытие соединений** - проверяет корректное закрытие всех соединений

### Consumer тесты

1. **Успешное потребление** - проверяет получение и коммит сообщения
2. **Ошибка получения** - проверяет обработку ошибок FetchMessage
3. **Ошибка коммита** - проверяет обработку ошибок CommitMessages
4. **Отмена контекста** - проверяет корректное завершение при отмене контекста

### ProducerPool тесты

1. **Создание пула** - проверяет инициализацию пула и воркеров
2. **Отправка через пул** - проверяет отправку сообщений через пул
3. **Переполнение буфера** - проверяет поведение при заполнении буфера
4. **Конкурентный доступ** - проверяет работу с несколькими горутинами
5. **Метрики** - проверяет обновление Prometheus метрик

### Интеграционные тесты

1. **End-to-end** - тестирует полный цикл Producer -> ProducerPool -> Consumer
2. **Обработка ошибок** - тестирует восстановление после ошибок
3. **Конкурентность** - тестирует работу под нагрузкой
4. **Метрики** - тестирует метрики в реальных сценариях

## Метрики

Тесты проверяют корректность обновления Prometheus метрик:

- `kafka_messages_produced_total` - количество отправленных сообщений
- `kafka_produce_errors_total` - количество ошибок отправки
- `kafka_messages_dropped_total` - количество отброшенных сообщений

## Лучшие практики

1. **Используйте моки** для изоляции тестов от внешних зависимостей
2. **Тестируйте граничные случаи** - ошибки, переполнения, отмены
3. **Проверяйте метрики** - убедитесь, что метрики обновляются корректно
4. **Используйте таймауты** - для асинхронных операций
5. **Тестируйте конкурентность** - проверяйте работу с несколькими горутинами

## Примеры использования

### Тестирование Producer с моками

```go
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
    
    mockWriter.On("WriteMessages", ctx, mock.AnythingOfType("[]kafka.Message")).Return(nil)
    
    err := producer.Produce(ctx, topic, key, value)
    
    assert.NoError(t, err)
    mockWriter.AssertExpectations(t)
}
```

### Тестирование ProducerPool

```go
func TestProducerPool_Produce_Success(t *testing.T) {
    mockProducer := &MockProducerInterface{}
    pool := NewProducerPool(mockProducer, 1, 5)
    defer pool.Close()
    
    topic := "test-topic"
    key := []byte("test-key")
    value := []byte("test-value")
    
    mockProducer.On("Produce", mock.Anything, topic, key, value).Return(nil)
    
    err := pool.Produce(topic, key, value)
    assert.NoError(t, err)
    
    time.Sleep(100 * time.Millisecond)
    mockProducer.AssertExpectations(t)
}
```

## Отладка тестов

### Включение логирования

```bash
go test -v -logtostderr ./internal/kafka/...
```

### Параллельное выполнение

```bash
go test -parallel 4 ./internal/kafka/...
```

### Таймауты

```bash
go test -timeout 30s ./internal/kafka/...
```

## CI/CD интеграция

Тесты автоматически запускаются в CI/CD пайплайне:

```yaml
- name: Run Kafka tests
  run: go test -v -race -cover ./internal/kafka/...
```

## Мониторинг покрытия

Регулярно проверяйте покрытие кода тестами:

```bash
go test -coverprofile=kafka_coverage.out ./internal/kafka/...
go tool cover -func=kafka_coverage.out
```

Цель - достичь покрытия не менее 80% для критических компонентов. 