package models

import "time"

type Department struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	ParentDepartmentID *string   `json:"parent_department_id"`
	HeadEmployeeID     *string   `json:"head_employee_id"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
