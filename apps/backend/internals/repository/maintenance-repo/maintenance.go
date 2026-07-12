package maintenance_repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

var (
	ErrMaintenanceNotFound = errors.New("maintenance request not found")
)

func (r *MaintenanceRepository) Create(ctx context.Context, m models.MaintenanceRequest) (*models.MaintenanceRequest, error) {
	var req models.MaintenanceRequest
	err := r.pool.QueryRow(ctx,
		`INSERT INTO maintenance_requests (asset_id, raised_by_employee_id, issue_description, priority, photo_url)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, asset_id, raised_by_employee_id, issue_description, priority, photo_url,
		           status, approved_by, approved_at, technician_name, resolved_at,
		           resolution_notes, created_at, updated_at`,
		m.AssetID, m.RaisedByEmployeeID, m.IssueDescription, m.Priority, m.PhotoURL,
	).Scan(
		&req.ID, &req.AssetID, &req.RaisedByEmployeeID, &req.IssueDescription,
		&req.Priority, &req.PhotoURL, &req.Status, &req.ApprovedBy,
		&req.ApprovedAt, &req.TechnicianName, &req.ResolvedAt,
		&req.ResolutionNotes, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create maintenance request: %w", err)
	}
	return &req, nil
}

type MaintenanceListFilters struct {
	AssetID  string
	Status   string
	Priority string
	Page     int
	PageSize int
}

type MaintenanceListResult struct {
	Requests   []models.MaintenanceDetail
	TotalCount int
}

func (r *MaintenanceRepository) List(ctx context.Context, filters MaintenanceListFilters) (*MaintenanceListResult, error) {
	conditions := []string{}
	args := []any{}
	argIdx := 1

	if filters.AssetID != "" {
		conditions = append(conditions, fmt.Sprintf("mr.asset_id = $%d", argIdx))
		args = append(args, filters.AssetID)
		argIdx++
	}
	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("mr.status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("mr.priority = $%d", argIdx))
		args = append(args, filters.Priority)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 50
	}
	offset := (filters.Page - 1) * filters.PageSize

	var totalCount int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM maintenance_requests mr"+whereClause, args...,
	).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count maintenance requests: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT mr.id, mr.asset_id, mr.raised_by_employee_id, mr.issue_description,
		        mr.priority, mr.photo_url, mr.status, mr.approved_by, mr.approved_at,
		        mr.technician_name, mr.resolved_at, mr.resolution_notes,
		        mr.created_at, mr.updated_at,
		        a.asset_tag, a.name, raiser.name, approver.name
		 FROM maintenance_requests mr
		 JOIN assets a ON a.id = mr.asset_id
		 JOIN employees raiser ON raiser.id = mr.raised_by_employee_id
		 LEFT JOIN employees approver ON approver.id = mr.approved_by
		 %s
		 ORDER BY mr.created_at DESC
		 LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1,
	)
	args = append(args, filters.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list maintenance requests: %w", err)
	}
	defer rows.Close()

	requests := []models.MaintenanceDetail{}
	for rows.Next() {
		var d models.MaintenanceDetail
		if err := rows.Scan(
			&d.ID, &d.AssetID, &d.RaisedByEmployeeID, &d.IssueDescription,
			&d.Priority, &d.PhotoURL, &d.Status, &d.ApprovedBy, &d.ApprovedAt,
			&d.TechnicianName, &d.ResolvedAt, &d.ResolutionNotes,
			&d.CreatedAt, &d.UpdatedAt,
			&d.AssetTag, &d.AssetName, &d.RaisedByName, &d.ApprovedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan maintenance request: %w", err)
		}
		requests = append(requests, d)
	}
	if requests == nil {
		requests = []models.MaintenanceDetail{}
	}
	return &MaintenanceListResult{Requests: requests, TotalCount: totalCount}, nil
}

func (r *MaintenanceRepository) GetByID(ctx context.Context, id string) (*models.MaintenanceDetail, error) {
	var d models.MaintenanceDetail
	err := r.pool.QueryRow(ctx,
		`SELECT mr.id, mr.asset_id, mr.raised_by_employee_id, mr.issue_description,
		        mr.priority, mr.photo_url, mr.status, mr.approved_by, mr.approved_at,
		        mr.technician_name, mr.resolved_at, mr.resolution_notes,
		        mr.created_at, mr.updated_at,
		        a.asset_tag, a.name, raiser.name, approver.name
		 FROM maintenance_requests mr
		 JOIN assets a ON a.id = mr.asset_id
		 JOIN employees raiser ON raiser.id = mr.raised_by_employee_id
		 LEFT JOIN employees approver ON approver.id = mr.approved_by
		 WHERE mr.id = $1`, id,
	).Scan(
		&d.ID, &d.AssetID, &d.RaisedByEmployeeID, &d.IssueDescription,
		&d.Priority, &d.PhotoURL, &d.Status, &d.ApprovedBy, &d.ApprovedAt,
		&d.TechnicianName, &d.ResolvedAt, &d.ResolutionNotes,
		&d.CreatedAt, &d.UpdatedAt,
		&d.AssetTag, &d.AssetName, &d.RaisedByName, &d.ApprovedByName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMaintenanceNotFound
		}
		return nil, fmt.Errorf("failed to get maintenance request: %w", err)
	}
	return &d, nil
}

func (r *MaintenanceRepository) GetCurrentAssetStatus(ctx context.Context, assetID string) (string, error) {
	var status string
	err := r.pool.QueryRow(ctx,
		`SELECT status FROM assets WHERE id = $1`, assetID,
	).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("asset not found")
		}
		return "", fmt.Errorf("failed to get asset status: %w", err)
	}
	return status, nil
}

func (r *MaintenanceRepository) UpdateAssetStatus(ctx context.Context, assetID, toStatus string, changedByID *string, reason string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE assets SET status = $1, updated_at = now() WHERE id = $2`,
		toStatus, assetID,
	)
	if err != nil {
		return fmt.Errorf("failed to update asset status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("asset not found")
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO asset_status_history (asset_id, from_status, to_status, changed_by, reason)
		 VALUES ($1, (SELECT status FROM assets WHERE id = $1), $2, $3, $4)`,
		assetID, toStatus, changedByID, reason,
	)
	if err != nil {
		return fmt.Errorf("failed to create status history: %w", err)
	}
	return nil
}

func (r *MaintenanceRepository) UpdateStatus(ctx context.Context, id, status string, fields map[string]interface{}) error {
	setClauses := []string{"status = $2", "updated_at = now()"}
	args := []any{id, status}
	argIdx := 3

	for col, val := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}

	query := fmt.Sprintf(
		"UPDATE maintenance_requests SET %s WHERE id = $1",
		strings.Join(setClauses, ", "),
	)

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update maintenance request status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrMaintenanceNotFound
	}
	return nil
}

func (r *MaintenanceRepository) ListByAsset(ctx context.Context, assetID string) ([]models.MaintenanceDetail, error) {
	result, err := r.List(ctx, MaintenanceListFilters{AssetID: assetID, PageSize: 100})
	if err != nil {
		return nil, err
	}
	return result.Requests, nil
}
