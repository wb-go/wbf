// Package redis provides a client wrapper for Redis operations.
package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/wb-go/wbf/retry"
)

// NoMatches is returned when Redis did not find any matching key.
const NoMatches = redis.Nil

// Client wraps the Redis client.
type Client struct {
	*redis.Client
}

// New creates a new Redis client.
func New(addr, password string, db int) *Client {
	return &Client{
		redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

// Get retrieves a value by key from Redis.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Set stores a value by key in Redis.
func (c *Client) Set(ctx context.Context, key string, value interface{}) error {
	return c.Client.Set(ctx, key, value, 0).Err()
}

// SetWithExpiration stores a value with a specified expiration time.
func (c *Client) SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// GetWithRetry retrieves a value using a retry strategy.
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

// SetWithRetry stores a value using a retry strategy.
func (c *Client) SetWithRetry(ctx context.Context, strategy retry.Strategy, key string, value interface{}) error {
	return retry.Do(func() error {
		return c.Set(ctx, key, value)
	}, strategy)
}

// BatchWriter performs batched writes to Redis asynchronously.
func (c *Client) BatchWriter(ctx context.Context, in <-chan [2]string) {
	go func() {
		for pair := range in {
			_ = c.Set(ctx, pair[0], pair[1]) // Errors can be logged if needed.
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
}

// Del removes a key from Redis.
func (c *Client) Del(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}

// DelWithRetry removes a key from Redis using a retry strategy.
func (c *Client) DelWithRetry(ctx context.Context, strategy retry.Strategy, key string) error {
	return retry.Do(func() error {
		return c.Del(ctx, key)
	}, strategy)
}
