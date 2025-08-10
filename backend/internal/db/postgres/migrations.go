package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func runMigrations(ctx context.Context, connStr string) error {
	migrationsPath := "file://internal/db/migrations"

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to create db instance for migrations: %w", err)
	}
	defer db.Close()

	if err = db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping db for migrations: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	sourceErr, dbErr := m.Close()
	if sourceErr != nil {
		return fmt.Errorf("migration source error: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("migration db error: %w", dbErr)
	}

	fmt.Println("migrations successfully applied")
	return nil
}
