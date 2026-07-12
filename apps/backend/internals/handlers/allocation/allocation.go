package allocation

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
	allocation_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/allocation-repo"
	transfer_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/transfer-repo"
)

type AllocationHandler struct {
	store *repository.StorageRegistry
	pool  *pgxpool.Pool
}

func NewAllocationHandler(store *repository.StorageRegistry, pool *pgxpool.Pool) *AllocationHandler {
	return &AllocationHandler{store: store, pool: pool}
}

type CreateAllocationRequest struct {
	AssetID            string  `json:"asset_id" binding:"required"`
	EmployeeID         *string `json:"employee_id"`
	DepartmentID       *string `json:"department_id"`
	ExpectedReturnDate *string `json:"expected_return_date"`
}

func (h *AllocationHandler) CreateHandler(c *gin.Context) {
	var req CreateAllocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	alloc := models.Allocation{
		AssetID:            req.AssetID,
		EmployeeID:         req.EmployeeID,
		DepartmentID:       req.DepartmentID,
		AllocatedBy:        empIDStr,
		ExpectedReturnDate: req.ExpectedReturnDate,
	}

	created, err := h.store.Allocation.Create(c.Request.Context(), alloc)
	if err != nil {
		if errors.Is(err, allocation_repo.ErrAllocationAlreadyActive) {
			active, err2 := h.store.Allocation.GetActiveByAssetID(c.Request.Context(), req.AssetID)
			if err2 == nil && active != nil {
				holderName := ""
				if active.EmployeeName != nil {
					holderName = *active.EmployeeName
				} else if active.DepartmentName != nil {
					holderName = *active.DepartmentName
				}
				c.JSON(http.StatusConflict, gin.H{
					"error":          "AlreadyAllocated",
					"message":        "currently held by " + holderName,
					"current_holder": active,
				})
				return
			}
			c.JSON(http.StatusConflict, gin.H{"error": "AlreadyAllocated", "message": "asset already has an active allocation"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create allocation"})
		return
	}

	if err := h.store.Allocation.UpdateAssetStatusTx(c.Request.Context(), created.ID, req.AssetID, "Allocated", &empIDStr, "Asset allocated"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "allocation created but failed to update asset status"})
		return
	}

	if req.EmployeeID != nil {
		if err := h.store.Allocation.UpdateAssetHolder(c.Request.Context(), req.AssetID, req.EmployeeID, nil); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "allocation created but failed to update asset holder"})
			return
		}
		_ = h.store.Notification.Create(c.Request.Context(), models.Notification{
			EmployeeID: *req.EmployeeID,
			Type:       "AssetAssigned",
			Message:    "Asset " + created.ID + " has been assigned to you",
			RelatedEntityType: strPtr("allocation"),
			RelatedEntityID:   &created.ID,
		})
	} else if req.DepartmentID != nil {
		if err := h.store.Allocation.UpdateAssetHolder(c.Request.Context(), req.AssetID, nil, req.DepartmentID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "allocation created but failed to update asset holder"})
			return
		}
	}

	_ = h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &empIDStr,
		Action:          "asset.allocate",
		EntityType:      "allocation",
		EntityID:        &created.ID,
		Metadata: map[string]interface{}{
			"asset_id":      req.AssetID,
			"employee_id":   req.EmployeeID,
			"department_id": req.DepartmentID,
		},
	})

	c.JSON(http.StatusCreated, gin.H{"allocation": created})
}

type ReturnAllocationRequest struct {
	ReturnConditionNotes *string `json:"return_condition_notes"`
}

func (h *AllocationHandler) ReturnHandler(c *gin.Context) {
	id := c.Param("id")

	var req ReturnAllocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	alloc, err := h.store.Allocation.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, allocation_repo.ErrAllocationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "allocation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get allocation"})
		return
	}

	now := time.Now()
	if err := h.store.Allocation.UpdateStatus(c.Request.Context(), id, "Returned", &now, req.ReturnConditionNotes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to return allocation"})
		return
	}

	if err := h.store.Allocation.UpdateAssetStatusTx(c.Request.Context(), id, alloc.AssetID, "Available", &empIDStr, "Asset returned"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "allocation returned but failed to update asset status"})
		return
	}

	if err := h.store.Allocation.UpdateAssetHolder(c.Request.Context(), alloc.AssetID, nil, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "allocation returned but failed to clear asset holder"})
		return
	}

	_ = h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &empIDStr,
		Action:          "asset.return",
		EntityType:      "allocation",
		EntityID:        &id,
		Metadata: map[string]interface{}{
			"asset_id": alloc.AssetID,
			"return_condition_notes": req.ReturnConditionNotes,
		},
	})

	c.JSON(http.StatusOK, gin.H{"message": "asset returned successfully"})
}

func (h *AllocationHandler) ListOverdueHandler(c *gin.Context) {
	allocations, err := h.store.Allocation.ListOverdue(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list overdue allocations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"allocations": allocations})
}

type CreateTransferRequest struct {
	AllocationID string `json:"allocation_id" binding:"required"`
	ToEmployeeID string `json:"to_employee_id" binding:"required"`
}

func (h *AllocationHandler) CreateTransferHandler(c *gin.Context) {
	var req CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	alloc, err := h.store.Allocation.GetByID(c.Request.Context(), req.AllocationID)
	if err != nil {
		if errors.Is(err, allocation_repo.ErrAllocationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "allocation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get allocation"})
		return
	}

	tr := models.TransferRequest{
		AssetID:        alloc.AssetID,
		AllocationID:   req.AllocationID,
		FromEmployeeID: alloc.EmployeeID,
		ToEmployeeID:   req.ToEmployeeID,
		RequestedBy:    empIDStr,
	}

	created, err := h.store.Transfer.Create(c.Request.Context(), tr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transfer request"})
		return
	}

	_ = h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &empIDStr,
		Action:          "transfer.request",
		EntityType:      "transfer_request",
		EntityID:        &created.ID,
		Metadata: map[string]interface{}{
			"allocation_id":   req.AllocationID,
			"to_employee_id":  req.ToEmployeeID,
			"from_employee_id": alloc.EmployeeID,
		},
	})

	c.JSON(http.StatusCreated, gin.H{"transfer": created})
}

func (h *AllocationHandler) ApproveTransferHandler(c *gin.Context) {
	id := c.Param("id")

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	tr, err := h.store.Transfer.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, transfer_repo.ErrTransferNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "transfer request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transfer request"})
		return
	}

	if tr.Status != "Requested" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transfer request is not in Requested status"})
		return
	}

	tx, err := h.pool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback(c.Request.Context())

	now := time.Now()
	if _, err := tx.Exec(c.Request.Context(),
		`UPDATE allocations SET status = 'Returned', returned_at = $1, updated_at = now() WHERE id = $2`,
		now, tr.AllocationID,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close old allocation"})
		return
	}

	var newAllocID string
	if err := tx.QueryRow(c.Request.Context(),
		`INSERT INTO allocations (asset_id, employee_id, allocated_by, expected_return_date)
		 VALUES ($1, $2, $3, NULL)
		 RETURNING id`,
		tr.AssetID, tr.ToEmployeeID, empIDStr,
	).Scan(&newAllocID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create new allocation"})
		return
	}

	if _, err := tx.Exec(c.Request.Context(),
		`UPDATE transfer_requests SET status = 'Approved', approved_by = $1, approved_at = $2, updated_at = now() WHERE id = $3`,
		empIDStr, now, id,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve transfer"})
		return
	}

	if _, err := tx.Exec(c.Request.Context(),
		`UPDATE assets SET status = 'Allocated', current_holder_employee_id = $1, updated_at = now() WHERE id = $2`,
		tr.ToEmployeeID, tr.AssetID,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update asset"})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	_ = h.store.Notification.Create(c.Request.Context(), models.Notification{
		EmployeeID: tr.ToEmployeeID,
		Type:       "TransferApproved",
		Message:    "Asset " + tr.AssetTag + " has been transferred to you",
		RelatedEntityType: strPtr("transfer"),
		RelatedEntityID:   &id,
	})

	_ = h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &empIDStr,
		Action:          "transfer.approve",
		EntityType:      "transfer_request",
		EntityID:        &id,
		Metadata: map[string]interface{}{
			"new_allocation_id": newAllocID,
			"to_employee_id":    tr.ToEmployeeID,
		},
	})

	c.JSON(http.StatusOK, gin.H{"message": "transfer approved", "new_allocation_id": newAllocID})
}

type RejectTransferRequest struct {
	Reason *string `json:"reason"`
}

func (h *AllocationHandler) RejectTransferHandler(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.Transfer.UpdateStatus(c.Request.Context(), id, "Rejected", nil, nil, nil); err != nil {
		if errors.Is(err, transfer_repo.ErrTransferNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "transfer request not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject transfer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer rejected"})
}

func (h *AllocationHandler) ListPendingTransfersHandler(c *gin.Context) {
	transfers, err := h.store.Transfer.ListPending(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list pending transfers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transfers": transfers})
}

func (h *AllocationHandler) ListMyAllocationsHandler(c *gin.Context) {
	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	allocations, err := h.store.Allocation.ListByEmployee(c.Request.Context(), empIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list allocations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"allocations": allocations})
}

func strPtr(s string) *string {
	return &s
}
