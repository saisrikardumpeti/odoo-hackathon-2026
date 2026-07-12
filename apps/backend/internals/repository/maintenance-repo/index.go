package maintenance_repo

import "github.com/jackc/pgx/v5/pgxpool"

type MaintenanceRepository struct {
	pool *pgxpool.Pool
}

func NewMaintenanceRepository(pool *pgxpool.Pool) *MaintenanceRepository {
	return &MaintenanceRepository{pool: pool}
}
