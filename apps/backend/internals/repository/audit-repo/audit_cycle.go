package audit_repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

var (
	ErrCycleNotFound   = errors.New("audit cycle not found")
	ErrCycleAlreadyClosed = errors.New("audit cycle is already closed")
	ErrItemNotFound    = errors.New("audit item not found")
	ErrNotAssignedAuditor = errors.New("not an assigned auditor for this item's cycle")
)

func (r *AuditRepository) CreateCycle(ctx context.Context, cycle models.AuditCycle, scopeDeptID, scopeLoc *string) (*models.AuditCycle, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var created models.AuditCycle
	err = tx.QueryRow(ctx,
		`INSERT INTO audit_cycles (name, scope_department_id, scope_location, start_date, end_date, status, created_by)
		 VALUES ($1, $2, $3, $4, $5, 'Draft', $6)
		 RETURNING id, name, scope_department_id, scope_location, start_date::text, end_date::text, status, created_by, closed_at, created_at, updated_at`,
		cycle.Name, cycle.ScopeDepartmentID, cycle.ScopeLocation, cycle.StartDate, cycle.EndDate, cycle.CreatedBy,
	).Scan(
		&created.ID, &created.Name, &created.ScopeDepartmentID, &created.ScopeLocation,
		&created.StartDate, &created.EndDate, &created.Status, &created.CreatedBy,
		&created.ClosedAt, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	whereClauses := []string{}
	args := []interface{}{}
	argIdx := 2

	if scopeDeptID != nil && *scopeDeptID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.current_holder_department_id = $%d", argIdx))
		args = append(args, *scopeDeptID)
		argIdx++
	}
	if scopeLoc != nil && *scopeLoc != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.location ILIKE $%d", argIdx))
		args = append(args, "%"+*scopeLoc+"%")
		argIdx++
	}

	if len(whereClauses) == 0 {
		whereClauses = append(whereClauses, "true")
	}

	query := fmt.Sprintf(
		`INSERT INTO audit_items (audit_cycle_id, asset_id)
		 SELECT $1, a.id FROM assets a WHERE %s
		 ON CONFLICT (audit_cycle_id, asset_id) DO NOTHING`,
		strings.Join(whereClauses, " AND "),
	)

	insertArgs := append([]interface{}{created.ID}, args...)
	_, err = tx.Exec(ctx, query, insertArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-populate audit items: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *AuditRepository) GetCycleByID(ctx context.Context, id string) (*models.AuditCycleDetail, error) {
	var detail models.AuditCycleDetail
	err := r.pool.QueryRow(ctx,
		`SELECT ac.id, ac.name, ac.scope_department_id, ac.scope_location,
		        ac.start_date::text, ac.end_date::text, ac.status, ac.created_by, ac.closed_at,
		        ac.created_at, ac.updated_at,
		        d.name AS scope_department_name,
		        e.name AS created_by_name
		 FROM audit_cycles ac
		 LEFT JOIN departments d ON d.id = ac.scope_department_id
		 LEFT JOIN employees e ON e.id = ac.created_by
		 WHERE ac.id = $1`, id,
	).Scan(
		&detail.ID, &detail.Name, &detail.ScopeDepartmentID, &detail.ScopeLocation,
		&detail.StartDate, &detail.EndDate, &detail.Status, &detail.CreatedBy, &detail.ClosedAt,
		&detail.CreatedAt, &detail.UpdatedAt,
		&detail.ScopeDepartmentName, &detail.CreatedByName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCycleNotFound
		}
		return nil, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT e.id, e.name, e.email, e.department_id, e.role, e.status, e.created_at, e.updated_at
		 FROM audit_cycle_auditors aca
		 JOIN employees e ON e.id = aca.employee_id
		 WHERE aca.audit_cycle_id = $1`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var emp models.Employee
		if err := rows.Scan(&emp.ID, &emp.Name, &emp.Email, &emp.DepartmentID, &emp.Role, &emp.Status, &emp.CreatedAt, &emp.UpdatedAt); err != nil {
			return nil, err
		}
		detail.AssignedAuditors = append(detail.AssignedAuditors, emp)
	}
	if detail.AssignedAuditors == nil {
		detail.AssignedAuditors = []models.Employee{}
	}

	var itemCount, verifiedCount, missingCount, damagedCount int
	r.pool.QueryRow(ctx,
		`SELECT COUNT(*),
		        COUNT(*) FILTER (WHERE result = 'Verified'),
		        COUNT(*) FILTER (WHERE result = 'Missing'),
		        COUNT(*) FILTER (WHERE result = 'Damaged')
		 FROM audit_items WHERE audit_cycle_id = $1`, id,
	).Scan(&itemCount, &verifiedCount, &missingCount, &damagedCount)
	detail.ItemCount = &itemCount
	detail.VerifiedCount = &verifiedCount
	detail.MissingCount = &missingCount
	detail.DamagedCount = &damagedCount

	return &detail, nil
}

func (r *AuditRepository) ListCycles(ctx context.Context) ([]models.AuditCycleDetail, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ac.id, ac.name, ac.scope_department_id, ac.scope_location,
		        ac.start_date::text, ac.end_date::text, ac.status, ac.created_by, ac.closed_at,
		        ac.created_at, ac.updated_at,
		        d.name AS scope_department_name,
		        e.name AS created_by_name
		 FROM audit_cycles ac
		 LEFT JOIN departments d ON d.id = ac.scope_department_id
		 LEFT JOIN employees e ON e.id = ac.created_by
		 ORDER BY ac.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cycles []models.AuditCycleDetail
	for rows.Next() {
		var c models.AuditCycleDetail
		if err := rows.Scan(
			&c.ID, &c.Name, &c.ScopeDepartmentID, &c.ScopeLocation,
			&c.StartDate, &c.EndDate, &c.Status, &c.CreatedBy, &c.ClosedAt,
			&c.CreatedAt, &c.UpdatedAt,
			&c.ScopeDepartmentName, &c.CreatedByName,
		); err != nil {
			return nil, err
		}
		var itemCount int
		r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_items WHERE audit_cycle_id = $1`, c.ID).Scan(&itemCount)
		c.ItemCount = &itemCount
		cycles = append(cycles, c)
	}
	if cycles == nil {
		return []models.AuditCycleDetail{}, nil
	}
	return cycles, nil
}

func (r *AuditRepository) AssignAuditors(ctx context.Context, cycleID string, employeeIDs []string) error {
	if len(employeeIDs) == 0 {
		return nil
	}

	values := []string{}
	args := []interface{}{cycleID}
	for i, empID := range employeeIDs {
		values = append(values, fmt.Sprintf("($1, $%d)", i+2))
		args = append(args, empID)
	}

	query := fmt.Sprintf(
		`INSERT INTO audit_cycle_auditors (audit_cycle_id, employee_id) VALUES %s ON CONFLICT DO NOTHING`,
		strings.Join(values, ", "),
	)

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *AuditRepository) GetAssignedAuditorIDs(ctx context.Context, cycleID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT employee_id FROM audit_cycle_auditors WHERE audit_cycle_id = $1`, cycleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if ids == nil {
		return []string{}, nil
	}
	return ids, nil
}

func (r *AuditRepository) CloseCycle(ctx context.Context, cycleID, closedBy string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var currentStatus string
	err = tx.QueryRow(ctx, `SELECT status FROM audit_cycles WHERE id = $1 FOR UPDATE`, cycleID).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCycleNotFound
		}
		return err
	}
	if currentStatus == "Closed" {
		return ErrCycleAlreadyClosed
	}

	_, err = tx.Exec(ctx,
		`UPDATE audit_cycles SET status = 'Closed', closed_at = now(), updated_at = now() WHERE id = $1`, cycleID,
	)
	if err != nil {
		return err
	}

	rows, err := tx.Query(ctx,
		`SELECT ai.asset_id FROM audit_items ai
		 WHERE ai.audit_cycle_id = $1 AND ai.result = 'Missing'`, cycleID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var missingAssetIDs []string
	for rows.Next() {
		var assetID string
		if err := rows.Scan(&assetID); err != nil {
			return err
		}
		missingAssetIDs = append(missingAssetIDs, assetID)
	}

	for _, assetID := range missingAssetIDs {
		var fromStatus string
		err = tx.QueryRow(ctx,
			`UPDATE assets SET status = 'Lost', updated_at = now() WHERE id = $1 RETURNING status`, assetID,
		).Scan(&fromStatus)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO asset_status_history (asset_id, from_status, to_status, changed_by, reason)
			 VALUES ($1, $2, 'Lost', $3, 'Confirmed missing via audit cycle closure')`,
			assetID, fromStatus, closedBy,
		)
		if err != nil {
			return err
		}
	}

	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"missing_asset_count": len(missingAssetIDs),
		"closed_by":           closedBy,
	})
	_, err = tx.Exec(ctx,
		`INSERT INTO activity_logs (actor_employee_id, action, entity_type, entity_id, metadata)
		 VALUES ($1, 'audit.close', 'audit_cycle', $2, $3)`,
		closedBy, cycleID, metadataJSON,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
