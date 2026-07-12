package allocation_repo

import "github.com/jackc/pgx/v5/pgxpool"

type AllocationRepository struct {
	pool *pgxpool.Pool
}

func NewAllocationRepository(pool *pgxpool.Pool) *AllocationRepository {
	return &AllocationRepository{pool: pool}
}
