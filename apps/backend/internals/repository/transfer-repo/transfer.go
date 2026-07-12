package transfer_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

var (
	ErrTransferNotFound = errors.New("transfer request not found")
)

func (r *TransferRepository) Create(ctx context.Context, t models.TransferRequest) (*models.TransferRequest, error) {
	var tr models.TransferRequest
	err := r.pool.QueryRow(ctx,
		`INSERT INTO transfer_requests (asset_id, allocation_id, from_employee_id, to_employee_id, requested_by)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, asset_id, allocation_id, from_employee_id, to_employee_id,
		           requested_by, status, approved_by, approved_at, created_at, updated_at`,
		t.AssetID, t.AllocationID, t.FromEmployeeID, t.ToEmployeeID, t.RequestedBy,
	).Scan(
		&tr.ID, &tr.AssetID, &tr.AllocationID, &tr.FromEmployeeID, &tr.ToEmployeeID,
		&tr.RequestedBy, &tr.Status, &tr.ApprovedBy, &tr.ApprovedAt, &tr.CreatedAt, &tr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer request: %w", err)
	}
	return &tr, nil
}

func (r *TransferRepository) GetByID(ctx context.Context, id string) (*models.TransferRequestDetail, error) {
	var t models.TransferRequestDetail
	err := r.pool.QueryRow(ctx,
		`SELECT tr.id, tr.asset_id, tr.allocation_id, tr.from_employee_id, tr.to_employee_id,
		        tr.requested_by, tr.status, tr.approved_by, tr.approved_at, tr.created_at, tr.updated_at,
		        as2.asset_tag, as2.name,
		        from_e.name, to_e.name, req_e.name
		 FROM transfer_requests tr
		 JOIN assets as2 ON as2.id = tr.asset_id
		 LEFT JOIN employees from_e ON from_e.id = tr.from_employee_id
		 JOIN employees to_e ON to_e.id = tr.to_employee_id
		 JOIN employees req_e ON req_e.id = tr.requested_by
		 WHERE tr.id = $1`, id,
	).Scan(
		&t.ID, &t.AssetID, &t.AllocationID, &t.FromEmployeeID, &t.ToEmployeeID,
		&t.RequestedBy, &t.Status, &t.ApprovedBy, &t.ApprovedAt, &t.CreatedAt, &t.UpdatedAt,
		&t.AssetTag, &t.AssetName,
		&t.FromEmployeeName, &t.ToEmployeeName, &t.RequestedByName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransferNotFound
		}
		return nil, fmt.Errorf("failed to get transfer request: %w", err)
	}
	return &t, nil
}

func (r *TransferRepository) ListPending(ctx context.Context) ([]models.TransferRequestDetail, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT tr.id, tr.asset_id, tr.allocation_id, tr.from_employee_id, tr.to_employee_id,
		        tr.requested_by, tr.status, tr.approved_by, tr.approved_at, tr.created_at, tr.updated_at,
		        as2.asset_tag, as2.name,
		        from_e.name, to_e.name, req_e.name
		 FROM transfer_requests tr
		 JOIN assets as2 ON as2.id = tr.asset_id
		 LEFT JOIN employees from_e ON from_e.id = tr.from_employee_id
		 JOIN employees to_e ON to_e.id = tr.to_employee_id
		 JOIN employees req_e ON req_e.id = tr.requested_by
		 WHERE tr.status = 'Requested'
		 ORDER BY tr.created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending transfers: %w", err)
	}
	defer rows.Close()

	var transfers []models.TransferRequestDetail
	for rows.Next() {
		var t models.TransferRequestDetail
		if err := rows.Scan(
			&t.ID, &t.AssetID, &t.AllocationID, &t.FromEmployeeID, &t.ToEmployeeID,
			&t.RequestedBy, &t.Status, &t.ApprovedBy, &t.ApprovedAt, &t.CreatedAt, &t.UpdatedAt,
			&t.AssetTag, &t.AssetName,
			&t.FromEmployeeName, &t.ToEmployeeName, &t.RequestedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transfer request: %w", err)
		}
		transfers = append(transfers, t)
	}
	if transfers == nil {
		transfers = []models.TransferRequestDetail{}
	}
	return transfers, nil
}

func (r *TransferRepository) UpdateStatus(ctx context.Context, id, status string, approvedBy *string, approvedAt *time.Time, _ /* rejectedReason */ *string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE transfer_requests SET status = $1, approved_by = $2, approved_at = $3, updated_at = now() WHERE id = $4`,
		status, approvedBy, approvedAt, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update transfer status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTransferNotFound
	}
	return nil
}
