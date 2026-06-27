package kafka

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

// Publisher отправляет сообщения в Kafka.
type Publisher struct {
	writer *kafka.Writer
	logger *slog.Logger
}

// NewPublisher создаёт нового Publisher'а.
func NewPublisher(brokers []string, topic string, logger *slog.Logger) (*Publisher, error) {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &Publisher{writer: w, logger: logger}, nil
}

// Publish отправляет сообщение в Kafka.
func (p *Publisher) Publish(ctx context.Context, key, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

// Close закрывает Writer.
func (p *Publisher) Close() error {
	return p.writer.Close()
}

type MockPublisher struct {
	Logger *slog.Logger
}

func (m *MockPublisher) Publish(ctx context.Context, key, value []byte) error {
	m.Logger.Info("mock publish", "key", string(key), "value", string(value))
	return nil
}

func (m *MockPublisher) Close() error {
	return nil
}
