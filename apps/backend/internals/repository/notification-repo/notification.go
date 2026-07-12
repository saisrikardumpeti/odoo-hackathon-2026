package notification_repo

import (
	"context"
	"fmt"

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

func (r *NotificationRepository) ListByEmployee(ctx context.Context, employeeID string, unreadOnly bool) ([]models.Notification, error) {
	query := `SELECT id, employee_id, type, message, related_entity_type, related_entity_id, is_read, created_at
		 FROM notifications WHERE employee_id = $1`
	args := []interface{}{employeeID}

	if unreadOnly {
		query += ` AND is_read = false`
	}
	query += ` ORDER BY created_at DESC`

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
	return notifications, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE notifications SET is_read = true WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}
