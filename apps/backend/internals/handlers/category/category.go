package category

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
	category_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/category-repo"
)

type CategoryHandler struct {
	store *repository.StorageRegistry
}

func NewCategoryHandler(store *repository.StorageRegistry) *CategoryHandler {
	return &CategoryHandler{store: store}
}

type CreateCategoryRequest struct {
	Name         string                 `json:"name" binding:"required"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

type UpdateCategoryRequest struct {
	Name         string                 `json:"name" binding:"required"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func (h *CategoryHandler) ListHandler(c *gin.Context) {
	categories, err := h.store.Category.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list categories"})
		return
	}
	if categories == nil {
		categories = []models.AssetCategory{}
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func (h *CategoryHandler) CreateHandler(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.CustomFields == nil {
		req.CustomFields = make(map[string]interface{})
	}

	cat := models.AssetCategory{
		Name:         req.Name,
		CustomFields: req.CustomFields,
	}

	created, err := h.store.Category.Create(c.Request.Context(), cat)
	if err != nil {
		if errors.Is(err, category_repo.ErrCategoryNameExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "category name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"category": created})
}

func (h *CategoryHandler) UpdateHandler(c *gin.Context) {
	id := c.Param("id")
	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.CustomFields == nil {
		req.CustomFields = make(map[string]interface{})
	}

	cat := models.AssetCategory{
		ID:           id,
		Name:         req.Name,
		CustomFields: req.CustomFields,
	}

	updated, err := h.store.Category.Update(c.Request.Context(), cat)
	if err != nil {
		if errors.Is(err, category_repo.ErrCategoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}
		if errors.Is(err, category_repo.ErrCategoryNameExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "category name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"category": updated})
}