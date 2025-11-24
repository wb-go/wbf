package rabbitmq

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/retry"
)

// ClientConfig — конфигурация клиента RabbitMQ.
type ClientConfig struct {
	URL            string
	ConnectionName string // для идентификации в RabbitMQ UI
	ConnectTimeout time.Duration
	Heartbeat      time.Duration
	PublishRetry   retry.Strategy
	ConsumeRetry   retry.Strategy
}

// PublishOption — функциональная опция для публикации.
type PublishOption func(*amqp091.Publishing)

func WithExpiration(d time.Duration) PublishOption {
	return func(p *amqp091.Publishing) {
		if d > 0 {
			p.Expiration = d.Truncate(time.Millisecond).String()
		}
	}
}

func WithHeaders(headers amqp091.Table) PublishOption {
	return func(p *amqp091.Publishing) {
		p.Headers = headers
	}
}

// MessageHandler обрабатывает сообщение. Возвращает ошибку → NACK, nil → ACK.
type MessageHandler func(context.Context, amqp091.Delivery) error

// ConsumerConfig — конфигурация потребителя.
type ConsumerConfig struct {
	Queue         string
	ConsumerTag   string
	AutoAck       bool
	Ask           AskConfig
	Nack          NackConfig
	Args          amqp091.Table
	Workers       int
	PrefetchCount int
}

type AskConfig struct {
	Multiple bool
}

type NackConfig struct {
	Multiple bool
	Requeue  bool
}
