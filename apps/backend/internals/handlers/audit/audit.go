package audit

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
	audit_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/audit-repo"
)

type AuditHandler struct {
	store *repository.StorageRegistry
}

func NewAuditHandler(store *repository.StorageRegistry) *AuditHandler {
	return &AuditHandler{store: store}
}

type CreateCycleRequest struct {
	Name              string  `json:"name" binding:"required"`
	ScopeDepartmentID *string `json:"scope_department_id"`
	ScopeLocation     *string `json:"scope_location"`
	StartDate         string  `json:"start_date" binding:"required"`
	EndDate           string  `json:"end_date" binding:"required"`
}

func (h *AuditHandler) CreateCycleHandler(c *gin.Context) {
	var req CreateCycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	cycle := models.AuditCycle{
		Name:      req.Name,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		CreatedBy: empIDStr,
	}

	created, err := h.store.Audit.CreateCycle(c.Request.Context(), cycle, req.ScopeDepartmentID, req.ScopeLocation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create audit cycle"})
		return
	}

	_ = h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &empIDStr,
		Action:          "audit.create",
		EntityType:      "audit_cycle",
		EntityID:        &created.ID,
		Metadata: map[string]interface{}{
			"name": req.Name,
		},
	})

	c.JSON(http.StatusCreated, gin.H{"audit_cycle": created})
}

type AssignAuditorsRequest struct {
	EmployeeIDs []string `json:"employee_ids" binding:"required"`
}

func (h *AuditHandler) AssignAuditorsHandler(c *gin.Context) {
	cycleID := c.Param("id")

	var req AssignAuditorsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.store.Audit.GetCycleByID(c.Request.Context(), cycleID)
	if err != nil {
		if errors.Is(err, audit_repo.ErrCycleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "audit cycle not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get audit cycle"})
		return
	}

	if err := h.store.Audit.AssignAuditors(c.Request.Context(), cycleID, req.EmployeeIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign auditors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "auditors assigned"})
}

func (h *AuditHandler) ListCyclesHandler(c *gin.Context) {
	cycles, err := h.store.Audit.ListCycles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list audit cycles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"audit_cycles": cycles})
}

func (h *AuditHandler) GetCycleHandler(c *gin.Context) {
	cycleID := c.Param("id")

	cycle, err := h.store.Audit.GetCycleByID(c.Request.Context(), cycleID)
	if err != nil {
		if errors.Is(err, audit_repo.ErrCycleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "audit cycle not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get audit cycle"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"audit_cycle": cycle})
}

func (h *AuditHandler) ListItemsHandler(c *gin.Context) {
	cycleID := c.Param("id")

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)
	roleStr, _ := c.Get("role")

	myItems := c.Query("my_items")

	var filterAuditorID *string
	if myItems == "true" {
		filterAuditorID = &empIDStr
	}

	items, err := h.store.Audit.ListItems(c.Request.Context(), cycleID, filterAuditorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list audit items"})
		return
	}

	if myItems != "true" && roleStr != "Admin" {
		cycle, err := h.store.Audit.GetCycleByID(c.Request.Context(), cycleID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "cycle not found"})
			return
		}
		isAssigned := false
		for _, a := range cycle.AssignedAuditors {
			if a.ID == empIDStr {
				isAssigned = true
				break
			}
		}
		if !isAssigned {
			c.JSON(http.StatusForbidden, gin.H{"error": "not assigned to this cycle"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

type PatchItemRequest struct {
	Result *string `json:"result"`
	Notes  *string `json:"notes"`
}

func (h *AuditHandler) PatchItemHandler(c *gin.Context) {
	itemID := c.Param("id")

	var req PatchItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Result != nil {
		valid := false
		for _, r := range []string{"Verified", "Missing", "Damaged"} {
			if *req.Result == r {
				valid = true
				break
			}
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "result must be Verified, Missing, or Damaged"})
			return
		}
	}

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	if err := h.store.Audit.VerifyItem(c.Request.Context(), itemID, empIDStr, req.Result, req.Notes); err != nil {
		if errors.Is(err, audit_repo.ErrItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "audit item not found"})
			return
		}
		if errors.Is(err, audit_repo.ErrNotAssignedAuditor) {
			c.JSON(http.StatusForbidden, gin.H{"error": "not an assigned auditor for this cycle"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "audit item updated"})
}

func (h *AuditHandler) CloseCycleHandler(c *gin.Context) {
	cycleID := c.Param("id")

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	if err := h.store.Audit.CloseCycle(c.Request.Context(), cycleID, empIDStr); err != nil {
		if errors.Is(err, audit_repo.ErrCycleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "audit cycle not found"})
			return
		}
		if errors.Is(err, audit_repo.ErrCycleAlreadyClosed) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "audit cycle is already closed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close audit cycle"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "audit cycle closed"})
}

func (h *AuditHandler) ListDiscrepancyReportsHandler(c *gin.Context) {
	cycleID := c.Query("cycle_id")
	resolvedStr := c.Query("resolved")

	var resolved *bool
	if resolvedStr == "true" {
		t := true
		resolved = &t
	} else if resolvedStr == "false" {
		f := false
		resolved = &f
	}

	var cycleIDPtr *string
	if cycleID != "" {
		cycleIDPtr = &cycleID
	}

	reports, err := h.store.Audit.ListDiscrepancyReports(c.Request.Context(), cycleIDPtr, resolved)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list discrepancy reports"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"discrepancy_reports": reports})
}

func (h *AuditHandler) ResolveDiscrepancyHandler(c *gin.Context) {
	id := c.Param("id")

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	if err := h.store.Audit.ResolveDiscrepancy(c.Request.Context(), id, empIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "discrepancy report resolved"})
}
