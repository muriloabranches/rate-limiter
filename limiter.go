package main

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	client         *redis.Client
	rateLimitIP    int
	rateLimitToken int
	blockDuration  time.Duration
}

func NewRateLimiter(client *redis.Client, rateLimitIP, rateLimitToken int, blockDuration time.Duration) *RateLimiter {
	log.Println("Creating new RateLimiter")
	return &RateLimiter{
		client:         client,
		rateLimitIP:    rateLimitIP,
		rateLimitToken: rateLimitToken,
		blockDuration:  blockDuration,
	}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int) (bool, error) {
	log.Printf("Checking rate limit for key: %s", key)
	val, err := rl.client.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		log.Printf("Error getting value for key %s: %v", key, err)
		return false, err
	}

	if val >= limit {
		log.Printf("Rate limit exceeded for key: %s", key)
		return false, nil
	}

	pipe := rl.client.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Second)
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("Error incrementing value for key %s: %v", key, err)
		return false, err
	}

	log.Printf("Request allowed for key: %s", key)
	return true, nil
}

func (rl *RateLimiter) Block(ctx context.Context, key string, limit int) error {
	log.Printf("Blocking key: %s", key)
	ttl, err := rl.client.TTL(ctx, key).Result()
	if err != nil {
		log.Printf("Error getting TTL for key %s: %v", key, err)
		return err
	}

	if ttl > 0 {
		log.Printf("Key %s is already blocked, TTL: %v", key, ttl)
		return nil
	}

	log.Printf("Setting block duration for key %s: %v", key, rl.blockDuration)
	return rl.client.Set(ctx, key, limit, rl.blockDuration).Err()
}
