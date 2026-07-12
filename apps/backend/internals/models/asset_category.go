package models

import "time"

type AssetCategory struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}
