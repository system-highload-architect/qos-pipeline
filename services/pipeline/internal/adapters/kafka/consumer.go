package kafka

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer читает метрики из Kafka.
type Consumer struct {
	reader *kafka.Reader
	logger *slog.Logger
}

// NewConsumer создаёт нового Consumer'а.
func NewConsumer(brokers []string, topic, groupID string, logger *slog.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafka.LastOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	return &Consumer{reader: reader, logger: logger}
}

// Start запускает чтение сообщений в бесконечном цикле.
// Отправляет данные в канал `out`. Завершается при отмене ctx.
func (c *Consumer) Start(ctx context.Context, out chan<- []byte) {
	go func() {
		for {
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.logger.Error("fetch message error", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}
			select {
			case out <- msg.Value:
			case <-ctx.Done():
				return
			}
			// Коммит офсета
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("commit message error", "error", err)
			}
		}
	}()
}

// Close закрывает Reader.
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// MockConsumer — заглушка для тестов, если Kafka недоступна.
type MockConsumer struct {
	Logger *slog.Logger
	Data   [][]byte
}

func (m *MockConsumer) Start(ctx context.Context, out chan<- []byte) {
	go func() {
		for _, d := range m.Data {
			out <- d
			time.Sleep(500 * time.Millisecond)
		}
	}()
}
func (m *MockConsumer) Close() error { return nil }
