package kafka

import (
	"context"

	"github.com/pozedorum/wbf/retry"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	Writer *kafka.Writer
}

type Consumer struct {
	Reader *kafka.Reader
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		Writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Send(ctx context.Context, key, value []byte) error {
	return p.Writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

func (p *Producer) Close() error {
	return p.Writer.Close()
}

func (p *Producer) SendWithRetry(ctx context.Context, strat retry.Strategy, key, value []byte) error {
	return retry.Do(func() error {
		return p.Send(ctx, key, value)
	}, strat)
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (c *Consumer) Fetch(ctx context.Context) (kafka.Message, error) {
	return c.Reader.FetchMessage(ctx)
}

func (c *Consumer) Commit(ctx context.Context, msg kafka.Message) error {
	return c.Reader.CommitMessages(ctx, msg)
}

func (c *Consumer) Close() error {
	return c.Reader.Close()
}

func (c *Consumer) FetchWithRetry(ctx context.Context, strat retry.Strategy) (kafka.Message, error) {
	var msg kafka.Message
	err := retry.Do(func() error {
		m, e := c.Fetch(ctx)
		if e == nil {
			msg = m
		}
		return e
	}, strat)
	return msg, err
}

func (c *Consumer) StartConsuming(ctx context.Context, out chan<- kafka.Message, strat retry.Strategy) {
	go func() {
		defer close(out)
		for {
			msg, err := c.FetchWithRetry(ctx, strat)
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
