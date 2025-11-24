## WBF

WBF — это внутренний инфраструктурный фреймворк, минималистичный набор обёрток для работы с основными инфраструктурными сервисами в Go. Он поможет быстро и просто подключить к проекту такие штуки, как PostgreSQL, Redis, Kafka, логирование (через zerolog), конфиги (через viper), повторные попытки (retry) и удобную работу с горутинами и каналами.

Что внутри?
    * dbpg — работа с PostgreSQL (разделение на чтение/запись, пул соединений, батчи, автоматический retry)
    * redis — работа с Redis (поддержка батчей и retry)
    * kafka — Kafka (producer/consumer, автоматические retry, асинхронное потребление)
    * zlog — логирование на базе zerolog (по умолчанию в формате JSON)
    * config — загрузка конфигов через viper
    * retry — универсальный механизм повторных попыток для любых операций
    * ginext — обёртка для Gin (web-фреймворк)

### Как быстро начать?


#### PostgreSQL
```go
opts := &dbpg.Options{MaxOpenConns: 10, MaxIdleConns: 5}
db, err := dbpg.New(masterDSN, slaveDSNs, opts)
```

С автоматическим повтором запросов (через пакет retry):
```go
res, err := db.ExecWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: time.Second, Backoff: 2}, "UPDATE ...")
```

Пакетная запись через канал:

```go
ch := make(chan string)
go db.BatchExec(ctx, ch)
ch <- "INSERT ..."
close(ch)
```

Транзакции WithTx:

```go
err := db.WithTx(ctx, func(tx *sql.Tx) error {
    tx.ExecContext(ctx, "INSERT ...")
    tx.ExecContext(ctx, "UPDATE ...")
    return nil
})
```

#### Redis
```go
client := redis.New("localhost:6379", "", 0)
val, err := client.GetWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: time.Second, Backoff: 2}, "key")
```

Пакетная запись через канал:
```go
ch := make(chan [2]string)
go client.BatchWriter(ctx, ch)
ch <- [2]string{"key", "value"}
close(ch)
```


#### Kafka
```go
// Producer
producer := kafka.NewProducer([]string{"localhost:9092"}, "topic")
err := producer.SendWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: time.Second, Backoff: 2}, []byte("key"), []byte("value"))
```

```go
// Consumer
consumer := kafka.NewConsumer([]string{"localhost:9092"}, "topic", "group")
msgCh := make(chan kafka.Message)
consumer.StartConsuming(ctx, msgCh, retry.Strategy{Attempts: 3, Delay: time.Second, Backoff: 2})
for msg := range msgCh {
    // обработка сообщения
}
```

#### Логирование
```go
zlog.Init()
zlog.Logger.Info().Msg("Hello")
```

#### Конфиги
```go
cfg := config.New()
_ = cfg.Load("config.yaml")
val := cfg.GetString("some.key")
```

#### Повторные попытки (retry)
```go
err := retry.Do(func() error {
    // ваш код
    return nil
}, retry.Strategy{Attempts: 3, Delay: time.Second, Backoff: 2})
```
а так же

```go
ctx := context.Background()
err := retry.DoContext(ctx, retry.Strategy{Attempts: 3, Delay: time.Second, Backoff: 2},
    func(context.Context) error {
	//ваш код
	retrun nil
})
```


### rabbitmq

Описание и документация [rabbitmq_doc.md](docs/rabbitmq_doc.md)

## TODO
  * Написать тесты
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
