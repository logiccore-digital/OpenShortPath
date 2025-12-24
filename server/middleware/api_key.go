package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"openshortpath/server/constants"
	"openshortpath/server/models"
	"openshortpath/server/utils"
)

type APIKeyMiddleware struct {
	db *gorm.DB
}

func NewAPIKeyMiddleware(db *gorm.DB) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		db: db,
	}
}

// OptionalAuth is a middleware that optionally validates API keys
// If a valid API key is provided, it extracts the user_id and scopes and stores them in the context
// If no key or invalid key is provided, the request continues without setting user_id or scopes
func (m *APIKeyMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Bearer token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			c.Next()
			return
		}

		// Check if it's an API key (starts with osp_sk_)
		if !strings.HasPrefix(tokenString, utils.APIKeyPrefix) {
			c.Next()
			return
		}

		// Validate API key
		userID, scopes, err := m.validateAPIKey(tokenString)
		if err != nil {
			// Silently ignore validation errors (optional auth)
			c.Next()
			return
		}

		// Store user ID, scopes, and authentication method in context
		c.Set(constants.ContextKeyUserID, userID)
		c.Set(constants.ContextKeyScopes, scopes)
		c.Set(constants.ContextKeyAuthMethod, constants.AuthMethodAPIKey)
		c.Next()
	}
}

// ValidateAPIKey validates the API key and returns the user_id and scopes
// This is exported so it can be used by JWT middleware
func (m *APIKeyMiddleware) ValidateAPIKey(key string) (string, []string, error) {
	return m.validateAPIKey(key)
}

// validateAPIKey validates the API key and returns the user_id and scopes
func (m *APIKeyMiddleware) validateAPIKey(key string) (string, []string, error) {
	// Query all API keys (we need to check hashes, so we can't use direct lookup)
	// In production, you might want to add a prefix field for faster filtering
	var apiKeys []models.APIKey
	if err := m.db.Find(&apiKeys).Error; err != nil {
		return "", nil, err
	}

	// Iterate through keys and verify hash
	for _, apiKey := range apiKeys {
		valid, err := utils.VerifyPassword(key, apiKey.HashedKey)
		if err != nil {
			continue // Skip invalid hashes
		}
		if valid {
			return apiKey.UserID, apiKey.Scopes, nil
		}
	}

	return "", nil, gorm.ErrRecordNotFound
}
