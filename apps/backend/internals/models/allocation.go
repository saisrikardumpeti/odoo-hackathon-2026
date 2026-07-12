package models

import "time"

type Allocation struct {
	ID                   string    `json:"id"`
	AssetID              string    `json:"asset_id"`
	EmployeeID           *string   `json:"employee_id"`
	DepartmentID         *string   `json:"department_id"`
	AllocatedBy          string    `json:"allocated_by"`
	AllocatedAt          time.Time `json:"allocated_at"`
	ExpectedReturnDate   *string   `json:"expected_return_date"`
	ReturnedAt           *time.Time `json:"returned_at"`
	ReturnConditionNotes *string   `json:"return_condition_notes"`
	Status               string    `json:"status"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type AllocationDetail struct {
	Allocation
	AssetTag          string  `json:"asset_tag"`
	AssetName         string  `json:"asset_name"`
	EmployeeName      *string `json:"employee_name"`
	DepartmentName    *string `json:"department_name"`
	AllocatedByName   string  `json:"allocated_by_name"`
}

type TransferRequest struct {
	ID              string     `json:"id"`
	AssetID         string     `json:"asset_id"`
	AllocationID    string     `json:"allocation_id"`
	FromEmployeeID  *string    `json:"from_employee_id"`
	ToEmployeeID    string     `json:"to_employee_id"`
	RequestedBy     string     `json:"requested_by"`
	Status          string     `json:"status"`
	ApprovedBy      *string    `json:"approved_by"`
	ApprovedAt      *time.Time `json:"approved_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type TransferRequestDetail struct {
	TransferRequest
	AssetTag        string  `json:"asset_tag"`
	AssetName       string  `json:"asset_name"`
	FromEmployeeName *string `json:"from_employee_name"`
	ToEmployeeName  string  `json:"to_employee_name"`
	RequestedByName string  `json:"requested_by_name"`
}
