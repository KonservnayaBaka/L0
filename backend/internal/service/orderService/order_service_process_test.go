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

func TestOrderService_ProcessAndCache(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	tests := []struct {
		name            string
		setupMocks      func(pg *mocks.OrderPostgresRepositoryInterface, r *mocks.OrderRedisRepoInterface)
		input           *models.FullOrder
		expectErr       bool
		expectRedisCall bool
	}{
		{
			name: "success path",
			setupMocks: func(pg *mocks.OrderPostgresRepositoryInterface, r *mocks.OrderRedisRepoInterface) {
				tx := &mocks.PgxTx{}
				pg.On("BeginTx", mock.Anything).Return(tx, nil)
				pg.On("SaveOrderDataTx", mock.Anything, tx, mock.Anything).Return(nil)
				pg.On("SaveDeliveryDataTx", mock.Anything, tx, mock.Anything).Return(nil)
				pg.On("SavePaymentDataTx", mock.Anything, tx, mock.Anything).Return(nil)
				pg.On("SaveItemsDataTx", mock.Anything, tx, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)

				r.On("SetOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			input:           &models.FullOrder{Order: models.Order{OrderUID: "ok1"}, Items: []models.Item{{}}},
			expectErr:       false,
			expectRedisCall: true,
		},
		{
			name: "BeginTx error",
			setupMocks: func(pg *mocks.OrderPostgresRepositoryInterface, r *mocks.OrderRedisRepoInterface) {
				pg.On("BeginTx", mock.Anything).Return(nil, errors.New("begin fail"))
			},
			input:           &models.FullOrder{Order: models.Order{OrderUID: "ok2"}},
			expectErr:       true,
			expectRedisCall: false,
		},
		{
			name: "SaveItemsDataTx fails",
			setupMocks: func(pg *mocks.OrderPostgresRepositoryInterface, r *mocks.OrderRedisRepoInterface) {
				tx := &mocks.PgxTx{}
				pg.On("BeginTx", mock.Anything).Return(tx, nil)
				pg.On("SaveOrderDataTx", mock.Anything, tx, mock.Anything).Return(nil)
				pg.On("SaveDeliveryDataTx", mock.Anything, tx, mock.Anything).Return(nil)
				pg.On("SavePaymentDataTx", mock.Anything, tx, mock.Anything).Return(nil)
				pg.On("SaveItemsDataTx", mock.Anything, tx, mock.Anything).Return(errors.New("item fail"))
				tx.On("Rollback", mock.Anything).Return(nil)
			},
			input:           &models.FullOrder{Order: models.Order{OrderUID: "ok3"}, Items: []models.Item{{}}},
			expectErr:       true,
			expectRedisCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgMock := &mocks.OrderPostgresRepositoryInterface{}
			rMock := &mocks.OrderRedisRepoInterface{}
			tt.setupMocks(pgMock, rMock)

			service := svc.NewOrderService(pgMock, rMock, logger, time.Hour)
			err := service.ProcessAndCache(ctx, tt.input)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			pgMock.AssertExpectations(t)
			rMock.AssertExpectations(t)
		})
	}
}
