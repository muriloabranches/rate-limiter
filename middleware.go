package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var ctx = context.Background()

type RateLimiter struct {
	persistence    Persistence
	ipRateLimit    int
	tokenRateLimit int
	blockDuration  time.Duration
}

func NewRateLimiter(persistence Persistence) *RateLimiter {
	godotenv.Load()

	ipRateLimit, _ := strconv.Atoi(os.Getenv("IP_RATE_LIMIT"))
	tokenRateLimit, _ := strconv.Atoi(os.Getenv("TOKEN_RATE_LIMIT"))
	blockDuration, _ := strconv.Atoi(os.Getenv("BLOCK_DURATION"))

	return &RateLimiter{
		persistence:    persistence,
		ipRateLimit:    ipRateLimit,
		tokenRateLimit: tokenRateLimit,
		blockDuration:  time.Duration(blockDuration) * time.Second,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		token := r.Header.Get("API_KEY")
		var key string
		var limit int

		if token != "" {
			key = token
			limit = rl.tokenRateLimit
		} else {
			key = ip
			limit = rl.ipRateLimit
		}

		remaining, reset, allowed := rl.allowRequest(key, limit)
		if !allowed {
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
			log.Printf("Blocked request from %s (token: %s), retry after: %s", ip, token, reset)
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(reset.Seconds())))
		log.Printf("Allowed request from %s (token: %s), remaining: %d", ip, token, remaining)
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allowRequest(key string, limit int) (int, time.Duration, bool) {
	if rl.isBlocked(key) {
		ttl, err := rl.persistence.TTL(ctx, key)
		if err != nil {
			return 0, 0, false
		}
		return 0, ttl, false
	}

	val, err := rl.persistence.Get(ctx, key)
	if err != nil && err != redis.Nil {
		return 0, 0, false
	}

	ttl, err := rl.persistence.TTL(ctx, key)
	if err != nil {
		return 0, 0, false
	}

	if ttl > 0 && val >= limit {
		rl.blockKey(key)
		return 0, rl.blockDuration, false
	}

	if ttl <= 0 {
		rl.persistence.Set(ctx, key, 0, time.Second)
		val = 0
	}

	rl.persistence.Incr(ctx, key)
	rl.persistence.Expire(ctx, key, time.Second)
	remaining := limit - val - 1
	reset := time.Second

	if remaining < 0 {
		rl.blockKey(key)
		return 0, rl.blockDuration, false
	}

	return remaining, reset, true
}

func (rl *RateLimiter) isBlocked(key string) bool {
	val, err := rl.persistence.Get(ctx, key)
	if err != nil && err != redis.Nil {
		return false
	}

	ttl, err := rl.persistence.TTL(ctx, key)
	if err != nil {
		return false
	}

	return ttl > 0 && val == -1
}

func (rl *RateLimiter) blockKey(key string) {
	rl.persistence.Set(ctx, key, -1, rl.blockDuration)
}
