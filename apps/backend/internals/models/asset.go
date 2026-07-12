package models

import (
	"encoding/json"
	"time"
)

type Asset struct {
	ID                        string    `json:"id"`
	AssetTag                  string    `json:"asset_tag"`
	Name                      string    `json:"name"`
	CategoryID                string    `json:"category_id"`
	SerialNumber              *string   `json:"serial_number"`
	AcquisitionDate           *string   `json:"acquisition_date"`
	AcquisitionCost           *float64  `json:"acquisition_cost"`
	Condition                 *string   `json:"condition"`
	Location                  *string   `json:"location"`
	IsBookable                bool      `json:"is_bookable"`
	Status                    string    `json:"status"`
	CurrentHolderEmployeeID   *string   `json:"current_holder_employee_id"`
	CurrentHolderDepartmentID *string   `json:"current_holder_department_id"`
	QRCode                    *string                `json:"qr_code"`
	CustomFields              map[string]interface{} `json:"custom_fields"`
	CreatedAt                 time.Time              `json:"created_at"`
	UpdatedAt                 time.Time              `json:"updated_at"`
}

type AssetDocument struct {
	ID         string    `json:"id"`
	AssetID    string    `json:"asset_id"`
	URL        string    `json:"url"`
	Type       string    `json:"type"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type AssetStatusHistory struct {
	ID         string    `json:"id"`
	AssetID    string    `json:"asset_id"`
	FromStatus *string   `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	ChangedBy  *string   `json:"changed_by"`
	Reason     *string   `json:"reason"`
	ChangedAt  time.Time `json:"changed_at"`
}

type HistoryEvent struct {
	Timestamp time.Time       `json:"timestamp"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
}

type AssetDetail struct {
	Asset
	CategoryName              *string `json:"category_name"`
	CurrentHolderName         *string `json:"current_holder_name"`
	CurrentHolderDepartmentName *string `json:"current_holder_department_name"`
}

type AssetListItem struct {
	ID                        string  `json:"id"`
	AssetTag                  string  `json:"asset_tag"`
	Name                      string  `json:"name"`
	CategoryName              string  `json:"category_name"`
	SerialNumber              *string `json:"serial_number"`
	Status                    string  `json:"status"`
	Location                  *string `json:"location"`
	CurrentHolderDepartmentID *string `json:"current_holder_department_id"`
	IsBookable                bool    `json:"is_bookable"`
}
