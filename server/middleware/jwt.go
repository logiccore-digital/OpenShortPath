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
	"openshortpath/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTMiddleware struct {
	cfg         *config.JWT
	apiKeyMw    *APIKeyMiddleware
}

func NewJWTMiddleware(cfg *config.JWT) *JWTMiddleware {
	return &JWTMiddleware{
		cfg: cfg,
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

		// Store user ID in context
		c.Set(constants.ContextKeyUserID, userID)
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

		// Store user ID in context
		c.Set(constants.ContextKeyUserID, userID)
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

	// Store user ID and scopes in context
	c.Set(constants.ContextKeyUserID, userID)
	c.Set(constants.ContextKeyScopes, scopes)
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
