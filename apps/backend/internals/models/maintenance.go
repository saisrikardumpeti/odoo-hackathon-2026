package models

import "time"

type MaintenanceRequest struct {
	ID                string    `json:"id"`
	AssetID           string    `json:"asset_id"`
	RaisedByEmployeeID string   `json:"raised_by_employee_id"`
	IssueDescription  string    `json:"issue_description"`
	Priority          string    `json:"priority"`
	PhotoURL          *string   `json:"photo_url"`
	Status            string    `json:"status"`
	ApprovedBy        *string   `json:"approved_by"`
	ApprovedAt        *time.Time `json:"approved_at"`
	TechnicianName    *string   `json:"technician_name"`
	ResolvedAt        *time.Time `json:"resolved_at"`
	ResolutionNotes   *string   `json:"resolution_notes"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type MaintenanceDetail struct {
	MaintenanceRequest
	AssetTag         string  `json:"asset_tag"`
	AssetName        string  `json:"asset_name"`
	RaisedByName     *string `json:"raised_by_name"`
	ApprovedByName   *string `json:"approved_by_name"`
}

var ValidMaintenanceTransitions = map[string][]string{
	"Pending":            {"Approved", "Rejected"},
	"Approved":           {"TechnicianAssigned", "Rejected"},
	"Rejected":           {},
	"TechnicianAssigned": {"InProgress"},
	"InProgress":         {"Resolved"},
	"Resolved":           {},
}

func IsValidMaintenanceTransition(from, to string) bool {
	allowed, ok := ValidMaintenanceTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
