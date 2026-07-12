package maintenance

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
	maintenance_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/maintenance-repo"
)

type MaintenanceHandler struct {
	store *repository.StorageRegistry
}

func NewMaintenanceHandler(store *repository.StorageRegistry) *MaintenanceHandler {
	return &MaintenanceHandler{store: store}
}

type CreateMaintenanceRequest struct {
	AssetID          string  `json:"asset_id" binding:"required"`
	IssueDescription string  `json:"issue_description" binding:"required"`
	Priority         string  `json:"priority" binding:"required"`
	PhotoURL         *string `json:"photo_url"`
}

func (h *MaintenanceHandler) CreateHandler(c *gin.Context) {
	var req CreateMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	m := models.MaintenanceRequest{
		AssetID:            req.AssetID,
		RaisedByEmployeeID: empIDStr,
		IssueDescription:   req.IssueDescription,
		Priority:           req.Priority,
		PhotoURL:           req.PhotoURL,
	}

	created, err := h.store.Maintenance.Create(c.Request.Context(), m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create maintenance request"})
		return
	}

	_ = h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &empIDStr,
		Action:          "maintenance.create",
		EntityType:      "maintenance_request",
		EntityID:        &created.ID,
		Metadata: map[string]interface{}{
			"asset_id":  req.AssetID,
			"priority":  req.Priority,
		},
	})

	c.JSON(http.StatusCreated, gin.H{"maintenance": created})
}

func (h *MaintenanceHandler) ListHandler(c *gin.Context) {
	filters := maintenance_repo.MaintenanceListFilters{
		AssetID:  c.Query("asset_id"),
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
	}

	result, err := h.store.Maintenance.List(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list maintenance requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"maintenance_requests": result.Requests,
		"total_count":          result.TotalCount,
	})
}

type RejectRequest struct {
	Reason *string `json:"reason"`
}

type AssignTechnicianRequest struct {
	TechnicianName string `json:"technician_name" binding:"required"`
}

type ResolveRequest struct {
	ResolutionNotes string `json:"resolution_notes" binding:"required"`
}

func transitionWithValidation(h *MaintenanceHandler, c *gin.Context, targetStatus string, updateFields map[string]interface{}, assetStatusUpdate func(*models.MaintenanceDetail) (string, string, error)) {
	id := c.Param("id")

	reqDetail, err := h.store.Maintenance.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, maintenance_repo.ErrMaintenanceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "maintenance request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get maintenance request"})
		return
	}

	if !models.IsValidMaintenanceTransition(reqDetail.Status, targetStatus) {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "invalid_transition",
			"message": "cannot transition from " + reqDetail.Status + " to " + targetStatus,
		})
		return
	}

	if assetStatusUpdate != nil {
		newAssetStatus, reason, err := assetStatusUpdate(reqDetail)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		employeeID, _ := c.Get("employee_id")
		empIDStr, _ := employeeID.(string)

		if err := h.store.Maintenance.UpdateAssetStatus(c.Request.Context(), reqDetail.AssetID, newAssetStatus, &empIDStr, reason); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update asset status"})
			return
		}
	}

	if err := h.store.Maintenance.UpdateStatus(c.Request.Context(), id, targetStatus, updateFields); err != nil {
		if errors.Is(err, maintenance_repo.ErrMaintenanceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "maintenance request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update maintenance request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "maintenance request " + targetStatus + " successfully"})
}

func (h *MaintenanceHandler) ApproveHandler(c *gin.Context) {
	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)
	now := time.Now()
	id := c.Param("id")

	reqDetail, err := h.store.Maintenance.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, maintenance_repo.ErrMaintenanceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "maintenance request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get maintenance request"})
		return
	}

	if !models.IsValidMaintenanceTransition(reqDetail.Status, "Approved") {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "invalid_transition",
			"message": "cannot transition from " + reqDetail.Status + " to Approved",
		})
		return
	}

	currentStatus, err := h.store.Maintenance.GetCurrentAssetStatus(c.Request.Context(), reqDetail.AssetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check asset status"})
		return
	}
	if currentStatus == "Disposed" || currentStatus == "Retired" {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "invalid_asset_status",
			"message": "asset is " + currentStatus + " and cannot be set to UnderMaintenance",
		})
		return
	}

	if err := h.store.Maintenance.UpdateAssetStatus(c.Request.Context(), reqDetail.AssetID, "UnderMaintenance", &empIDStr, "Maintenance approved"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update asset status"})
		return
	}

	fields := map[string]interface{}{
		"approved_by": empIDStr,
		"approved_at": now,
	}
	if err := h.store.Maintenance.UpdateStatus(c.Request.Context(), id, "Approved", fields); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve maintenance request"})
		return
	}

	_ = h.store.Notification.Create(c.Request.Context(), models.Notification{
		EmployeeID: reqDetail.RaisedByEmployeeID,
		Type:       "MaintenanceApproved",
		Message:    "Your maintenance request has been approved",
		RelatedEntityType: strPtr("maintenance_request"),
		RelatedEntityID:   &id,
	})

	_ = h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &empIDStr,
		Action:          "maintenance.approve",
		EntityType:      "maintenance_request",
		EntityID:        &id,
		Metadata: map[string]interface{}{
			"asset_id": reqDetail.AssetID,
		},
	})

	c.JSON(http.StatusOK, gin.H{"message": "maintenance request approved successfully"})
}

func (h *MaintenanceHandler) RejectHandler(c *gin.Context) {
	var req RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)
	id := c.Param("id")

	reqDetail, err := h.store.Maintenance.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, maintenance_repo.ErrMaintenanceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "maintenance request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get maintenance request"})
		return
	}

	if !models.IsValidMaintenanceTransition(reqDetail.Status, "Rejected") {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "invalid_transition",
			"message": "cannot transition from " + reqDetail.Status + " to Rejected",
		})
		return
	}

	fields := map[string]interface{}{
		"approved_by": empIDStr,
		"approved_at": time.Now(),
	}
	if req.Reason != nil {
		fields["resolution_notes"] = *req.Reason
	}

	if err := h.store.Maintenance.UpdateStatus(c.Request.Context(), id, "Rejected", fields); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject maintenance request"})
		return
	}

	_ = h.store.Notification.Create(c.Request.Context(), models.Notification{
		EmployeeID: reqDetail.RaisedByEmployeeID,
		Type:       "MaintenanceRejected",
		Message:    "Your maintenance request has been rejected",
		RelatedEntityType: strPtr("maintenance_request"),
		RelatedEntityID:   &id,
	})

	c.JSON(http.StatusOK, gin.H{"message": "maintenance request rejected successfully"})
}

func (h *MaintenanceHandler) AssignTechnicianHandler(c *gin.Context) {
	var req AssignTechnicianRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fields := map[string]interface{}{
		"technician_name": req.TechnicianName,
	}

	transitionWithValidation(h, c, "TechnicianAssigned", fields, nil)
}

func (h *MaintenanceHandler) StartHandler(c *gin.Context) {
	transitionWithValidation(h, c, "InProgress", nil, nil)
}

func (h *MaintenanceHandler) ResolveHandler(c *gin.Context) {
	var req ResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	fields := map[string]interface{}{
		"resolution_notes": req.ResolutionNotes,
		"resolved_at":      now,
	}

	transitionWithValidation(h, c, "Resolved", fields, func(reqDetail *models.MaintenanceDetail) (string, string, error) {
		return "Available", "Maintenance resolved", nil
	})
}

func strPtr(s string) *string {
	return &s
}
