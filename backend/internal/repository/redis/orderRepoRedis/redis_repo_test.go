package orderRepoRedis_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	//"github.com/go-redis/redis/v8"
	redismock "github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"log/slog"

	"wbL0/internal/models"
	orderRepoRedis "wbL0/internal/repository/redis/orderRepoRedis"
)

func TestOrderRedisRepo_GetSet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()

	logger := slog.Default()
	repo := orderRepoRedis.NewRedisRepo(rdb, logger)

	// case 1: key missing -> should return (nil, nil)
	mock.ExpectGet("order:missing").RedisNil()

	got, err := repo.GetOrder(ctx, "missing")
	assert.NoError(t, err)
	assert.Nil(t, got)
	assert.NoError(t, mock.ExpectationsWereMet())

	// case 2: SetOrder + GetOrder returns stored JSON
	fo := &models.FullOrder{
		Order: models.Order{
			OrderUID:    "r1",
			TrackNumber: "t123",
			Entry:       "e",
			Locale:      "ru",
			CustomerID:  "c1",
		},
		Delivery: models.Delivery{
			OrderUID: "r1",
			Name:     "N",
			Phone:    "+7",
			Zip:      123,
			City:     "City",
			Address:  "Addr",
			Region:   "R",
			Email:    "e@mail",
		},
	}

	// prepare expected JSON
	data, err := json.Marshal(fo)
	assert.NoError(t, err)

	ttl := 10 * time.Minute

	// Expect Set to be called with exact bytes and provided TTL
	mock.ExpectSet("order:r1", data, ttl).SetVal("OK")
	// And then Get returns the same bytes
	mock.ExpectGet("order:r1").SetVal(string(data))

	// Call SetOrder
	err = repo.SetOrder(ctx, fo, ttl)
	assert.NoError(t, err)

	// Call GetOrder
	got2, err := repo.GetOrder(ctx, "r1")
	assert.NoError(t, err)
	if assert.NotNil(t, got2) {
		assert.Equal(t, fo.Order.OrderUID, got2.Order.OrderUID)
		assert.Equal(t, fo.Delivery.Name, got2.Delivery.Name)
	}

	// ensure all expectations met
	assert.NoError(t, mock.ExpectationsWereMet())

	// cleanup client
	_ = rdb.Close()
}
