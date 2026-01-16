package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"openshortpath/server/config"
	"openshortpath/server/constants"
)

// generateTestRSAKeyPair generates a test RSA key pair for testing RS256
func generateTestRSAKeyPair(t *testing.T) (*rsa.PrivateKey, string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return privateKey, string(publicKeyPEM)
}

// generateTestTokenHS256 generates a test JWT token signed with HS256
func generateTestTokenHS256(secretKey string, sub string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": sub,
	})
	return token.SignedString([]byte(secretKey))
}

// generateTestTokenRS256 generates a test JWT token signed with RS256
func generateTestTokenRS256(privateKey *rsa.PrivateKey, sub string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": sub,
	})
	return token.SignedString(privateKey)
}

func TestJWTMiddleware_OptionalAuth_NoAuthorizationHeader(t *testing.T) {
	// Setup
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	// Execute
	handler(c)

	// Assert - should continue without setting user_id
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTMiddleware_OptionalAuth_InvalidAuthorizationFormat(t *testing.T) {
	// Setup
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	// Test cases
	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "token123"},
		{"Wrong prefix", "Basic token123"},
		{"Empty token", "Bearer "},
		{"Empty header", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.header != "" {
				c.Request.Header.Set("Authorization", tc.header)
			}

			handler(c)

			_, exists := c.Get(constants.ContextKeyUserID)
			assert.False(t, exists)
		})
	}
}

func TestJWTMiddleware_OptionalAuth_ValidHS256Token(t *testing.T) {
	// Setup
	secretKey := "test-secret-key"
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: secretKey,
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	// Generate valid token
	tokenString, err := generateTestTokenHS256(secretKey, "user123")
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute
	handler(c)

	// Assert
	userID, exists := c.Get(constants.ContextKeyUserID)
	assert.True(t, exists)
	assert.Equal(t, "user123", userID)
}

func TestJWTMiddleware_OptionalAuth_ValidRS256Token(t *testing.T) {
	// Setup
	privateKey, publicKeyPEM := generateTestRSAKeyPair(t)
	cfg := &config.JWT{
		Algorithm: "RS256",
		PublicKey: publicKeyPEM,
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	// Generate valid token
	tokenString, err := generateTestTokenRS256(privateKey, "user456")
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute
	handler(c)

	// Assert
	userID, exists := c.Get(constants.ContextKeyUserID)
	assert.True(t, exists)
	assert.Equal(t, "user456", userID)
}

func TestJWTMiddleware_OptionalAuth_InvalidTokenSignature(t *testing.T) {
	// Setup
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	// Generate token with wrong secret
	wrongToken, err := generateTestTokenHS256("wrong-secret", "user123")
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+wrongToken)

	// Execute
	handler(c)

	// Assert - should continue without setting user_id (optional auth)
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

func TestJWTMiddleware_OptionalAuth_MissingSubClaim(t *testing.T) {
	// Setup
	secretKey := "test-secret-key"
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: secretKey,
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	// Generate token without sub claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"other": "claim",
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute
	handler(c)

	// Assert - should continue without setting user_id
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

func TestJWTMiddleware_OptionalAuth_EmptySubClaim(t *testing.T) {
	// Setup
	secretKey := "test-secret-key"
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: secretKey,
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	// Generate token with empty sub claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "",
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute
	handler(c)

	// Assert - should continue without setting user_id
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

func TestJWTMiddleware_OptionalAuth_InvalidTokenFormat(t *testing.T) {
	// Setup
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	// Setup Gin context with invalid token
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid.token.here")

	// Execute
	handler(c)

	// Assert - should continue without setting user_id
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

func TestJWTMiddleware_OptionalAuth_WrongAlgorithm(t *testing.T) {
	// Setup - configured for HS256 but token uses RS256
	privateKey, _ := generateTestRSAKeyPair(t)
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	// Generate RS256 token
	tokenString, err := generateTestTokenRS256(privateKey, "user123")
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute
	handler(c)

	// Assert - should continue without setting user_id
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

func TestJWTMiddleware_OptionalAuth_MissingSecretKey(t *testing.T) {
	// Setup
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "", // Missing secret key
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	// Generate token with some secret
	tokenString, err := generateTestTokenHS256("some-secret", "user123")
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute
	handler(c)

	// Assert - should continue without setting user_id
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

func TestJWTMiddleware_OptionalAuth_MissingPublicKey(t *testing.T) {
	// Setup
	cfg := &config.JWT{
		Algorithm: "RS256",
		PublicKey: "", // Missing public key
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	privateKey, _ := generateTestRSAKeyPair(t)
	tokenString, err := generateTestTokenRS256(privateKey, "user123")
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute
	handler(c)

	// Assert - should continue without setting user_id
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

func TestJWTMiddleware_OptionalAuth_UnsupportedAlgorithm(t *testing.T) {
	// Setup
	cfg := &config.JWT{
		Algorithm: "ES256", // Unsupported algorithm
		SecretKey: "test-secret",
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer some.token.here")

	// Execute
	handler(c)

	// Assert - should continue without setting user_id
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

func TestJWTMiddleware_OptionalAuth_InvalidPublicKeyFormat(t *testing.T) {
	// Setup
	cfg := &config.JWT{
		Algorithm: "RS256",
		PublicKey: "invalid-pem-format",
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	privateKey, _ := generateTestRSAKeyPair(t)
	tokenString, err := generateTestTokenRS256(privateKey, "user123")
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute
	handler(c)

	// Assert - should continue without setting user_id
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

func TestJWTMiddleware_OptionalAuth_NonStringSubClaim(t *testing.T) {
	// Setup
	secretKey := "test-secret-key"
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: secretKey,
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.OptionalAuth()

	// Generate token with non-string sub claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": 123, // Should be string
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute
	handler(c)

	// Assert - should continue without setting user_id
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

func TestJWTMiddleware_OptionalAuth_NilConfig(t *testing.T) {
	// Setup - nil config
	middleware := NewJWTMiddleware(nil, nil, "local")
	handler := middleware.OptionalAuth()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer some.token.here")

	// Execute
	handler(c)

	// Assert - should continue without setting user_id
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

// RequireAuth tests
func TestJWTMiddleware_RequireAuth_NoAuthorizationHeader(t *testing.T) {
	// Setup
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.RequireAuth()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	// Execute
	handler(c)

	// Assert - should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
	assert.True(t, c.IsAborted())
}

func TestJWTMiddleware_RequireAuth_InvalidAuthorizationFormat(t *testing.T) {
	// Setup
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.RequireAuth()

	// Test cases
	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "token123"},
		{"Wrong prefix", "Basic token123"},
		{"Empty token", "Bearer "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			c.Request.Header.Set("Authorization", tc.header)

			handler(c)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			_, exists := c.Get(constants.ContextKeyUserID)
			assert.False(t, exists)
			assert.True(t, c.IsAborted())
		})
	}
}

func TestJWTMiddleware_RequireAuth_ValidHS256Token(t *testing.T) {
	// Setup
	secretKey := "test-secret-key"
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: secretKey,
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.RequireAuth()

	// Generate valid token
	tokenString, err := generateTestTokenHS256(secretKey, "user123")
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute
	handler(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	userID, exists := c.Get(constants.ContextKeyUserID)
	assert.True(t, exists)
	assert.Equal(t, "user123", userID)
	assert.False(t, c.IsAborted())
}

func TestJWTMiddleware_RequireAuth_ValidRS256Token(t *testing.T) {
	// Setup
	privateKey, publicKeyPEM := generateTestRSAKeyPair(t)
	cfg := &config.JWT{
		Algorithm: "RS256",
		PublicKey: publicKeyPEM,
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.RequireAuth()

	// Generate valid token
	tokenString, err := generateTestTokenRS256(privateKey, "user456")
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute
	handler(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	userID, exists := c.Get(constants.ContextKeyUserID)
	assert.True(t, exists)
	assert.Equal(t, "user456", userID)
	assert.False(t, c.IsAborted())
}

func TestJWTMiddleware_RequireAuth_InvalidTokenSignature(t *testing.T) {
	// Setup
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.RequireAuth()

	// Generate token with wrong secret
	wrongToken, err := generateTestTokenHS256("wrong-secret", "user123")
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+wrongToken)

	// Execute
	handler(c)

	// Assert - should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
	assert.True(t, c.IsAborted())
}

func TestJWTMiddleware_RequireAuth_MissingSubClaim(t *testing.T) {
	// Setup
	secretKey := "test-secret-key"
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: secretKey,
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.RequireAuth()

	// Generate token without sub claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"other": "claim",
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	assert.NoError(t, err)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Execute
	handler(c)

	// Assert - should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
	assert.True(t, c.IsAborted())
}

func TestJWTMiddleware_RequireAuth_InvalidTokenFormat(t *testing.T) {
	// Setup
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	middleware := NewJWTMiddleware(cfg, nil, "local")
	handler := middleware.RequireAuth()

	// Setup Gin context with invalid token
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid.token.here")

	// Execute
	handler(c)

	// Assert - should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
	assert.True(t, c.IsAborted())
}

func TestJWTMiddleware_RequireAuth_NilConfig(t *testing.T) {
	// Setup - nil config
	middleware := NewJWTMiddleware(nil, nil, "local")
	handler := middleware.RequireAuth()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer some.token.here")

	// Execute
	handler(c)

	// Assert - should return 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
	assert.True(t, c.IsAborted())
}
