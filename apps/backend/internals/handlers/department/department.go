package department

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
	department_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/department-repo"
)

type DepartmentHandler struct {
	store *repository.StorageRegistry
}

func NewDepartmentHandler(store *repository.StorageRegistry) *DepartmentHandler {
	return &DepartmentHandler{store: store}
}

type CreateDepartmentRequest struct {
	Name               string  `json:"name" binding:"required"`
	ParentDepartmentID *string `json:"parent_department_id"`
	HeadEmployeeID     *string `json:"head_employee_id"`
}

type UpdateDepartmentRequest struct {
	Name               string  `json:"name" binding:"required"`
	ParentDepartmentID *string `json:"parent_department_id"`
	HeadEmployeeID     *string `json:"head_employee_id"`
}

func (h *DepartmentHandler) ListHandler(c *gin.Context) {
	departments, err := h.store.Department.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list departments"})
		return
	}
	if departments == nil {
		departments = []models.Department{}
	}
	c.JSON(http.StatusOK, gin.H{"departments": departments})
}

func (h *DepartmentHandler) CreateHandler(c *gin.Context) {
	var req CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dept := models.Department{
		Name:               req.Name,
		ParentDepartmentID: req.ParentDepartmentID,
		HeadEmployeeID:     req.HeadEmployeeID,
	}

	created, err := h.store.Department.Create(c.Request.Context(), dept)
	if err != nil {
		if errors.Is(err, department_repo.ErrDepartmentNameExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "department name already exists"})
			return
		}
		if errors.Is(err, department_repo.ErrParentDepartmentMissing) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "parent department not found"})
			return
		}
		if errors.Is(err, department_repo.ErrHeadEmployeeMissing) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "head employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create department"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"department": created})
}

func (h *DepartmentHandler) UpdateHandler(c *gin.Context) {
	id := c.Param("id")
	var req UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dept := models.Department{
		ID:                 id,
		Name:               req.Name,
		ParentDepartmentID: req.ParentDepartmentID,
		HeadEmployeeID:     req.HeadEmployeeID,
	}

	updated, err := h.store.Department.Update(c.Request.Context(), dept)
	if err != nil {
		if errors.Is(err, department_repo.ErrDepartmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
			return
		}
		if errors.Is(err, department_repo.ErrDepartmentNameExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "department name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update department"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"department": updated})
}

type DeactivateRequest struct {
	Force bool `json:"force"`
}

func (h *DepartmentHandler) DeactivateHandler(c *gin.Context) {
	id := c.Param("id")

	var req DeactivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Force = false
	}

	activeCount, err := h.store.Department.GetActiveEmployeeCount(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check active employees"})
		return
	}

	if activeCount > 0 && !req.Force {
		c.JSON(http.StatusConflict, gin.H{
			"error":                 "department has active employees",
			"active_employee_count": activeCount,
			"requires_confirmation": true,
		})
		return
	}

	if err := h.store.Department.Deactivate(c.Request.Context(), id); err != nil {
		if errors.Is(err, department_repo.ErrDepartmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "department not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate department"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "department deactivated"})
}
