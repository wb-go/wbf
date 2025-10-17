// Package kafka предоставляет клиенты для работы с Apache Kafka.
package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"

	"github.com/wb-go/wbf/retry"
)

// Producer представляет Kafka продюсер.
type Producer struct {
	Writer *kafka.Writer
}

// Consumer представляет Kafka консьюмер.
type Consumer struct {
	Reader *kafka.Reader
}

// NewProducer создает новый Kafka продюсер.
func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		Writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

// Send отправляет сообщение в Kafka.
func (p *Producer) Send(ctx context.Context, key, value []byte) error {
	return p.Writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

// Close закрывает соединение с Kafka.
func (p *Producer) Close() error {
	return p.Writer.Close()
}

// SendWithRetry отправляет сообщение с стратегией повторных попыток.
func (p *Producer) SendWithRetry(ctx context.Context, strategy retry.Strategy, key, value []byte) error {
	return retry.Do(func() error {
		return p.Send(ctx, key, value)
	}, strategy)
}

// NewConsumer создает новый Kafka консьюмер.
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

// Fetch получает сообщение из Kafka.
func (c *Consumer) Fetch(ctx context.Context) (kafka.Message, error) {
	return c.Reader.FetchMessage(ctx)
}

// Commit подтверждает обработку сообщения.
func (c *Consumer) Commit(ctx context.Context, msg kafka.Message) error {
	return c.Reader.CommitMessages(ctx, msg)
}

// Close закрывает соединение с Kafka.
func (c *Consumer) Close() error {
	return c.Reader.Close()
}

// FetchWithRetry получает сообщение с стратегией повторных попыток.
func (c *Consumer) FetchWithRetry(ctx context.Context, strategy retry.Strategy) (kafka.Message, error) {
	var msg kafka.Message
	err := retry.Do(func() error {
		m, e := c.Fetch(ctx)
		if e == nil {
			msg = m
		}
		return e
	}, strategy)
	return msg, err
}

// StartConsuming запускает процесс потребления сообщений.
func (c *Consumer) StartConsuming(ctx context.Context, out chan<- kafka.Message, strategy retry.Strategy) {
	go func() {
		defer close(out)
		for {
			msg, err := c.FetchWithRetry(ctx, strategy)
			if err != nil {
				// Можно добавить логирование ошибки
				break
			}
			select {
			case out <- msg:
			case <-ctx.Done():
				return
			}
		}
	}()
}
