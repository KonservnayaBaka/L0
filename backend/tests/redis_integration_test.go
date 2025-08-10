package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"

	"wbL0/internal/config"
	redisPkg "wbL0/internal/db/redis"
)

func TestRedis_ConnectionSetGet(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("dockertest.NewPool: %v", err)
	}

	resource, err := pool.Run("redis", "7-alpine", nil)
	if err != nil {
		t.Fatalf("could not start redis: %v", err)
	}
	defer func() { _ = pool.Purge(resource) }()

	var redisHostPort string
	if err := pool.Retry(func() error {
		redisHostPort = resource.GetPort("6379/tcp")
		rdb := redis.NewClient(&redis.Options{Addr: "localhost:" + redisHostPort})
		defer rdb.Close()
		return rdb.Ping(context.Background()).Err()
	}); err != nil {
		t.Fatalf("redis did not start: %v", err)
	}

	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: func() int {
				p, _ := resource.GetPort("6379/tcp"), 0
				_ = p
				return 6379
			}(),
			Password: "",
			DB:       0,
			TTL:      1 * time.Hour,
		},
	}

	var portInt int
	if _, err := fmt.Sscanf(redisHostPort, "%d", &portInt); err == nil {
		cfg.Redis.Port = portInt
	}

	rdb := redisPkg.NewRedisClient(context.Background(), cfg)
	defer func() { _ = rdb.Close() }()

	ctx := context.Background()
	if err := rdb.Set(ctx, "integration:test:key", "value", 10*time.Second).Err(); err != nil {
		t.Fatalf("redis SET failed: %v", err)
	}

	val, err := rdb.Get(ctx, "integration:test:key").Result()
	if err != nil {
		t.Fatalf("redis GET failed: %v", err)
	}
	if val != "value" {
		t.Fatalf("unexpected value: %s", val)
	}
}
