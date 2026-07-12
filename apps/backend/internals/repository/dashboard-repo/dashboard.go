package dashboard_repo

import (
	"context"
	"fmt"
	"time"

	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

func (r *DashboardRepository) GetKPIs(ctx context.Context, employeeID *string, role string) (*models.KPIsResponse, error) {
	var kpi models.KPIsResponse

	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM assets WHERE status = 'Available'`).Scan(&kpi.AssetsAvailable); err != nil {
		return nil, fmt.Errorf("failed to count available assets: %w", err)
	}

	assetAllocatedQuery := `SELECT COUNT(*) FROM allocations WHERE status = 'Active'`
	var allocArgs []interface{}
	if employeeID != nil && role == "Employee" {
		assetAllocatedQuery += ` AND employee_id = $1`
		allocArgs = append(allocArgs, *employeeID)
	}
	if err := r.pool.QueryRow(ctx, assetAllocatedQuery, allocArgs...).Scan(&kpi.AssetsAllocated); err != nil {
		return nil, fmt.Errorf("failed to count allocated assets: %w", err)
	}

	todayStart := time.Now().Truncate(24 * time.Hour)
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM maintenance_requests WHERE created_at >= $1 OR updated_at >= $1`, todayStart,
	).Scan(&kpi.MaintenanceToday); err != nil {
		return nil, fmt.Errorf("failed to count maintenance today: %w", err)
	}

	bookingQuery := `SELECT COUNT(*) FROM bookings WHERE status IN ('Upcoming', 'Ongoing')`
	var bookingArgs []interface{}
	if employeeID != nil && role == "Employee" {
		bookingQuery += ` AND booked_by_employee_id = $1`
		bookingArgs = append(bookingArgs, *employeeID)
	}
	if err := r.pool.QueryRow(ctx, bookingQuery, bookingArgs...).Scan(&kpi.ActiveBookings); err != nil {
		return nil, fmt.Errorf("failed to count active bookings: %w", err)
	}

	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM transfer_requests WHERE status = 'Requested'`,
	).Scan(&kpi.PendingTransfers); err != nil {
		return nil, fmt.Errorf("failed to count pending transfers: %w", err)
	}

	upcomingQuery := `SELECT COUNT(*) FROM allocations WHERE status = 'Active' AND expected_return_date IS NOT NULL AND expected_return_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '7 days'`
	var upcomingArgs []interface{}
	if employeeID != nil && role == "Employee" {
		upcomingQuery += ` AND employee_id = $1`
		upcomingArgs = append(upcomingArgs, *employeeID)
	}
	if err := r.pool.QueryRow(ctx, upcomingQuery, upcomingArgs...).Scan(&kpi.UpcomingReturns); err != nil {
		return nil, fmt.Errorf("failed to count upcoming returns: %w", err)
	}

	return &kpi, nil
}

func (r *DashboardRepository) GetOverdue(ctx context.Context, employeeID *string, role string) ([]models.OverdueItem, error) {
	query := `SELECT al.id, al.asset_id, a.asset_tag, a.name, al.employee_id, e.name, al.expected_return_date::text,
	             (CURRENT_DATE - al.expected_return_date)::int
	      FROM allocations al
	      JOIN assets a ON a.id = al.asset_id
	      LEFT JOIN employees e ON e.id = al.employee_id
	      WHERE al.status = 'Active' AND al.expected_return_date < CURRENT_DATE`

	var args []interface{}
	argIdx := 1

	if employeeID != nil && role == "Employee" {
		query += fmt.Sprintf(` AND al.employee_id = $%d`, argIdx)
		args = append(args, *employeeID)
		argIdx++
	}

	query += ` ORDER BY al.expected_return_date ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list overdue items: %w", err)
	}
	defer rows.Close()

	var items []models.OverdueItem
	for rows.Next() {
		var item models.OverdueItem
		item.Type = "allocation"
		if err := rows.Scan(&item.ID, &item.AssetID, &item.AssetTag, &item.AssetName, &item.EmployeeID, &item.EmployeeName, &item.ExpectedReturnDate, &item.DaysOverdue); err != nil {
			return nil, fmt.Errorf("failed to scan overdue item: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.OverdueItem{}
	}
	return items, nil
}

func (r *DashboardRepository) GetUpcoming(ctx context.Context, employeeID *string, role string, windowDays int) ([]models.UpcomingItem, error) {
	query := `SELECT al.id, al.asset_id, a.asset_tag, a.name, al.employee_id, e.name, al.expected_return_date::text,
	             (al.expected_return_date - CURRENT_DATE)::int
	      FROM allocations al
	      JOIN assets a ON a.id = al.asset_id
	      LEFT JOIN employees e ON e.id = al.employee_id
	      WHERE al.status = 'Active' AND al.expected_return_date IS NOT NULL
	        AND al.expected_return_date >= CURRENT_DATE
	        AND al.expected_return_date <= CURRENT_DATE + $1 * INTERVAL '1 day'`

	var args []interface{}
	args = append(args, windowDays)
	argIdx := 2

	if employeeID != nil && role == "Employee" {
		query += fmt.Sprintf(` AND al.employee_id = $%d`, argIdx)
		args = append(args, *employeeID)
		argIdx++
	}

	query += ` ORDER BY al.expected_return_date ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list upcoming items: %w", err)
	}
	defer rows.Close()

	var items []models.UpcomingItem
	for rows.Next() {
		var item models.UpcomingItem
		item.Type = "allocation"
		if err := rows.Scan(&item.ID, &item.AssetID, &item.AssetTag, &item.AssetName, &item.EmployeeID, &item.EmployeeName, &item.ExpectedDate, &item.DaysUntilDue); err != nil {
			return nil, fmt.Errorf("failed to scan upcoming item: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.UpcomingItem{}
	}
	return items, nil
}

func (r *DashboardRepository) GetRecentActivity(ctx context.Context, employeeID *string, role string, limit int) ([]models.RecentActivityItem, error) {
	query := `SELECT al.id, al.action, al.entity_type, e.name, al.created_at::text
	          FROM activity_logs al
	          LEFT JOIN employees e ON e.id = al.actor_employee_id`

	var args []interface{}
	if employeeID != nil && role == "Employee" {
		query += ` WHERE al.actor_employee_id = $1`
		args = append(args, *employeeID)
	}

	query += ` ORDER BY al.created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list recent activity: %w", err)
	}
	defer rows.Close()

	var items []models.RecentActivityItem
	for rows.Next() {
		var item models.RecentActivityItem
		if err := rows.Scan(&item.ID, &item.Action, &item.EntityType, &item.ActorName, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan recent activity: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.RecentActivityItem{}
	}
	return items, nil
}
