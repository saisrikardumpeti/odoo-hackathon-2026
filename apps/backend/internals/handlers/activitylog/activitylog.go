package activitylog

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	activity_log_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/activity-log-repo"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
)

type ActivityLogHandler struct {
	store *repository.StorageRegistry
}

func NewActivityLogHandler(store *repository.StorageRegistry) *ActivityLogHandler {
	return &ActivityLogHandler{store: store}
}

func (h *ActivityLogHandler) ListHandler(c *gin.Context) {
	filters := activity_log_repo.ActivityLogFilters{
		Page:     1,
		PageSize: 20,
	}

	if actor := c.Query("actor"); actor != "" {
		filters.ActorEmployeeID = &actor
	}
	if action := c.Query("action"); action != "" {
		filters.Action = &action
	}
	if entityType := c.Query("entity_type"); entityType != "" {
		filters.EntityType = &entityType
	}
	if entityID := c.Query("entity_id"); entityID != "" {
		filters.EntityID = &entityID
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			filters.DateFrom = &t
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse(time.RFC3339, dateTo); err == nil {
			filters.DateTo = &t
		}
	}
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			filters.Page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			filters.PageSize = v
		}
	}

	result, err := h.store.ActivityLog.List(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list activity logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      result.Logs,
		"total":     result.Total,
		"page":      filters.Page,
		"page_size": filters.PageSize,
	})
}
