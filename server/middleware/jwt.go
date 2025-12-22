package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"openshortpath/server/config"
	"openshortpath/server/constants"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTMiddleware struct {
	cfg *config.JWT
}

func NewJWTMiddleware(cfg *config.JWT) *JWTMiddleware {
	return &JWTMiddleware{
		cfg: cfg,
	}
}

// OptionalAuth is a middleware that optionally validates JWT tokens
// If a valid token is provided, it extracts the 'sub' claim and stores it as 'user_id' in the context
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

		// Parse and validate token
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
