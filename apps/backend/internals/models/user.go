package models

import "time"

type User struct {
	ID               string     `json:"id"`
	EmployeeID       string     `json:"employee_id"`
	PasswordHash     string     `json:"-"`
	RefreshTokenHash *string    `json:"-"`
	LastLoginAt      *time.Time `json:"last_login_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
