package middleware

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"openshortpath/server/constants"
	"openshortpath/server/models"
	"openshortpath/server/utils"
)

func setupTestDBForAPIKey(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql db: %v", err)
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm db: %v", err)
	}

	return db, mock, sqlDB
}

func TestAPIKeyMiddleware_OptionalAuth_NoAuthorizationHeader(t *testing.T) {
	db, _, sqlDB := setupTestDBForAPIKey(t)
	defer sqlDB.Close()

	middleware := NewAPIKeyMiddleware(db)
	handler := middleware.OptionalAuth()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	handler(c)

	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
	_, exists = c.Get(constants.ContextKeyScopes)
	assert.False(t, exists)
}

func TestAPIKeyMiddleware_OptionalAuth_InvalidFormat(t *testing.T) {
	db, _, sqlDB := setupTestDBForAPIKey(t)
	defer sqlDB.Close()

	middleware := NewAPIKeyMiddleware(db)
	handler := middleware.OptionalAuth()

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
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			c.Request = req

			handler(c)

			_, exists := c.Get(constants.ContextKeyUserID)
			assert.False(t, exists)
		})
	}
}

func TestAPIKeyMiddleware_OptionalAuth_NotAPIKey(t *testing.T) {
	db, _, sqlDB := setupTestDBForAPIKey(t)
	defer sqlDB.Close()

	middleware := NewAPIKeyMiddleware(db)
	handler := middleware.OptionalAuth()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer jwt_token_here")
	c.Request = req

	handler(c)

	// Should not set user_id or scopes for non-API-key tokens
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

func TestAPIKeyMiddleware_OptionalAuth_ValidAPIKey(t *testing.T) {
	db, mock, sqlDB := setupTestDBForAPIKey(t)
	defer sqlDB.Close()

	// Generate a test API key
	apiKey, err := utils.GenerateAPIKey()
	assert.NoError(t, err)

	// Hash the key
	hashedKey, err := utils.HashPassword(apiKey)
	assert.NoError(t, err)

	// Create test API key record
	userID := uuid.New().String()
	apiKeyID := uuid.New().String()
	scopes := []string{"shorten_url", "read_urls"}

	apiKeyRecord := models.APIKey{
		ID:        apiKeyID,
		UserID:    userID,
		HashedKey: hashedKey,
		Scopes:    scopes,
	}

	// Mock database query - Find all API keys
	// Scopes need to be JSON bytes for proper scanning
	scopesJSON, _ := json.Marshal(apiKeyRecord.Scopes)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "hashed_key", "scopes", "created_at", "updated_at"}).
		AddRow(apiKeyRecord.ID, apiKeyRecord.UserID, apiKeyRecord.HashedKey, scopesJSON, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "api_keys"`).
		WillReturnRows(rows)

	middleware := NewAPIKeyMiddleware(db)
	handler := middleware.OptionalAuth()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	c.Request = req

	handler(c)

	// Should set user_id and scopes
	userIDValue, exists := c.Get(constants.ContextKeyUserID)
	assert.True(t, exists)
	assert.Equal(t, userID, userIDValue)

	scopesValue, exists := c.Get(constants.ContextKeyScopes)
	assert.True(t, exists)
	assert.Equal(t, scopes, scopesValue)
}

func TestAPIKeyMiddleware_OptionalAuth_InvalidAPIKey(t *testing.T) {
	db, mock, sqlDB := setupTestDBForAPIKey(t)
	defer sqlDB.Close()

	// Generate a test API key
	apiKey, err := utils.GenerateAPIKey()
	assert.NoError(t, err)

	// Mock database query - return empty result
	rows := sqlmock.NewRows([]string{"id", "user_id", "hashed_key", "scopes", "created_at", "updated_at"})
	mock.ExpectQuery(`SELECT (.+) FROM "api_keys"`).
		WillReturnRows(rows)

	middleware := NewAPIKeyMiddleware(db)
	handler := middleware.OptionalAuth()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	c.Request = req

	handler(c)

	// Should not set user_id or scopes for invalid key
	_, exists := c.Get(constants.ContextKeyUserID)
	assert.False(t, exists)
}

func TestAPIKeyMiddleware_ValidateAPIKey_ValidKey(t *testing.T) {
	db, mock, sqlDB := setupTestDBForAPIKey(t)
	defer sqlDB.Close()

	// Generate a test API key
	apiKey, err := utils.GenerateAPIKey()
	assert.NoError(t, err)

	// Hash the key
	hashedKey, err := utils.HashPassword(apiKey)
	assert.NoError(t, err)

	userID := uuid.New().String()
	apiKeyID := uuid.New().String()
	scopes := []string{"shorten_url", "read_urls"}

	// Mock database query - Scopes need to be JSON bytes
	scopesJSON, _ := json.Marshal(scopes)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "hashed_key", "scopes", "created_at", "updated_at"}).
		AddRow(apiKeyID, userID, hashedKey, scopesJSON, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "api_keys"`).
		WillReturnRows(rows)

	middleware := NewAPIKeyMiddleware(db)

	returnedUserID, returnedScopes, err := middleware.ValidateAPIKey(apiKey)
	assert.NoError(t, err)
	assert.Equal(t, userID, returnedUserID)
	assert.Equal(t, scopes, returnedScopes)
}

func TestAPIKeyMiddleware_ValidateAPIKey_InvalidKey(t *testing.T) {
	db, mock, sqlDB := setupTestDBForAPIKey(t)
	defer sqlDB.Close()

	// Generate a test API key
	apiKey, err := utils.GenerateAPIKey()
	assert.NoError(t, err)

	// Mock database query - return empty result
	rows := sqlmock.NewRows([]string{"id", "user_id", "hashed_key", "scopes", "created_at", "updated_at"})
	mock.ExpectQuery(`SELECT (.+) FROM "api_keys"`).
		WillReturnRows(rows)

	middleware := NewAPIKeyMiddleware(db)

	_, _, err = middleware.ValidateAPIKey(apiKey)
	assert.Error(t, err)
}

