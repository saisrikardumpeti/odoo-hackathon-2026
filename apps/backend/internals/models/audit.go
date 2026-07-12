package models

import "time"

type AuditCycle struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	ScopeDepartmentID  *string    `json:"scope_department_id"`
	ScopeLocation      *string    `json:"scope_location"`
	StartDate          string     `json:"start_date"`
	EndDate            string     `json:"end_date"`
	Status             string     `json:"status"`
	CreatedBy          string     `json:"created_by"`
	ClosedAt           *time.Time `json:"closed_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type AuditCycleDetail struct {
	AuditCycle
	ScopeDepartmentName *string `json:"scope_department_name"`
	CreatedByName       *string `json:"created_by_name"`
	AssignedAuditors    []Employee `json:"assigned_auditors,omitempty"`
	ItemCount           *int    `json:"item_count"`
	VerifiedCount       *int    `json:"verified_count"`
	MissingCount        *int    `json:"missing_count"`
	DamagedCount        *int    `json:"damaged_count"`
}

type AuditItem struct {
	ID           string     `json:"id"`
	AuditCycleID string     `json:"audit_cycle_id"`
	AssetID      string     `json:"asset_id"`
	AuditorID    *string    `json:"auditor_id"`
	Result       *string    `json:"result"`
	Notes        *string    `json:"notes"`
	VerifiedAt   *time.Time `json:"verified_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type AuditItemDetail struct {
	AuditItem
	AssetTag     string  `json:"asset_tag"`
	AssetName    string  `json:"asset_name"`
	AssetStatus  string  `json:"asset_status"`
	AssetLocation *string `json:"asset_location"`
}

type DiscrepancyReport struct {
	ID            string     `json:"id"`
	AuditCycleID  string     `json:"audit_cycle_id"`
	AssetID       string     `json:"asset_id"`
	AuditItemID   string     `json:"audit_item_id"`
	IssueType     string     `json:"issue_type"`
	Resolved      bool       `json:"resolved"`
	ResolvedBy    *string    `json:"resolved_by"`
	ResolvedAt    *time.Time `json:"resolved_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type DiscrepancyReportDetail struct {
	DiscrepancyReport
	CycleName  string  `json:"cycle_name"`
	AssetTag   string  `json:"asset_tag"`
	AssetName  string  `json:"asset_name"`
	ResolvedByName *string `json:"resolved_by_name"`
}
