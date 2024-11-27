package main

import (
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
)

func main() {
	log.Println("Starting Redis client...")
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	persistence := NewRedisPersistence(client)
	rateLimiter := NewRateLimiter(persistence)

	log.Println("Setting up HTTP handler...")
	http.Handle("/", RateLimiterMiddleware(rateLimiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})))

	log.Println("Starting HTTP server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
