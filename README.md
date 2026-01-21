![wbf banner](assets/banner.png)

<h3 align="center">Минималистичный фреймворк для работы с базовыми инфраструктурными штуками.</h3> 

<h1></h1>

<br>

WBF — это готовый набор обёрток для стандартной инфраструктуры. С его помощью можно быстро интегрировать в проект базу данных (PostgreSQL), кэширование (Redis), брокера сообщений (Kafka/RabbitMQ), систему логирования (Zerolog/Logger) и загрузчик конфигураций (Viper/CleanEnv).

<br>

## Пакеты:

* [dbpg](/dbpg/dbpg.go) — пакет для работы с PostgreSQL, реализующий архитектуру «мастер-реплика» с балансировкой нагрузки на чтение, пулом соединений и встроенной поддержкой повторных попыток. 

* [pgxdriver](/dbpg/pgx-driver/postgres.go) — пакет-обёртка над pgx/v5 с настраиваемым пулом соединений, встроенным retry-механизмом при подключении, транзакционным менеджером, batch/bulk-операциями и интеграцией с Squirrel.

* [redis](/redis/redis.go) — пакет-обёртка над go-redis со встроенной поддержкой повторных попыток, асинхронным батчевым выполнением операций записи и упрощённым API.

* [kafka](/kafka/kafka.go) — пакет для работы с Apache Kafka, предоставляющий готовых продюсера и консьюмера с автоматическими повторами и асинхронной обработкой сообщений.

* [kafkav2](/kafka/kafka-v2/processor.go) — улучшенный пакет для работы с Apache Kafka, предоставляющий готовый producer, отказоустойчивый consumer с process retry + jitter, возможность работы с DLQ и улучшенное логирование.

* [dlq](kafka/dlq/dlq.go) — компонент Dead Letter Queue для Kafka, предназначенный для надёжного сохранения сообщений, поддерживает сериализацию оригинального тела сообщения в base64, метаданные (топик, попытка, временная метка), а также fallback-механизм при ошибках маршалинга.

* [rabbitmq](/rabbitmq/client.go) — пакет для работы с RabbitMQ, предоставляющий готовые клиенты для публикации и обработки сообщений с автоматическим переподключением, настраиваемыми стратегиями повторных попыток и поддержкой многопоточной обработки.

* [zlog](/zlog/zlog.go) — пакет для структурированного логирования на базе zerolog, предоставляющий готовый глобальный логгер с настройкой формата вывода (JSON или консоль), уровнями логирования и автоматическим добавлением временных меток.

* [logger](logger/logger.go) — пакет для логирования с возможность настройки, унифицированными интерфейсами для zap/slog/zerolog/logrus, с поддержкой request_id, ротации через lumberjack и структурированием атрибутов.

* [config](/config/config.go) — пакет для работы с конфигурацией, реализующий загрузку настроек из различных источников через Viper, включая .env файлы, YAML/JSON конфиги, переменные окружения и командные флаги.

* [config/cleanenvport](config/cleanenv-port) — порт популярной библиотеки для работы с конфигураций, обеспечивающий строго типизированную загрузку с валидацией через validator и поддержку флага --config / CONFIG_PATH.
    
* [retry](/retry/retry.go) — пакет для реализации повторных попыток выполнения операций, предоставляющий настраиваемые стратегии с экспоненциальным бэк-оффом, поддержкой контекста для graceful shutdown и универсальным интерфейсом для любых функций.

* [ginext](/ginext/ginext.go) — пакет-обёртка для веб-фреймворка Gin с полной поддержкой всех HTTP-методов, middleware и удобной настройкой режимов работы.

* [helpers](/helpers) — пакет для мелких вспомогательных функций общего назначения.

<br>


## Примеры использования

### PostgreSQL

#### dbpg

Инициализация подключения с настройками пула соединений:
```go
opts := &dbpg.Options{MaxOpenConns: 10, MaxIdleConns: 5} 
db, err := dbpg.New(masterDSN, slaveDSNs, opts)
```

<br>

Запрос с автоматическим повторением при ошибках (через пакет retry):
```go
query := "UPDATE..."
strategy := retry.Strategy{Attempts: 3, Delay: 5 * time.Second, Backoff: 2}

res, err := db.ExecWithRetry(ctx, strategy, query)
```

<br>

Пакетная запись через канал:
```go
ch := make(chan string)
go db.BatchExec(ctx, ch)
ch <- "INSERT ..."
close(ch)
```

<br>

Транзакция с автоматическим rollback/commit:
```go
err := db.WithTx(ctx, func(tx *sql.Tx) error {
    tx.ExecContext(ctx, "INSERT ...")
    tx.ExecContext(ctx, "UPDATE ...")
    return nil
})
```

<br>

#### pgx-drvier

Подключение с retry и настройкой пула:

```go
pg, err := pgxdriver.New(
    dsn,
    log,
    pgxdriver.MaxPoolSize(50),
    pgxdriver.MaxConnAttempts(5),
    pgxdriver.BaseRetryDelay(100*time.Millisecond),
)
if err != nil {
    log.Fatal("Failed to connect to PostgreSQL:", err)
}
defer pg.Close()
```

<br>

Работа с транзакциями с автоматическим retry:
```go
tm, err := transaction.NewManager(
    pg,
    log,
    transaction.MaxAttempts(5),
    transaction.BaseRetryDelay(10*time.Millisecond),
)
if err != nil {
    return err
}

err = tm.ExecuteInTransaction(ctx, "transfer", func(tx pgxdriver.QueryExecuter) error {
    _, err := tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
    if err != nil {
        return err
    }
    _, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)
    return err
})
```
<br>

Массовая вставка через BulkInsert:
```go
columns := []string{"name", "email"}
data := [][]any{
    {"Alice", "alice@example.com"},
    {"Bob", "bob@example.com"},
}
count, err := pgxdriver.BulkInsert(ctx, pg, "users", columns, data)
```




### Redis

Подключение и чтение с ретраями:
```go
client := redis.New("localhost:6379", "", 0)
strategy := retry.Strategy{Attempts: 3, Delay: 5 * time.Second, Backoff: 2}

val, err := client.GetWithRetry(ctx, strategy, "key")
```

<br>


Подключение с конфигурацией памяти:
```go
options := redis.Options{
    Address:   "localhost:6379",
    Password:  "",                    
    MaxMemory: "100mb",               
    Policy:    "allkeys-lru",        
}

client, err := redis.Connect(options)
```

<br>

Запись с TTL и ретраями:
```go
strategy := retry.Strategy{Attempts: 3, Delay: 2 * time.Second, Backoff: 2}
key := "abobaUUID"
value := "pending"
expiration := time.Hour

if err := client.SetWithExpirationAndRetry(ctx, strategy, key, value, expiration); err != nil {
    return err
}
```

<br>

Пакетная запись через канал:
```go
ch := make(chan [2]string)
go client.BatchWriter(ctx, ch)
ch <- [2]string{"key", "value"}
close(ch)
```

<br>

### Kafka

Producer — отправка сообщений с автоматическим повторением при ошибках:
```go
producer := kafka.NewProducer([]string{"localhost:9092"}, "topic")
strategy := retry.Strategy{Attempts: 3, Delay: 5 * time.Second, Backoff: 2}

err := producer.SendWithRetry(ctx, strategy, []byte("key"), []byte("value"))
```

#### Kafkav2
```go
producer := kafkav2.NewProducer(brokers, "orders", log)
```

<br>

Consumer — асинхронная обработка сообщений с повторами:
```go
msgCh := make(chan kafka.Message)
consumer := kafka.NewConsumer([]string{"localhost:9092"}, "topic", "group")
strategy := retry.Strategy{Attempts: 3, Delay: 5 * time.Second, Backoff: 2}

consumer.StartConsuming(ctx, msgCh, strategy)

for msg := range msgCh {
    // обработка сообщения
}
```
#### Kafkav2
```go
    consumer := kafkav2.NewConsumer(brokers, "orders", "order-processor", log)
```
<br>
Processor управляет жизненным циклом обработки Kafka сообщений, включающий retry/backoff механизмы и DLQ fallback

```go
// Основной consumer
consumer := kafkav2.NewConsumer(brokers, "orders", "order-processor", log)

// DLQ-продюсер
dlqProducer := kafkav2.NewProducer(brokers, "dlq-orders", log)
dlqClient := dlq.New(dlqProducer, log)

// Процессор с retry и jitter
processor, err := kafkav2.NewProcessor(
    consumer,
    dlqClient,
    log,
    kafkav2.WithMaxAttempts(5),
    kafkav2.WithBaseRetryDelay(150*time.Millisecond),
    kafkav2.WithMaxRetryDelay(5*time.Second),
)
if err != nil {
    log.Fatal(err)
}

// Запуск обработки
processor.Start(ctx, func(ctx context.Context, msg kafka.Message) error {
    // Обработка сообщения
    if err := processOrder(msg.Value); err != nil {
        return fmt.Errorf("обработка заказа %d: %w", msg.Offset, err)
    }
    return nil
})
```

<br>

### Логирование

#### zlog

```go
zlog.Init()
zlog.Logger.Info().Msg("Hello")
```



#### logger

Инициализация с ротацией и уровнем:
```go
log, err := logger.InitLogger(
    logger.ZapEngine,
    "order-service",
    "prod",
    logger.WithLevel(logger.InfoLevel),
    logger.WithRotation("logs/app.log", 100, 5, 30),
)
if err != nil {
    fmt.Fprintf(os.Stderr, "Ошибка инициализации логгера: %v\n", err)
    os.Exit(1)
}
```

<br>

Логирование с контекстом (request_id):

```go
ctx = logger.SetRequestID(ctx, logger.GenerateRequestID())
log.Ctx(ctx).Info("Начата обработка заказа", "order_id", 123)
```

<br>

Структурированный вывод:

```go
log.LogAttrs(ctx, logger.ErrorLevel, "Ошибка обработки",
    logger.String("error_type", "validation"),
    logger.Int("order_id", 123),
    logger.Any("payload", msg.Value),
)
```


### Конфиги

#### Viper
```go
cfg := config.New()
_ = cfg.Load("config.yaml")
val := cfg.GetString("some.key")
```

#### CleanEnvPort

Загрузка с валидацией через --config или CONFIG_PATH:
```go
type Config struct {
    Server struct {
        Host string `yaml:"host" env:"SERVER_HOST" validate:"required"`
        Port int    `yaml:"port" env:"SERVER_PORT" validate:"required,min=1,max=65535"`
    } `yaml:"server"`
    DB struct {
        DSN string `yaml:"dsn" env:"DATABASE_DSN" validate:"required"`
    } `yaml:"database"`
}

var cfg Config
if err := cleanenvport.Load(&cfg); err != nil {
    log.Fatalf("Ошибка загрузки конфигурации: %v", err)
}
```

<br>

Прямая загрузка из файла:
```go
if err := cleanenvport.LoadPath("./config.yaml", &cfg); err != nil {
    log.Fatalf("Ошибка загрузки конфигурации: %v", err)
}
```

<br>

### Повторные попытки (retry)
```go
ctx := context.Background()
strategy := retry.Strategy{Attempts: 3, Delay: 5 * time.Second, Backoff: 2}

err := retry.Do(func() error { return nil }, strategy)
err := retry.DoContext(ctx, strategy, func() error { retrun nil })
```

<br>

### rabbitmq

Описание и документация: [rabbitmq_doc.md](docs/rabbitmq_doc.md)

<br>

## TODO
  * Написать тесты (like that's ever gonna happen)
  * Добавить больше примеров использования
  * Сделать middleware и метрики

## Требования к качеству кода и коммитам

### Pre-commit hooks

В проекте используется [pre-commit](https://pre-commit.com/) для автоматической проверки кода и сообщений коммитов:
- **conventional commits** — все коммиты должны соответствовать [conventionalcommits.org](https://www.conventionalcommits.org/ru/v1.0.0/)
- **golangci-lint** — код должен проходить все проверки линтера

#### Установка и настройка:
1. Установите pre-commit: `pip install pre-commit` или `brew install pre-commit`
2. Установите хуки: `pre-commit install`
3. Для проверки вручную: `pre-commit run --all-files`

## Линтеры

В проекте используется [golangci-lint](https://golangci-lint.run/):
- Конфиг: `.golangci.yml`
- Проверяются стиль, ошибки, best practices
- Перед коммитом и в CI код должен проходить все проверки линтера

## Импорт

Для использования импортируйте пакеты так:

```go
import "github.com/wb-go/wbf/dbpg"
import "github.com/wb-go/wbf/redis"
import "github.com/wb-go/wbf/kafka"
// и т.д.
```

## Лицензия

Этот проект распространяется под лицензией Apache License 2.0. См. файл [LICENSE](LICENSE).
