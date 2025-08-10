package orderRepoRedis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log/slog"
	"time"
	"wbL0/internal/models"
)

//go:generate mockery --name=OrderRedisRepoInterface --dir=. --srcpkg=./internal/repository/redis/orderRepoRedis --output=../../../mocks --outpkg=mocks --case=underscore
type OrderRedisRepoInterface interface {
	GetOrder(ctx context.Context, orderUID string) (*models.FullOrder, error)
	SetOrder(ctx context.Context, order *models.FullOrder, ttl time.Duration) error
	RestoreOrders(ctx context.Context, orders []*models.FullOrder, ttl time.Duration) error
}

type OrderRedisRepo struct {
	rdb *redis.Client
	log *slog.Logger
}

func NewRedisRepo(rdb *redis.Client, log *slog.Logger) *OrderRedisRepo {
	return &OrderRedisRepo{rdb: rdb, log: log}
}

func (r *OrderRedisRepo) GetOrder(ctx context.Context, orderUID string) (*models.FullOrder, error) {
	const op = "OrderRedisRepo.GetOrder"
	key := fmt.Sprintf("order:%s", orderUID)

	raw, err := r.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		r.log.Warn("failed to get order from redis", "op", op, "err", err)
		return nil, err
	}

	var fo models.FullOrder
	if err := json.Unmarshal(raw, &fo); err != nil {
		r.log.Warn("failed to unmarshal order from redis", "op", op, "err", err)
		return nil, err
	}
	return &fo, nil
}

func (r *OrderRedisRepo) SetOrder(ctx context.Context, order *models.FullOrder, ttl time.Duration) error {
	const op = "OrderRedisRepo.SetOrder"
	key := fmt.Sprintf("order:%s", order.Order.OrderUID)

	data, err := json.Marshal(order)
	if err != nil {
		r.log.Warn("failed to marshal order for redis", "op", op, "err", err)
		return err
	}

	if err := r.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		r.log.Warn("failed to set order in redis", "op", op, "err", err)
		return err
	}
	return nil
}

func (r *OrderRedisRepo) RestoreOrders(ctx context.Context, orders []*models.FullOrder, ttl time.Duration) error {
	const op = "OrderRedisRepo.RestoreOrders"

	for _, fo := range orders {
		if err := r.SetOrder(ctx, fo, ttl); err != nil {
			r.log.Warn("failed to restore order to redis", "op", op, "err", err)
			continue
		}
	}
	r.log.Info("cache restored", "op", op, "count", len(orders))
	return nil
}
