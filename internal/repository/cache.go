package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is a thin Redis-backed key/value cache with JSON (de)serialisation.
// Permissions and user identities are cached here to cut DB load on hot paths.
type Cache struct {
	rdb *redis.Client
}

// NewCache returns a Cache bound to the given Redis client.
func NewCache(rdb *redis.Client) *Cache { return &Cache{rdb: rdb} }

// Available reports whether the cache client is configured.
func (c *Cache) Available() bool { return c.rdb != nil }

// Set serialises v as JSON and stores it with a TTL.
func (c *Cache) Set(ctx context.Context, key string, v any, ttl time.Duration) error {
	if c.rdb == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("cache marshal %s: %w", key, err)
	}
	return c.rdb.Set(ctx, key, b, ttl).Err()
}

// Get deserialises the value at key into v. Returns false (no error) on a miss.
func (c *Cache) Get(ctx context.Context, key string, v any) (bool, error) {
	if c.rdb == nil {
		return false, nil
	}
	b, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return false, fmt.Errorf("cache unmarshal %s: %w", key, err)
	}
	return true, nil
}

// Del removes one or more keys.
func (c *Cache) Del(ctx context.Context, keys ...string) error {
	if c.rdb == nil {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}

// Incr increments a counter key (used for rate limiting). It creates the key
// at 1 on first use with the given TTL.
func (c *Cache) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if c.rdb == nil {
		return 0, nil
	}
	n, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 {
		_ = c.rdb.Expire(ctx, key, ttl).Err()
	}
	return n, nil
}

// SetNX is the cache-based idempotency primitive: it stores key=v only if the
// key does not already exist, returning true if it acquired the slot.
func (c *Cache) SetNX(ctx context.Context, key string, v any, ttl time.Duration) (bool, error) {
	if c.rdb == nil {
		return true, nil // no Redis => fail open
	}
	b, err := json.Marshal(v)
	if err != nil {
		return false, err
	}
	return c.rdb.SetNX(ctx, key, b, ttl).Result()
}

// Cache key builders, centralised so callers agree on format.
const (
	keyPermPrefix = "perms:user:"
	keyUserPrefix = "user:"
	keyIdemPrefix = "idem:"
	keyRatePrefix = "rate:"
)

func PermKey(userID int64) string       { return fmt.Sprintf("%s%d", keyPermPrefix, userID) }
func UserKey(userID int64) string       { return fmt.Sprintf("%s%d", keyUserPrefix, userID) }
func IdemKey(k string) string           { return keyIdemPrefix + k }
func RateKey(kind, ident string) string { return fmt.Sprintf("%s%s:%s", keyRatePrefix, kind, ident) }
