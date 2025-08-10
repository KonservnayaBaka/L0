package orderRepoPostgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"

	mocks "wbL0/internal/mocks"
	"wbL0/internal/models"
	orderRepoPostgres "wbL0/internal/repository/postgres/orderRepoPostgres"
)

func anyArgs(n int) []interface{} {
	args := make([]interface{}, n)
	for i := 0; i < n; i++ {
		args[i] = mock.Anything
	}
	return args
}

func TestSaveDataTx(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.Default()
	repo := orderRepoPostgres.NewPostgresRepository(nil, logger)

	now := time.Now()

	tests := []struct {
		name       string
		execArgs   int
		input      interface{}
		saveFunc   func(context.Context, orderRepoPostgres.PgxTx, interface{}) error
		wantErr    bool
		mockErr    error
		commandTag pgconn.CommandTag
	}{
		{
			name:     "SaveOrderDataTx success",
			execArgs: 13,
			input: &models.Order{
				OrderUID:    "u1",
				TrackNumber: "t1",
				Entry:       "e",
				Locale:      "ru",
				CustomerID:  "c1",
				DateCreated: now,
			},
			saveFunc: func(ctx context.Context, tx orderRepoPostgres.PgxTx, v interface{}) error {
				return repo.SaveOrderDataTx(ctx, tx, v.(*models.Order))
			},
			wantErr:    false,
			mockErr:    nil,
			commandTag: pgconn.NewCommandTag("INSERT 0 1"),
		},
		{
			name:     "SaveOrderDataTx failure",
			execArgs: 13,
			input: &models.Order{
				OrderUID:    "u1",
				TrackNumber: "t1",
				Entry:       "e",
				Locale:      "ru",
				CustomerID:  "c1",
				DateCreated: now,
			},
			saveFunc: func(ctx context.Context, tx orderRepoPostgres.PgxTx, v interface{}) error {
				return repo.SaveOrderDataTx(ctx, tx, v.(*models.Order))
			},
			wantErr:    true,
			mockErr:    errors.New("db error"),
			commandTag: pgconn.NewCommandTag(""),
		},
		{
			name:     "SaveItemsDataTx success",
			execArgs: 14,
			input: &models.Item{
				OrderUID:    "u1",
				ChrtID:      1,
				TrackNumber: "t1",
				Price:       10,
				Rid:         "rid",
				Name:        "it",
				Sale:        0,
				Size:        "M",
				TotalPrice:  10,
				NmID:        100,
				Brand:       "b",
				Status:      1,
			},
			saveFunc: func(ctx context.Context, tx orderRepoPostgres.PgxTx, v interface{}) error {
				return repo.SaveItemsDataTx(ctx, tx, v.(*models.Item))
			},
			wantErr:    false,
			mockErr:    nil,
			commandTag: pgconn.NewCommandTag("INSERT 0 1"),
		},
		{
			name:     "SaveItemsDataTx failure",
			execArgs: 14,
			input: &models.Item{
				OrderUID:    "u1",
				ChrtID:      1,
				TrackNumber: "t1",
				Price:       10,
				Rid:         "rid",
				Name:        "it",
				Sale:        0,
				Size:        "M",
				TotalPrice:  10,
				NmID:        100,
				Brand:       "b",
				Status:      1,
			},
			saveFunc: func(ctx context.Context, tx orderRepoPostgres.PgxTx, v interface{}) error {
				return repo.SaveItemsDataTx(ctx, tx, v.(*models.Item))
			},
			wantErr:    true,
			mockErr:    errors.New("exec fail"),
			commandTag: pgconn.NewCommandTag(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := mocks.NewPgxTx(t)
			tx.On("Exec", anyArgs(tt.execArgs)...).
				Return(tt.commandTag, tt.mockErr)

			err := tt.saveFunc(ctx, tx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tx.AssertCalled(t, "Exec", anyArgs(tt.execArgs)...)
		})
	}
}
