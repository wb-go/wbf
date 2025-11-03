# 🐇 Пакет `rabbitmq`

Пакет `rabbitmq` предоставляет высокоуровневую обёртку над библиотекой [amqp091-go](https://pkg.go.dev/github.com/rabbitmq/amqp091-go) для упрощённой и надёжной работы с RabbitMQ.  
Он обеспечивает удобные средства для создания подключений, публикации и потребления сообщений с поддержкой **retry-стратегий**, **graceful shutdown** и **функциональных опций** для конфигурации сообщений.

---

## 📘 Описание

Основная цель пакета - предоставить удобный API для:
- создания и управления подключением к RabbitMQ;
- надёжной публикации сообщений с автоматическими повторами;
- обработки сообщений в виде consumer’ов с контролем ack/nack;
- гиькой настройки задержек и стратегии повторов через `retry.Strategy`.

Пакет поддерживает работу с контекстом (`context.Context`) и корректное завершение при остановке приложения.

---

## ⚙️ Структуры и типы

### `ClientConfig`
Конфигурация для клиента RabbitMQ.

```go
type ClientConfig struct {
    URL            string        // AMQP URL для подключения
    ConnectionName string        // Имя подключения (отображается в RabbitMQ UI)
    ConnectTimeout time.Duration // Таймаут на установку соединения
    Heartbeat      time.Duration // Интервал heartbeat
    PublishRetry   retry.Strategy // Стратегия повторов для публикации
    ConsumeRetry   retry.Strategy // Стратегия повторов для потребления
}
```

---

### `RabbitClient`
Главная структура, управляющая соединением с RabbitMQ.

```go
type RabbitClient struct {
    config ClientConfig
    conn   *amqp091.Connection
    mu     sync.RWMutex
    notify chan *amqp091.Error
    ctx    context.Context
    cancel context.CancelFunc
    closed atomic.Bool
}
```

**Основные методы:**

- `NewClient(cfg ClientConfig) (*RabbitClient, error)`  
  Создаёт новое соединение с RabbitMQ с учётом параметров `ClientConfig` и контекста.

- `GetChannel() (*amqp091.Channel, error)`  
  возвращает новый канал.

- `Close() error`  
  Корректно закрывает соединение и канал.

- `Healthy() bool`  
  проверяет, живо ли соединение.

- `Context() context.Context`  
  возвращает контекст клиента.

- `DeclareExchange(name, kind string, durable, autoDelete, internal bool, args amqp091.Table) error`  
  Объявляет exchange.

- `DeclareQueue(queueName, exchangeName, routingKey string, queueDurable, queueAutoDelete bool, exchangeDurable bool, queueArgs amqp091.Table) error`  
  объявляет очередь и привязывает её к exchange.
---

   Методы `DeclareExchange` и `DeclareQueue` для удобства локальной разработки

---
### `Publisher`
Объект, отвечающий за отправку сообщений в RabbitMQ.

```go
type Publisher struct {
    client      *RabbitClient
    exchange    string
    contentType string
}
```

**Основные функции и методы:**

- ` NewPublisher(client *RabbitClient, exchange, contentType string) *Publisher`  
  Создаёт новый экземпляр Publisher.

- `Publish(ctx context.Context,	body []byte, routingKey string, opts ...PublishOption) error`  
  Отправляет сообщение в указанный `routingKey`.  
  Поддерживает функциональные опции (`WithExpiration`, `WithHeaders`) и стратегию повторных попыток при ошибках.

---

### `Consumer`
Отвечает за приём и обработку сообщений из очередей RabbitMQ.

```go
type Consumer struct {
    client  *RabbitClient
    config  ConsumerConfig
    handler MessageHandler
}
```

**ConsumerConfig**
```go
type ConsumerConfig struct {
	Queue string  // имя очереди
	ConsumerTag string //имя обработчика лучше оставить пустым
	AutoAck bool // авто подтверждение 
	Ask AskConfig // настройки Ask (Multiple)
	Nack NackConfig // настройки Nack (Multiple, Requeue)
	Args amqp091.Table // метаданные для rabbit
}
```

**Основные функции и методы:**

- `NewConsumer(client *RabbitClient, cfg ConsumerConfig, handler MessageHandler) *Consumer`  
  Создаёт нового консьюмера, использующего заданную стратегию повторов и обработчик сообщений.
  обработчик `handler` метод вида `func(context.Context, amqp091.Delivery) error` для обработки каждого из сообщений
  при ошибке обработчика — выполняет `NACK` и может повторить попытку в соответствии со стратегией.

- `Start(ctx context.Context) error`  
  запускает консьюмер на выполнение.

---

### `MessageHandler`
Тип обработчика сообщений.

```go
type MessageHandler func(ctx context.Context, d amqp.Delivery) error
```

Если обработчик возвращает `nil`, сообщение считается успешно обработанным (`ACK`).  
Если возвращает ошибку — сообщение помечается как неудавшееся (`NACK`).

---

### `PublishOption`
Функциональные опции, применяемые к сообщению перед отправкой.

```go
type PublishOption func(*amqp.Publishing)
```

**Доступные опции:**

- `WithExpiration(d time.Duration) PublishOption`  
  Устанавливает срок жизни сообщения.

- `WithHeaders(headers amqp091.Table) PublishOption`  
  Добавляет пользовательские заголовки.

---

### Ошибки
Пакет может определять и возвращать собственные типы ошибок (например, `ErrClientClosed`, `ErrChannelLost`).  
Они используются для более точной диагностики при работе с брокером.

---

## 💡 Использование (Usage)

### Создание клиента

```go
strategy := retry.Strategy{
    Attempts: 3,
    Delay:    3 * time.Second,
    Backoff:  2,
}
cfg := rabbitmq.ClientConfig{
    URL:            "amqp://guest:guest@localhost:5672/",
    ConnectionName: "my-service",
    ConnectTimeout: 5 * time.Second,
    Heartbeat:      10 * time.Second,
    PublishRetry: strategy,
    ConsumeRetry: strategy,
}

client, err := rabbitmq.NewClient(config)
if err != nil {
    log.Fatalf("Ошибка подключения к RabbitMQ: %v", err)
}
defer client.Close()
```

---

### Публикация сообщения

```go
publisher := rabbitmq.NewPublisher(client, "MyTestExchange", "application/json")

ctx := context.Backgroung()

bodyMsg := []byte(`{"event":"user_registered","id":123}`)
routingKey := "MyTestRoutingKey"
err = publisher.Publish(
    ctx,
    bodyMsg,
    routingKey,
    rabbitmq.WithExpiration(5*time.Minute),
    rabbitmq.WithHeaders(amqp.Table{"x-service": "auth"}),
)
if err != nil {
    log.Printf("Ошибка публикации: %v", err)
}
```

---

### Потребление сообщений

```go
done := make(chan struct{}) // имитация ожидания

handler := func(ctx context.Context, d amqp.Delivery) error {
    log.Printf("Получено сообщение: %s", string(d.Body))
    // Обработка...
    return nil // вернуть ошибку, если нужно NACK
}

queueArgs := amqp091.Table{
    "x-dead-letter-exchange":    "dlx",          // exchange для DLQ
    "x-dead-letter-routing-key": "test.queue.dlq", // routing key для DLQ
}

consumerCfg := rabbitmq.ConsumerConfig{
	Queue: "my-queue", 
	Args: queueArgs,
}

consumer := rabbitmq.NewConsumer(client, consumerCfg, handler)

go func() {
    if err := err := consumer.Start(ctx); err != nil {
        log.Fatalf("Ошибка при потреблении сообщений: %v", err)
    }
    done <- struct{}{}
}

<-done // завершаем
```

---

## 🧠 Примечания

- Все операции выполняются с поддержкой `context.Context`, что позволяет корректно завершать работу при остановке сервиса.  
- `retry.Strategy` используется для повторов при ошибках как в `Publish`, так и в `Consume`.  
- При необходимости можно внедрить собственную стратегию логирования, метрик или трейсинга на уровне обработчиков.

