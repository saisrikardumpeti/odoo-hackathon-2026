package report

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
)

type ReportHandler struct {
	store *repository.StorageRegistry
}

func NewReportHandler(store *repository.StorageRegistry) *ReportHandler {
	return &ReportHandler{store: store}
}

func parseDateRange(c *gin.Context) (*time.Time, *time.Time) {
	var from, to *time.Time
	if f := c.Query("from"); f != "" {
		if t, err := time.Parse(time.RFC3339, f); err == nil {
			from = &t
		}
	}
	if t := c.Query("to"); t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			to = &parsed
		}
	}
	return from, to
}

func (h *ReportHandler) GetUtilizationHandler(c *gin.Context) {
	from, to := parseDateRange(c)
	idleDays := 30
	if d := c.Query("idle_days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			idleDays = parsed
		}
	}

	items, err := h.store.Report.GetUtilization(c.Request.Context(), from, to, idleDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch utilization report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"utilization": items})
}

func (h *ReportHandler) GetMaintenanceFrequencyHandler(c *gin.Context) {
	from, to := parseDateRange(c)

	assetItems, catItems, err := h.store.Report.GetMaintenanceFrequency(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch maintenance frequency report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"by_asset":    assetItems,
		"by_category": catItems,
	})
}

func (h *ReportHandler) GetRetirementWatchlistHandler(c *gin.Context) {
	threshold := 5.0
	if t := c.Query("age_years"); t != "" {
		if parsed, err := strconv.ParseFloat(t, 64); err == nil && parsed > 0 {
			threshold = parsed
		}
	}

	items, err := h.store.Report.GetRetirementWatchlist(c.Request.Context(), threshold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch retirement watchlist"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"retirement_watchlist": items})
}

func (h *ReportHandler) GetAllocationSummaryHandler(c *gin.Context) {
	items, err := h.store.Report.GetAllocationSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch allocation summary"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"allocation_summary": items})
}

func (h *ReportHandler) GetBookingHeatmapHandler(c *gin.Context) {
	from, to := parseDateRange(c)

	items, err := h.store.Report.GetBookingHeatmap(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch booking heatmap"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"heatmap": items})
}

func (h *ReportHandler) ExportHandler(c *gin.Context) {
	reportType := c.Query("type")
	from, to := parseDateRange(c)

	var records [][]string
	var headers []string

	switch reportType {
	case "utilization":
		idleDays := 30
		if d := c.Query("idle_days"); d != "" {
			if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
				idleDays = parsed
			}
		}
		items, err := h.store.Report.GetUtilization(c.Request.Context(), from, to, idleDays)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch utilization report"})
			return
		}
		headers = []string{"Asset ID", "Asset Tag", "Asset Name", "Category", "Allocations", "Bookings", "Total Activity", "Last Activity", "Days Idle"}
		for _, item := range items {
			lastAct := ""
			if item.LastActivity != nil {
				lastAct = *item.LastActivity
			}
			daysIdle := ""
			if item.DaysIdle != nil {
				daysIdle = strconv.Itoa(*item.DaysIdle)
			}
			records = append(records, []string{
				item.AssetID, item.AssetTag, item.AssetName, item.CategoryName,
				strconv.Itoa(item.AllocationCount), strconv.Itoa(item.BookingCount),
				strconv.Itoa(item.TotalActivity), lastAct, daysIdle,
			})
		}

	case "maintenance-frequency":
		assetItems, catItems, err := h.store.Report.GetMaintenanceFrequency(c.Request.Context(), from, to)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch maintenance frequency report"})
			return
		}
		_ = catItems
		headers = []string{"Asset ID", "Asset Tag", "Asset Name", "Category", "Maintenance Count"}
		for _, item := range assetItems {
			records = append(records, []string{
				item.AssetID, item.AssetTag, item.AssetName, item.CategoryName,
				strconv.Itoa(item.Count),
			})
		}

	case "retirement-watchlist":
		threshold := 5.0
		if t := c.Query("age_years"); t != "" {
			if parsed, err := strconv.ParseFloat(t, 64); err == nil && parsed > 0 {
				threshold = parsed
			}
		}
		items, err := h.store.Report.GetRetirementWatchlist(c.Request.Context(), threshold)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch retirement watchlist"})
			return
		}
		headers = []string{"Asset ID", "Asset Tag", "Asset Name", "Category", "Acquisition Date", "Age (Years)", "Status"}
		for _, item := range items {
			acqDate := ""
			if item.AcquisitionDate != nil {
				acqDate = *item.AcquisitionDate
			}
			age := ""
			if item.AgeYears != nil {
				age = fmt.Sprintf("%.1f", *item.AgeYears)
			}
			records = append(records, []string{
				item.AssetID, item.AssetTag, item.AssetName, item.CategoryName,
				acqDate, age, item.Status,
			})
		}

	case "allocation-summary":
		items, err := h.store.Report.GetAllocationSummary(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch allocation summary"})
			return
		}
		headers = []string{"Department", "Active Allocations"}
		for _, item := range items {
			records = append(records, []string{
				item.DepartmentName, strconv.Itoa(item.AssetCount),
			})
		}

	case "booking-heatmap":
		items, err := h.store.Report.GetBookingHeatmap(c.Request.Context(), from, to)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch booking heatmap"})
			return
		}
		dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		headers = []string{"Day", "Hour", "Bookings"}
		for _, item := range items {
			dayName := "Unknown"
			if item.DayOfWeek >= 0 && item.DayOfWeek <= 6 {
				dayName = dayNames[item.DayOfWeek]
			}
			records = append(records, []string{
				dayName, strconv.Itoa(item.Hour), strconv.Itoa(item.Count),
			})
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report type: " + reportType + ". Valid types: utilization, maintenance-frequency, retirement-watchlist, allocation-summary, booking-heatmap"})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-report.csv", reportType))

	writer := csv.NewWriter(c.Writer)
	if err := writer.Write(headers); err != nil {
		return
	}
	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return
		}
	}
	writer.Flush()
}
