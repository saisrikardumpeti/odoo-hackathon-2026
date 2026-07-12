package activity_log_repo

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

type ActivityLogFilters struct {
	ActorEmployeeID *string
	Action          *string
	EntityType      *string
	EntityID        *string
	DateFrom        *time.Time
	DateTo          *time.Time
	Page            int
	PageSize        int
}

type ActivityLogListResult struct {
	Logs  []models.ActivityLogDetail
	Total int
}

type ActivityLogRepository struct {
	pool *pgxpool.Pool
}

func NewActivityLogRepository(pool *pgxpool.Pool) *ActivityLogRepository {
	return &ActivityLogRepository{pool: pool}
}
