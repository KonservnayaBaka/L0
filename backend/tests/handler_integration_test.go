package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log/slog"

	"github.com/gin-gonic/gin"
	"wbL0/internal/config"
	"wbL0/internal/db/postgres"
	redisPkg "wbL0/internal/db/redis"
	orderHandler "wbL0/internal/http/handler/orderHandler"
	"wbL0/internal/http/routes"
	"wbL0/internal/models"
	orderRepoPostgres "wbL0/internal/repository/postgres/orderRepoPostgres"
	orderRepoRedis "wbL0/internal/repository/redis/orderRepoRedis"
	svc "wbL0/internal/service/orderService"
)

func startPgAndRedis(t *testing.T, pool *dockertest.Pool) (cfg *config.Config, pgPool *pgxpool.Pool, rdb *redis.Client, cleanup func()) {
	pgOpts := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=1234",
			"POSTGRES_DB=wbL0",
		},
	}
	pgRes, err := pool.RunWithOptions(pgOpts, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}

	redisRes, err := pool.Run("redis", "7-alpine", nil)
	if err != nil {
		_ = pool.Purge(pgRes)
		t.Fatalf("start redis: %v", err)
	}

	cleanup = func() {
		_ = pool.Purge(pgRes)
		_ = pool.Purge(redisRes)
	}

	var pgPortStr, rPortStr string
	if err := pool.Retry(func() error {
		pgPortStr = pgRes.GetPort("5432/tcp")
		connStr := fmt.Sprintf("postgres://postgres:1234@localhost:%s/postgres?sslmode=disable", pgPortStr)
		db, err := pgxpool.New(context.Background(), connStr)
		if err != nil {
			return err
		}
		defer db.Close()
		return db.Ping(context.Background())
	}); err != nil {
		cleanup()
		t.Fatalf("postgres didn't come up: %v", err)
	}
	if err := pool.Retry(func() error {
		rPortStr = redisRes.GetPort("6379/tcp")
		rdb := redis.NewClient(&redis.Options{Addr: "localhost:" + rPortStr})
		defer rdb.Close()
		return rdb.Ping(context.Background()).Err()
	}); err != nil {
		cleanup()
		t.Fatalf("redis didn't come up: %v", err)
	}

	pgPort, _ := strconv.Atoi(pgPortStr)
	redisPort, _ := strconv.Atoi(rPortStr)

	cfg = &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     pgPort,
			User:     "postgres",
			Password: "1234",
			DBName:   "wbL0",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: redisPort,
			DB:   0,
			TTL:  24 * time.Hour,
		},
		Server: config.ServerConfig{
			Port:    0,
			Timeout: 5 * time.Second,
		},
	}

	origWd, err := os.Getwd()
	if err != nil {
		cleanup()
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	if err := os.Chdir(".."); err != nil {
		cleanup()
		t.Fatalf("failed to chdir to project root: %v", err)
	}

	pgPool = postgres.MustLoad(context.Background(), cfg)
	rdb = redisPkg.NewRedisClient(context.Background(), cfg)

	return cfg, pgPool, rdb, cleanup
}

func TestHandler_GetOrderInfo(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("dockertest.NewPool: %v", err)
	}

	cfg, pgPool, rdb, cleanup := startPgAndRedis(t, pool)
	defer cleanup()
	defer pgPool.Close()
	defer func() { _ = rdb.Close() }()

	logger := slog.Default()
	pgRepo := orderRepoPostgres.NewPostgresRepository(pgPool, logger)
	redisRepo := orderRepoRedis.NewRedisRepo(rdb, logger)
	service := svc.NewOrderService(pgRepo, redisRepo, logger, cfg.Redis.TTL)

	fullOrder := &models.FullOrder{
		Order: models.Order{
			OrderUID:    "handler-uid-1",
			TrackNumber: "track-1",
			Entry:       "entry",
			Locale:      "ru",
			CustomerID:  "cust",
			DateCreated: time.Now(),
		},
		Delivery: models.Delivery{
			OrderUID: "handler-uid-1",
			Name:     "John",
			Phone:    "+700",
			Zip:      1,
			City:     "City",
			Address:  "Addr",
			Region:   "R",
			Email:    "e@mail",
		},
		Payment: models.Payment{
			OrderUID:     "handler-uid-1",
			Transaction:  "tx",
			RequestID:    "req",
			Currency:     "RUB",
			Provider:     "p",
			Amount:       10,
			PaymentDt:    time.Now(),
			Bank:         "bank",
			DeliveryCost: 1,
			GoodsTotal:   9,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				OrderUID:    "handler-uid-1",
				ChrtID:      1,
				TrackNumber: "track-1",
				Price:       10,
				Rid:         "r",
				Name:        "it",
				Sale:        0,
				Size:        "L",
				TotalPrice:  10,
				NmID:        1,
				Brand:       "b",
				Status:      1,
			},
		},
	}

	ctx := context.Background()
	if err := service.ProcessAndCache(ctx, fullOrder); err != nil {
		t.Fatalf("ProcessAndCache failed: %v", err)
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := orderHandler.NewOrderHandler(service, logger)
	routes.InitRoutes(engine, *handler)

	ts := httptest.NewServer(engine)
	defer ts.Close()

	url := fmt.Sprintf("%s/order/%s", ts.URL, fullOrder.Order.OrderUID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("http.Get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var got models.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if got.OrderUID != fullOrder.Order.OrderUID {
		t.Fatalf("unexpected order uid: got=%s want=%s", got.OrderUID, fullOrder.Order.OrderUID)
	}
	if len(got.Items) != 1 {
		t.Fatalf("unexpected items count: %d", len(got.Items))
	}
}
