package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"openshortpath/server/constants"
	"openshortpath/server/services"
)

// RateLimitMiddleware creates a middleware that enforces rate limiting
// based on IP address (for anonymous users) or user plan (for authenticated users)
// Rate limits are skipped for JWT-authenticated requests, but still apply for API key authentication
func RateLimitMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract client IP address
		clientIP := services.GetClientIP(c)

		// Check if user is authenticated
		userIDValue, exists := c.Get(constants.ContextKeyUserID)
		var userID string
		if exists {
			if userIDStr, ok := userIDValue.(string); ok {
				userID = userIDStr
			}
		}

		// Check authentication method
		authMethodValue, authMethodExists := c.Get(constants.ContextKeyAuthMethod)
		var authMethod string
		if authMethodExists {
			if authMethodStr, ok := authMethodValue.(string); ok {
				authMethod = authMethodStr
			}
		}

		// Skip rate limiting for JWT-authenticated requests
		if authMethod == constants.AuthMethodJWT {
			// JWT authentication - skip rate limit check
			c.Next()
			return
		}

		var rateLimitInfo *services.RateLimitInfo
		var err error
		var limitType string
		var identifier string
		var limitPerHour int

		if userID != "" {
			// Authenticated user (API key) - apply user-level rate limiting
			plan, err := services.GetUserPlan(db, userID)
			if err != nil {
				// If we can't get the plan, default to hobbyist limit
				plan = constants.PlanHobbyist
			}

			limitPerHour = services.GetRateLimitForPlan(plan)
			limitType = constants.RateLimitTypeUser
			identifier = userID

			// Check rate limit
			rateLimitInfo, err = services.CheckRateLimit(db, identifier, limitType, limitPerHour)
		} else {
			// Anonymous user - apply IP-level rate limiting
			limitPerHour = 5 // Anonymous users: 5 requests per hour per IP
			limitType = constants.RateLimitTypeIP
			identifier = clientIP

			// Check rate limit
			rateLimitInfo, err = services.CheckRateLimit(db, identifier, limitType, limitPerHour)
		}

		if err != nil {
			// Log error but don't block the request
			// In production, you might want to log this properly
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limit check failed",
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		if rateLimitInfo.Limit > 0 {
			c.Header("X-RateLimit-Limit", strconv.Itoa(rateLimitInfo.Limit))
			if rateLimitInfo.Remaining >= 0 {
				c.Header("X-RateLimit-Remaining", strconv.Itoa(rateLimitInfo.Remaining))
			}
			if !rateLimitInfo.Reset.IsZero() {
				c.Header("X-RateLimit-Reset", strconv.FormatInt(rateLimitInfo.Reset.Unix(), 10))
			}
		}

		if rateLimitInfo.Exceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		// Rate limit check passed, continue to next handler
		c.Next()
	}
}
