package rabbitmq

import (
	"context"
	"errors"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"github.com/wb-go/wbf/zlog"
)

type Consumer struct {
	client  *RabbitClient
	config  ConsumerConfig
	handler MessageHandler
}

func NewConsumer(client *RabbitClient, cfg ConsumerConfig, handler MessageHandler) *Consumer {
	if cfg.ConsumerTag == "" {
		cfg.ConsumerTag = "consumer"
	}
	return &Consumer{
		client:  client,
		config:  cfg,
		handler: handler,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	fmt.Println("start consumer " + c.config.ConsumerTag)
	for {
		if err := c.consumeOnce(ctx); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			select {
			case <-ctx.Done():
				return nil
			case <-c.client.Context().Done():
				return nil
			default:
			}
			// Повтор с задержкой
			if !c.client.backoffWait(ctx, c.client.config.ConsumeRetry.Delay) {
				return nil
			}
		} else {
			return nil
		}
	}
}

func (c *Consumer) consumeOnce(ctx context.Context) error {
	ch, err := c.client.GetChannel()
	if err != nil {
		return err
	}
	defer func(ch *amqp091.Channel) {
		_ = ch.Close()
	}(ch)

	msgs, err := ch.Consume(
		c.config.Queue,
		c.config.ConsumerTag,
		c.config.AutoAck,
		false, // эксклюзивность 1 или несколько слушателей
		false,
		false, // ждать подтверждения, что слушатель зарегистрирован
		c.config.Args,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.client.Context().Done():
			return ErrClientClosed
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("message channel closed unexpectedly")
			}
			// возможо добавлю логирование autoask сценария
			if !c.config.AutoAck {
				// Ручное подтверждение
				if err := c.handler(ctx, msg); err != nil {
					if nackErr := msg.Nack(c.config.Nack.Multiple, c.config.Nack.Requeue); nackErr != nil {
						zlog.Logger.Error().Msgf("NACK failed: %v", nackErr)
					}
				} else {
					if ackErr := msg.Ack(c.config.Ask.Multiple); ackErr != nil {
						zlog.Logger.Error().Msgf("ACK failed: %v", ackErr)
					}
				}
			}
		}
	}
}
