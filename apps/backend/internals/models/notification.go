package models

import "time"

type Notification struct {
	ID                string    `json:"id"`
	EmployeeID        string    `json:"employee_id"`
	Type              string    `json:"type"`
	Message           string    `json:"message"`
	RelatedEntityType *string   `json:"related_entity_type"`
	RelatedEntityID   *string   `json:"related_entity_id"`
	IsRead            bool      `json:"is_read"`
	CreatedAt         time.Time `json:"created_at"`
}
