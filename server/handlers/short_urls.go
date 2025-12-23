package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"openshortpath/server/config"
	"openshortpath/server/constants"
	"openshortpath/server/models"
)

type ShortURLsHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

type UpdateShortURLRequest struct {
	URL    string `json:"url,omitempty"`
	Slug   string `json:"slug,omitempty"`
	Domain string `json:"domain,omitempty"`
}

type ListResponse struct {
	URLs       []models.ShortURL `json:"urls"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	Total      int64              `json:"total"`
	TotalPages int                `json:"total_pages"`
}

func NewShortURLsHandler(db *gorm.DB, cfg *config.Config) *ShortURLsHandler {
	return &ShortURLsHandler{
		db:  db,
		cfg: cfg,
	}
}

// isValidDomain checks if the domain exists in the available short domains list
func isValidDomainForUpdate(domain string, availableDomains []string) bool {
	for _, availableDomain := range availableDomains {
		if domain == availableDomain {
			return true
		}
	}
	return false
}

// List returns a paginated list of shortened URLs for the authenticated user
func (h *ShortURLsHandler) List(c *gin.Context) {
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
	if err := h.db.Model(&models.ShortURL{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": err.Error(),
		})
		return
	}

	// Query paginated results
	var urls []models.ShortURL
	if err := h.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&urls).Error; err != nil {
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

	response := ListResponse{
		URLs:       urls,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// Get returns a single shortened URL by ID
func (h *ShortURLsHandler) Get(c *gin.Context) {
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

	// Find the ShortURL by ID and user_id
	var shortURL models.ShortURL
	result := h.db.Where("id = ? AND user_id = ?", id, userID).First(&shortURL)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Short URL not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, shortURL)
}

// Update updates a shortened URL by ID
func (h *ShortURLsHandler) Update(c *gin.Context) {
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

	// Find the ShortURL by ID and user_id
	var shortURL models.ShortURL
	result := h.db.Where("id = ? AND user_id = ?", id, userID).First(&shortURL)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Short URL not found",
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
	var req UpdateShortURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Update fields if provided
	updateFields := make(map[string]interface{})

	if req.URL != "" {
		updateFields["url"] = req.URL
	}

	if req.Domain != "" {
		// Validate domain if it's being changed
		if !isValidDomainForUpdate(req.Domain, h.cfg.AvailableShortDomains) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Domain '%s' is not in the list of available short domains", req.Domain),
			})
			return
		}
		updateFields["domain"] = req.Domain
	}

	if req.Slug != "" {
		// Check for slug conflict if slug is being changed
		newDomain := req.Domain
		if newDomain == "" {
			newDomain = shortURL.Domain
		}

		var existing models.ShortURL
		conflictResult := h.db.Where("domain = ? AND slug = ? AND id != ?", newDomain, req.Slug, id).First(&existing)
		if conflictResult.Error == nil {
			// Conflict found
			c.JSON(http.StatusConflict, gin.H{
				"error": fmt.Sprintf("Short URL with domain '%s' and slug '%s' already exists", newDomain, req.Slug),
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
		updateFields["slug"] = req.Slug
	}

	// If no fields to update, return the existing record
	if len(updateFields) == 0 {
		c.JSON(http.StatusOK, shortURL)
		return
	}

	// Update the record
	if err := h.db.Model(&shortURL).Updates(updateFields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update short URL",
			"details": err.Error(),
		})
		return
	}

	// Reload the record to get updated values
	if err := h.db.Where("id = ?", id).First(&shortURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reload updated short URL",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, shortURL)
}

// Delete deletes a shortened URL by ID
func (h *ShortURLsHandler) Delete(c *gin.Context) {
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

	// Find the ShortURL by ID and user_id to verify ownership
	var shortURL models.ShortURL
	result := h.db.Where("id = ? AND user_id = ?", id, userID).First(&shortURL)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Short URL not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
		return
	}

	// Delete the record
	if err := h.db.Delete(&shortURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete short URL",
			"details": err.Error(),
		})
		return
	}

	c.AbortWithStatus(http.StatusNoContent)
}

