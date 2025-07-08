package redis

import (
	"context"

	"wbf/retry"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	*redis.Client
}

func New(addr, password string, db int) *Client {
	return &Client{
		redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value interface{}) error {
	return c.Client.Set(ctx, key, value, 0).Err()
}

func (c *Client) GetWithRetry(ctx context.Context, strat retry.Strategy, key string) (string, error) {
	var val string
	err := retry.Do(func() error {
		v, e := c.Get(ctx, key)
		if e == nil {
			val = v
		}
		return e
	}, strat)
	return val, err
}

func (c *Client) SetWithRetry(ctx context.Context, strat retry.Strategy, key string, value interface{}) error {
	return retry.Do(func() error {
		return c.Set(ctx, key, value)
	}, strat)
}

func (c *Client) BatchWriter(ctx context.Context, in <-chan [2]string) {
	go func() {
		for pair := range in {
			_ = c.Set(ctx, pair[0], pair[1]) // Ошибки можно логировать
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
}
