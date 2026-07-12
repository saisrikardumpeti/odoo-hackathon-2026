package dashboard

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
)

type DashboardHandler struct {
	store *repository.StorageRegistry
}

func NewDashboardHandler(store *repository.StorageRegistry) *DashboardHandler {
	return &DashboardHandler{store: store}
}

func (h *DashboardHandler) GetKPIsHandler(c *gin.Context) {
	role := c.GetString("role")
	empID := c.GetString("employee_id")

	var empIDPtr *string
	if role == "Employee" {
		empIDPtr = &empID
	}

	kpis, err := h.store.Dashboard.GetKPIs(c.Request.Context(), empIDPtr, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch KPIs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"kpis": kpis})
}

func (h *DashboardHandler) GetOverdueHandler(c *gin.Context) {
	role := c.GetString("role")
	empID := c.GetString("employee_id")

	var empIDPtr *string
	if role == "Employee" {
		empIDPtr = &empID
	}

	items, err := h.store.Dashboard.GetOverdue(c.Request.Context(), empIDPtr, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch overdue items"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"overdue": items})
}

func (h *DashboardHandler) GetUpcomingHandler(c *gin.Context) {
	role := c.GetString("role")
	empID := c.GetString("employee_id")

	var empIDPtr *string
	if role == "Employee" {
		empIDPtr = &empID
	}

	windowDays := 7
	if w := c.Query("window_days"); w != "" {
		if parsed, err := strconv.Atoi(w); err == nil && parsed > 0 {
			windowDays = parsed
		}
	}

	items, err := h.store.Dashboard.GetUpcoming(c.Request.Context(), empIDPtr, role, windowDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch upcoming items"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"upcoming": items})
}

func (h *DashboardHandler) GetRecentActivityHandler(c *gin.Context) {
	role := c.GetString("role")
	empID := c.GetString("employee_id")

	var empIDPtr *string
	if role == "Employee" {
		empIDPtr = &empID
	}

	items, err := h.store.Dashboard.GetRecentActivity(c.Request.Context(), empIDPtr, role, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recent activity"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"activity": items})
}
