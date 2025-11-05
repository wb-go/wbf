package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"

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
	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}
	return &Consumer{
		client:  client,
		config:  cfg,
		handler: handler,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	zlog.Logger.Info().Msgf("Starting consumer %s", c.config.ConsumerTag)
	for {
		err := c.consumeOnce(ctx)
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return nil
		case <-c.client.Context().Done():
			return ErrClientClosed
		default:
		}

		if !c.client.backoffWait(ctx, c.client.config.ConsumeRetry.Delay) {
			return nil
		}
	}
}
func (c *Consumer) consumeOnce(ctx context.Context) error {
	ch, err := c.client.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer func(ch *amqp091.Channel) {
		_ = ch.Close()
	}(ch)

	if c.config.PrefetchCount > 0 {
		if err := ch.Qos(c.config.PrefetchCount, 0, false); err != nil {
			return fmt.Errorf("failed to set QoS: %w", err)
		}
	}

	msgs, err := ch.Consume(
		c.config.Queue,
		c.config.ConsumerTag,
		c.config.AutoAck,
		false, // exclusive
		false, // no-local
		false, // no-wait
		c.config.Args,
	)
	if err != nil {
		return fmt.Errorf("failed to consume: %w", err)
	}

	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	for i := 0; i < c.config.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.worker(workerCtx, msgs)
		}()
	}
	select {
	case <-ctx.Done():
		cancel()
		wg.Wait()
		return ctx.Err()
	case <-c.client.Context().Done():
		cancel()
		wg.Wait()
		return ErrClientClosed
	case _, ok := <-msgs:
		if !ok {
			cancel()
			wg.Wait()
			return errors.New("message channel closed unexpectedly")
		}
	}

	wg.Wait()
	return nil
}

func (c *Consumer) worker(ctx context.Context, msgs <-chan amqp091.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}

			if c.config.AutoAck {
				// В режиме AutoAck не нужно подтверждать
				if err := c.handler(ctx, msg); err != nil {
					zlog.Logger.Warn().Err(err).Str("consumer", c.config.ConsumerTag).Msg("AutoAck handler failed")
				}
			} else {
				if err := c.handler(ctx, msg); err != nil {
					// NACK с настройками
					if nackErr := msg.Nack(c.config.Nack.Multiple, c.config.Nack.Requeue); nackErr != nil {
						zlog.Logger.Error().Err(nackErr).Msg("NACK failed")
					}
				} else {
					// ACK с настройками
					if ackErr := msg.Ack(c.config.Ask.Multiple); ackErr != nil {
						zlog.Logger.Error().Err(ackErr).Msg("ACK failed")
					}
				}
			}
		}
	}
}
