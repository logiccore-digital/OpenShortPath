package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"openshortpath/server/constants"
)

// RequireScope is a middleware that checks if the required scope exists in the context
// If scopes are present (from API key), it checks if the required scope is included
// If no scopes are present (JWT token), it allows the request (backward compatibility)
func RequireScope(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get scopes from context
		scopesValue, exists := c.Get(constants.ContextKeyScopes)
		if !exists {
			// No scopes in context means JWT authentication (backward compatibility)
			// JWT tokens have full access, so allow the request
			c.Next()
			return
		}

		// Convert scopes to []string
		scopes, ok := scopesValue.([]string)
		if !ok {
			// Invalid scopes type, deny access
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid scopes in context",
			})
			c.Abort()
			return
		}

		// Check if required scope is present
		for _, s := range scopes {
			if s == scope {
				// Scope found, allow request
				c.Next()
				return
			}
		}

		// Scope not found, deny access
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions. Required scope: " + scope,
		})
		c.Abort()
	}
}
