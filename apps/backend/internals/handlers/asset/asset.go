package asset

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
	asset_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/asset-repo"
)

type AssetHandler struct {
	store *repository.StorageRegistry
}

func NewAssetHandler(store *repository.StorageRegistry) *AssetHandler {
	return &AssetHandler{store: store}
}

type RegisterAssetRequest struct {
	Name            string                 `json:"name" binding:"required"`
	CategoryID      string                 `json:"category_id" binding:"required"`
	SerialNumber    *string                `json:"serial_number"`
	AcquisitionDate *string                `json:"acquisition_date"`
	AcquisitionCost *float64               `json:"acquisition_cost"`
	Condition       *string                `json:"condition"`
	Location        *string                `json:"location"`
	IsBookable      bool                   `json:"is_bookable"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
}

func (h *AssetHandler) RegisterHandler(c *gin.Context) {
	var req RegisterAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employeeID, _ := c.Get("employee_id")
	empIDStr, _ := employeeID.(string)

	if req.CustomFields == nil {
		req.CustomFields = make(map[string]interface{})
	}
	asset := models.Asset{
		Name:            req.Name,
		CategoryID:      req.CategoryID,
		SerialNumber:    req.SerialNumber,
		AcquisitionDate: req.AcquisitionDate,
		AcquisitionCost: req.AcquisitionCost,
		Condition:       req.Condition,
		Location:        req.Location,
		IsBookable:      req.IsBookable,
		CustomFields:    req.CustomFields,
	}

	created, err := h.store.Asset.Create(c.Request.Context(), asset)
	if err != nil {
		if errors.Is(err, asset_repo.ErrCategoryNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register asset"})
		return
	}

	status := "Available"
	history := models.AssetStatusHistory{
		AssetID:    created.ID,
		FromStatus: nil,
		ToStatus:   status,
		ChangedBy:  &empIDStr,
		Reason:     strPtr("Asset registered"),
	}
	if err := h.store.Asset.CreateStatusHistory(c.Request.Context(), history); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "asset created but failed to log status history"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"asset": created})
}

type ListAssetsQuery struct {
	AssetTag     string `form:"asset_tag"`
	SerialNumber string `form:"serial_number"`
	CategoryID   string `form:"category_id"`
	Status       string `form:"status"`
	DepartmentID string `form:"department"`
	Location     string `form:"location"`
	Page         string `form:"page"`
	PageSize     string `form:"page_size"`
}

func (h *AssetHandler) ListHandler(c *gin.Context) {
	var q ListAssetsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page, _ := strconv.Atoi(q.Page)
	pageSize, _ := strconv.Atoi(q.PageSize)

	filters := asset_repo.AssetListFilters{
		AssetTag:     q.AssetTag,
		SerialNumber: q.SerialNumber,
		CategoryID:   q.CategoryID,
		Status:       q.Status,
		DepartmentID: q.DepartmentID,
		Location:     q.Location,
		Page:         page,
		PageSize:     pageSize,
	}

	result, err := h.store.Asset.List(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list assets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assets":      result.Assets,
		"total_count": result.TotalCount,
		"page":        page,
		"page_size":   pageSize,
	})
}

type AssetDetailResponse struct {
	models.AssetDetail
	Documents []models.AssetDocument `json:"documents"`
}

func (h *AssetHandler) GetHandler(c *gin.Context) {
	id := c.Param("id")

	asset, err := h.store.Asset.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, asset_repo.ErrAssetNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get asset"})
		return
	}

	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	if roleStr == "Employee" {
		asset.AcquisitionCost = nil
	}

	c.JSON(http.StatusOK, gin.H{"asset": asset})
}

func (h *AssetHandler) GetHistoryHandler(c *gin.Context) {
	id := c.Param("id")

	events, err := h.store.Asset.GetHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get asset history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": events})
}

func (h *AssetHandler) UploadDocumentHandler(c *gin.Context) {
	id := c.Param("id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	docType := c.PostForm("type")
	if docType == "" {
		docType = "document"
	}
	if docType != "photo" && docType != "document" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be 'photo' or 'document'"})
		return
	}

	uploadDir := "uploads/assets"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := fmt.Sprintf("%s_%d%s", id, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	doc := models.AssetDocument{
		AssetID: id,
		URL:     fmt.Sprintf("/uploads/assets/%s", filename),
		Type:    docType,
	}

	created, err := h.store.Asset.CreateDocument(c.Request.Context(), doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file saved but failed to record document"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"document": created})
}

func strPtr(s string) *string {
	return &s
}
