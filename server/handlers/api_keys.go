package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"openshortpath/server/constants"
	"openshortpath/server/models"
	"openshortpath/server/utils"
)

type APIKeysHandler struct {
	db *gorm.DB
}

type CreateAPIKeyRequest struct {
	Scopes []string `json:"scopes" binding:"required"`
}

type CreateAPIKeyResponse struct {
	ID        string   `json:"id"`
	Key       string   `json:"key"` // Only shown once during creation
	Scopes    []string `json:"scopes"`
	CreatedAt string   `json:"created_at"`
}

type APIKeyListItem struct {
	ID        string   `json:"id"`
	Scopes    []string `json:"scopes"`
	CreatedAt string   `json:"created_at"`
}

type ListAPIKeysResponse struct {
	Keys []APIKeyListItem `json:"keys"`
}

func NewAPIKeysHandler(db *gorm.DB) *APIKeysHandler {
	return &APIKeysHandler{
		db: db,
	}
}

// CreateAPIKey handles POST /api/v1/api-keys
// Creates a new API key for the authenticated user
func (h *APIKeysHandler) CreateAPIKey(c *gin.Context) {
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
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate scopes
	validScopes := map[string]bool{
		"shorten_url": true,
		"read_urls":   true,
		"write_urls":  true,
	}
	for _, scope := range req.Scopes {
		if !validScopes[scope] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid scope: " + scope,
			})
			return
		}
	}

	// Generate API key
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate API key",
			"details": err.Error(),
		})
		return
	}

	// Hash the API key
	hashedKey, err := utils.HashPassword(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to hash API key",
			"details": err.Error(),
		})
		return
	}

	// Create API key record
	id := uuid.New().String()
	apiKeyRecord := models.APIKey{
		ID:        id,
		UserID:    userID,
		HashedKey: hashedKey,
		Scopes:    req.Scopes,
	}

	if err := h.db.Create(&apiKeyRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create API key",
			"details": err.Error(),
		})
		return
	}

	// Return response with the plain key (shown only once)
	response := CreateAPIKeyResponse{
		ID:        id,
		Key:       apiKey,
		Scopes:    req.Scopes,
		CreatedAt: apiKeyRecord.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusCreated, response)
}

// ListAPIKeys handles GET /api/v1/api-keys
// Returns a list of API keys for the authenticated user (without key values)
func (h *APIKeysHandler) ListAPIKeys(c *gin.Context) {
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

	// Query API keys for this user
	var apiKeys []models.APIKey
	if err := h.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&apiKeys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": err.Error(),
		})
		return
	}

	// Build response (without key values)
	items := make([]APIKeyListItem, len(apiKeys))
	for i, key := range apiKeys {
		items[i] = APIKeyListItem{
			ID:        key.ID,
			Scopes:    key.Scopes,
			CreatedAt: key.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	response := ListAPIKeysResponse{
		Keys: items,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteAPIKey handles DELETE /api/v1/api-keys/:id
// Deletes an API key if it belongs to the authenticated user
func (h *APIKeysHandler) DeleteAPIKey(c *gin.Context) {
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

	// Get ID from URL parameter
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID parameter is required",
		})
		return
	}

	// Find the API key by ID and user_id to verify ownership
	var apiKey models.APIKey
	result := h.db.Where("id = ? AND user_id = ?", id, userID).First(&apiKey)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API key not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
		return
	}

	// Delete the API key
	if err := h.db.Delete(&apiKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete API key",
			"details": err.Error(),
		})
		return
	}

	c.AbortWithStatus(http.StatusNoContent)
}

