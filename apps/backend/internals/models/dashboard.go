package models

type KPIsResponse struct {
	AssetsAvailable  int `json:"assets_available"`
	AssetsAllocated  int `json:"assets_allocated"`
	MaintenanceToday int `json:"maintenance_today"`
	ActiveBookings   int `json:"active_bookings"`
	PendingTransfers int `json:"pending_transfers"`
	UpcomingReturns  int `json:"upcoming_returns"`
}

type OverdueItem struct {
	Type              string  `json:"type"`
	ID                string  `json:"id"`
	AssetID           string  `json:"asset_id"`
	AssetTag          string  `json:"asset_tag"`
	AssetName         string  `json:"asset_name"`
	EmployeeID        *string `json:"employee_id"`
	EmployeeName      *string `json:"employee_name"`
	ExpectedReturnDate *string `json:"expected_return_date"`
	DaysOverdue       int     `json:"days_overdue"`
}

type RecentActivityItem struct {
	ID        string `json:"id"`
	Action    string `json:"action"`
	EntityType string `json:"entity_type"`
	ActorName *string `json:"actor_name"`
	CreatedAt string `json:"created_at"`
}

type UpcomingItem struct {
	Type             string  `json:"type"`
	ID               string  `json:"id"`
	AssetID          string  `json:"asset_id"`
	AssetTag         string  `json:"asset_tag"`
	AssetName        string  `json:"asset_name"`
	EmployeeID       *string `json:"employee_id"`
	EmployeeName     *string `json:"employee_name"`
	ExpectedDate     string  `json:"expected_date"`
	DaysUntilDue     int     `json:"days_until_due"`
}
