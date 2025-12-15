// Package redis provides a client wrapper for Redis operations.
package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/wb-go/wbf/retry"
)

// NoMatches is returned when Redis did not find any matching key.
const NoMatches = redis.Nil

// Validation errors.
var (
	ErrAddressRequired = errors.New("redis address is required")
	ErrInvalidMemory   = errors.New("invalid maxmemory format")
	ErrInvalidPolicy   = errors.New("invalid maxmemory-policy")
)

// Client wraps the Redis client.
type Client struct {
	*redis.Client
}

// Options contains configuration for Redis connection.
type Options struct {
	Address   string // Redis server address (host:port)
	Password  string // Redis password (optional)
	MaxMemory string // Max memory limit (e.g., "100mb", "1gb")
	Policy    string // Memory eviction policy
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

// Connect creates a new Redis client with validated options.
func Connect(options Options) (*Client, error) {
	if err := validateOptions(options); err != nil {
		return nil, err
	}
	client := &Client{
		redis.NewClient(&redis.Options{
			Addr:     options.Address,
			Password: options.Password,
		}),
	}
	ctx := context.Background()
	client.ConfigSet(ctx, "maxmemory", options.MaxMemory)
	client.ConfigSet(ctx, "maxmemory-policy", options.Policy)
	return client, client.Ping(ctx)
}

// validateOptions validates Redis connection options.
func validateOptions(options Options) error {
	if options.Address == "" {
		return ErrAddressRequired
	}
	if options.MaxMemory != "" {
		maxMem := strings.ToLower(options.MaxMemory)
		if !strings.HasSuffix(maxMem, "mb") && !strings.HasSuffix(maxMem, "gb") {
			return ErrInvalidMemory
		}
	}
	if options.Policy != "" {
		switch options.Policy {
		case "noeviction", "allkeys-lru", "volatile-lru",
			"allkeys-random", "volatile-random", "volatile-ttl":
		default:
			return ErrInvalidPolicy
		}
	}
	return nil
}

// Ping tests the Redis connection.
func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}

// Get retrieves a value by key from Redis.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Set stores a value by key in Redis.
func (c *Client) Set(ctx context.Context, key string, value any) error {
	return c.Client.Set(ctx, key, value, 0).Err()
}

// SetWithExpiration stores a value with a specified expiration time.
func (c *Client) SetWithExpiration(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// SetWithExpirationAndRetry stores a value with expiration using a retry strategy.
func (c *Client) SetWithExpirationAndRetry(ctx context.Context, strategy retry.Strategy,
	key string, value any, expiration time.Duration) error {
	return retry.DoContext(ctx, strategy, func() error {
		return c.Client.Set(ctx, key, value, expiration).Err()
	})
}

// Expire sets the expiration time for a key.
// If expiration is 0, the key will not expire.
// If expiration is negative, the key will be deleted immediately.
// Returns an error if the operation fails.
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.Client.Expire(ctx, key, expiration).Err()
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
func (c *Client) SetWithRetry(ctx context.Context, strategy retry.Strategy, key string, value any) error {
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

// Close closes the client, releasing any open resources.
func (c *Client) Close() error {
	return c.Client.Close()
}
