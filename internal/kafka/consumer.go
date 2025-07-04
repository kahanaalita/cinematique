package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer wraps a kafka.Reader for consuming messages.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(cfg ConsumerConfig) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.BrokerAddress},
		GroupID:  cfg.GroupID,
		Topic:    cfg.Topic,
		MinBytes: cfg.MinBytes,
		MaxBytes: cfg.MaxBytes,
		MaxWait:  cfg.MaxWait,
		// Добавлены другие важные настройки для стабильности
		CommitInterval:    0, // Ручной коммит
		HeartbeatInterval: 3 * time.Second,
		SessionTimeout:    30 * time.Second,
		RebalanceTimeout:  30 * time.Second,
	})
	return &Consumer{reader: reader}
}

// ConsumeMessages consumes messages from Kafka and logs them.
func (c *Consumer) ConsumeMessages(ctx context.Context) {
	log.Printf("Starting Kafka consumer for topic: %s, groupID: %s", c.reader.Config().Topic, c.reader.Config().GroupID)
	defer log.Printf("Stopping consumer for topic: %s", c.reader.Config().Topic)

	for {
		// FetchMessage будет автоматически обрабатывать переподключения.
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				// Контекст отменён, это ожидаемое завершение работы.
				break
			}
			// Библиотека сама будет пытаться переподключиться, поэтому здесь просто логируем ошибку.
			log.Printf("Error fetching message from Kafka: %v. Library will handle reconnect.", err)
			continue
		}

		log.Printf("Получено сообщение Kafka - Тема: %s, Раздел: %d, Смещение: %d, Ключ: %s, Значение: %s\n",
			m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("Ошибка при подтверждении сообщения в Kafka: %v", err)
		}
	}
}

// Close закрывает потребитель Kafka.
func (c *Consumer) Close() error {
	log.Printf("Closing Kafka reader for topic: %s", c.reader.Config().Topic)
	return c.reader.Close()
}
