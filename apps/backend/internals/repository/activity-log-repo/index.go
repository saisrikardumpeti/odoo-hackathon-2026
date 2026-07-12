package activity_log_repo

import "github.com/jackc/pgx/v5/pgxpool"

type ActivityLogRepository struct {
	pool *pgxpool.Pool
}

func NewActivityLogRepository(pool *pgxpool.Pool) *ActivityLogRepository {
	return &ActivityLogRepository{pool: pool}
}
