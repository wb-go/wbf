// Package rabbitmq это обертка над github.com/rabbitmq/amqp091-go
package rabbitmq

import (
	"context"
	"fmt"
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

// WithExpiration - опция для указания время истечения.
func WithExpiration(d time.Duration) PublishOption {
	return func(p *amqp091.Publishing) {
		if d > 0 {
			p.Expiration = fmt.Sprintf("%d", d.Milliseconds())
		}
	}
}

// WithHeaders - опция для указания заголовков.
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

// AskConfig - настройки Ask.
type AskConfig struct {
	Multiple bool
}

// NackConfig - настройки Nack.
type NackConfig struct {
	Multiple bool
	Requeue  bool
}
