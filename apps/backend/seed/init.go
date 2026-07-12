package seed

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

//go:embed init.sql
var initSQL string

const defaultAdminEmail = "admin@assetflow.local"
const defaultAdminPassword = "admin123"

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

	if err := seedDefaultAdmin(ctx, pool); err != nil {
		return fmt.Errorf("failed to seed default admin: %w", err)
	}

	return nil
}

func seedDefaultAdmin(ctx context.Context, pool *pgxpool.Pool) error {
	var adminExists bool
	err := pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM employees WHERE email = $1)",
		defaultAdminEmail,
	).Scan(&adminExists)
	if err != nil {
		return err
	}
	if adminExists {
		log.Println("Default admin already exists, skipping seed")
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	var employeeID string
	err = pool.QueryRow(ctx,
		`INSERT INTO employees (name, email, role, status)
		 VALUES ('System Admin', $1, 'Admin', 'Active')
		 RETURNING id`,
		defaultAdminEmail,
	).Scan(&employeeID)
	if err != nil {
		return fmt.Errorf("failed to insert admin employee: %w", err)
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO users (employee_id, password_hash) VALUES ($1, $2)`,
		employeeID, string(hash),
	)
	if err != nil {
		return fmt.Errorf("failed to insert admin user: %w", err)
	}

	log.Printf("Default admin created: %s / %s", defaultAdminEmail, defaultAdminPassword)
	return nil
}
