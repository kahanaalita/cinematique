package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

// ProducerConfig содержит настройки для Kafka-продюсера.
type ProducerConfig struct {
	BrokerAddress string
	Balancer      kafka.Balancer
	BatchTimeout  time.Duration
	BatchSize     int
	RequiredAcks  kafka.RequiredAcks
	MaxAttempts   int
	Compression   kafka.Compression
	Async         bool
	DLQTopic      string // Топик для Dead Letter Queue
}

// NewProducerConfig создает конфиг продюсера со значениями по умолчанию.
func NewProducerConfig(brokerAddress string) ProducerConfig {
	return ProducerConfig{
		BrokerAddress: brokerAddress,
		Balancer:      &kafka.LeastBytes{},
		BatchTimeout:  10 * time.Millisecond,
		BatchSize:     100,
		RequiredAcks:  kafka.RequireAll,
		MaxAttempts:   5,
		Compression:   kafka.Snappy,
		Async:         false, // По умолчанию синхронный режим для большей надежности
		DLQTopic:      "dead-letter-queue",
	}
}

// ConsumerConfig содержит настройки для Kafka-консьюмера.
type ConsumerConfig struct {
	BrokerAddress string
	GroupID       string
	Topic         string
	MaxBytes      int
	MinBytes      int
	MaxWait       time.Duration
}

// NewConsumerConfig создает конфиг консьюмера со значениями по умолчанию.
func NewConsumerConfig(brokerAddress, groupID, topic string) ConsumerConfig {
	return ConsumerConfig{
		BrokerAddress: brokerAddress,
		GroupID:       groupID,
		Topic:         topic,
		MinBytes:      10e3, // 10KB
		MaxBytes:      10e6, // 10MB
		MaxWait:       1 * time.Second,
	}
}
