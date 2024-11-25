package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

func RateLimiterMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()
			ip := r.RemoteAddr
			token := r.Header.Get("API_KEY")

			var key string
			var limit int

			if token != "" {
				key = "token:" + token
				limit = rl.rateLimitToken
			} else {
				key = "ip:" + ip
				limit = rl.rateLimitIP
			}

			allowed, err := rl.Allow(ctx, key, limit)
			if err != nil {
				log.Printf("Error checking rate limit for key %s: %v", key, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			remaining, err := rl.client.Get(ctx, key).Int()
			if err != nil && err != redis.Nil {
				log.Printf("Error getting remaining requests for key %s: %v", key, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(limit-remaining))

			ttl, err := rl.client.TTL(ctx, key).Result()
			if err != nil {
				log.Printf("Error getting TTL for key %s: %v", key, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			resetTime := time.Now().Add(ttl).Unix()
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

			// resetTime := time.Now().Add(time.Second).Unix()
			// w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

			if !allowed {
				log.Printf("Rate limit exceeded for key %s", key)
				err := rl.Block(ctx, key, limit)
				if err != nil {
					log.Printf("Error blocking key %s: %v", key, err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				http.Error(w, "You have reached the maximum number of requests or actions allowed within a certain time frame. Please try again after "+strconv.FormatInt(resetTime, 10), http.StatusTooManyRequests)
				return
			}

			log.Printf("Request allowed for key %s", key)
			next.ServeHTTP(w, r)
		})
	}
}
