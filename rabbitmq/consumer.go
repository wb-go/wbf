package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/retry"
)

const maxDelay = 1 * time.Hour

// Consumer - обертка над RabbitMQ-клиентом для получения сообщений из обменника.
type Consumer struct {
	client  *RabbitClient
	config  ConsumerConfig
	handler MessageHandler
}

// NewConsumer конструктор Consumer.
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

// Start запускает консьюмера. При разрыве соединения автоматически
// восстанавливает подключение с возрастающими задержками между попытками.
func (c *Consumer) Start(ctx context.Context) error {
	currentDelay := c.client.config.ReconnectStrat.Delay

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.client.Context().Done():
			return ErrClientClosed
		default:
		}

		if !c.client.Healthy() {
			c.backoffWait(ctx, c.client.config.ReconnectStrat, &currentDelay)
			continue
		}

		currentDelay = c.client.config.ReconnectStrat.Delay

		if err := c.consume(ctx); err != nil {
			continue
		}

		return nil
	}
}

// Healthy проверяет состояние клиента RabbitMQ.
// Клиент считается здоровым, если не закрыт явно и имеет активное AMQP-соединение.
func (c *RabbitClient) Healthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.closed.Load() && c.conn != nil && !c.conn.IsClosed()
}

// backoffWait ждёт currentDelay секунд, экспоненциально увеличивая задержку при каждом вызове.
// Прерывается при отмене основного контекста или при закрытии клиента RabbitMQ.
func (c *Consumer) backoffWait(ctx context.Context, strategy retry.Strategy, currentDelay *time.Duration) {
	timer := time.NewTimer(*currentDelay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		return
	case <-c.client.Context().Done():
		if !timer.Stop() {
			<-timer.C
		}
		return
	case <-timer.C:
		newDelay := min(time.Duration(float64(*currentDelay)*strategy.Backoff), maxDelay)
		*currentDelay = newDelay
		return
	}
}

// consume подключается к очереди и стартует воркеров для обработки сообщений.
// Работает до отмены контекста, закрытия клиента или обрыва связи с брокером.
func (c *Consumer) consume(ctx context.Context) error {
	ch, err := c.client.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer func(ch *amqp091.Channel) {
		_ = ch.Close()
	}(ch)

	if c.config.PrefetchCount > 0 {
		if err := ch.Qos(c.config.PrefetchCount, 0, false); err != nil {
			return fmt.Errorf("failed to set prefetch count to %d: %w",
				c.config.PrefetchCount, err)
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
		return fmt.Errorf("failed to start consumer %q on queue %q: %w",
			c.config.ConsumerTag, c.config.Queue, err)
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
	default:
		wg.Wait()
		return ErrWorkersTerminated
	}
}

// worker читает сообщения из канала msgs и передаёт их на обработку в processDelivery.
// Завершается при закрытии канала msgs (потеря соединения) или отмене контекста.
func (c *Consumer) worker(ctx context.Context, msgs <-chan amqp091.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			c.processDelivery(ctx, msg)
		}
	}
}

// processDelivery обрабатывает одно сообщение в соответствии с настройками консьюмера.
func (c *Consumer) processDelivery(ctx context.Context, msg amqp091.Delivery) {
	if c.config.AutoAck {
		if err := retry.DoContext(ctx, c.client.config.ConsumingStrat,
			func() error { return c.handler(ctx, msg) }); err != nil {
			log.Printf("WARN: AutoAck handler failed for consumer %q: %v", c.config.ConsumerTag, err)
		}
		return
	}

	if err := retry.DoContext(ctx, c.client.config.ConsumingStrat,
		func() error { return c.handler(ctx, msg) }); err != nil {
		if nackErr := msg.Nack(c.config.Nack.Multiple, c.config.Nack.Requeue); nackErr != nil {
			log.Printf("ERROR: Failed to send NACK: %v", nackErr)
		}
	} else {
		if ackErr := msg.Ack(c.config.Ask.Multiple); ackErr != nil {
			log.Printf("ERROR: Failed to send ACK: %v", ackErr)
		}
	}
}
