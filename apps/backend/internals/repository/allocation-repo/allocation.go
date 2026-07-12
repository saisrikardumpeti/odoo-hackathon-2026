package allocation_repo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

var (
	ErrAllocationNotFound       = errors.New("allocation not found")
	ErrAllocationAlreadyActive  = errors.New("asset already has an active allocation")
	ErrAssetNotFound            = errors.New("asset not found")
)

func (r *AllocationRepository) Create(ctx context.Context, a models.Allocation) (*models.Allocation, error) {
	var alloc models.Allocation
	err := r.pool.QueryRow(ctx,
		`INSERT INTO allocations (asset_id, employee_id, department_id, allocated_by, expected_return_date)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, asset_id, employee_id, department_id, allocated_by, allocated_at,
		           expected_return_date::text, returned_at, return_condition_notes, status, created_at, updated_at`,
		a.AssetID, a.EmployeeID, a.DepartmentID, a.AllocatedBy, a.ExpectedReturnDate,
	).Scan(
		&alloc.ID, &alloc.AssetID, &alloc.EmployeeID, &alloc.DepartmentID, &alloc.AllocatedBy,
		&alloc.AllocatedAt, &alloc.ExpectedReturnDate, &alloc.ReturnedAt, &alloc.ReturnConditionNotes,
		&alloc.Status, &alloc.CreatedAt, &alloc.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrAllocationAlreadyActive
		}
		return nil, fmt.Errorf("failed to create allocation: %w", err)
	}
	return &alloc, nil
}

func (r *AllocationRepository) GetByID(ctx context.Context, id string) (*models.AllocationDetail, error) {
	var a models.AllocationDetail
	err := r.pool.QueryRow(ctx,
		`SELECT al.id, al.asset_id, al.employee_id, al.department_id, al.allocated_by,
		        al.allocated_at, al.expected_return_date::text, al.returned_at,
		        al.return_condition_notes, al.status, al.created_at, al.updated_at,
		        as2.asset_tag, as2.name,
		        e.name, d.name, allocator.name
		 FROM allocations al
		 JOIN assets as2 ON as2.id = al.asset_id
		 LEFT JOIN employees e ON e.id = al.employee_id
		 LEFT JOIN departments d ON d.id = al.department_id
		 JOIN employees allocator ON allocator.id = al.allocated_by
		 WHERE al.id = $1`, id,
	).Scan(
		&a.ID, &a.AssetID, &a.EmployeeID, &a.DepartmentID, &a.AllocatedBy,
		&a.AllocatedAt, &a.ExpectedReturnDate, &a.ReturnedAt,
		&a.ReturnConditionNotes, &a.Status, &a.CreatedAt, &a.UpdatedAt,
		&a.AssetTag, &a.AssetName,
		&a.EmployeeName, &a.DepartmentName, &a.AllocatedByName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAllocationNotFound
		}
		return nil, fmt.Errorf("failed to get allocation: %w", err)
	}
	return &a, nil
}

func (r *AllocationRepository) GetActiveByAssetID(ctx context.Context, assetID string) (*models.AllocationDetail, error) {
	var a models.AllocationDetail
	err := r.pool.QueryRow(ctx,
		`SELECT al.id, al.asset_id, al.employee_id, al.department_id, al.allocated_by,
		        al.allocated_at, al.expected_return_date::text, al.returned_at,
		        al.return_condition_notes, al.status, al.created_at, al.updated_at,
		        as2.asset_tag, as2.name,
		        e.name, d.name, allocator.name
		 FROM allocations al
		 JOIN assets as2 ON as2.id = al.asset_id
		 LEFT JOIN employees e ON e.id = al.employee_id
		 LEFT JOIN departments d ON d.id = al.department_id
		 JOIN employees allocator ON allocator.id = al.allocated_by
		 WHERE al.asset_id = $1 AND al.status = 'Active'`, assetID,
	).Scan(
		&a.ID, &a.AssetID, &a.EmployeeID, &a.DepartmentID, &a.AllocatedBy,
		&a.AllocatedAt, &a.ExpectedReturnDate, &a.ReturnedAt,
		&a.ReturnConditionNotes, &a.Status, &a.CreatedAt, &a.UpdatedAt,
		&a.AssetTag, &a.AssetName,
		&a.EmployeeName, &a.DepartmentName, &a.AllocatedByName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAllocationNotFound
		}
		return nil, fmt.Errorf("failed to get active allocation: %w", err)
	}
	return &a, nil
}

func (r *AllocationRepository) UpdateStatus(ctx context.Context, id, status string, returnedAt *time.Time, returnConditionNotes *string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE allocations SET status = $1, returned_at = $2, return_condition_notes = $3, updated_at = now() WHERE id = $4`,
		status, returnedAt, returnConditionNotes, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update allocation status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAllocationNotFound
	}
	return nil
}

func (r *AllocationRepository) ListOverdue(ctx context.Context) ([]models.AllocationDetail, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT al.id, al.asset_id, al.employee_id, al.department_id, al.allocated_by,
		        al.allocated_at, al.expected_return_date::text, al.returned_at,
		        al.return_condition_notes, al.status, al.created_at, al.updated_at,
		        as2.asset_tag, as2.name,
		        e.name, d.name, allocator.name
		 FROM allocations al
		 JOIN assets as2 ON as2.id = al.asset_id
		 LEFT JOIN employees e ON e.id = al.employee_id
		 LEFT JOIN departments d ON d.id = al.department_id
		 JOIN employees allocator ON allocator.id = al.allocated_by
		 WHERE al.status = 'Active' AND al.expected_return_date < now()
		 ORDER BY al.expected_return_date ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list overdue allocations: %w", err)
	}
	defer rows.Close()

	var allocations []models.AllocationDetail
	for rows.Next() {
		var a models.AllocationDetail
		if err := rows.Scan(
			&a.ID, &a.AssetID, &a.EmployeeID, &a.DepartmentID, &a.AllocatedBy,
			&a.AllocatedAt, &a.ExpectedReturnDate, &a.ReturnedAt,
			&a.ReturnConditionNotes, &a.Status, &a.CreatedAt, &a.UpdatedAt,
			&a.AssetTag, &a.AssetName,
			&a.EmployeeName, &a.DepartmentName, &a.AllocatedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan overdue allocation: %w", err)
		}
		allocations = append(allocations, a)
	}
	if allocations == nil {
		allocations = []models.AllocationDetail{}
	}
	return allocations, nil
}

func (r *AllocationRepository) ListByEmployee(ctx context.Context, employeeID string) ([]models.AllocationDetail, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT al.id, al.asset_id, al.employee_id, al.department_id, al.allocated_by,
		        al.allocated_at, al.expected_return_date::text, al.returned_at,
		        al.return_condition_notes, al.status, al.created_at, al.updated_at,
		        as2.asset_tag, as2.name,
		        e.name, d.name, allocator.name
		 FROM allocations al
		 JOIN assets as2 ON as2.id = al.asset_id
		 LEFT JOIN employees e ON e.id = al.employee_id
		 LEFT JOIN departments d ON d.id = al.department_id
		 JOIN employees allocator ON allocator.id = al.allocated_by
		 WHERE al.employee_id = $1
		 ORDER BY al.created_at DESC`, employeeID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list allocations by employee: %w", err)
	}
	defer rows.Close()

	var allocations []models.AllocationDetail
	for rows.Next() {
		var a models.AllocationDetail
		if err := rows.Scan(
			&a.ID, &a.AssetID, &a.EmployeeID, &a.DepartmentID, &a.AllocatedBy,
			&a.AllocatedAt, &a.ExpectedReturnDate, &a.ReturnedAt,
			&a.ReturnConditionNotes, &a.Status, &a.CreatedAt, &a.UpdatedAt,
			&a.AssetTag, &a.AssetName,
			&a.EmployeeName, &a.DepartmentName, &a.AllocatedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan allocation: %w", err)
		}
		allocations = append(allocations, a)
	}
	if allocations == nil {
		allocations = []models.AllocationDetail{}
	}
	return allocations, nil
}

func (r *AllocationRepository) UpdateAssetStatusTx(ctx context.Context, allocationID, assetID, status string, changedByID *string, reason string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE assets SET status = $1, updated_at = now() WHERE id = $2`,
		status, assetID,
	)
	if err != nil {
		return fmt.Errorf("failed to update asset status: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO asset_status_history (asset_id, from_status, to_status, changed_by, reason)
		 VALUES ($1, (SELECT status FROM assets WHERE id = $1), $2, $3, $4)`,
		assetID, status, changedByID, reason,
	)
	if err != nil {
		return fmt.Errorf("failed to create status history: %w", err)
	}
	return nil
}

func (r *AllocationRepository) UpdateAssetHolder(ctx context.Context, assetID string, employeeID *string, departmentID *string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE assets SET current_holder_employee_id = $1, current_holder_department_id = $2, updated_at = now() WHERE id = $3`,
		employeeID, departmentID, assetID,
	)
	if err != nil {
		return fmt.Errorf("failed to update asset holder: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "unique constraint") ||
		strings.Contains(err.Error(), "duplicate key value")
}
