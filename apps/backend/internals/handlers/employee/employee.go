package employee

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
	employee_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/employee-repo"
)

type EmployeeHandler struct {
	store *repository.StorageRegistry
}

func NewEmployeeHandler(store *repository.StorageRegistry) *EmployeeHandler {
	return &EmployeeHandler{store: store}
}

type UpdateEmployeeRequest struct {
	Name         string  `json:"name"`
	DepartmentID *string `json:"department_id"`
	Status       string  `json:"status"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=DepartmentHead AssetManager Employee"`
}

func (h *EmployeeHandler) ListHandler(c *gin.Context) {
	departmentID := c.Query("department_id")
	role := c.Query("role")
	status := c.Query("status")

	employees, err := h.store.Employee.List(c.Request.Context(), departmentID, role, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list employees"})
		return
	}
	if employees == nil {
		employees = []models.Employee{}
	}
	c.JSON(http.StatusOK, gin.H{"employees": employees})
}

func (h *EmployeeHandler) UpdateHandler(c *gin.Context) {
	id := c.Param("id")
	var req UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emp := models.Employee{
		ID:           id,
		Name:         req.Name,
		DepartmentID: req.DepartmentID,
		Status:       req.Status,
	}

	updated, err := h.store.Employee.Update(c.Request.Context(), emp)
	if err != nil {
		if errors.Is(err, employee_repo.ErrEmployeeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update employee"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"employee": updated})
}

func (h *EmployeeHandler) UpdateRoleHandler(c *gin.Context) {
	id := c.Param("id")
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID, _ := c.Get("employee_id")
	actorIDStr := actorID.(string)

	employee, err := h.store.Employee.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, employee_repo.ErrEmployeeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get employee"})
		return
	}

	oldRole := employee.Role

	if err := h.store.Employee.UpdateRole(c.Request.Context(), id, req.Role); err != nil {
		if errors.Is(err, employee_repo.ErrEmployeeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}

	if err := h.store.ActivityLog.Create(c.Request.Context(), models.ActivityLog{
		ActorEmployeeID: &actorIDStr,
		Action:          "employee.role.change",
		EntityType:      "employee",
		EntityID:        &id,
		Metadata: map[string]interface{}{
			"from_role": oldRole,
			"to_role":   req.Role,
		},
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log role change"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "role updated",
		"from_role": oldRole,
		"to_role":   req.Role,
	})
}
