package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

// Producer wraps a kafka.Writer for sending messages.
type Producer struct {
	writer   *kafka.Writer
	dlqWriter *kafka.Writer // Опциональный writer для DLQ
}

// NewProducer создаёт новый Kafka-продюсер.
func NewProducer(cfg ProducerConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.BrokerAddress),
		Balancer:     cfg.Balancer,
		BatchTimeout: cfg.BatchTimeout,
		BatchSize:    cfg.BatchSize,
		RequiredAcks: cfg.RequiredAcks,
		MaxAttempts:  cfg.MaxAttempts,
		Compression:  cfg.Compression,
		Async:        cfg.Async,
	}

	var dlqWriter *kafka.Writer
	if cfg.DLQTopic != "" {
		dlqWriter = &kafka.Writer{
			Addr:        kafka.TCP(cfg.BrokerAddress),
			Topic:       cfg.DLQTopic,
			Balancer:    &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
		}
	}

	return &Producer{writer: writer, dlqWriter: dlqWriter}
}

// Produce отправляет сообщение в указанную Kafka-тему.
// Если после всех попыток сообщение не удается доставить, оно отправляется в DLQ (если настроено).
func (p *Producer) Produce(ctx context.Context, topic string, key, value []byte) error {
	message := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	err := p.writer.WriteMessages(ctx, message)
	if err != nil {
		log.Printf("Failed to write message to topic %s after multiple retries: %v", topic, err)
		// Попытка отправить в DLQ
		if p.dlqWriter != nil {
			dlqMessage := kafka.Message{
				// DLQ сообщение содержит исходное сообщение и причину ошибки
				Key:   key,
				Value: []byte(fmt.Sprintf("original_topic: %s, error: %v, message: %s", topic, err, string(value))),
			}
			if dlqErr := p.dlqWriter.WriteMessages(context.Background(), dlqMessage); dlqErr != nil {
				log.Printf("CRITICAL: Failed to write message to DLQ topic %s: %v", p.dlqWriter.Stats().Topic, dlqErr)
				return fmt.Errorf("failed to write to topic %s and DLQ: %w", topic, err) // Возвращаем обе ошибки
			}
			log.Printf("Message successfully sent to DLQ topic: %s", p.dlqWriter.Stats().Topic)
		}
		return err // Возвращаем исходную ошибку
	}
	return nil
}

// Close closes the Kafka producer and the DLQ writer.
func (p *Producer) Close() error {
	log.Println("Closing main Kafka writer...")
	err := p.writer.Close()
	if err != nil {
		log.Printf("Error closing main Kafka writer: %v", err)
	}

	if p.dlqWriter != nil {
		log.Println("Closing DLQ Kafka writer...")
		if dlqErr := p.dlqWriter.Close(); dlqErr != nil {
			log.Printf("Error closing DLQ Kafka writer: %v", dlqErr)
			// Возвращаем первую ошибку, если обе не nil
			if err == nil {
				return dlqErr
			}
		}
	}
	return err
}
