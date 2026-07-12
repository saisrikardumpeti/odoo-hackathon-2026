package models

type UtilizationReportItem struct {
	AssetID         string  `json:"asset_id"`
	AssetTag        string  `json:"asset_tag"`
	AssetName       string  `json:"asset_name"`
	CategoryName    string  `json:"category_name"`
	AllocationCount int     `json:"allocation_count"`
	BookingCount    int     `json:"booking_count"`
	TotalActivity   int     `json:"total_activity"`
	LastActivity    *string `json:"last_activity"`
	DaysIdle        *int    `json:"days_idle"`
}

type MaintenanceFrequencyItem struct {
	AssetID      string `json:"asset_id"`
	AssetTag     string `json:"asset_tag"`
	AssetName    string `json:"asset_name"`
	CategoryName string `json:"category_name"`
	Count        int    `json:"count"`
}

type MaintenanceCategoryItem struct {
	CategoryName string `json:"category_name"`
	Count        int    `json:"count"`
}

type RetirementWatchlistItem struct {
	AssetID         string   `json:"asset_id"`
	AssetTag        string   `json:"asset_tag"`
	AssetName       string   `json:"asset_name"`
	CategoryName    string   `json:"category_name"`
	AcquisitionDate *string  `json:"acquisition_date"`
	AgeYears        *float64 `json:"age_years"`
	Status          string   `json:"status"`
}

type AllocationSummaryItem struct {
	DepartmentName string `json:"department_name"`
	DepartmentID   string `json:"department_id"`
	AssetCount     int    `json:"asset_count"`
}

type BookingHeatmapItem struct {
	DayOfWeek int `json:"day_of_week"`
	Hour      int `json:"hour"`
	Count     int `json:"count"`
}
