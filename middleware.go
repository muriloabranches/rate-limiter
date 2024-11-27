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
	persistence     Persistence
	ipRateLimit     int
	tokenRateLimit  int
	blockDuration   time.Duration
	rateLimitWindow time.Duration
}

func NewRateLimiter(persistence Persistence) *RateLimiter {
	godotenv.Load()

	ipRateLimit, _ := strconv.Atoi(os.Getenv("IP_RATE_LIMIT"))
	tokenRateLimit, _ := strconv.Atoi(os.Getenv("TOKEN_RATE_LIMIT"))
	blockDuration, _ := strconv.Atoi(os.Getenv("BLOCK_DURATION"))
	rateLimitWindow, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_WINDOW"))

	log.Printf("Initializing RateLimiter with IP rate limit: %d, Token rate limit: %d, Block duration: %d seconds, Rate limit window: %d seconds", ipRateLimit, tokenRateLimit, blockDuration, rateLimitWindow)

	return &RateLimiter{
		persistence:     persistence,
		ipRateLimit:     ipRateLimit,
		tokenRateLimit:  tokenRateLimit,
		blockDuration:   time.Duration(blockDuration) * time.Second,
		rateLimitWindow: time.Duration(rateLimitWindow) * time.Second,
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
			log.Printf("Error checking TTL for blocked key %s: %v", key, err)
			return 0, 0, false
		}
		log.Printf("Key %s is blocked, retry after: %s", key, ttl)
		return 0, ttl, false
	}

	val, err := rl.persistence.Get(ctx, key)
	if err != nil && err != redis.Nil {
		log.Printf("Error getting value for key %s: %v", key, err)
		return 0, 0, false
	}

	ttl, err := rl.persistence.TTL(ctx, key)
	if err != nil {
		log.Printf("Error checking TTL for key %s: %v", key, err)
		return 0, 0, false
	}

	if ttl > 0 && val >= limit {
		rl.blockKey(key)
		log.Printf("Key %s has exceeded the limit, blocking for %s", key, rl.blockDuration)
		return 0, rl.blockDuration, false
	}

	if ttl <= 0 {
		rl.persistence.Set(ctx, key, 0, rl.rateLimitWindow)
		val = 0
		ttl = rl.rateLimitWindow
	}

	rl.persistence.Incr(ctx, key)
	rl.persistence.Expire(ctx, key, rl.rateLimitWindow)
	remaining := limit - val - 1
	reset := ttl

	if remaining < 0 {
		rl.blockKey(key)
		log.Printf("Key %s has exceeded the limit after increment, blocking for %s", key, rl.blockDuration)
		return 0, rl.blockDuration, false
	}

	log.Printf("Key %s is allowed, remaining: %d, reset after: %s", key, remaining, reset)
	return remaining, reset, true
}

func (rl *RateLimiter) isBlocked(key string) bool {
	val, err := rl.persistence.Get(ctx, key)
	if err != nil && err != redis.Nil {
		log.Printf("Error getting value for key %s: %v", key, err)
		return false
	}

	ttl, err := rl.persistence.TTL(ctx, key)
	if err != nil {
		log.Printf("Error checking TTL for key %s: %v", key, err)
		return false
	}

	isBlocked := ttl > 0 && val == -1
	if isBlocked {
		log.Printf("Key %s is currently blocked", key)
	}
	return isBlocked
}

func (rl *RateLimiter) blockKey(key string) {
	log.Printf("Blocking key %s for %s", key, rl.blockDuration)
	rl.persistence.Set(ctx, key, -1, rl.blockDuration)
}
