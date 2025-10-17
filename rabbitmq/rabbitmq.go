// Package rabbitmq provides a high-level wrapper over amqp091-go
// for working with RabbitMQ. Currently supports only basic functionality :).
package rabbitmq

import (
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/retry"
)

type Connection = amqp091.Connection
type Channel = amqp091.Channel
type Queue = amqp091.Queue

type QueueManager struct {
	channel *Channel
}

type QueueConfig struct {
	Durable    bool          // If true, the queue is persisted on RabbitMQ restart
	AutoDelete bool          // If true, the queue is deleted when unused
	Exclusive  bool          // If true, the queue is exclusive to a single connection
	NoWait     bool          // If true, the server will not send a confirmation
	Args       amqp091.Table // Additional arguments
}

type Publisher struct {
	channel  *Channel
	exchange string
}

type PublishingOptions struct {
	Mandatory  bool          // If true, message is returned if there is no matching queue
	Immediate  bool          // If true, message is returned if there is no active consumer
	Expiration time.Duration // Message TTL
	Headers    amqp091.Table // Message headers
}

type ConsumerConfig struct {
	Queue     string        // Queue name
	Consumer  string        // Consumer tag
	AutoAck   bool          // Automatically acknowledge messages
	Exclusive bool          // Exclusive access to the queue
	NoLocal   bool          // Not supported in RabbitMQ
	NoWait    bool          // If true, the server will not send a confirmation
	Args      amqp091.Table // Additional arguments
}

type Consumer struct {
	channel *Channel
	config  *ConsumerConfig
}

type Exchange struct {
	name       string        // Exchange name
	kind       string        // Exchange type: direct, fanout, topic, headers
	Durable    bool          // If true, exchange is persisted after restart
	AutoDelete bool          // If true, exchange is deleted when unused
	Internal   bool          // If true, exchange cannot be published directly
	NoWait     bool          // If true, no server confirmation is expected
	Args       amqp091.Table // Additional arguments
}

// Name returns the exchange name
func (e *Exchange) Name() string {
	return e.name
}

// Kind returns the exchange type
func (e *Exchange) Kind() string {
	return e.name
}

/*
NewExchange creates a new Exchange instance.

name - exchange name

kind - exchange type: direct, fanout, topic, headers
*/
func NewExchange(name, kind string) *Exchange {
	return &Exchange{
		name: name,
		kind: kind,
	}
}

/*
NewConsumer creates a new Consumer instance.

ch - AMQP channel

config - consumer configuration
*/
func NewConsumer(ch *Channel, config *ConsumerConfig) *Consumer {
	return &Consumer{
		channel: ch,
		config:  config,
	}
}

/*
NewConsumerConfig creates a default consumer configuration.

queue - name of the queue to subscribe to
*/
func NewConsumerConfig(queue string) *ConsumerConfig {
	return &ConsumerConfig{
		Queue: queue,
	}
}

/*
NewPublisher creates a new Publisher instance.

ch - AMQP channel

exchange - exchange name
*/
func NewPublisher(ch *Channel, exchange string) *Publisher {
	return &Publisher{
		channel:  ch,
		exchange: exchange,
	}
}

/*
NewQueueManager creates a new QueueManager instance.

channel - AMQP channel used for queue management
*/
func NewQueueManager(channel *Channel) *QueueManager {
	return &QueueManager{
		channel: channel,
	}
}

/*
Connect establishes a connection to RabbitMQ with retry attempts.

url - connection string

retries - number of attempts

pause - delay between attempts
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
BindToChannel declares an exchange on the given AMQP channel.

ch - AMQP channel
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
DeclareQueue declares a queue with a given name and configuration.

name - queue name

config - optional configuration parameters
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
Publish sends a message with a given routingKey to the exchange associated with Publisher.

body - message body

exchange - target exchange

contentType - MIME content type

options - optional publishing options
*/
func (p *Publisher) Publish(body []byte, routingKey, contentType string, options ...PublishingOptions) error {
	var option PublishingOptions

	if len(options) > 0 {
		option = options[0]
	}

	pub := amqp091.Publishing{
		Headers:     option.Headers,
		ContentType: contentType,
		Body:        body,
	}

	if option.Expiration > 0 {
		pub.Expiration = fmt.Sprintf("%d", option.Expiration.Milliseconds())
	}

	return p.channel.Publish(
		p.exchange,
		routingKey,
		option.Mandatory,
		option.Immediate,
		pub,
	)
}

/*
PublishWithRetry publishes a message with retry attempts on failure.

body - message body

exchange - target exchange

contentType - MIME content type

strategy - retry strategy
*/
func (p *Publisher) PublishWithRetry(body []byte, routingKey, contentType string, strategy retry.Strategy, options ...PublishingOptions) error {
	return retry.Do(func() error {
		return p.Publish(body, routingKey, contentType, options...)
	}, strategy)
}

/*
Consume starts message consumption and sends messages into the provided channel.

msgChan - channel to receive message bodies
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

/*
ConsumeWithRetry attempts to consume messages with a retry strategy on failure.

msgChan - channel to receive message bodies
*/
func (c *Consumer) ConsumeWithRetry(msgChan chan []byte, strategy retry.Strategy) error {
	return retry.Do(func() error {
		return c.Consume(msgChan)
	}, strategy)
}
