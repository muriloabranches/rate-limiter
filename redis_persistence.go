package main

import (
	"context"
	"log"
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
	log.Printf("Getting value for key: %s", key)
	value, err := rp.client.Get(ctx, key).Int()
	if err != nil {
		log.Printf("Error getting value for key %s: %v", key, err)
	}
	return value, err
}

func (rp *RedisPersistence) Incr(ctx context.Context, key string) error {
	log.Printf("Incrementing value for key: %s", key)
	err := rp.client.Incr(ctx, key).Err()
	if err != nil {
		log.Printf("Error incrementing value for key %s: %v", key, err)
	}
	return err
}

func (rp *RedisPersistence) Expire(ctx context.Context, key string, expiration time.Duration) error {
	log.Printf("Setting expiration for key: %s to %v", key, expiration)
	err := rp.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		log.Printf("Error setting expiration for key %s: %v", key, err)
	}
	return err
}

func (rp *RedisPersistence) Set(ctx context.Context, key string, value int, expiration time.Duration) error {
	log.Printf("Setting value for key: %s to %d with expiration %v", key, value, expiration)
	err := rp.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Printf("Error setting value for key %s: %v", key, err)
	}
	return err
}

func (rp *RedisPersistence) TTL(ctx context.Context, key string) (time.Duration, error) {
	log.Printf("Getting TTL for key: %s", key)
	ttl, err := rp.client.TTL(ctx, key).Result()
	if err != nil {
		log.Printf("Error getting TTL for key %s: %v", key, err)
	}
	return ttl, err
}
