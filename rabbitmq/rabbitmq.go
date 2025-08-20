// Package rabbitmq предоставляет высокоуровневую обертку над библиотекой amqp091-go
// для работы с RabbitMQ. Пока только базовая функциональность :).
package rabbitmq

import (
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Connection = amqp091.Connection
type Channel = amqp091.Channel
type Queue = amqp091.Queue

type QueueManager struct {
	channel *Channel
}

type QueueConfig struct {
	Durable    bool          // Если true, то очередь сохраняется при перезапуске RabbitMQ
	AutoDelete bool          // Если true, то очередь удаляется при отсутствии подписчиков
	Exclusive  bool          // Если true, то очередь доступна только одному соединению
	NoWait     bool          // Если true, то не ждем подтверждения от сервера
	Args       amqp091.Table // Доп аргументы
}

type Publisher struct {
	channel    *Channel
	exchange   string
	routingKey string
}

type PublishingOptions struct {
	Mandatory  bool          // Если true, то сообщение будет возвращено при отсутствии очереди
	Immediate  bool          // Если true, то сообщение будет возвращено при отсутствии потребителя
	Expiration time.Duration // TTL сообщения
}

type ConsumerConfig struct {
	Queue     string        // Имя очереди
	Consumer  string        // Идентификатор потребителя
	AutoAck   bool          // Автоматическое подтверждение сообщений
	Exclusive bool          // Эксклюзивный доступ к очереди
	NoLocal   bool          // Не поддерживается в RabbitMQ
	NoWait    bool          // Если true, то не ждем подтверждения от сервера
	Args      amqp091.Table // Доп аргументы
}

type Consumer struct {
	channel *Channel
	config  *ConsumerConfig
}

type Exchange struct {
	name       string        // Название обменника
	kind       string        // Тип обменника: direct, fanout, topic, headers
	Durable    bool          // Усли true, то обменник сохранится при перезагрузке сервера
	AutoDelete bool          // Если true, то обменник удалится когда все очереди отвяжутся
	Internal   bool          // Усли true, то обменник нельзя использовать для публикации напрямую
	NoWait     bool          // Если true, то не ждем подтверждения от сервера
	Args       amqp091.Table // Доп аргументы
}

// Name возвращает название обменника
func (e *Exchange) Name() string {
	return e.name
}

// Kind возвращает тип обменника
func (e *Exchange) Kind() string {
	return e.name
}

/*
NewExchange создате новый экземпляр Exchange

name - название обменника,

kind - тип обменника: direct, fanout, topic, headers
*/
func NewExchange(name, kind string) *Exchange {
	return &Exchange{
		name: name,
		kind: kind,
	}
}

/*
NewConsumer создает новый экземпляр Consumer.

ch - канал AMQP,

config - конфигурация потребителя.
*/
func NewConsumer(ch *Channel, config *ConsumerConfig) *Consumer {
	return &Consumer{
		channel: ch,
		config:  config,
	}
}

/*
NewConsumerConfig создает конфигурацию потребителя с настройками по умолчанию.

queue - имя очереди для подписки.
*/
func NewConsumerConfig(queue string) *ConsumerConfig {
	return &ConsumerConfig{
		Queue: queue,
	}
}

/*
NewPublisher создает новый экземпляр Publisher.

ch - канал AMQP,

routingKey - "адрес" для доставки сообщений.
*/
func NewPublisher(ch *Channel, exhange, routingKey string) *Publisher {
	return &Publisher{
		channel:    ch,
		exchange:   exhange,
		routingKey: routingKey,
	}
}

/*
NewQueueManager создает новый экземпляр QueueManager.

channel - канал AMQP для управления очередями.
*/
func NewQueueManager(channel *Channel) *QueueManager {
	return &QueueManager{
		channel: channel,
	}
}

/*
Connect устанавливает соединение с RabbitMQ с автоматическими повторами.

url - строка подключения,

retries - количество попыток,

pause - задержка между попытками.
*/
func Connect(url string, retries int, pause time.Duration) (*Connection, error) {
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
func (e *Exchange) BindToChannel(ch *Channel) error {
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
func (qm *QueueManager) DeclareQueue(name string, config ...QueueConfig) (Queue, error) {
	cfg := &QueueConfig{}

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
Publish публикует сообщение с заданным routingKey в exchange, связанный с Publisher.

body - тело сообщения,

exchange - точка обмена,

contentType - тип контента,

options - необязательные параметры публикации.
*/
func (p *Publisher) Publish(body []byte, contentType string, options ...PublishingOptions) error {
	var option PublishingOptions

	if len(options) > 0 {
		option = options[0]
	}

	pub := amqp091.Publishing{
		ContentType: contentType,
		Body:        body,
	}

	if option.Expiration > 0 {
		pub.Expiration = fmt.Sprintf("%d", option.Expiration.Milliseconds())
	}

	return p.channel.Publish(
		p.exchange,
		p.routingKey,
		option.Mandatory,
		option.Immediate,
		pub,
	)
}

/*
Consume начинает потребление сообщений и отправляет их в указанный канал.

msgChan - канал для получения сообщений.
*/
func (c *Consumer) Consume(msgChan chan []byte) error {
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

				if err = msg.Nack(false, true); err != nil {
					log.Printf("Failed to nack message: %v", err)
				}
			}
		}

		msgChan <- msg.Body
	}

	return nil
}
