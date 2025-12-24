package handlers

import (
	"database/sql"
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

	// Mock database query (without namespace)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "domain", "slug", "url", "user_id", "namespace_id", "created_at", "updated_at"}).
		AddRow(uuid.New().String(), "example.com", "abc123", "https://example.com/target", "", nil, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", "abc123").
		WillReturnRows(rows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/abc123", nil)
	c.Request.Host = "example.com"
	c.Request.URL.Path = "/abc123"

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
	c.Request.URL.Path = "/abc123"

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
	c.Request.URL.Path = "/"

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

	// Mock database query returning no rows (without namespace)
	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", "nonexistent").
		WillReturnError(gorm.ErrRecordNotFound)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	c.Request.Host = "example.com"
	c.Request.URL.Path = "/nonexistent"

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
	c.Request.URL.Path = "/abc123"

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

	// Mock database query (without namespace)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "domain", "slug", "url", "user_id", "namespace_id", "created_at", "updated_at"}).
		AddRow(uuid.New().String(), "localhost:3000", "test123", "https://example.com/target", "", nil, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("localhost:3000", "test123").
		WillReturnRows(rows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test123", nil)
	c.Request.Host = "localhost:3000"
	c.Request.URL.Path = "/test123"

	// Execute
	handler.Redirect(c)

	// Assert
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://example.com/target", w.Header().Get("Location"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedirectHandler_RedirectWithNamespace_Success(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewRedirectHandler(db, cfg)
	namespaceID := uuid.New().String()
	now := time.Now()

	// Mock namespace query
	namespaceRows := sqlmock.NewRows([]string{"id", "name", "domain", "user_id", "created_at", "updated_at"}).
		AddRow(namespaceID, "my-namespace", "example.com", uuid.New().String(), now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs("example.com", "my-namespace").
		WillReturnRows(namespaceRows)

	// Mock short URL query
	shortURLRows := sqlmock.NewRows([]string{"id", "domain", "slug", "url", "user_id", "namespace_id", "created_at", "updated_at"}).
		AddRow(uuid.New().String(), "example.com", "abc123", "https://example.com/target", "", namespaceID, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", namespaceID, "abc123").
		WillReturnRows(shortURLRows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/my-namespace/abc123", nil)
	c.Request.Host = "example.com"
	c.Request.URL.Path = "/my-namespace/abc123"

	// Execute
	handler.Redirect(c)

	// Assert
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://example.com/target", w.Header().Get("Location"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedirectHandler_RedirectWithNamespace_NamespaceNotFound(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewRedirectHandler(db, cfg)

	// Mock namespace query returning no rows
	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs("example.com", "nonexistent-namespace").
		WillReturnError(gorm.ErrRecordNotFound)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/nonexistent-namespace/abc123", nil)
	c.Request.Host = "example.com"
	c.Request.URL.Path = "/nonexistent-namespace/abc123"

	// Execute
	handler.Redirect(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedirectHandler_RedirectWithNamespace_ShortURLNotFound(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewRedirectHandler(db, cfg)
	namespaceID := uuid.New().String()
	now := time.Now()

	// Mock namespace query
	namespaceRows := sqlmock.NewRows([]string{"id", "name", "domain", "user_id", "created_at", "updated_at"}).
		AddRow(namespaceID, "my-namespace", "example.com", uuid.New().String(), now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs("example.com", "my-namespace").
		WillReturnRows(namespaceRows)

	// Mock short URL query returning no rows
	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", namespaceID, "nonexistent-slug").
		WillReturnError(gorm.ErrRecordNotFound)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/my-namespace/nonexistent-slug", nil)
	c.Request.Host = "example.com"
	c.Request.URL.Path = "/my-namespace/nonexistent-slug"

	// Execute
	handler.Redirect(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
