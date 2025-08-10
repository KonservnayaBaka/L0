package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log/slog"
	"strconv"

	"wbL0/internal/config"
	"wbL0/internal/db/postgres"
	"wbL0/internal/models"
	orderRepoPostgres "wbL0/internal/repository/postgres/orderRepoPostgres"
)

func startPostgresForTest(t *testing.T, pool *dockertest.Pool) (host string, port int, cleanup func()) {
	opts := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=1234",
			"POSTGRES_DB=wbL0",
		},
	}
	resource, err := pool.RunWithOptions(opts, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("could not start postgres container: %v", err)
	}

	cleanup = func() { _ = pool.Purge(resource) }

	var hostPort string
	if err := pool.Retry(func() error {
		hostPort = resource.GetPort("5432/tcp")
		connStr := fmt.Sprintf("postgres://postgres:1234@localhost:%s/postgres?sslmode=disable", hostPort)
		db, err := sql.Open("pgx", connStr)
		if err != nil {
			return err
		}
		defer db.Close()
		return db.Ping()
	}); err != nil {
		cleanup()
		t.Fatalf("postgres did not start: %v", err)
	}

	var portInt int
	if p, err := strconv.Atoi(hostPort); err == nil {
		portInt = p
	} else {
		t.Fatalf("failed to parse postgres host port: %v", err)
	}

	return "localhost", portInt, cleanup
}

func TestPostgres_SaveAndGetFullOrder(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("dockertest.NewPool: %v", err)
	}

	host, port, cleanup := startPostgresForTest(t, pool)
	defer cleanup()

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     host,
			Port:     port,
			User:     "postgres",
			Password: "1234",
			DBName:   "wbL0",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{TTL: 1 * time.Hour},
	}

	restoreWd := EnsureRepoRoot(t)
	defer restoreWd()

	ctx := context.Background()
	poolDB := postgres.MustLoad(ctx, cfg)
	defer poolDB.Close()

	logger := slog.Default()
	repo := orderRepoPostgres.NewPostgresRepository(poolDB, logger)

	order := models.Order{
		OrderUID:    "pg-test-uid-1",
		TrackNumber: "t-1",
		Entry:       "entry",
		Locale:      "ru",
		CustomerID:  "cust",
		DateCreated: time.Now(),
	}
	delivery := models.Delivery{
		OrderUID: order.OrderUID,
		Name:     "X",
		Phone:    "+7000",
		Zip:      123,
		City:     "City",
		Address:  "Addr",
		Region:   "R",
		Email:    "e@mail",
	}
	payment := models.Payment{
		OrderUID:     order.OrderUID,
		Transaction:  "tx",
		RequestID:    "r",
		Currency:     "RUB",
		Provider:     "p",
		Amount:       1,
		PaymentDt:    time.Now(),
		Bank:         "bank",
		DeliveryCost: 1,
		GoodsTotal:   0,
		CustomFee:    0,
	}
	item := models.Item{
		OrderUID:    order.OrderUID,
		ChrtID:      1,
		TrackNumber: "t-1",
		Price:       1,
		Rid:         "rid",
		Name:        "it",
		Sale:        0,
		Size:        "M",
		TotalPrice:  1,
		NmID:        10,
		Brand:       "b",
		Status:      1,
	}

	tx, err := repo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}
	defer tx.Rollback(ctx)

	if err := repo.SaveOrderDataTx(ctx, tx, &order); err != nil {
		t.Fatalf("SaveOrderDataTx failed: %v", err)
	}
	if err := repo.SaveDeliveryDataTx(ctx, tx, &delivery); err != nil {
		t.Fatalf("SaveDeliveryDataTx failed: %v", err)
	}
	if err := repo.SavePaymentDataTx(ctx, tx, &payment); err != nil {
		t.Fatalf("SavePaymentDataTx failed: %v", err)
	}
	if err := repo.SaveItemsDataTx(ctx, tx, &item); err != nil {
		t.Fatalf("SaveItemsDataTx failed: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("tx commit failed: %v", err)
	}

	got, err := repo.GetFullOrderByUID(ctx, order.OrderUID)
	if err != nil {
		t.Fatalf("GetFullOrderByUID failed: %v", err)
	}
	if got.Order.OrderUID != order.OrderUID {
		t.Fatalf("unexpected order uid got=%s want=%s", got.Order.OrderUID, order.OrderUID)
	}
	if len(got.Items) != 1 {
		t.Fatalf("unexpected items count: %d", len(got.Items))
	}
}
