package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrKeyNotFound is returned when the requested cache key does not exist.
var ErrKeyNotFound = errors.New("cache key not found")

// ClientInterface allows mocking cache operations in tests.
type ClientInterface interface {
	Set(ctx context.Context, addr, password string, opts SetOptions) error
	Get(ctx context.Context, addr, password, key string) (string, error)
}

// Client performs Redis cache operations (SET/GET) against a given address.
// It is stateless; each operation uses the provided endpoint and password.
type Client struct{}

// Ensure Client implements ClientInterface.
var _ ClientInterface = (*Client)(nil)

// NewClient returns a cache client (no connection pooling; we connect per request per instance).
func NewClient() *Client {
	return &Client{}
}

// SetOptions configures a SET operation.
type SetOptions struct {
	Key        string
	Value      string
	TTLSeconds int // 0 means no expiry
}

// Set stores a key-value pair in Redis at the given address.
// If TTLSeconds > 0, the key will expire after that many seconds.
func (c *Client) Set(ctx context.Context, addr, password string, opts SetOptions) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	defer rdb.Close()

	var expiration time.Duration
	if opts.TTLSeconds > 0 {
		expiration = time.Duration(opts.TTLSeconds) * time.Second
	}
	if err := rdb.Set(ctx, opts.Key, opts.Value, expiration).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

// Get retrieves the value for key from Redis at the given address.
// Returns ErrKeyNotFound if the key does not exist.
func (c *Client) Get(ctx context.Context, addr, password, key string) (string, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	defer rdb.Close()

	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("redis get: %w", err)
	}
	return val, nil
}
