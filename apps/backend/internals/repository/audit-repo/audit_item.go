package audit_repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

func (r *AuditRepository) ListItems(ctx context.Context, cycleID string, filterAuditorID *string) ([]models.AuditItemDetail, error) {
	query := `SELECT ai.id, ai.audit_cycle_id, ai.asset_id, ai.auditor_id, ai.result, ai.notes,
	                 ai.verified_at, ai.created_at, ai.updated_at,
	                 a.asset_tag, a.name AS asset_name, a.status AS asset_status, a.location AS asset_location
	          FROM audit_items ai
	          JOIN assets a ON a.id = ai.asset_id
	          WHERE ai.audit_cycle_id = $1`
	args := []interface{}{cycleID}
	argIdx := 2

	if filterAuditorID != nil && *filterAuditorID != "" {
		query += fmt.Sprintf(" AND ai.auditor_id = $%d", argIdx)
		args = append(args, *filterAuditorID)
		argIdx++
	}

	query += " ORDER BY a.asset_tag"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AuditItemDetail
	for rows.Next() {
		var item models.AuditItemDetail
		if err := rows.Scan(
			&item.ID, &item.AuditCycleID, &item.AssetID, &item.AuditorID,
			&item.Result, &item.Notes, &item.VerifiedAt, &item.CreatedAt, &item.UpdatedAt,
			&item.AssetTag, &item.AssetName, &item.AssetStatus, &item.AssetLocation,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		return []models.AuditItemDetail{}, nil
	}
	return items, nil
}

func (r *AuditRepository) GetItemByID(ctx context.Context, id string) (*models.AuditItem, error) {
	var item models.AuditItem
	err := r.pool.QueryRow(ctx,
		`SELECT id, audit_cycle_id, asset_id, auditor_id, result, notes, verified_at, created_at, updated_at
		 FROM audit_items WHERE id = $1`, id,
	).Scan(&item.ID, &item.AuditCycleID, &item.AssetID, &item.AuditorID,
		&item.Result, &item.Notes, &item.VerifiedAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *AuditRepository) VerifyItem(ctx context.Context, itemID, auditorID string, result, notes *string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var cycleStatus string
	var itemCycleID string
	var assetID string
	err = tx.QueryRow(ctx,
		`SELECT ai.audit_cycle_id, ai.asset_id, ai.result, ac.status
		 FROM audit_items ai
		 JOIN audit_cycles ac ON ac.id = ai.audit_cycle_id
		 WHERE ai.id = $1 FOR UPDATE OF ai`, itemID,
	).Scan(&itemCycleID, &assetID, nil, &cycleStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrItemNotFound
		}
		return err
	}

	if cycleStatus == "Closed" {
		return errors.New("cannot update items in a closed cycle")
	}

	var isAssigned bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM audit_cycle_auditors WHERE audit_cycle_id = $1 AND employee_id = $2)`,
		itemCycleID, auditorID,
	).Scan(&isAssigned)
	if err != nil {
		return err
	}
	if !isAssigned {
		return ErrNotAssignedAuditor
	}

	if result != nil && *result != "" {
		_, err = tx.Exec(ctx,
			`UPDATE audit_items SET result = $1, auditor_id = $2, notes = COALESCE($3, notes), verified_at = now(), updated_at = now()
			 WHERE id = $4`,
			*result, auditorID, notes, itemID,
		)
	} else {
		_, err = tx.Exec(ctx,
			`UPDATE audit_items SET auditor_id = $1, notes = COALESCE($2, notes), updated_at = now()
			 WHERE id = $3`,
			auditorID, notes, itemID,
		)
	}
	if err != nil {
		return err
	}

	if result != nil && (*result == "Missing" || *result == "Damaged") {
		_, err = tx.Exec(ctx,
			`INSERT INTO discrepancy_reports (audit_cycle_id, asset_id, audit_item_id, issue_type)
			 VALUES ($1, $2, $3, $4)`,
			itemCycleID, assetID, itemID, *result,
		)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				_, err = tx.Exec(ctx,
					`UPDATE discrepancy_reports SET issue_type = $1, resolved = false, resolved_by = NULL, resolved_at = NULL, updated_at = now()
					 WHERE audit_item_id = $2`,
					*result, itemID,
				)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO notifications (employee_id, type, message, related_entity_type, related_entity_id)
			 VALUES ($1, 'AuditDiscrepancyFlagged',
			         'Audit flagged asset as ' || $2 || ' during cycle audit',
			         'audit_cycle', $3)`,
			auditorID, *result, itemCycleID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *AuditRepository) ListDiscrepancyReports(ctx context.Context, cycleID *string, resolved *bool) ([]models.DiscrepancyReportDetail, error) {
	query := `SELECT dr.id, dr.audit_cycle_id, dr.asset_id, dr.audit_item_id, dr.issue_type,
	                 dr.resolved, dr.resolved_by, dr.resolved_at, dr.created_at, dr.updated_at,
	                 ac.name AS cycle_name, a.asset_tag, a.name AS asset_name,
	                 e.name AS resolved_by_name
	          FROM discrepancy_reports dr
	          JOIN audit_cycles ac ON ac.id = dr.audit_cycle_id
	          JOIN assets a ON a.id = dr.asset_id
	          LEFT JOIN employees e ON e.id = dr.resolved_by
	          WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if cycleID != nil && *cycleID != "" {
		query += fmt.Sprintf(" AND dr.audit_cycle_id = $%d", argIdx)
		args = append(args, *cycleID)
		argIdx++
	}

	if resolved != nil {
		query += fmt.Sprintf(" AND dr.resolved = $%d", argIdx)
		args = append(args, *resolved)
		argIdx++
	}

	query += " ORDER BY dr.created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []models.DiscrepancyReportDetail
	for rows.Next() {
		var r models.DiscrepancyReportDetail
		if err := rows.Scan(
			&r.ID, &r.AuditCycleID, &r.AssetID, &r.AuditItemID, &r.IssueType,
			&r.Resolved, &r.ResolvedBy, &r.ResolvedAt, &r.CreatedAt, &r.UpdatedAt,
			&r.CycleName, &r.AssetTag, &r.AssetName, &r.ResolvedByName,
		); err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	if reports == nil {
		return []models.DiscrepancyReportDetail{}, nil
	}
	return reports, nil
}

func (r *AuditRepository) ResolveDiscrepancy(ctx context.Context, id, resolvedBy string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE discrepancy_reports SET resolved = true, resolved_by = $1, resolved_at = now(), updated_at = now()
		 WHERE id = $2 AND resolved = false`,
		resolvedBy, id,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("discrepancy report not found or already resolved")
	}
	return nil
}
