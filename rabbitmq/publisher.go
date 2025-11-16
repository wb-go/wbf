package rabbitmq

import (
	"context"

	"github.com/rabbitmq/amqp091-go"

	"github.com/wb-go/wbf/retry"
)

type Publisher struct {
	client      *RabbitClient
	exchange    string
	contentType string
}

func NewPublisher(client *RabbitClient, exchange, contentType string) *Publisher {
	return &Publisher{
		client:      client,
		exchange:    exchange,
		contentType: contentType,
	}
}

func (p *Publisher) Publish(
	ctx context.Context,
	body []byte,
	routingKey string,
	opts ...PublishOption,
) error {
	return retry.DoContext(ctx, p.client.config.PublishRetry, func() error {
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

		err = ch.PublishWithContext(ctx, p.exchange, routingKey, false, false, pub) //mandatory и immediate не используются практически пока так
		if err != nil {
			return err
		}

		return nil
	})
}
