package main

import (
	"net/http"

	"github.com/go-redis/redis/v8"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	persistence := NewRedisPersistence(client)
	rateLimiter := NewRateLimiter(persistence)

	http.Handle("/", rateLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})))

	http.ListenAndServe(":8080", nil)
}
