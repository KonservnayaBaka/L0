package postgres

import (
	"context"
	"fmt"
	"wbL0/internal/config"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func MustLoad(ctx context.Context, cfg *config.Config) *pgxpool.Pool {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port,
		cfg.Database.DBName, cfg.Database.SSLMode)

	err := runMigrations(ctx, connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	if err := pool.Ping(ctx); err != nil {
		panic(fmt.Sprintf("failed to ping database: %v", err))
	}

	return pool
}
