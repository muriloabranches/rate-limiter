package main

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	RateLimitIP    int
	RateLimitToken int
	BlockDuration  int
}

func LoadConfig() (*Config, error) {
	log.Println("Loading configuration")
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Printf("Error parsing REDIS_DB: %v", err)
		return nil, err
	}
	rateLimitIP, err := strconv.Atoi(os.Getenv("RATE_LIMIT_IP"))
	if err != nil {
		log.Printf("Error parsing RATE_LIMIT_IP: %v", err)
		return nil, err
	}
	rateLimitToken, err := strconv.Atoi(os.Getenv("RATE_LIMIT_TOKEN"))
	if err != nil {
		log.Printf("Error parsing RATE_LIMIT_TOKEN: %v", err)
		return nil, err
	}
	blockDuration, err := strconv.Atoi(os.Getenv("BLOCK_DURATION"))
	if err != nil {
		log.Printf("Error parsing BLOCK_DURATION: %v", err)
		return nil, err
	}

	log.Println("Configuration loaded successfully")
	return &Config{
		RedisAddr:      os.Getenv("REDIS_ADDR"),
		RedisPassword:  os.Getenv("REDIS_PASSWORD"),
		RedisDB:        db,
		RateLimitIP:    rateLimitIP,
		RateLimitToken: rateLimitToken,
		BlockDuration:  blockDuration,
	}, nil
}
