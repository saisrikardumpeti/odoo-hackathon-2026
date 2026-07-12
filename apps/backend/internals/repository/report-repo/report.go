package report_repo

import (
	"context"
	"fmt"
	"time"

	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

func (r *ReportRepository) GetUtilization(ctx context.Context, from, to *time.Time, idleDays int) ([]models.UtilizationReportItem, error) {
	query := `
		SELECT
			a.id, a.asset_tag, a.name, COALESCE(ac.name, 'Uncategorized'),
			COALESCE(alloc_counts.cnt, 0),
			COALESCE(book_counts.cnt, 0),
			COALESCE(alloc_counts.cnt, 0) + COALESCE(book_counts.cnt, 0),
			GREATEST(
				COALESCE(alloc_last.last_alloc, '1970-01-01'::timestamptz),
				COALESCE(book_last.last_book, '1970-01-01'::timestamptz)
			) AS last_activity
		FROM assets a
		LEFT JOIN asset_categories ac ON ac.id = a.category_id
		LEFT JOIN (
			SELECT asset_id, COUNT(*) AS cnt, MAX(created_at) AS last_alloc
			FROM allocations
			WHERE ($1::timestamptz IS NULL OR created_at >= $1::timestamptz)
			  AND ($2::timestamptz IS NULL OR created_at <= $2::timestamptz)
			GROUP BY asset_id
		) alloc_counts ON alloc_counts.asset_id = a.id
		LEFT JOIN (
			SELECT resource_asset_id, COUNT(*) AS cnt, MAX(created_at) AS last_book
			FROM bookings
			WHERE ($1::timestamptz IS NULL OR created_at >= $1::timestamptz)
			  AND ($2::timestamptz IS NULL OR created_at <= $2::timestamptz)
			GROUP BY resource_asset_id
		) book_counts ON book_counts.resource_asset_id = a.id
		LEFT JOIN LATERAL (
			SELECT created_at AS last_alloc
			FROM allocations
			WHERE asset_id = a.id
			ORDER BY created_at DESC
			LIMIT 1
		) alloc_last ON true
		LEFT JOIN LATERAL (
			SELECT created_at AS last_book
			FROM bookings
			WHERE resource_asset_id = a.id
			ORDER BY created_at DESC
			LIMIT 1
		) book_last ON true
		ORDER BY total_activity DESC
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query utilization: %w", err)
	}
	defer rows.Close()

	var items []models.UtilizationReportItem
	cutoff := time.Now().AddDate(0, 0, -idleDays)

	for rows.Next() {
		var item models.UtilizationReportItem
		var lastActivity time.Time
		if err := rows.Scan(
			&item.AssetID, &item.AssetTag, &item.AssetName, &item.CategoryName,
			&item.AllocationCount, &item.BookingCount, &item.TotalActivity,
			&lastActivity,
		); err != nil {
			return nil, fmt.Errorf("failed to scan utilization item: %w", err)
		}

		lastStr := lastActivity.Format(time.RFC3339)
		item.LastActivity = &lastStr

		if lastActivity.Before(cutoff) {
			days := int(time.Since(lastActivity).Hours() / 24)
			item.DaysIdle = &days
		}

		items = append(items, item)
	}
	if items == nil {
		items = []models.UtilizationReportItem{}
	}
	return items, nil
}

func (r *ReportRepository) GetMaintenanceFrequency(ctx context.Context, from, to *time.Time) ([]models.MaintenanceFrequencyItem, []models.MaintenanceCategoryItem, error) {
	assetQuery := `
		SELECT a.id, a.asset_tag, a.name, COALESCE(ac.name, 'Uncategorized'), COUNT(mr.id)::int
		FROM maintenance_requests mr
		JOIN assets a ON a.id = mr.asset_id
		LEFT JOIN asset_categories ac ON ac.id = a.category_id
		WHERE ($1::timestamptz IS NULL OR mr.created_at >= $1::timestamptz)
		  AND ($2::timestamptz IS NULL OR mr.created_at <= $2::timestamptz)
		GROUP BY a.id, a.asset_tag, a.name, ac.name
		ORDER BY COUNT(mr.id) DESC
	`

	assetRows, err := r.pool.Query(ctx, assetQuery, from, to)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query maintenance by asset: %w", err)
	}
	defer assetRows.Close()

	var assetItems []models.MaintenanceFrequencyItem
	for assetRows.Next() {
		var item models.MaintenanceFrequencyItem
		if err := assetRows.Scan(&item.AssetID, &item.AssetTag, &item.AssetName, &item.CategoryName, &item.Count); err != nil {
			return nil, nil, fmt.Errorf("failed to scan maintenance by asset: %w", err)
		}
		assetItems = append(assetItems, item)
	}
	if assetItems == nil {
		assetItems = []models.MaintenanceFrequencyItem{}
	}

	catQuery := `
		SELECT COALESCE(ac.name, 'Uncategorized'), COUNT(mr.id)::int
		FROM maintenance_requests mr
		JOIN assets a ON a.id = mr.asset_id
		LEFT JOIN asset_categories ac ON ac.id = a.category_id
		WHERE ($1::timestamptz IS NULL OR mr.created_at >= $1::timestamptz)
		  AND ($2::timestamptz IS NULL OR mr.created_at <= $2::timestamptz)
		GROUP BY ac.name
		ORDER BY COUNT(mr.id) DESC
	`

	catRows, err := r.pool.Query(ctx, catQuery, from, to)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query maintenance by category: %w", err)
	}
	defer catRows.Close()

	var catItems []models.MaintenanceCategoryItem
	for catRows.Next() {
		var item models.MaintenanceCategoryItem
		if err := catRows.Scan(&item.CategoryName, &item.Count); err != nil {
			return nil, nil, fmt.Errorf("failed to scan maintenance by category: %w", err)
		}
		catItems = append(catItems, item)
	}
	if catItems == nil {
		catItems = []models.MaintenanceCategoryItem{}
	}

	return assetItems, catItems, nil
}

func (r *ReportRepository) GetRetirementWatchlist(ctx context.Context, ageYearsThreshold float64) ([]models.RetirementWatchlistItem, error) {
	query := `
		SELECT a.id, a.asset_tag, a.name, COALESCE(ac.name, 'Uncategorized'),
		       a.acquisition_date::text,
		       EXTRACT(YEAR FROM age(CURRENT_DATE, a.acquisition_date))::float8,
		       a.status::text
		FROM assets a
		LEFT JOIN asset_categories ac ON ac.id = a.category_id
		WHERE a.status NOT IN ('Retired', 'Disposed')
		  AND a.acquisition_date IS NOT NULL
		  AND EXTRACT(YEAR FROM age(CURRENT_DATE, a.acquisition_date)) >= $1
		ORDER BY age_years DESC
	`

	rows, err := r.pool.Query(ctx, query, ageYearsThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to query retirement watchlist: %w", err)
	}
	defer rows.Close()

	var items []models.RetirementWatchlistItem
	for rows.Next() {
		var item models.RetirementWatchlistItem
		if err := rows.Scan(
			&item.AssetID, &item.AssetTag, &item.AssetName, &item.CategoryName,
			&item.AcquisitionDate, &item.AgeYears, &item.Status,
		); err != nil {
			return nil, fmt.Errorf("failed to scan retirement watchlist item: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.RetirementWatchlistItem{}
	}
	return items, nil
}

func (r *ReportRepository) GetAllocationSummary(ctx context.Context) ([]models.AllocationSummaryItem, error) {
	query := `
		SELECT d.id, d.name, COUNT(al.id)::int
		FROM departments d
		LEFT JOIN allocations al ON al.department_id = d.id AND al.status = 'Active'
		GROUP BY d.id, d.name
		ORDER BY COUNT(al.id) DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query allocation summary: %w", err)
	}
	defer rows.Close()

	var items []models.AllocationSummaryItem
	for rows.Next() {
		var item models.AllocationSummaryItem
		if err := rows.Scan(&item.DepartmentID, &item.DepartmentName, &item.AssetCount); err != nil {
			return nil, fmt.Errorf("failed to scan allocation summary item: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.AllocationSummaryItem{}
	}
	return items, nil
}

func (r *ReportRepository) GetBookingHeatmap(ctx context.Context, from, to *time.Time) ([]models.BookingHeatmapItem, error) {
	query := `
		SELECT EXTRACT(DOW FROM start_time)::int AS day_of_week,
		       EXTRACT(HOUR FROM start_time)::int AS hour,
		       COUNT(*)::int
		FROM bookings
		WHERE status <> 'Cancelled'
		  AND ($1::timestamptz IS NULL OR start_time >= $1::timestamptz)
		  AND ($2::timestamptz IS NULL OR end_time <= $2::timestamptz)
		GROUP BY day_of_week, hour
		ORDER BY day_of_week, hour
	`

	rows, err := r.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query booking heatmap: %w", err)
	}
	defer rows.Close()

	var items []models.BookingHeatmapItem
	for rows.Next() {
		var item models.BookingHeatmapItem
		if err := rows.Scan(&item.DayOfWeek, &item.Hour, &item.Count); err != nil {
			return nil, fmt.Errorf("failed to scan booking heatmap item: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.BookingHeatmapItem{}
	}
	return items, nil
}
