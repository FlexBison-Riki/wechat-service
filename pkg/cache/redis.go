package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/silenceper/wechat/v2/cache"
)

// RedisCache implements cache.Cache interface using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisCache{client: client}
}

// Get implements cache.Cache (no error return)
func (c *RedisCache) Get(key string) interface{} {
	val, _ := c.client.Get(context.Background(), key).Result()
	return val
}

// Set stores a value in cache with TTL
func (c *RedisCache) Set(key string, val interface{}, ttl time.Duration) error {
	return c.client.Set(context.Background(), key, val, ttl).Err()
}

// IsExist checks if a key exists
func (c *RedisCache) IsExist(key string) bool {
	result, _ := c.client.Exists(context.Background(), key).Result()
	return result > 0
}

// Delete removes a key from cache
func (c *RedisCache) Delete(key string) error {
	return c.client.Del(context.Background(), key).Err()
}

// MemoryCache is a simple in-memory cache using Redis (miniredis in tests)
type MemoryCache struct {
	client *redis.Client
}

// NewMemoryCache creates a memory cache for development
func NewMemoryCache() *MemoryCache {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	return &MemoryCache{client: client}
}

// Get implements cache.Cache
func (c *MemoryCache) Get(key string) interface{} {
	val, _ := c.client.Get(context.Background(), key).Result()
	return val
}

// Set stores a value in memory cache with TTL
func (c *MemoryCache) Set(key string, val interface{}, ttl time.Duration) error {
	return c.client.Set(context.Background(), key, val, ttl).Err()
}

// IsExist checks if a key exists
func (c *MemoryCache) IsExist(key string) bool {
	result, _ := c.client.Exists(context.Background(), key).Result()
	return result > 0
}

// Delete removes a key from memory cache
func (c *MemoryCache) Delete(key string) error {
	return c.client.Del(context.Background(), key).Err()
}

var _ cache.Cache = (*RedisCache)(nil)
var _ cache.Cache = (*MemoryCache)(nil)
