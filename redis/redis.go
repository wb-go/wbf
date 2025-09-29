// Package redis предоставляет клиент для работы с Redis.
package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/wb-go/wbf/retry"
)

// Client оборачивает Redis клиент.
type Client struct {
	*redis.Client
}

// New создает новый Redis клиент.
func New(addr, password string, db int) *Client {
	return &Client{
		redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

// Get получает значение по ключу из Redis.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Set устанавливает значение по ключу в Redis.
func (c *Client) Set(ctx context.Context, key string, value interface{}) error {
	return c.Client.Set(ctx, key, value, 0).Err()
}

func (c *Client) SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// GetWithRetry получает значение с стратегией повторных попыток.
func (c *Client) GetWithRetry(ctx context.Context, strategy retry.Strategy, key string) (string, error) {
	var val string
	err := retry.Do(func() error {
		v, e := c.Get(ctx, key)
		if e == nil {
			val = v
		}
		return e
	}, strategy)
	return val, err
}

// SetWithRetry устанавливает значение с стратегией повторных попыток.
func (c *Client) SetWithRetry(ctx context.Context, strategy retry.Strategy, key string, value interface{}) error {
	return retry.Do(func() error {
		return c.Set(ctx, key, value)
	}, strategy)
}

// BatchWriter выполняет пакетную запись в Redis.
func (c *Client) BatchWriter(ctx context.Context, in <-chan [2]string) {
	go func() {
		for pair := range in {
			_ = c.Set(ctx, pair[0], pair[1]) // Ошибки можно логировать.
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
}
