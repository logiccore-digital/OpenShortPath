package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"openshortpath/server/config"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm db: %v", err)
	}

	return gormDB, mock, sqlDB
}

func TestRedirectHandler_Redirect_Success(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewRedirectHandler(db, cfg)

	// Mock database query
	now := time.Now()
	rows := sqlmock.NewRows([]string{"domain", "slug", "url", "user_id", "created_at", "updated_at"}).
		AddRow("example.com", "abc123", "https://example.com/target", "", now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", "abc123").
		WillReturnRows(rows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/abc123", nil)
	c.Request.Host = "example.com"
	c.Params = gin.Params{gin.Param{Key: "slug", Value: "abc123"}}

	// Execute
	handler.Redirect(c)

	// Assert
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://example.com/target", w.Header().Get("Location"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedirectHandler_Redirect_InvalidDomain(t *testing.T) {
	// Setup
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewRedirectHandler(db, cfg)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/abc123", nil)
	c.Request.Host = "invalid-domain.com"
	c.Params = gin.Params{gin.Param{Key: "slug", Value: "abc123"}}

	// Execute
	handler.Redirect(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRedirectHandler_Redirect_EmptySlug(t *testing.T) {
	// Setup
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewRedirectHandler(db, cfg)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Host = "example.com"
	c.Params = gin.Params{gin.Param{Key: "slug", Value: ""}}

	// Execute
	handler.Redirect(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRedirectHandler_Redirect_SlugNotFound(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewRedirectHandler(db, cfg)

	// Mock database query returning no rows
	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", "nonexistent").
		WillReturnError(gorm.ErrRecordNotFound)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	c.Request.Host = "example.com"
	c.Params = gin.Params{gin.Param{Key: "slug", Value: "nonexistent"}}

	// Execute
	handler.Redirect(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedirectHandler_Redirect_DatabaseError(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewRedirectHandler(db, cfg)

	// Mock database query returning an error
	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", "abc123").
		WillReturnError(sql.ErrConnDone)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/abc123", nil)
	c.Request.Host = "example.com"
	c.Params = gin.Params{gin.Param{Key: "slug", Value: "abc123"}}

	// Execute
	handler.Redirect(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedirectHandler_Redirect_LocalhostDomain(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewRedirectHandler(db, cfg)

	// Mock database query
	now := time.Now()
	rows := sqlmock.NewRows([]string{"domain", "slug", "url", "user_id", "created_at", "updated_at"}).
		AddRow("localhost:3000", "test123", "https://example.com/target", "", now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("localhost:3000", "test123").
		WillReturnRows(rows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test123", nil)
	c.Request.Host = "localhost:3000"
	c.Params = gin.Params{gin.Param{Key: "slug", Value: "test123"}}

	// Execute
	handler.Redirect(c)

	// Assert
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://example.com/target", w.Header().Get("Location"))
	assert.NoError(t, mock.ExpectationsWereMet())
}
