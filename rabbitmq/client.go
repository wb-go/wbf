// Package rabbitmq это обертка над github.com/rabbitmq/amqp091-go
package rabbitmq

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	defaultConnectTimeout = 10 * time.Second
	defaultHeartbeat      = 10 * time.Second
)

// RabbitClient основная структура клиента.
type RabbitClient struct {
	config ClientConfig
	conn   *amqp091.Connection
	mu     sync.RWMutex
	notify chan *amqp091.Error
	ctx    context.Context
	cancel context.CancelFunc
	closed atomic.Bool
}

// NewClient создаёт и инициализирует нового клиента RabbitMQ.
// Проверяет обязательные параметры, устанавливает значения по умолчанию,
// создаёт контекст и выполняет первичное подключение к брокеру.
func NewClient(cfg ClientConfig) (*RabbitClient, error) {
	if cfg.URL == "" {
		return nil, ErrMissingURL
	}
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = defaultConnectTimeout
	}
	if cfg.Heartbeat == 0 {
		cfg.Heartbeat = defaultHeartbeat
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &RabbitClient{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	if err := client.connect(); err != nil {
		cancel()
		return nil, fmt.Errorf("initial connect failed: %w", err)
	}

	return client, nil
}

// connect устанавливает соединение с RabbitMQ согласно настройкам клиента.
// Создаёт новое AMQP-соединение с конфигурацией таймаутов и heartbeat, заменяет
// существующее соединение (если есть) и запускает горутину для мониторинга состояния.
func (c *RabbitClient) connect() error {
	dialer := &net.Dialer{
		Timeout:   c.config.ConnectTimeout,
		KeepAlive: c.config.Heartbeat,
	}

	amqpConf := amqp091.Config{
		Heartbeat: c.config.Heartbeat,
		Dial:      func(network, addr string) (net.Conn, error) { return dialer.Dial(network, addr) },
		Locale:    "en_US",
	}

	conn, err := amqp091.DialConfig(c.config.URL, amqpConf)
	if err != nil {
		return err
	}

	c.mu.Lock()

	oldConn := c.conn
	c.conn = conn
	c.notify = make(chan *amqp091.Error, 1)
	conn.NotifyClose(c.notify)

	c.mu.Unlock()

	if oldConn != nil {
		_ = oldConn.Close()
	}

	go c.watchConnection()

	return nil
}

// watchConnection отслеживает закрытие соединения с RabbitMQ.
// При получении ошибки из канала (если клиент не закрыт явно)
// запускает цикл переподключения.
func (c *RabbitClient) watchConnection() {
	select {
	case <-c.ctx.Done():
		return
	case err := <-c.notify:
		if err != nil && !c.closed.Load() {
			go c.reconnectLoop()
		}
	}
}

// reconnectLoop выполняет повторные попытки подключения к RabbitMQ с экспоненциальной задержкой.
// Цикл продолжается до успешного подключения или явного закрытия клиента.
func (c *RabbitClient) reconnectLoop() {
	delay := c.config.ReconnectStrat.Delay
	for attempt := 0; !c.closed.Load(); attempt++ {
		if err := c.connect(); err == nil {
			return
		}
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(delay):
		}
		delay = time.Duration(float64(delay) * c.config.ReconnectStrat.Backoff)
	}
}

// GetChannel возвращает новый AMQP-канал для работы с RabbitMQ.
// Проверяет, что клиент не закрыт и соединение активно.
func (c *RabbitClient) GetChannel() (*amqp091.Channel, error) {
	if c.closed.Load() {
		return nil, ErrClientClosed
	}

	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return nil, ErrChannelLost
	}

	return conn.Channel()
}

// Close выполняет graceful shutdown клиента RabbitMQ:
// Устанавливает флаг closed, отменяет контекст и закрывает AMQP-соединение.
func (c *RabbitClient) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}
	c.cancel()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Context возвращает контекст клиента.
func (c *RabbitClient) Context() context.Context {
	return c.ctx
}

// DeclareExchange объявляет exchange в RabbitMQ.
// Создаёт временный канал, объявляет exchange с указанными
// параметрами и закрывает канал после использования.
func (c *RabbitClient) DeclareExchange(name, kind string, durable, autoDelete,
	internal bool, args amqp091.Table) error {
	ch, err := c.GetChannel()
	if err != nil {
		return err
	}
	defer func(ch *amqp091.Channel) {
		_ = ch.Close()
	}(ch)

	return ch.ExchangeDeclare(name, kind, durable, autoDelete, internal, false, args)
}

// DeclareQueue объявляет очередь и настраивает её связь с обменником.
func (c *RabbitClient) DeclareQueue(
	queueName, exchangeName, routingKey string,
	queueDurable, queueAutoDelete bool,
	exchangeDurable bool,
	queueArgs amqp091.Table,
) error {
	ch, err := c.GetChannel()
	if err != nil {
		return err
	}
	defer func(ch *amqp091.Channel) {
		_ = ch.Close()
	}(ch)

	// Объявляем очередь
	_, err = ch.QueueDeclare(
		queueName, queueDurable, queueAutoDelete, false, false, queueArgs,
	)
	if err != nil {
		return err
	}

	// Привязываем очередь к exchange
	return ch.QueueBind(queueName, routingKey, exchangeName, false, nil)
}
