package models

import "time"

type ActivityLog struct {
	ID              string                 `json:"id"`
	ActorEmployeeID *string                `json:"actor_employee_id"`
	Action          string                 `json:"action"`
	EntityType      string                 `json:"entity_type"`
	EntityID        *string                `json:"entity_id"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
}

type ActivityLogDetail struct {
	ActivityLog
	ActorName *string `json:"actor_name"`
}
