package orderRepoPostgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"wbL0/internal/models"
)

//go:generate mockery --name=PgxTx --dir=. --output=../../../mocks --outpkg=mocks --case=underscore
type PgxTx interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

//go:generate mockery --name=OrderPostgresRepositoryInterface --dir=. --output=../../../mocks --outpkg=mocks --case=underscore
type OrderPostgresRepositoryInterface interface {
	BeginTx(ctx context.Context) (PgxTx, error)
	SaveOrderDataTx(ctx context.Context, tx PgxTx, order *models.Order) error
	SaveDeliveryDataTx(ctx context.Context, tx PgxTx, delivery *models.Delivery) error
	SavePaymentDataTx(ctx context.Context, tx PgxTx, payment *models.Payment) error
	SaveItemsDataTx(ctx context.Context, tx PgxTx, item *models.Item) error
	GetOrderInfoByUid(ctx context.Context, orderUID string) (*models.Order, error)
	GetAllFullOrders(ctx context.Context) ([]*models.FullOrder, error)
	GetFullOrderByUID(ctx context.Context, orderUID string) (*models.FullOrder, error)
}
type OrderPostgresRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewPostgresRepository(pool *pgxpool.Pool, log *slog.Logger) *OrderPostgresRepository {
	return &OrderPostgresRepository{pool: pool, log: log}
}

func (r *OrderPostgresRepository) BeginTx(ctx context.Context) (PgxTx, error) {
	const op = "OrderPostgresRepository.BeginTx"
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.log.Error("failed to begin transaction", "op", op, "err", err)
		return nil, err
	}
	return tx, nil
}
