package services

import (
	"context"
	"fmt"
	"time"

	config "github.com/Neratus/geoguide/internal/repository/config"
	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(cfg *config.Config) (*RateLimiter, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Database.Redis.Host, cfg.Database.Redis.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Database.Redis.Password,
		DB:       cfg.Database.Redis.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &RateLimiter{client: client}, nil
}

func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		rl.client.Expire(ctx, key, window)
	}
	return count <= int64(limit), nil
}
