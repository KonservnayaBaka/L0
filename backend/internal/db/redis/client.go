package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"wbL0/internal/config"
)

func NewRedisClient(ctx context.Context, cfg *config.Config) *redis.Client {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("failed to connect to redis: %w", err))
	}
	return rdb
}
