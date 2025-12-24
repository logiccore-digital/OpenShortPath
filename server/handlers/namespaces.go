package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"openshortpath/server/config"
	"openshortpath/server/constants"
	"openshortpath/server/models"
)

type NamespacesHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

type CreateNamespaceRequest struct {
	Name   string `json:"name" binding:"required"`
	Domain string `json:"domain" binding:"required"`
}

type UpdateNamespaceRequest struct {
	Name   string `json:"name,omitempty"`
	Domain string `json:"domain,omitempty"`
}

type ListNamespacesResponse struct {
	Namespaces []models.Namespace `json:"namespaces"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	Total      int64              `json:"total"`
	TotalPages int                `json:"total_pages"`
}

func NewNamespacesHandler(db *gorm.DB, cfg *config.Config) *NamespacesHandler {
	return &NamespacesHandler{
		db:  db,
		cfg: cfg,
	}
}

// isValidDomain checks if the domain exists in the available short domains list
func isValidDomainForNamespace(domain string, availableDomains []string) bool {
	for _, availableDomain := range availableDomains {
		if domain == availableDomain {
			return true
		}
	}
	return false
}

// isValidNamespaceName validates that the namespace name is lowercase alphanumerical
// with optional hyphens or underscores, and has a maximum length of 32 characters
func isValidNamespaceName(name string) bool {
	// Check length (max 32 characters)
	if len(name) > 32 {
		return false
	}
	// Check if name is reserved
	if IsReservedNamespaceName(name) {
		return false
	}
	// Pattern: lowercase letters, digits, hyphens, and underscores only
	matched, _ := regexp.MatchString(`^[a-z0-9_-]+$`, name)
	return matched
}

// CreateNamespace handles POST /api/v1/namespaces
func (h *NamespacesHandler) CreateNamespace(c *gin.Context) {
	// Get user ID from context (set by RequireAuth middleware)
	userIDValue, exists := c.Get(constants.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user ID in context",
		})
		return
	}

	// Parse request body
	var req CreateNamespaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate domain
	if !isValidDomainForNamespace(req.Domain, h.cfg.AvailableShortDomains) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Domain '%s' is not in the list of available short domains", req.Domain),
		})
		return
	}

	// Validate namespace name format
	if !isValidNamespaceName(req.Name) {
		if IsReservedNamespaceName(req.Name) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Namespace name '%s' is reserved and cannot be used", req.Name),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace name must be lowercase alphanumerical with optional hyphens or underscores, and must be 32 characters or less",
		})
		return
	}

	// Check for existing namespace with same (domain, name) combination
	var existing models.Namespace
	result := h.db.Where("domain = ? AND name = ?", req.Domain, req.Name).First(&existing)
	if result.Error == nil {
		// Record exists
		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("Namespace with domain '%s' and name '%s' already exists", req.Domain, req.Name),
		})
		return
	}
	if result.Error != gorm.ErrRecordNotFound {
		// Database error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
		return
	}

	// Generate UUID for ID
	id := uuid.New().String()

	// Create new Namespace record
	namespace := models.Namespace{
		ID:     id,
		Name:   req.Name,
		Domain: req.Domain,
		UserID: userID,
	}

	if err := h.db.Create(&namespace).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create namespace",
			"details": err.Error(),
		})
		return
	}

	// Return the created namespace
	c.JSON(http.StatusCreated, namespace)
}

// ListNamespaces handles GET /api/v1/namespaces
func (h *NamespacesHandler) ListNamespaces(c *gin.Context) {
	// Get user ID from context (set by RequireAuth middleware)
	userIDValue, exists := c.Get(constants.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user ID in context",
		})
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > 100 {
				limit = 100
			} else {
				limit = l
			}
		}
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Query total count
	var total int64
	if err := h.db.Model(&models.Namespace{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": err.Error(),
		})
		return
	}

	// Query paginated results
	var namespaces []models.Namespace
	if err := h.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&namespaces).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": err.Error(),
		})
		return
	}

	// Calculate total pages
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	response := ListNamespacesResponse{
		Namespaces: namespaces,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// GetNamespace handles GET /api/v1/namespaces/:id
func (h *NamespacesHandler) GetNamespace(c *gin.Context) {
	// Get user ID from context
	userIDValue, exists := c.Get(constants.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user ID in context",
		})
		return
	}

	// Get ID from URL parameter
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID parameter is required",
		})
		return
	}

	// Find the Namespace by ID and user_id
	var namespace models.Namespace
	result := h.db.Where("id = ? AND user_id = ?", id, userID).First(&namespace)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Namespace not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, namespace)
}

// UpdateNamespace handles PUT /api/v1/namespaces/:id
func (h *NamespacesHandler) UpdateNamespace(c *gin.Context) {
	// Get user ID from context
	userIDValue, exists := c.Get(constants.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user ID in context",
		})
		return
	}

	// Get ID from URL parameter
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID parameter is required",
		})
		return
	}

	// Find the Namespace by ID and user_id
	var namespace models.Namespace
	result := h.db.Where("id = ? AND user_id = ?", id, userID).First(&namespace)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Namespace not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
		return
	}

	// Parse request body
	var req UpdateNamespaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Update fields if provided
	updateFields := make(map[string]interface{})

	if req.Domain != "" {
		// Validate domain if it's being changed
		if !isValidDomainForNamespace(req.Domain, h.cfg.AvailableShortDomains) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Domain '%s' is not in the list of available short domains", req.Domain),
			})
			return
		}
		updateFields["domain"] = req.Domain
	}

	if req.Name != "" {
		// Validate namespace name format
		if !isValidNamespaceName(req.Name) {
			if IsReservedNamespaceName(req.Name) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Namespace name '%s' is reserved and cannot be used", req.Name),
				})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Namespace name must be lowercase alphanumerical with optional hyphens or underscores, and must be 32 characters or less",
			})
			return
		}

		// Check for name conflict if name is being changed
		newDomain := req.Domain
		if newDomain == "" {
			newDomain = namespace.Domain
		}

		var existing models.Namespace
		conflictResult := h.db.Where("domain = ? AND name = ? AND id != ?", newDomain, req.Name, id).First(&existing)
		if conflictResult.Error == nil {
			// Conflict found
			c.JSON(http.StatusConflict, gin.H{
				"error": fmt.Sprintf("Namespace with domain '%s' and name '%s' already exists", newDomain, req.Name),
			})
			return
		}
		if conflictResult.Error != gorm.ErrRecordNotFound {
			// Database error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Database error",
				"details": conflictResult.Error.Error(),
			})
			return
		}
		updateFields["name"] = req.Name
	}

	// If no fields to update, return the existing record
	if len(updateFields) == 0 {
		c.JSON(http.StatusOK, namespace)
		return
	}

	// Update the record
	if err := h.db.Model(&namespace).Updates(updateFields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update namespace",
			"details": err.Error(),
		})
		return
	}

	// Reload the record to get updated values
	if err := h.db.Where("id = ?", id).First(&namespace).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reload updated namespace",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, namespace)
}

// DeleteNamespace handles DELETE /api/v1/namespaces/:id
func (h *NamespacesHandler) DeleteNamespace(c *gin.Context) {
	// Get user ID from context
	userIDValue, exists := c.Get(constants.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user ID in context",
		})
		return
	}

	// Get ID from URL parameter
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID parameter is required",
		})
		return
	}

	// Find the Namespace by ID and user_id to verify ownership
	var namespace models.Namespace
	result := h.db.Where("id = ? AND user_id = ?", id, userID).First(&namespace)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Namespace not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
		return
	}

	// Delete all short URLs that reference this namespace
	if err := h.db.Where("namespace_id = ?", id).Delete(&models.ShortURL{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete associated short URLs",
			"details": err.Error(),
		})
		return
	}

	// Delete the namespace record
	if err := h.db.Delete(&namespace).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete namespace",
			"details": err.Error(),
		})
		return
	}

	c.AbortWithStatus(http.StatusNoContent)
}
