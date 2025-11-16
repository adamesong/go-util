package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// go-redis v9 usage: https://redis.io/docs/clients/go/

// RedisClient is a wrapper for the go-redis client.
// It holds a long-lived client that manages a connection pool.
type RedisClient struct {
	Client *redis.Client
}

var ctx = context.Background()

// NewRedisClient creates a new RedisClient.
// It establishes a connection and pings the server to ensure it's alive.
func NewRedisClient(addr string, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test the connection
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis connection test failed: %w", err)
	}

	return &RedisClient{Client: client}, nil
}

// Close closes the underlying redis client and release resources.
func (r *RedisClient) Close() error {
	return r.Client.Close()
}

// Keys finds all keys matching the given pattern.
// Warning: KEYS can be slow on a large database.
func (r *RedisClient) Keys(pattern string) ([]string, error) {
	return r.Client.Keys(ctx, pattern).Result()
}

// Set sets a key-value pair. A duration of 0 means no expiration.
func (r *RedisClient) Set(key string, value interface{}, duration time.Duration) error {
	return r.Client.Set(ctx, key, value, duration).Err()
}

// Get retrieves a value by key. Returns redis.Nil error if key does not exist.
func (r *RedisClient) Get(key string) ([]byte, error) {
	return r.Client.Get(ctx, key).Bytes()
}

// MGet retrieves multiple values by keys.
func (r *RedisClient) MGet(keys ...string) ([]interface{}, error) {
	return r.Client.MGet(ctx, keys...).Result()
}

// SetNX sets a key-value pair only if the key does not exist.
func (r *RedisClient) SetNX(key string, value interface{}, duration time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, key, value, duration).Result()
}

// SetXX sets a key-value pair only if the key already exists.
func (r *RedisClient) SetXX(key string, value interface{}, duration time.Duration) (bool, error) {
	return r.Client.SetXX(ctx, key, value, duration).Result()
}

// Exists checks if one or more keys exist.
func (r *RedisClient) Exists(key ...string) (int64, error) {
	return r.Client.Exists(ctx, key...).Result()
}

// GetSet sets a new value for a key and returns the old value.
func (r *RedisClient) GetSet(key string, value interface{}) ([]byte, error) {
	return r.Client.GetSet(ctx, key, value).Bytes()
}

// Delete deletes one or more keys. Returns the number of keys that were removed.
func (r *RedisClient) Delete(keys ...string) (int64, error) {
	return r.Client.Del(ctx, keys...).Result()
}

// TTL returns the remaining time to live of a key.
func (r *RedisClient) TTL(key string) (time.Duration, error) {
	return r.Client.TTL(ctx, key).Result()
}

// Expire sets a new expiration for a key.
func (r *RedisClient) Expire(key string, duration time.Duration) (bool, error) {
	return r.Client.Expire(ctx, key, duration).Result()
}

// LikeDeletes deletes keys matching a pattern. Warning: KEYS can be slow in production.
func (r *RedisClient) LikeDeletes(key string) (int64, error) {
	keys, err := r.Client.Keys(ctx, "*"+key+"*").Result()
	if err != nil {
		return 0, err
	}
	if len(keys) == 0 {
		return 0, nil
	}
	return r.Delete(keys...)
}

// RPush appends one or more values to a list.
func (r *RedisClient) RPush(key string, values ...interface{}) (int64, error) {
	return r.Client.RPush(ctx, key, values...).Result()
}

// RPushX appends a value to a list, only if the list exists.
func (r *RedisClient) RPushX(key string, value interface{}) (int64, error) {
	return r.Client.RPushX(ctx, key, value).Result()
}

// Z is a type alias for redis.Z, representing a member in a sorted set.
type Z = redis.Z

// ZAdd adds one or more members to a sorted set.
func (r *RedisClient) ZAdd(key string, members ...Z) (int64, error) {
	return r.Client.ZAdd(ctx, key, members...).Result()
}

// ZRange returns a range of members from a sorted set, by index.
func (r *RedisClient) ZRange(key string, start, stop int64) ([]string, error) {
	return r.Client.ZRange(ctx, key, start, stop).Result()
}

// ZRevRange returns a range of members from a sorted set, by index, in reverse order.
func (r *RedisClient) ZRevRange(key string, start, stop int64) ([]string, error) {
	return r.Client.ZRevRange(ctx, key, start, stop).Result()
}

// HMSet sets multiple hash fields to multiple values.
func (r *RedisClient) HMSet(key string, fields map[string]interface{}) (bool, error) {
	// Note: HMSet is deprecated in Redis 4.0.0. Consider using HSet with multiple field-value pairs.
	return r.Client.HMSet(ctx, key, fields).Result()
}

// MSet sets multiple key-value pairs.
func (r *RedisClient) MSet(pairs ...interface{}) (string, error) {
	return r.Client.MSet(ctx, pairs...).Result()
}

// MSetNX sets multiple key-value pairs, only if none of the keys exist.
func (r *RedisClient) MSetNX(pairs ...interface{}) (bool, error) {
	return r.Client.MSetNX(ctx, pairs...).Result()
}

// MExpire sets an expiration for multiple keys using a pipeline.
func (r *RedisClient) MExpire(keys []string, duration time.Duration) error {
	pl := r.Client.Pipeline()
	for _, key := range keys {
		pl.Expire(ctx, key, duration)
	}
	_, err := pl.Exec(ctx)
	return err
}

// PFAdd adds elements to a HyperLogLog.
func (r *RedisClient) PFAdd(key string, els ...interface{}) (int64, error) {
	return r.Client.PFAdd(ctx, key, els...).Result()
}

// PFCount returns the approximate cardinality of the set observed by the HyperLogLog.
func (r *RedisClient) PFCount(keys ...string) (int64, error) {
	return r.Client.PFCount(ctx, keys...).Result()
}

// MPFCount counts the cardinality of multiple HyperLogLogs using a pipeline.
func (r *RedisClient) MPFCount(keys []string) (map[string]int64, error) {
	resultMap := make(map[string]int64)
	pl := r.Client.Pipeline()
	cmds := make([]*redis.IntCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pl.PFCount(ctx, key)
	}
	_, err := pl.Exec(ctx)
	// redis.Nil is not an error from Exec, so no need to check for it here.
	if err != nil {
		return nil, err
	}

	for i, cmd := range cmds {
		n, err := cmd.Result()
		if err != nil {
			// A single command can fail. The original code logged and continued.
			// We will do the same by skipping the failed command.
			continue
		}
		resultMap[keys[i]] = n
	}
	return resultMap, nil
}

// Incr increments the integer value of a key by one.
func (r *RedisClient) Incr(key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}
