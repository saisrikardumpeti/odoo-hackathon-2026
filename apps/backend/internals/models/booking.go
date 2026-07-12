package models

import "time"

type Booking struct {
	ID                 string    `json:"id"`
	ResourceAssetID    string    `json:"resource_asset_id"`
	BookedByEmployeeID string    `json:"booked_by_employee_id"`
	StartTime          time.Time `json:"start_time"`
	EndTime            time.Time `json:"end_time"`
	Purpose            *string   `json:"purpose"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type BookingDetail struct {
	Booking
	AssetName    string  `json:"asset_name"`
	AssetTag     string  `json:"asset_tag"`
	BookedByName *string `json:"booked_by_name"`
}
