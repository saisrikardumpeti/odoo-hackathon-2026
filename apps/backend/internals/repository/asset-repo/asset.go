package asset_repo
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

var (
	ErrAssetNotFound      = errors.New("asset not found")
	ErrCategoryNotFound   = errors.New("category not found")
)

func (r *AssetRepository) Create(ctx context.Context, asset models.Asset) (*models.Asset, error) {
	var a models.Asset

	customFieldsJSON, err := json.Marshal(asset.CustomFields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal custom_fields: %w", err)
	}
	if asset.CustomFields == nil {
		customFieldsJSON = []byte("{}")
	}

	err = r.pool.QueryRow(ctx,
		`INSERT INTO assets (name, category_id, serial_number, acquisition_date, acquisition_cost, condition, location, is_bookable, custom_fields)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, asset_tag, name, category_id, serial_number, acquisition_date::text, acquisition_cost,
		           condition, location, is_bookable, status, current_holder_employee_id,
		           current_holder_department_id, qr_code, custom_fields, created_at, updated_at`,
		asset.Name, asset.CategoryID, asset.SerialNumber, asset.AcquisitionDate,
		asset.AcquisitionCost, asset.Condition, asset.Location, asset.IsBookable, customFieldsJSON,
	).Scan(
		&a.ID, &a.AssetTag, &a.Name, &a.CategoryID, &a.SerialNumber,
		&a.AcquisitionDate, &a.AcquisitionCost, &a.Condition, &a.Location,
		&a.IsBookable, &a.Status, &a.CurrentHolderEmployeeID,
		&a.CurrentHolderDepartmentID, &a.QRCode, &customFieldsJSON, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if isCategoryFKViolation(err) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}
	if err := json.Unmarshal(customFieldsJSON, &a.CustomFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
	}
	return &a, nil
}

func (r *AssetRepository) CreateStatusHistory(ctx context.Context, h models.AssetStatusHistory) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO asset_status_history (asset_id, from_status, to_status, changed_by, reason)
		 VALUES ($1, $2, $3, $4, $5)`,
		h.AssetID, h.FromStatus, h.ToStatus, h.ChangedBy, h.Reason,
	)
	if err != nil {
		return fmt.Errorf("failed to create status history: %w", err)
	}
	return nil
}

type AssetListFilters struct {
	AssetTag        string
	SerialNumber    string
	CategoryID      string
	Status          string
	DepartmentID    string
	Location        string
	Page            int
	PageSize        int
}

type AssetListResult struct {
	Assets     []models.AssetListItem
	TotalCount int
}

func (r *AssetRepository) List(ctx context.Context, filters AssetListFilters) (*AssetListResult, error) {
	conditions := []string{}
	args := []any{}
	argIdx := 1

	if filters.AssetTag != "" {
		conditions = append(conditions, fmt.Sprintf("a.asset_tag ILIKE $%d", argIdx))
		args = append(args, "%"+filters.AssetTag+"%")
		argIdx++
	}
	if filters.SerialNumber != "" {
		conditions = append(conditions, fmt.Sprintf("a.serial_number ILIKE $%d", argIdx))
		args = append(args, "%"+filters.SerialNumber+"%")
		argIdx++
	}
	if filters.CategoryID != "" {
		conditions = append(conditions, fmt.Sprintf("a.category_id = $%d", argIdx))
		args = append(args, filters.CategoryID)
		argIdx++
	}
	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("a.status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.DepartmentID != "" {
		conditions = append(conditions, fmt.Sprintf("a.current_holder_department_id = $%d", argIdx))
		args = append(args, filters.DepartmentID)
		argIdx++
	}
	if filters.Location != "" {
		conditions = append(conditions, fmt.Sprintf("a.location ILIKE $%d", argIdx))
		args = append(args, "%"+filters.Location+"%")
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
		filters.PageSize = 20
	}
	offset := (filters.Page - 1) * filters.PageSize

	var totalCount int
	countQuery := "SELECT COUNT(*) FROM assets a" + whereClause
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count assets: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT a.id, a.asset_tag, a.name, COALESCE(c.name, '') as category_name,
		        a.serial_number, a.status, a.location, a.current_holder_department_id, a.is_bookable
		 FROM assets a
		 LEFT JOIN asset_categories c ON c.id = a.category_id
		 %s
		 ORDER BY a.created_at DESC
		 LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1,
	)
	args = append(args, filters.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}
	defer rows.Close()

	assets := []models.AssetListItem{}
	for rows.Next() {
		var item models.AssetListItem
		if err := rows.Scan(
			&item.ID, &item.AssetTag, &item.Name, &item.CategoryName,
			&item.SerialNumber, &item.Status, &item.Location,
			&item.CurrentHolderDepartmentID, &item.IsBookable,
		); err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, item)
	}

	return &AssetListResult{Assets: assets, TotalCount: totalCount}, nil
}

func (r *AssetRepository) GetByID(ctx context.Context, id string) (*models.AssetDetail, error) {
	var a models.AssetDetail
	var catName, holderName, deptName *string
	var cfJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT a.id, a.asset_tag, a.name, a.category_id, a.serial_number,
		        a.acquisition_date::text, a.acquisition_cost, a.condition, a.location,
		        a.is_bookable, a.status, a.current_holder_employee_id,
		        a.current_holder_department_id, a.qr_code, a.custom_fields, a.created_at, a.updated_at,
		        c.name,
		        e.name,
		        d.name
		 FROM assets a
		 LEFT JOIN asset_categories c ON c.id = a.category_id
		 LEFT JOIN employees e ON e.id = a.current_holder_employee_id
		 LEFT JOIN departments d ON d.id = a.current_holder_department_id
		 WHERE a.id = $1`, id,
	).Scan(
		&a.ID, &a.AssetTag, &a.Name, &a.CategoryID, &a.SerialNumber,
		&a.AcquisitionDate, &a.AcquisitionCost, &a.Condition, &a.Location,
		&a.IsBookable, &a.Status, &a.CurrentHolderEmployeeID,
		&a.CurrentHolderDepartmentID, &a.QRCode, &cfJSON, &a.CreatedAt, &a.UpdatedAt,
		&catName, &holderName, &deptName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}
	a.CategoryName = catName
	a.CurrentHolderName = holderName
	a.CurrentHolderDepartmentName = deptName
	if err := json.Unmarshal(cfJSON, &a.CustomFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
	}
	if a.CustomFields == nil {
		a.CustomFields = map[string]interface{}{}
	}
	return &a, nil
}

func (r *AssetRepository) GetHistory(ctx context.Context, assetID string) ([]models.HistoryEvent, error) {
	query := `
		SELECT changed_at, 'status_change', jsonb_build_object(
			'id', id,
			'from_status', from_status,
			'to_status', to_status,
			'changed_by', changed_by,
			'reason', reason
		)
		FROM asset_status_history
		WHERE asset_id = $1
	UNION ALL
		SELECT allocated_at, 'allocation', jsonb_build_object(
			'id', id,
			'employee_id', employee_id,
			'department_id', department_id,
			'allocated_by', allocated_by,
			'expected_return_date', expected_return_date,
			'returned_at', returned_at,
			'status', status
		)
		FROM allocations
		WHERE asset_id = $1
	UNION ALL
		SELECT created_at, 'maintenance', jsonb_build_object(
			'id', id,
			'raised_by_employee_id', raised_by_employee_id,
			'issue_description', issue_description,
			'priority', priority,
			'status', status,
			'resolved_at', resolved_at,
			'resolution_notes', resolution_notes
		)
		FROM maintenance_requests
		WHERE asset_id = $1
	ORDER BY 1 ASC
	`

	rows, err := r.pool.Query(ctx, query, assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset history: %w", err)
	}
	defer rows.Close()

	events := []models.HistoryEvent{}
	for rows.Next() {
		var e models.HistoryEvent
		var ts time.Time
		var data []byte
		if err := rows.Scan(&ts, &e.Type, &data); err != nil {
			return nil, fmt.Errorf("failed to scan history event: %w", err)
		}
		e.Timestamp = ts
		e.Data = json.RawMessage(data)
		events = append(events, e)
	}

	if events == nil {
		events = []models.HistoryEvent{}
	}
	return events, nil
}

func (r *AssetRepository) CreateDocument(ctx context.Context, doc models.AssetDocument) (*models.AssetDocument, error) {
	var d models.AssetDocument
	err := r.pool.QueryRow(ctx,
		`INSERT INTO asset_documents (asset_id, url, type)
		 VALUES ($1, $2, $3)
		 RETURNING id, asset_id, url, type, uploaded_at`,
		doc.AssetID, doc.URL, doc.Type,
	).Scan(&d.ID, &d.AssetID, &d.URL, &d.Type, &d.UploadedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}
	return &d, nil
}

func isCategoryFKViolation(err error) bool {
	return strings.Contains(err.Error(), "violates foreign key constraint") &&
		strings.Contains(err.Error(), "category_id")
}

