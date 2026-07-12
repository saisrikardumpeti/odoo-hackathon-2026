package department_repo

import "github.com/jackc/pgx/v5/pgxpool"

type DepartmentRepository struct {
	pool *pgxpool.Pool
}

func NewDepartmentRepository(pool *pgxpool.Pool) *DepartmentRepository {
	return &DepartmentRepository{pool: pool}
}
