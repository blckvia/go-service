package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	GetInt(ctx context.Context, key string) (int, error)
	SetInt(ctx context.Context, key string, value int, ttl time.Duration) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) GetInt(ctx context.Context, key string) (int, error) {
	value, err := r.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

func (r *RedisCache) SetInt(ctx context.Context, key string, value int, expiration time.Duration) error {
	return r.client.Set(ctx, key, strconv.Itoa(value), expiration).Err()
}
