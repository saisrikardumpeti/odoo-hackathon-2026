package notification_repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

func (r *NotificationRepository) Create(ctx context.Context, n models.Notification) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO notifications (employee_id, type, message, related_entity_type, related_entity_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		n.EmployeeID, n.Type, n.Message, n.RelatedEntityType, n.RelatedEntityID,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

func (r *NotificationRepository) ListByEmployee(ctx context.Context, employeeID string, unreadOnly bool, page, pageSize int) (*NotificationListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int
	countQuery := `SELECT COUNT(*) FROM notifications WHERE employee_id = $1`
	if unreadOnly {
		countQuery += ` AND is_read = false`
	}
	err := r.pool.QueryRow(ctx, countQuery, employeeID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	var unreadCount int
	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE employee_id = $1 AND is_read = false`,
		employeeID,
	).Scan(&unreadCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	query := `SELECT id, employee_id, type, message, related_entity_type, related_entity_id, is_read, created_at
		 FROM notifications WHERE employee_id = $1`
	args := []interface{}{employeeID}

	if unreadOnly {
		query += ` AND is_read = false`
	}
	query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.EmployeeID, &n.Type, &n.Message, &n.RelatedEntityType, &n.RelatedEntityID, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}
	if notifications == nil {
		notifications = []models.Notification{}
	}

	return &NotificationListResult{
		Notifications: notifications,
		Total:         total,
		UnreadCount:   unreadCount,
	}, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE notifications SET is_read = true WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *NotificationRepository) MarkReadAll(ctx context.Context, employeeID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE notifications SET is_read = true WHERE employee_id = $1 AND is_read = false`, employeeID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

func (r *NotificationRepository) Exists(ctx context.Context, employeeID, notifType string, relatedEntityID string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE employee_id = $1 AND type = $2 AND related_entity_id = $3`,
		employeeID, notifType, relatedEntityID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check notification exists: %w", err)
	}
	return count > 0, nil
}

func (r *NotificationRepository) UnreadCount(ctx context.Context, employeeID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE employee_id = $1 AND is_read = false`,
		employeeID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}
	return count, nil
}


