package main

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisPersistence struct {
	client *redis.Client
}

func NewRedisPersistence(client *redis.Client) *RedisPersistence {
	return &RedisPersistence{client: client}
}

func (rp *RedisPersistence) Get(ctx context.Context, key string) (int, error) {
	return rp.client.Get(ctx, key).Int()
}

func (rp *RedisPersistence) Incr(ctx context.Context, key string) error {
	return rp.client.Incr(ctx, key).Err()
}

func (rp *RedisPersistence) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return rp.client.Expire(ctx, key, expiration).Err()
}

func (rp *RedisPersistence) Set(ctx context.Context, key string, value int, expiration time.Duration) error {
	return rp.client.Set(ctx, key, value, expiration).Err()
}

func (rp *RedisPersistence) TTL(ctx context.Context, key string) (time.Duration, error) {
	return rp.client.TTL(ctx, key).Result()
}
