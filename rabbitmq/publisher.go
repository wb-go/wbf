package rabbitmq

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/retry"
)

// Publisher - обертка над RabbitMQ-клиентом для публикации сообщений в обменник.
type Publisher struct {
	client      *RabbitClient
	exchange    string
	contentType string
}

// NewPublisher конструктор Publisher.
func NewPublisher(client *RabbitClient, exchange, contentType string) *Publisher {
	return &Publisher{
		client:      client,
		exchange:    exchange,
		contentType: contentType,
	}
}

// GetExchangeName возвращает имя обменника, который использует publisher.
func (p *Publisher) GetExchangeName() string {
	return p.exchange
}

// Publish отправляет сообщение в указанный exchange с заданным routing key.
// Использует стратегию повторных попыток (ProducingStrat) при ошибках.
// Автоматически управляет временными каналами и применяет дополнительные опции публикации.
func (p *Publisher) Publish(
	ctx context.Context,
	body []byte,
	routingKey string,
	opts ...PublishOption,
) error {
	return retry.DoContext(ctx, p.client.config.ProducingStrat, func() error {
		ch, err := p.client.GetChannel()
		if err != nil {
			return err
		}
		defer func(ch *amqp091.Channel) {
			_ = ch.Close()
		}(ch)

		pub := amqp091.Publishing{
			ContentType: p.contentType,
			Body:        body,
		}

		for _, opt := range opts {
			opt(&pub)
		}
		// mandatory и immediate не используются практически пока так.
		err = ch.PublishWithContext(ctx, p.exchange, routingKey, false, false, pub)
		if err != nil {
			return err
		}

		return nil
	})
}
