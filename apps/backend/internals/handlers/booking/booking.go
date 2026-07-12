package booking

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
	booking_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/booking-repo"
)

type BookingHandler struct {
	store *repository.StorageRegistry
}

func NewBookingHandler(store *repository.StorageRegistry) *BookingHandler {
	return &BookingHandler{store: store}
}

func (h *BookingHandler) ListByResourceHandler(c *gin.Context) {
	assetID := c.Param("assetId")

	var from, to *time.Time
	if fromStr := c.Query("from"); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' query param, use RFC3339"})
			return
		}
		from = &t
	}
	if toStr := c.Query("to"); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' query param, use RFC3339"})
			return
		}
		to = &t
	}

	bookings, err := h.store.Booking.ListByResource(c.Request.Context(), assetID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

type CreateBookingRequest struct {
	ResourceAssetID string `json:"resource_asset_id" binding:"required"`
	StartTime       string `json:"start_time" binding:"required"`
	EndTime         string `json:"end_time" binding:"required"`
	Purpose         string `json:"purpose"`
}

func (h *BookingHandler) CreateHandler(c *gin.Context) {
	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time, use RFC3339 format"})
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time, use RFC3339 format"})
		return
	}

	if !endTime.After(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
		return
	}

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	var purpose *string
	if req.Purpose != "" {
		purpose = &req.Purpose
	}

	booking := models.Booking{
		ResourceAssetID:    req.ResourceAssetID,
		BookedByEmployeeID: empIDStr,
		StartTime:          startTime,
		EndTime:            endTime,
		Purpose:            purpose,
	}

	created, err := h.store.Booking.Create(c.Request.Context(), booking)
	if err != nil {
		if errors.Is(err, booking_repo.ErrBookingOverlap) {
			conflicts, err2 := h.store.Booking.FindConflicting(c.Request.Context(), req.ResourceAssetID, startTime, endTime, nil)
			if err2 == nil && len(conflicts) > 0 {
				c.JSON(http.StatusConflict, gin.H{
					"error":               "BookingOverlap",
					"message":             "booking overlaps with an existing booking",
					"conflicting_bookings": conflicts,
				})
				return
			}
			c.JSON(http.StatusConflict, gin.H{
				"error":   "BookingOverlap",
				"message": "booking overlaps with an existing booking",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create booking"})
		return
	}

	_ = h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &empIDStr,
		Action:          "booking.create",
		EntityType:      "booking",
		EntityID:        &created.ID,
		Metadata: map[string]interface{}{
			"resource_asset_id": req.ResourceAssetID,
			"start_time":        req.StartTime,
			"end_time":          req.EndTime,
		},
	})

	c.JSON(http.StatusCreated, gin.H{"booking": created})
}

func (h *BookingHandler) CancelHandler(c *gin.Context) {
	id := c.Param("id")

	booking, err := h.store.Booking.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, booking_repo.ErrBookingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get booking"})
		return
	}

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)
	role, _ := c.Get("role")
	roleStr, _ := role.(string)

	if booking.BookedByEmployeeID != empIDStr && roleStr != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only cancel your own bookings"})
		return
	}

	if booking.Status == "Cancelled" || booking.Status == "Completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking is already " + booking.Status})
		return
	}

	if err := h.store.Booking.Cancel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel booking"})
		return
	}

	_ = h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &empIDStr,
		Action:          "booking.cancel",
		EntityType:      "booking",
		EntityID:        &id,
		Metadata: map[string]interface{}{
			"resource_asset_id": booking.ResourceAssetID,
		},
	})

	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled"})
}

type RescheduleBookingRequest struct {
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

func (h *BookingHandler) RescheduleHandler(c *gin.Context) {
	id := c.Param("id")

	booking, err := h.store.Booking.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, booking_repo.ErrBookingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get booking"})
		return
	}

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)
	role, _ := c.Get("role")
	roleStr, _ := role.(string)

	if booking.BookedByEmployeeID != empIDStr && roleStr != "Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only reschedule your own bookings"})
		return
	}

	if booking.Status == "Cancelled" || booking.Status == "Completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot reschedule a " + booking.Status + " booking"})
		return
	}

	var req RescheduleBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newStart, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time, use RFC3339 format"})
		return
	}
	newEnd, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time, use RFC3339 format"})
		return
	}

	if !newEnd.After(newStart) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_time must be after start_time"})
		return
	}

	excludeID := &id
	conflicts, err := h.store.Booking.FindConflicting(c.Request.Context(), booking.ResourceAssetID, newStart, newEnd, excludeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check for conflicts"})
		return
	}
	if len(conflicts) > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":               "BookingOverlap",
			"message":             "new time range overlaps with an existing booking",
			"conflicting_bookings": conflicts,
		})
		return
	}

	if err := h.store.Booking.Reschedule(c.Request.Context(), id, newStart, newEnd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reschedule booking"})
		return
	}

	_ = h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &empIDStr,
		Action:          "booking.reschedule",
		EntityType:      "booking",
		EntityID:        &id,
		Metadata: map[string]interface{}{
			"old_start": booking.StartTime,
			"old_end":   booking.EndTime,
			"new_start": req.StartTime,
			"new_end":   req.EndTime,
		},
	})

	c.JSON(http.StatusOK, gin.H{"message": "booking rescheduled"})
}

func (h *BookingHandler) ListMyBookingsHandler(c *gin.Context) {
	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	bookings, err := h.store.Booking.ListByBooker(c.Request.Context(), empIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}
