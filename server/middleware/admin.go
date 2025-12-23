package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AdminMiddleware struct {
	adminPassword string
}

func NewAdminMiddleware(adminPassword string) *AdminMiddleware {
	return &AdminMiddleware{
		adminPassword: adminPassword,
	}
}

// RequireAdmin is a middleware that validates the admin password from the Authorization header
// The admin password should be passed as a Bearer token: Authorization: Bearer <admin_password>
// Returns 401 Unauthorized if password doesn't match or is missing
func (m *AdminMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Bearer token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing Authorization header",
			})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization header format. Expected: Bearer <admin_password>",
			})
			c.Abort()
			return
		}

		providedPassword := parts[1]
		if providedPassword == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing admin password in Authorization header",
			})
			c.Abort()
			return
		}

		// Compare passwords using constant-time comparison to prevent timing attacks
		if m.adminPassword == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Admin password not configured",
			})
			c.Abort()
			return
		}

		// Use constant-time comparison
		if subtle.ConstantTimeCompare([]byte(providedPassword), []byte(m.adminPassword)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid admin password",
			})
			c.Abort()
			return
		}

		// Authentication successful, continue to next handler
		c.Next()
	}
}

