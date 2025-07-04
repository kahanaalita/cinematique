package kafka

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var ErrBufferFull = errors.New("producer pool buffer is full")

// Интерфейс для продюсера
// ProducerInterface описывает методы для отправки сообщений и закрытия продюсера
// Используется для моков и реальных реализаций
type ProducerInterface interface {
	Produce(ctx context.Context, topic string, key, value []byte) error
	Close() error
}

// KafkaEvent описывает событие для отправки в Kafka
type KafkaEvent struct {
	Topic string
	Key   []byte
	Value []byte
}

// Метрики для мониторинга
var (
	KafkaProduceErrorsTotal    = prometheus.NewCounter(prometheus.CounterOpts{Name: "kafka_produce_errors_total", Help: "Total number of Kafka produce errors."})
	KafkaMessagesProducedTotal = prometheus.NewCounter(prometheus.CounterOpts{Name: "kafka_messages_produced_total", Help: "Total number of Kafka messages produced."})
	KafkaMessagesDroppedTotal  = prometheus.NewCounter(prometheus.CounterOpts{Name: "kafka_messages_dropped_total", Help: "Total number of Kafka messages dropped due to buffer full."})
)

func init() {
	prometheus.MustRegister(KafkaProduceErrorsTotal)
	prometheus.MustRegister(KafkaMessagesProducedTotal)
	prometheus.MustRegister(KafkaMessagesDroppedTotal)
}

type ProducerPool struct {
	producer ProducerInterface
	events   chan KafkaEvent
	wg       sync.WaitGroup
}

func NewProducerPool(producer ProducerInterface, workers, bufSize int) *ProducerPool {
	pool := &ProducerPool{
		producer: producer,
		events:   make(chan KafkaEvent, bufSize),
	}
	pool.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go pool.worker()
	}
	return pool
}

func (p *ProducerPool) worker() {
	defer p.wg.Done()
	for event := range p.events {
		// Используем встроенный в продюсер механизм ретраев и DLQ
		if err := p.producer.Produce(context.Background(), event.Topic, event.Key, event.Value); err != nil {
			// Ошибка уже залогирована в самом продюсере, здесь достаточно метрики
			KafkaProduceErrorsTotal.Inc()
		} else {
			KafkaMessagesProducedTotal.Inc()
		}
	}
}

func (p *ProducerPool) Produce(topic string, key, value []byte) error {
	select {
	case p.events <- KafkaEvent{Topic: topic, Key: key, Value: value}:
		return nil
	default:
		KafkaMessagesDroppedTotal.Inc()
		log.Println("failed to queue message: buffer is full")
		return ErrBufferFull
	}
}

func (p *ProducerPool) Close() {
	log.Println("Closing producer pool...")
	close(p.events) // Закрываем канал, чтобы воркеры завершили работу после обработки оставшихся событий
	p.wg.Wait()     // Ждем, пока все воркеры закончат

	if err := p.producer.Close(); err != nil {
		log.Printf("Error closing producer: %v", err)
	}
	log.Println("Producer pool closed.")
}
