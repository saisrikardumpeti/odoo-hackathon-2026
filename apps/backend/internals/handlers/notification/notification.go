package notification

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
)

type NotificationHandler struct {
	store *repository.StorageRegistry
}

func NewNotificationHandler(store *repository.StorageRegistry) *NotificationHandler {
	return &NotificationHandler{store: store}
}

func (h *NotificationHandler) ListHandler(c *gin.Context) {
	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	unreadOnly := c.Query("is_read") == "false"

	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	result, err := h.store.Notification.ListByEmployee(c.Request.Context(), empIDStr, unreadOnly, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": result.Notifications,
		"total":         result.Total,
		"page":          page,
		"page_size":     pageSize,
		"unread_count":  result.UnreadCount,
	})
}

func (h *NotificationHandler) MarkReadHandler(c *gin.Context) {
	id := c.Param("id")

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	if err := h.store.Notification.MarkRead(c.Request.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notification as read"})
		return
	}

	_ = h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &empIDStr,
		Action:          "notification.read",
		EntityType:      "notification",
		EntityID:        &id,
		Metadata:        map[string]interface{}{},
	})

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

func (h *NotificationHandler) MarkReadAllHandler(c *gin.Context) {
	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	if err := h.store.Notification.MarkReadAll(c.Request.Context(), empIDStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark all notifications as read"})
		return
	}

	_ = h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &empIDStr,
		Action:          "notification.read_all",
		EntityType:      "notification",
		EntityID:        nil,
		Metadata:        map[string]interface{}{},
	})

	c.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
}

func (h *NotificationHandler) UnreadCountHandler(c *gin.Context) {
	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	count, err := h.store.Notification.UnreadCount(c.Request.Context(), empIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}
