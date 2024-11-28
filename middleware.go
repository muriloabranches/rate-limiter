package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

func RateLimiterMiddleware(rl *RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
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

		remaining, reset, allowed := rl.AllowRequest(key, limit)
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
