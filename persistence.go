package main

import (
	"context"
	"time"
)

type Persistence interface {
	Get(ctx context.Context, key string) (int, error)
	Incr(ctx context.Context, key string) error
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Set(ctx context.Context, key string, value int, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
}
