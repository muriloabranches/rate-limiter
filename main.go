package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	rl := NewRateLimiter(client, cfg.RateLimitIP, cfg.RateLimitToken, time.Duration(cfg.BlockDuration)*time.Second)
	log.Println("Rate limiter configured")

	r := mux.NewRouter()
	r.Use(RateLimiterMiddleware(rl))

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	log.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", r)
}
