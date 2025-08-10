package orderService_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"

	mocks "wbL0/internal/mocks"
	"wbL0/internal/models"
	svc "wbL0/internal/service/orderService"
)

func TestOrderService_GetOrder(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		orderUID      string
		setupMocks    func(postgres *mocks.OrderPostgresRepositoryInterface, redis *mocks.OrderRedisRepoInterface)
		expectedOrder *models.FullOrder
		expectedError error
	}{
		{
			name:     "Order found in Redis",
			orderUID: "order123",
			setupMocks: func(pg *mocks.OrderPostgresRepositoryInterface, r *mocks.OrderRedisRepoInterface) {
				r.On("GetOrder", ctx, "order123").Return(&models.FullOrder{
					Order: models.Order{OrderUID: "order123"},
				}, nil)
			},
			expectedOrder: &models.FullOrder{Order: models.Order{OrderUID: "order123"}},
			expectedError: nil,
		},
		{
			name:     "Order not in Redis, found in Postgres",
			orderUID: "order456",
			setupMocks: func(pg *mocks.OrderPostgresRepositoryInterface, r *mocks.OrderRedisRepoInterface) {
				r.On("GetOrder", ctx, "order456").Return(nil, nil)
				pg.On("GetFullOrderByUID", ctx, "order456").Return(&models.FullOrder{
					Order: models.Order{OrderUID: "order456"},
				}, nil)
				r.On("SetOrder", ctx, mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)
			},
			expectedOrder: &models.FullOrder{Order: models.Order{OrderUID: "order456"}},
			expectedError: nil,
		},
		{
			name:     "Order not in Redis, Postgres error",
			orderUID: "order789",
			setupMocks: func(pg *mocks.OrderPostgresRepositoryInterface, r *mocks.OrderRedisRepoInterface) {
				r.On("GetOrder", ctx, "order789").Return(nil, nil)
				pg.On("GetFullOrderByUID", ctx, "order789").Return(nil, errors.New("postgres error"))
			},
			expectedOrder: nil,
			expectedError: errors.New("postgres error"),
		},
		{
			name:     "Redis error, fallback to Postgres",
			orderUID: "order101",
			setupMocks: func(pg *mocks.OrderPostgresRepositoryInterface, r *mocks.OrderRedisRepoInterface) {
				r.On("GetOrder", ctx, "order101").Return(nil, errors.New("redis error"))
				pg.On("GetFullOrderByUID", ctx, "order101").Return(&models.FullOrder{
					Order: models.Order{OrderUID: "order101"},
				}, nil)
				r.On("SetOrder", ctx, mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)
			},
			expectedOrder: &models.FullOrder{Order: models.Order{OrderUID: "order101"}},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgMock := &mocks.OrderPostgresRepositoryInterface{}
			rMock := &mocks.OrderRedisRepoInterface{}
			tt.setupMocks(pgMock, rMock)

			service := svc.NewOrderService(pgMock, rMock, slog.Default(), time.Hour)

			order, err := service.GetOrder(ctx, tt.orderUID)

			assert.Equal(t, tt.expectedOrder, order)
			assert.Equal(t, tt.expectedError, err)

			pgMock.AssertExpectations(t)
			rMock.AssertExpectations(t)
		})
	}
}
