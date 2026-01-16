package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"

	"openshortpath/server/config"
	"openshortpath/server/constants"
	"openshortpath/server/models"
	"openshortpath/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type JWTMiddleware struct {
	cfg          *config.JWT
	apiKeyMw     *APIKeyMiddleware
	db           *gorm.DB
	authProvider string
}

func NewJWTMiddleware(cfg *config.JWT, db *gorm.DB, authProvider string) *JWTMiddleware {
	return &JWTMiddleware{
		cfg:          cfg,
		db:           db,
		authProvider: authProvider,
	}
}

// SetAPIKeyMiddleware sets the API key middleware for unified authentication
func (m *JWTMiddleware) SetAPIKeyMiddleware(apiKeyMw *APIKeyMiddleware) {
	m.apiKeyMw = apiKeyMw
}

// OptionalAuth is a middleware that optionally validates JWT tokens or API keys
// If a valid API key is provided (starts with osp_sk_), it uses API key authentication
// Otherwise, if a valid JWT token is provided, it extracts the 'sub' claim and stores it as 'user_id' in the context
// If no token or invalid token is provided, the request continues without setting user_id
func (m *JWTMiddleware) OptionalAuth() gin.HandlerFunc {
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
		if strings.HasPrefix(tokenString, utils.APIKeyPrefix) {
			// Delegate to API key middleware if available
			if m.apiKeyMw != nil {
				m.apiKeyMw.OptionalAuth()(c)
				return
			}
			// If API key middleware not set, continue (will fail validation later)
			c.Next()
			return
		}

		// Parse and validate JWT token
		userID, err := m.validateToken(tokenString)
		if err != nil {
			// Silently ignore validation errors (optional auth)
			c.Next()
			return
		}

		// Ensure user exists in database for external JWT providers
		if err := m.ensureUserExists(userID); err != nil {
			// Log error but don't fail the request (optional auth)
			// The user will be created on the next request or when they hit a protected endpoint
			c.Next()
			return
		}

		// Store user ID and authentication method in context
		c.Set(constants.ContextKeyUserID, userID)
		c.Set(constants.ContextKeyAuthMethod, constants.AuthMethodJWT)
		c.Next()
	}
}

// validateToken validates the JWT token and returns the 'sub' claim as user ID
func (m *JWTMiddleware) validateToken(tokenString string) (string, error) {
	if m.cfg == nil {
		return "", fmt.Errorf("JWT config not provided")
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		switch m.cfg.Algorithm {
		case "HS256":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			if m.cfg.SecretKey == "" {
				return nil, fmt.Errorf("secret key not provided for HS256")
			}
			return []byte(m.cfg.SecretKey), nil

		case "RS256":
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			if m.cfg.PublicKey == "" {
				return nil, fmt.Errorf("public key not provided for RS256")
			}
			// Parse PEM-encoded public key
			publicKey, err := parseRSAPublicKey(m.cfg.PublicKey)
			if err != nil {
				return nil, fmt.Errorf("failed to parse public key: %w", err)
			}
			return publicKey, nil

		default:
			return nil, fmt.Errorf("unsupported algorithm: %s", m.cfg.Algorithm)
		}
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	// Extract 'sub' claim
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to extract claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", fmt.Errorf("'sub' claim is missing or invalid")
	}

	return sub, nil
}

// ensureUserExists ensures a user exists in the database for external JWT providers
// If the user doesn't exist and authProvider is "external_jwt" or "clerk", it creates a new user
// with UserID and Username both set to the sub claim
func (m *JWTMiddleware) ensureUserExists(userID string) error {
	// Only auto-create users for external JWT providers, not for local auth
	if m.authProvider != "external_jwt" && m.authProvider != "clerk" {
		return nil
	}

	// Check if database is available
	if m.db == nil {
		return fmt.Errorf("database not available")
	}

	// Use FirstOrCreate to atomically check and create user
	// This handles race conditions where multiple requests might try to create the same user concurrently
	username := userID
	user := models.User{
		UserID:         userID,
		Username:       &username,
		HashedPassword: nil, // No password for external auth
		Active:         true,
		Plan:           constants.PlanHobbyist,
	}

	// FirstOrCreate will find the user if it exists, or create it if it doesn't
	// The Where clause specifies the condition to find existing user
	result := m.db.Where("user_id = ?", userID).FirstOrCreate(&user)
	if result.Error != nil {
		// Check if it's a unique constraint violation on username
		// This can happen if another request created the user with the same username
		// In that case, the user might exist but with a different user_id, or there was a race
		if strings.Contains(result.Error.Error(), "UNIQUE constraint") || 
		   strings.Contains(result.Error.Error(), "duplicate key") {
			// Try to find the user by user_id (which should be unique)
			var existingUser models.User
			if err := m.db.Where("user_id = ?", userID).First(&existingUser).Error; err == nil {
				// User exists, that's fine
				return nil
			}
		}
		return fmt.Errorf("failed to ensure user exists: %w", result.Error)
	}

	return nil
}

// RequireAuth is a middleware that requires JWT token or API key authentication
// If a valid API key is provided (starts with osp_sk_), it uses API key authentication
// Otherwise, if a valid JWT token is provided, it extracts the 'sub' claim and stores it as 'user_id' in the context
// If no token or invalid token is provided, it returns 401 Unauthorized
func (m *JWTMiddleware) RequireAuth() gin.HandlerFunc {
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
				"error": "Invalid Authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing token in Authorization header",
			})
			c.Abort()
			return
		}

		// Check if it's an API key (starts with osp_sk_)
		if strings.HasPrefix(tokenString, utils.APIKeyPrefix) {
			// Delegate to API key middleware if available
			if m.apiKeyMw != nil {
				// Use a wrapper to convert optional auth to required auth
				m.requireAPIKeyAuth(c, tokenString)
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key authentication not configured",
			})
			c.Abort()
			return
		}

		// Parse and validate JWT token
		userID, err := m.validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Ensure user exists in database for external JWT providers
		if err := m.ensureUserExists(userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to ensure user exists",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Store user ID and authentication method in context
		c.Set(constants.ContextKeyUserID, userID)
		c.Set(constants.ContextKeyAuthMethod, constants.AuthMethodJWT)
		c.Next()
	}
}

// requireAPIKeyAuth validates an API key and requires it to be valid
func (m *JWTMiddleware) requireAPIKeyAuth(c *gin.Context, key string) {
	userID, scopes, err := m.apiKeyMw.ValidateAPIKey(key)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired API key",
		})
		c.Abort()
		return
	}

	// Store user ID, scopes, and authentication method in context
	c.Set(constants.ContextKeyUserID, userID)
	c.Set(constants.ContextKeyScopes, scopes)
	c.Set(constants.ContextKeyAuthMethod, constants.AuthMethodAPIKey)
	c.Next()
}

// parseRSAPublicKey parses a PEM-encoded RSA public key
func parseRSAPublicKey(pemString string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}
