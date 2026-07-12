package seed

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed init.sql
var initSQL string

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	var exists bool
	err := pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'employees')",
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check schema state: %w", err)
	}
	if exists {
		log.Println("Database schema already initialized, skipping migration")
		return nil
	}

	_, err = pool.Exec(ctx, initSQL)
	if err != nil {
		return fmt.Errorf("failed to execute init.sql: %w", err)
	}
	log.Println("Database schema initialized successfully")
	return nil
}
