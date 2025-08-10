package orderService

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"wbL0/internal/models"
	"wbL0/internal/repository/postgres/orderRepoPostgres"
	"wbL0/internal/repository/redis/orderRepoRedis"
)

//go:generate mockery --name=OrderServiceInterface --dir=. --output=../../mocks --outpkg=mocks --case=underscore
type OrderServiceInterface interface {
	GetOrder(ctx context.Context, orderUID string) (*models.FullOrder, error)
	ProcessAndCache(ctx context.Context, fo *models.FullOrder) error
	RestoreCacheFromDB(ctx context.Context) error
}

type OrderService struct {
	repo      orderRepoPostgres.OrderPostgresRepositoryInterface
	redisRepo orderRepoRedis.OrderRedisRepoInterface
	log       *slog.Logger
	ttl       time.Duration
}

func NewOrderService(repo orderRepoPostgres.OrderPostgresRepositoryInterface, redisRepo orderRepoRedis.OrderRedisRepoInterface, log *slog.Logger, ttl time.Duration) *OrderService {
	return &OrderService{repo: repo, redisRepo: redisRepo, log: log, ttl: ttl}
}

func (s *OrderService) GetOrder(ctx context.Context, orderUID string) (*models.FullOrder, error) {
	const op = "OrderService.GetOrder"

	fo, err := s.redisRepo.GetOrder(ctx, orderUID)
	if err != nil {
		s.log.Warn("failed to get order from redis", "op", op, "orderUID", orderUID, "err", err)
	}
	if fo != nil {
		return fo, nil
	}

	fo, err = s.repo.GetFullOrderByUID(ctx, orderUID)
	if err != nil {
		s.log.Error("failed to get order from postgres", "op", op, "orderUID", orderUID, "err", err)
		return nil, err
	}

	if err := s.redisRepo.SetOrder(ctx, fo, s.ttl); err != nil {
		s.log.Warn("failed to cache order in redis", "op", op, "orderUID", orderUID, "err", err)
	}

	return fo, nil
}

func (s *OrderService) ProcessAndCache(ctx context.Context, fo *models.FullOrder) error {
	const op = "OrderService.ProcessAndCache"

	if fo.Order.OrderUID == "" {
		return fmt.Errorf("order_uid is empty")
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", "op", op, "err", err)
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.repo.SaveOrderDataTx(ctx, tx, &fo.Order); err != nil {
		s.log.Error("failed to save order data", "op", op, "err", err)
		return err
	}
	if err := s.repo.SaveDeliveryDataTx(ctx, tx, &fo.Delivery); err != nil {
		s.log.Error("failed to save delivery data", "op", op, "err", err)
		return err
	}
	if err := s.repo.SavePaymentDataTx(ctx, tx, &fo.Payment); err != nil {
		s.log.Error("failed to save payment data", "op", op, "err", err)
		return err
	}
	for i := range fo.Items {
		if err := s.repo.SaveItemsDataTx(ctx, tx, &fo.Items[i]); err != nil {
			s.log.Error("failed to save item data", "op", op, "item_index", i, "err", err)
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		s.log.Error("failed to commit transaction", "op", op, "err", err)
		return err
	}

	if err := s.redisRepo.SetOrder(ctx, fo, s.ttl); err != nil {
		s.log.Warn("failed to cache order in redis", "op", op, "err", err)
	}

	return nil
}

func (s *OrderService) RestoreCacheFromDB(ctx context.Context) error {
	const op = "OrderService.RestoreCacheFromDB"

	orders, err := s.repo.GetAllFullOrders(ctx)
	if err != nil {
		s.log.Error("failed to get all orders from postgres", "op", op, "err", err)
		return err
	}

	if err := s.redisRepo.RestoreOrders(ctx, orders, s.ttl); err != nil {
		s.log.Warn("failed to restore orders to redis", "op", op, "err", err)
	}

	s.log.Info("cache restored", "op", op, "count", len(orders))
	return nil
}
