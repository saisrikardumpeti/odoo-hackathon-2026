package activity_log_repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

func (r *ActivityLogRepository) Create(ctx context.Context, entry models.ActivityLog) error {
	rawMeta, err := json.Marshal(entry.Metadata)
	if err != nil {
		return err
	}
	if rawMeta == nil {
		rawMeta = []byte("{}")
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO activity_logs (actor_employee_id, action, entity_type, entity_id, metadata)
   VALUES ($1, $2, $3, $4, $5)`,
		entry.ActorEmployeeID, entry.Action, entry.EntityType, entry.EntityID, rawMeta,
	)
	return err
}

func (r *ActivityLogRepository) List(ctx context.Context, filters ActivityLogFilters) (*ActivityLogListResult, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}
	offset := (filters.Page - 1) * filters.PageSize

	var whereClauses []string
	var args []interface{}
	argIdx := 1

	if filters.ActorEmployeeID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.actor_employee_id = $%d", argIdx))
		args = append(args, *filters.ActorEmployeeID)
		argIdx++
	}
	if filters.Action != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.action = $%d", argIdx))
		args = append(args, *filters.Action)
		argIdx++
	}
	if filters.EntityType != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.entity_type = $%d", argIdx))
		args = append(args, *filters.EntityType)
		argIdx++
	}
	if filters.EntityID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.entity_id = $%d", argIdx))
		args = append(args, *filters.EntityID)
		argIdx++
	}
	if filters.DateFrom != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.created_at >= $%d", argIdx))
		args = append(args, *filters.DateFrom)
		argIdx++
	}
	if filters.DateTo != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.created_at <= $%d", argIdx))
		args = append(args, *filters.DateTo)
		argIdx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM activity_logs al` + whereSQL
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count activity logs: %w", err)
	}

	listQuery := fmt.Sprintf(`SELECT al.id, al.actor_employee_id, al.action, al.entity_type, al.entity_id,
		 al.metadata, al.created_at, e.name
		 FROM activity_logs al
		 LEFT JOIN employees e ON e.id = al.actor_employee_id%s
		 ORDER BY al.created_at DESC LIMIT $%d OFFSET $%d`, whereSQL, argIdx, argIdx+1)
	args = append(args, filters.PageSize, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list activity logs: %w", err)
	}
	defer rows.Close()

	var logs []models.ActivityLogDetail
	for rows.Next() {
		var log models.ActivityLogDetail
		var rawMeta []byte
		if err := rows.Scan(&log.ID, &log.ActorEmployeeID, &log.Action, &log.EntityType, &log.EntityID,
			&rawMeta, &log.CreatedAt, &log.ActorName); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				break
			}
			return nil, fmt.Errorf("failed to scan activity log: %w", err)
		}
		if rawMeta != nil {
			_ = json.Unmarshal(rawMeta, &log.Metadata)
		}
		if log.Metadata == nil {
			log.Metadata = map[string]interface{}{}
		}
		if name := extractEntityName(log.EntityType, log.Metadata); name != "" {
			log.EntityName = &name
		}
		logs = append(logs, log)
	}
	if logs == nil {
		logs = []models.ActivityLogDetail{}
	}

	return &ActivityLogListResult{
		Logs:  logs,
		Total: total,
	}, nil
}

func extractEntityName(entityType string, meta map[string]interface{}) string {
	switch entityType {
	case "asset":
		if tag, ok := meta["asset_tag"].(string); ok && tag != "" {
			return tag
		}
		if name, ok := meta["asset_name"].(string); ok && name != "" {
			return name
		}
	case "booking":
		if tag, ok := meta["asset_tag"].(string); ok && tag != "" {
			return tag
		}
	case "allocation":
		if tag, ok := meta["asset_tag"].(string); ok && tag != "" {
			return tag
		}
	case "maintenance":
		if tag, ok := meta["asset_tag"].(string); ok && tag != "" {
			return tag
		}
	case "employee":
		if name, ok := meta["employee_name"].(string); ok && name != "" {
			return name
		}
	case "department":
		if name, ok := meta["department_name"].(string); ok && name != "" {
			return name
		}
	case "audit_cycle":
		if name, ok := meta["cycle_name"].(string); ok && name != "" {
			return name
		}
	}
	return ""
}
