package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"openshortpath/server/config"
	"openshortpath/server/constants"
)

func TestShortenHandler_Shorten_Success(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortenHandler(db, cfg)

	// Mock database queries
	// First query: check for existing record (should return no rows)
	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock monthly link limit check for anonymous user (IP-based)
	// GetClientIP will return RemoteAddr, which we need to mock
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "monthly_link_limits"`).
		WithArgs(sqlmock.AnyArg(), "ip", sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec(`INSERT INTO "monthly_link_limits"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "ip", 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Second query: insert new record
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "short_urls"`).
		WithArgs(sqlmock.AnyArg(), "example.com", sqlmock.AnyArg(), "https://example.com/target", "", nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"domain": "example.com", "url": "https://example.com/target"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Shorten(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "example.com", response["domain"])
	assert.Equal(t, "https://example.com/target", response["url"])
	assert.NotEmpty(t, response["slug"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShortenHandler_Shorten_WithJWTToken(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortenHandler(db, cfg)
	userID := "user123"
	now := time.Now()

	// Mock database queries
	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock user query to get plan (for monthly limit check)
	userRows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "plan", "created_at", "updated_at"}).
		AddRow(userID, "testuser", nil, true, "hobbyist", now, now)
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs(userID).
		WillReturnRows(userRows)

	// Mock monthly link limit check transaction
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "monthly_link_limits"`).
		WithArgs(userID, "user", sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec(`INSERT INTO "monthly_link_limits"`).
		WithArgs(sqlmock.AnyArg(), userID, "user", 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "short_urls"`).
		WithArgs(sqlmock.AnyArg(), "example.com", sqlmock.AnyArg(), "https://example.com/target", userID, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Setup Gin context with user_id from JWT
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)

	reqBody := `{"domain": "example.com", "url": "https://example.com/target"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Shorten(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "user123", response["user_id"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShortenHandler_Shorten_WithCustomSlug(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortenHandler(db, cfg)

	// Mock database queries
	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", "custom-slug").
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock monthly link limit check for anonymous user (IP-based)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "monthly_link_limits"`).
		WithArgs(sqlmock.AnyArg(), "ip", sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec(`INSERT INTO "monthly_link_limits"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "ip", 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "short_urls"`).
		WithArgs(sqlmock.AnyArg(), "example.com", "custom-slug", "https://example.com/target", "", nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"domain": "example.com", "url": "https://example.com/target", "slug": "custom-slug"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Shorten(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "custom-slug", response["slug"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShortenHandler_Shorten_InvalidRequestBody(t *testing.T) {
	// Setup
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortenHandler(db, cfg)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"invalid": "json"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Shorten(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid request body")
}

func TestShortenHandler_Shorten_InvalidDomain(t *testing.T) {
	// Setup
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortenHandler(db, cfg)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"domain": "invalid-domain.com", "url": "https://example.com/target"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Shorten(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "not in the list of available short domains")
}

func TestShortenHandler_Shorten_DuplicateSlug(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortenHandler(db, cfg)

	// Mock database query - existing record found
	now := time.Now()
	rows := sqlmock.NewRows([]string{"domain", "slug", "url", "user_id", "created_at", "updated_at"}).
		AddRow("example.com", "existing", "https://example.com/old", "", now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", "existing").
		WillReturnRows(rows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"domain": "example.com", "url": "https://example.com/target", "slug": "existing"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Shorten(c)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "already exists")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShortenHandler_Shorten_DatabaseErrorOnCheck(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortenHandler(db, cfg)

	// Mock database query returning an error
	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"domain": "example.com", "url": "https://example.com/target"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Shorten(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShortenHandler_Shorten_DatabaseErrorOnInsert(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortenHandler(db, cfg)

	// Mock database queries
	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock monthly link limit check for anonymous user (IP-based)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "monthly_link_limits"`).
		WithArgs(sqlmock.AnyArg(), "ip", sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec(`INSERT INTO "monthly_link_limits"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "ip", 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "short_urls"`).
		WithArgs(sqlmock.AnyArg(), "example.com", sqlmock.AnyArg(), "https://example.com/target", "", nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"domain": "example.com", "url": "https://example.com/target"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Shorten(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to create short URL")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShortenHandler_Shorten_MissingRequiredFields(t *testing.T) {
	// Setup
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortenHandler(db, cfg)

	// Test missing domain
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"url": "https://example.com/target"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Shorten(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test missing URL
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)

	reqBody2 := `{"domain": "example.com"}`
	c2.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody2))
	c2.Request.Header.Set("Content-Type", "application/json")

	handler.Shorten(c2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}
