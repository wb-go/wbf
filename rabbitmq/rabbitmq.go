// Package rabbitmq предоставляет высокоуровневую обертку над библиотекой amqp091-go
// для работы с RabbitMQ. Пока только базовая функциональность :).
package rabbitmq

import (
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type queueManager struct {
	channel *amqp091.Channel
}

type queueConfig struct {
	Durable    bool          // Если true, то очередь сохраняется при перезапуске RabbitMQ
	AutoDelete bool          // Если true, то очередь удаляется при отсутствии подписчиков
	Exclusive  bool          // Если true, то очередь доступна только одному соединению
	NoWait     bool          // Если true, то не ждем подтверждения от сервера
	Args       amqp091.Table // Доп аргументы
}

type publisher struct {
	channel    *amqp091.Channel
	routingKey string
}

type publishingOptions struct {
	Mandatory bool // Если true, то сообщение будет возвращено при отсутствии очереди
	Immediate bool // Если true, то сообщение будет возвращено при отсутствии потребителя
}

type consumerConfig struct {
	Queue     string        // Имя очереди
	Consumer  string        // Идентификатор потребителя
	AutoAck   bool          // Автоматическое подтверждение сообщений
	Exclusive bool          // Эксклюзивный доступ к очереди
	NoLocal   bool          // Не поддерживается в RabbitMQ
	NoWait    bool          // Если true, то не ждем подтверждения от сервера
	Args      amqp091.Table // Доп аргументы
}

type consumer struct {
	channel *amqp091.Channel
	config  *consumerConfig
}

type exchange struct {
	name       string        // Название обменника
	kind       string        // Тип обменника: direct, fanout, topic, headers
	Durable    bool          // Усли true, то обменник сохранится при перезагрузке сервера
	AutoDelete bool          // Если true, то обменник удалится когда все очереди отвяжутся
	Internal   bool          // Усли true, то обменник нельзя использовать для публикации напрямую
	NoWait     bool          // Если true, то не ждем подтверждения от сервера
	Args       amqp091.Table // Доп аргументы
}

/*
NewExchange создате новый экземпляр Exchange

name - название обменника,

kind - тип обменника: direct, fanout, topic, headers
*/
func NewExhcange(name, kind string) *exchange {
	return &exchange{
		name: name,
		kind: kind,
	}
}

/*
NewConsumer создает новый экземпляр Consumer.

ch - канал AMQP,

config - конфигурация потребителя.
*/
func NewConsumer(ch *amqp091.Channel, config *consumerConfig) *consumer {
	return &consumer{
		channel: ch,
		config:  config,
	}
}

/*
NewConsumerConfig создает конфигурацию потребителя с настройками по умолчанию.

queue - имя очереди для подписки.
*/
func NewConsumerConfig(queue string) *consumerConfig {
	return &consumerConfig{
		Queue: queue,
	}
}

// NewPublishingOptions создает опции публикации с настройками по умолчанию.
func NewPublishingOptions() *publishingOptions {
	return &publishingOptions{}
}

/*
NewPublisher создает новый экземпляр Publisher.

ch - канал AMQP,

routingKey - "адрес" для доставки сообщений.
*/
func NewPublisher(ch *amqp091.Channel, routingKey string) *publisher {
	return &publisher{
		channel:    ch,
		routingKey: routingKey,
	}
}

// NewQueueConfig создает конфигурацию очереди с настройками по умолчанию.
func NewQueueConfig() *queueConfig {
	return &queueConfig{}
}

/*
NewQueueManager создает новый экземпляр QueueManager.

channel - канал AMQP для управления очередями.
*/
func NewQueueManager(channel *amqp091.Channel) *queueManager {
	return &queueManager{
		channel: channel,
	}
}

/*
Connect устанавливает соединение с RabbitMQ с автоматическими повторами.

url - строка подключения,

retries - количество попыток,

pause - задержка между попытками.
*/
func Connect(url string, retries int, pause time.Duration) (*amqp091.Connection, error) {
	var conn *amqp091.Connection
	var err error

	for i := 0; i < retries; i++ {
		conn, err = amqp091.Dial(url)
		if err == nil {
			return conn, nil
		}

		time.Sleep(pause)
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %v", retries, err)
}

/*
BindToChannel создает обменник.

ch - канал AMQP.
*/
func (e *exchange) BindToChannel(ch *amqp091.Channel) error {
	return ch.ExchangeDeclare(
		e.name,
		e.kind,
		e.Durable,
		e.AutoDelete,
		e.Internal,
		e.NoWait,
		e.Args,
	)
}

/*
DeclareQueue объявляет очередь с указанным именем и параметрами.

name - имя очереди,

config - необязательные параметры конфигурации.
*/
func (qm *queueManager) DeclareQueue(name string, config ...queueConfig) (amqp091.Queue, error) {
	cfg := NewQueueConfig()

	if len(config) > 0 {
		cfg = &config[0]
	}

	return qm.channel.QueueDeclare(
		name,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Exclusive,
		cfg.NoWait,
		cfg.Args,
	)
}

/*
Publish публикует сообщение в указанный exchange.

body - тело сообщения,

exchange - точка обмена,

contentType - тип контента,

options - необязательные параметры публикации.
*/
func (p *publisher) Publish(body []byte, exchange string, contentType string, options ...publishingOptions) error {
	var option publishingOptions

	if len(options) > 0 {
		option = options[0]
	}

	return p.channel.Publish(
		exchange,
		p.routingKey,
		option.Mandatory,
		option.Immediate,
		amqp091.Publishing{
			ContentType: contentType,
			Body:        body,
		},
	)
}

/*
Consume начинает потребление сообщений и отправляет их в указанный канал.

msgChan - канал для получения сообщений.
*/
func (c *consumer) Consume(msgChan chan []byte) error {
	msgs, err := c.channel.Consume(
		c.config.Queue,
		c.config.Consumer,
		c.config.AutoAck,
		c.config.Exclusive,
		c.config.NoLocal,
		c.config.NoWait,
		c.config.Args,
	)
	if err != nil {
		return err
	}

	for msg := range msgs {
		if !c.config.AutoAck {
			if err := msg.Ack(false); err != nil {
				log.Printf("Failed to ack message: %v", err)
				msg.Nack(false, true)
			}
		}

		msgChan <- msg.Body
	}

	return nil
}
