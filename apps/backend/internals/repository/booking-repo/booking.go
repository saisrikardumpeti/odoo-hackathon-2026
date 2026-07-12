package booking_repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

var (
	ErrBookingNotFound = errors.New("booking not found")
	ErrBookingOverlap  = errors.New("booking overlaps with an existing booking")
)

func (r *BookingRepository) ListByResource(ctx context.Context, assetID string, from, to *time.Time) ([]models.BookingDetail, error) {
	query := `SELECT b.id, b.resource_asset_id, b.booked_by_employee_id, b.start_time, b.end_time,
	                  b.purpose, b.status, b.created_at, b.updated_at,
	                  a.name, a.asset_tag, e.name
	           FROM bookings b
	           JOIN assets a ON a.id = b.resource_asset_id
	           JOIN employees e ON e.id = b.booked_by_employee_id
	           WHERE b.resource_asset_id = $1`
	args := []interface{}{assetID}
	argIdx := 2

	if from != nil {
		query += fmt.Sprintf(` AND b.end_time > $%d`, argIdx)
		args = append(args, *from)
		argIdx++
	}
	if to != nil {
		query += fmt.Sprintf(` AND b.start_time < $%d`, argIdx)
		args = append(args, *to)
		argIdx++
	}
	query += ` ORDER BY b.start_time ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings by resource: %w", err)
	}
	defer rows.Close()

	var bookings []models.BookingDetail
	for rows.Next() {
		var b models.BookingDetail
		if err := rows.Scan(
			&b.ID, &b.ResourceAssetID, &b.BookedByEmployeeID, &b.StartTime, &b.EndTime,
			&b.Purpose, &b.Status, &b.CreatedAt, &b.UpdatedAt,
			&b.AssetName, &b.AssetTag, &b.BookedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, b)
	}
	if bookings == nil {
		bookings = []models.BookingDetail{}
	}
	return bookings, nil
}

func (r *BookingRepository) Create(ctx context.Context, b models.Booking) (*models.Booking, error) {
	var booking models.Booking
	err := r.pool.QueryRow(ctx,
		`INSERT INTO bookings (resource_asset_id, booked_by_employee_id, start_time, end_time, purpose)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, resource_asset_id, booked_by_employee_id, start_time, end_time,
		           purpose, status, created_at, updated_at`,
		b.ResourceAssetID, b.BookedByEmployeeID, b.StartTime, b.EndTime, b.Purpose,
	).Scan(
		&booking.ID, &booking.ResourceAssetID, &booking.BookedByEmployeeID,
		&booking.StartTime, &booking.EndTime, &booking.Purpose,
		&booking.Status, &booking.CreatedAt, &booking.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			return nil, ErrBookingOverlap
		}
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}
	return &booking, nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id string) (*models.BookingDetail, error) {
	var b models.BookingDetail
	err := r.pool.QueryRow(ctx,
		`SELECT b.id, b.resource_asset_id, b.booked_by_employee_id, b.start_time, b.end_time,
		        b.purpose, b.status, b.created_at, b.updated_at,
		        a.name, a.asset_tag, e.name
		 FROM bookings b
		 JOIN assets a ON a.id = b.resource_asset_id
		 JOIN employees e ON e.id = b.booked_by_employee_id
		 WHERE b.id = $1`, id,
	).Scan(
		&b.ID, &b.ResourceAssetID, &b.BookedByEmployeeID, &b.StartTime, &b.EndTime,
		&b.Purpose, &b.Status, &b.CreatedAt, &b.UpdatedAt,
		&b.AssetName, &b.AssetTag, &b.BookedByName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	return &b, nil
}

func (r *BookingRepository) Cancel(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE bookings SET status = 'Cancelled', updated_at = now() WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrBookingNotFound
	}
	return nil
}

func (r *BookingRepository) Reschedule(ctx context.Context, id string, startTime, endTime time.Time) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE bookings SET start_time = $1, end_time = $2, updated_at = now() WHERE id = $3`,
		startTime, endTime, id,
	)
	if err != nil {
		return fmt.Errorf("failed to reschedule booking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrBookingNotFound
	}
	return nil
}

func (r *BookingRepository) ListByBooker(ctx context.Context, employeeID string) ([]models.BookingDetail, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT b.id, b.resource_asset_id, b.booked_by_employee_id, b.start_time, b.end_time,
		        b.purpose, b.status, b.created_at, b.updated_at,
		        a.name, a.asset_tag, e.name
		 FROM bookings b
		 JOIN assets a ON a.id = b.resource_asset_id
		 JOIN employees e ON e.id = b.booked_by_employee_id
		 WHERE b.booked_by_employee_id = $1
		 ORDER BY b.start_time DESC`, employeeID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings by booker: %w", err)
	}
	defer rows.Close()

	var bookings []models.BookingDetail
	for rows.Next() {
		var b models.BookingDetail
		if err := rows.Scan(
			&b.ID, &b.ResourceAssetID, &b.BookedByEmployeeID, &b.StartTime, &b.EndTime,
			&b.Purpose, &b.Status, &b.CreatedAt, &b.UpdatedAt,
			&b.AssetName, &b.AssetTag, &b.BookedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, b)
	}
	if bookings == nil {
		bookings = []models.BookingDetail{}
	}
	return bookings, nil
}

func (r *BookingRepository) FindConflicting(ctx context.Context, assetID string, startTime, endTime time.Time, excludeID *string) ([]models.BookingDetail, error) {
	query := `SELECT b.id, b.resource_asset_id, b.booked_by_employee_id, b.start_time, b.end_time,
	                  b.purpose, b.status, b.created_at, b.updated_at,
	                  a.name, a.asset_tag, e.name
	           FROM bookings b
	           JOIN assets a ON a.id = b.resource_asset_id
	           JOIN employees e ON e.id = b.booked_by_employee_id
	           WHERE b.resource_asset_id = $1
	             AND b.status <> 'Cancelled'
	             AND tstzrange(b.start_time, b.end_time) && tstzrange($2, $3)`
	args := []interface{}{assetID, startTime, endTime}

	if excludeID != nil {
		query += ` AND b.id <> $4`
		args = append(args, *excludeID)
	}
	query += ` ORDER BY b.start_time ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find conflicting bookings: %w", err)
	}
	defer rows.Close()

	var bookings []models.BookingDetail
	for rows.Next() {
		var b models.BookingDetail
		if err := rows.Scan(
			&b.ID, &b.ResourceAssetID, &b.BookedByEmployeeID, &b.StartTime, &b.EndTime,
			&b.Purpose, &b.Status, &b.CreatedAt, &b.UpdatedAt,
			&b.AssetName, &b.AssetTag, &b.BookedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan conflicting booking: %w", err)
		}
		bookings = append(bookings, b)
	}
	if bookings == nil {
		bookings = []models.BookingDetail{}
	}
	return bookings, nil
}

func (r *BookingRepository) TransitionStatuses(ctx context.Context) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE bookings SET status = 'Ongoing', updated_at = now()
		 WHERE status = 'Upcoming' AND start_time <= now()`,
	)
	if err != nil {
		return fmt.Errorf("failed to transition Upcoming->Ongoing: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`UPDATE bookings SET status = 'Completed', updated_at = now()
		 WHERE status = 'Ongoing' AND end_time <= now()`,
	)
	if err != nil {
		return fmt.Errorf("failed to transition Ongoing->Completed: %w", err)
	}

	return nil
}

func (r *BookingRepository) CreateReminders(ctx context.Context, beforeMinutes int) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO notifications (employee_id, type, message, related_entity_type, related_entity_id)
		 SELECT b.booked_by_employee_id, 'BookingReminder',
		        'Your booking for ' || a.name || ' starts at ' || b.start_time::text,
		        'booking', b.id
		 FROM bookings b
		 JOIN assets a ON a.id = b.resource_asset_id
		 WHERE b.status IN ('Upcoming', 'Ongoing')
		   AND b.start_time > now()
		   AND b.start_time <= now() + ($1 || ' minutes')::interval
		   AND NOT EXISTS (
		       SELECT 1 FROM notifications n
		       WHERE n.related_entity_type = 'booking'
		         AND n.related_entity_id = b.id
		         AND n.type = 'BookingReminder'
		   )`,
		fmt.Sprintf("%d", beforeMinutes),
	)
	if err != nil {
		return fmt.Errorf("failed to create reminders: %w", err)
	}
	return nil
}
