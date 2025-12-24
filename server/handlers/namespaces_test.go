package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"openshortpath/server/config"
	"openshortpath/server/constants"
	"openshortpath/server/models"
)

func TestNamespacesHandler_CreateNamespace_Success(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()

	// Mock database queries
	// First: check for existing namespace (should return no rows)
	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs("example.com", "my-namespace").
		WillReturnError(gorm.ErrRecordNotFound)

	// Second: insert new namespace
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "namespaces"`).
		WithArgs(sqlmock.AnyArg(), "my-namespace", "example.com", userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)

	reqBody := `{"name": "my-namespace", "domain": "example.com"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/namespaces", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateNamespace(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Namespace
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "my-namespace", response.Name)
	assert.Equal(t, "example.com", response.Domain)
	assert.Equal(t, userID, response.UserID)
}

func TestNamespacesHandler_CreateNamespace_InvalidName(t *testing.T) {
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()

	testCases := []struct {
		name    string
		reqBody string
		errorContains string
	}{
		{"Uppercase letters", `{"name": "MyNamespace", "domain": "example.com"}`, "lowercase alphanumerical"},
		{"Special characters", `{"name": "my@namespace", "domain": "example.com"}`, "lowercase alphanumerical"},
		{"Spaces", `{"name": "my namespace", "domain": "example.com"}`, "lowercase alphanumerical"},
		{"Too long", `{"name": "` + strings.Repeat("a", 33) + `", "domain": "example.com"}`, "lowercase alphanumerical"},
		{"Reserved: api", `{"name": "api", "domain": "example.com"}`, "reserved"},
		{"Reserved: dashboard", `{"name": "dashboard", "domain": "example.com"}`, "reserved"},
		{"Reserved: admin", `{"name": "admin", "domain": "example.com"}`, "reserved"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set(constants.ContextKeyUserID, userID)

			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/namespaces", strings.NewReader(tc.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateNamespace(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response["error"], tc.errorContains)
		})
	}
}

func TestNamespacesHandler_CreateNamespace_ValidNameFormats(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()

	testCases := []struct {
		name    string
		reqBody string
	}{
		{"Lowercase with hyphens", `{"name": "my-namespace", "domain": "example.com"}`},
		{"Lowercase with underscores", `{"name": "my_namespace", "domain": "example.com"}`},
		{"Lowercase with numbers", `{"name": "namespace123", "domain": "example.com"}`},
		{"Mixed hyphens and underscores", `{"name": "my-namespace_123", "domain": "example.com"}`},
		{"32 characters max", `{"name": "` + strings.Repeat("a", 32) + `", "domain": "example.com"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock database queries
			mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
				WithArgs("example.com", sqlmock.AnyArg()).
				WillReturnError(gorm.ErrRecordNotFound)

			mock.ExpectBegin()
			mock.ExpectExec(`INSERT INTO "namespaces"`).
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "example.com", userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set(constants.ContextKeyUserID, userID)

			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/namespaces", strings.NewReader(tc.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateNamespace(c)

			assert.Equal(t, http.StatusCreated, w.Code)
		})
	}
}

func TestNamespacesHandler_CreateNamespace_InvalidDomain(t *testing.T) {
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)

	reqBody := `{"name": "my-namespace", "domain": "invalid-domain.com"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/namespaces", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateNamespace(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "not in the list of available short domains")
}

func TestNamespacesHandler_CreateNamespace_Duplicate(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()
	now := time.Now()

	// Mock database query - existing namespace found
	rows := sqlmock.NewRows([]string{"id", "name", "domain", "user_id", "created_at", "updated_at"}).
		AddRow(uuid.New().String(), "my-namespace", "example.com", "other-user", now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs("example.com", "my-namespace").
		WillReturnRows(rows)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)

	reqBody := `{"name": "my-namespace", "domain": "example.com"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/namespaces", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateNamespace(c)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "already exists")
}

func TestNamespacesHandler_CreateNamespace_NoUserID(t *testing.T) {
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Don't set user_id in context

	reqBody := `{"name": "my-namespace", "domain": "example.com"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/namespaces", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateNamespace(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNamespacesHandler_ListNamespaces_Success(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()
	now := time.Now()
	id1 := uuid.New().String()
	id2 := uuid.New().String()

	// Count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "namespaces"`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Select query
	rows := sqlmock.NewRows([]string{"id", "name", "domain", "user_id", "created_at", "updated_at"}).
		AddRow(id1, "namespace1", "example.com", userID, now, now).
		AddRow(id2, "namespace2", "example.com", userID, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(userID).
		WillReturnRows(rows)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/namespaces", nil)

	handler.ListNamespaces(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListNamespacesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Namespaces, 2)
	assert.Equal(t, int64(2), response.Total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNamespacesHandler_GetNamespace_Success(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()
	namespaceID := uuid.New().String()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "name", "domain", "user_id", "created_at", "updated_at"}).
		AddRow(namespaceID, "my-namespace", "example.com", userID, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(namespaceID, userID).
		WillReturnRows(rows)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: namespaceID}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/"+namespaceID, nil)

	handler.GetNamespace(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Namespace
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, namespaceID, response.ID)
	assert.Equal(t, "my-namespace", response.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNamespacesHandler_GetNamespace_NotFound(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()
	namespaceID := uuid.New().String()

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(namespaceID, userID).
		WillReturnError(gorm.ErrRecordNotFound)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: namespaceID}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/"+namespaceID, nil)

	handler.GetNamespace(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNamespacesHandler_UpdateNamespace_Success(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()
	namespaceID := uuid.New().String()
	now := time.Now()

	// Mock find query
	rows := sqlmock.NewRows([]string{"id", "name", "domain", "user_id", "created_at", "updated_at"}).
		AddRow(namespaceID, "old-name", "example.com", userID, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(namespaceID, userID).
		WillReturnRows(rows)

	// Mock conflict check (no conflict)
	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs("example.com", "new-name", namespaceID).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock update
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "namespaces"`).
		WithArgs("new-name", sqlmock.AnyArg(), namespaceID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Mock reload
	updatedRows := sqlmock.NewRows([]string{"id", "name", "domain", "user_id", "created_at", "updated_at"}).
		AddRow(namespaceID, "new-name", "example.com", userID, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(namespaceID, namespaceID).
		WillReturnRows(updatedRows)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: namespaceID}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/namespaces/"+namespaceID, strings.NewReader(`{"name": "new-name"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateNamespace(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Namespace
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "new-name", response.Name)
}

func TestNamespacesHandler_UpdateNamespace_InvalidName(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()
	namespaceID := uuid.New().String()
	now := time.Now()

	// Mock find query
	rows := sqlmock.NewRows([]string{"id", "name", "domain", "user_id", "created_at", "updated_at"}).
		AddRow(namespaceID, "old-name", "example.com", userID, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(namespaceID, userID).
		WillReturnRows(rows)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: namespaceID}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/namespaces/"+namespaceID, strings.NewReader(`{"name": "InvalidName"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateNamespace(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "lowercase alphanumerical")
}

func TestNamespacesHandler_DeleteNamespace_Success(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()
	namespaceID := uuid.New().String()
	now := time.Now()

	// Mock find query
	rows := sqlmock.NewRows([]string{"id", "name", "domain", "user_id", "created_at", "updated_at"}).
		AddRow(namespaceID, "my-namespace", "example.com", userID, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(namespaceID, userID).
		WillReturnRows(rows)

	// Mock delete short URLs
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "short_urls"`).
		WithArgs(namespaceID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	// Mock delete namespace
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "namespaces"`).
		WithArgs(namespaceID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: namespaceID}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces/"+namespaceID, nil)

	handler.DeleteNamespace(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestNamespacesHandler_DeleteNamespace_WithShortURLs(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()
	namespaceID := uuid.New().String()
	now := time.Now()

	// Mock find query
	rows := sqlmock.NewRows([]string{"id", "name", "domain", "user_id", "created_at", "updated_at"}).
		AddRow(namespaceID, "my-namespace", "example.com", userID, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(namespaceID, userID).
		WillReturnRows(rows)

	// Mock delete short URLs (2 URLs deleted)
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "short_urls"`).
		WithArgs(namespaceID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	// Mock delete namespace
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "namespaces"`).
		WithArgs(namespaceID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: namespaceID}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces/"+namespaceID, nil)

	handler.DeleteNamespace(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestNamespacesHandler_DeleteNamespace_NotFound(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewNamespacesHandler(db, cfg)
	userID := uuid.New().String()
	namespaceID := uuid.New().String()

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(namespaceID, userID).
		WillReturnError(gorm.ErrRecordNotFound)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: namespaceID}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces/"+namespaceID, nil)

	handler.DeleteNamespace(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShortenHandler_WithNamespace_Success(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortenHandler(db, cfg)
	userID := uuid.New().String()
	namespaceID := uuid.New().String()
	now := time.Now()

	// Mock check for existing short URL (should return no rows) - this happens first
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
	// First, the transaction begins
	mock.ExpectBegin()
	// Then query for existing monthly limit record (should return no rows)
	mock.ExpectQuery(`SELECT (.+) FROM "monthly_link_limits"`).
		WithArgs(userID, "user", sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)
	// Then insert new monthly link limit record
	mock.ExpectExec(`INSERT INTO "monthly_link_limits"`).
		WithArgs(sqlmock.AnyArg(), userID, "user", 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Transaction commits
	mock.ExpectCommit()

	// Mock namespace ownership check
	namespaceRows := sqlmock.NewRows([]string{"id", "name", "domain", "user_id", "created_at", "updated_at"}).
		AddRow(namespaceID, "my-namespace", "example.com", userID, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(namespaceID, userID).
		WillReturnRows(namespaceRows)

	// Mock insert short URL
	// GORM order: id, domain, slug, url, user_id, namespace_id, created_at, updated_at
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "short_urls"`).
		WithArgs(sqlmock.AnyArg(), "example.com", sqlmock.AnyArg(), "https://example.com/target", userID, namespaceID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)

	reqBody := `{"domain": "example.com", "url": "https://example.com/target", "namespace_id": "` + namespaceID + `"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Shorten(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.ShortURL
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response.NamespaceID)
	assert.Equal(t, namespaceID, *response.NamespaceID)
}

func TestShortenHandler_WithNamespace_Unauthorized(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortenHandler(db, cfg)
	namespaceID := uuid.New().String()

	// Mock check for existing short URL (should return no rows) - this happens first
	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs("example.com", sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock monthly link limit check for anonymous user (IP-based)
	// This happens before namespace check, but since user is anonymous, it will check IP limit
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "monthly_link_limits"`).
		WithArgs(sqlmock.AnyArg(), "ip", sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec(`INSERT INTO "monthly_link_limits"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "ip", 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Don't set user_id in context

	reqBody := `{"domain": "example.com", "url": "https://example.com/target", "namespace_id": "` + namespaceID + `"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Shorten(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Authentication required")
}

func TestShortenHandler_WithNamespace_NotFound(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortenHandler(db, cfg)
	userID := uuid.New().String()
	namespaceID := uuid.New().String()
	now := time.Now()

	// Mock check for existing short URL (should return no rows) - this happens first
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

	// Mock namespace ownership check - not found
	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(namespaceID, userID).
		WillReturnError(gorm.ErrRecordNotFound)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)

	reqBody := `{"domain": "example.com", "url": "https://example.com/target", "namespace_id": "` + namespaceID + `"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Shorten(c)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "not found or you do not have permission")
}

func TestShortURLsHandler_Update_WithNamespace_Success(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortURLsHandler(db, cfg)
	userID := uuid.New().String()
	shortURLID := uuid.New().String()
	namespaceID := uuid.New().String()
	now := time.Now()

	// Mock find short URL
	rows := sqlmock.NewRows([]string{"id", "domain", "slug", "url", "user_id", "namespace_id", "created_at", "updated_at"}).
		AddRow(shortURLID, "example.com", "slug1", "https://example.com", userID, nil, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs(shortURLID, userID).
		WillReturnRows(rows)

	// Mock namespace ownership check
	namespaceRows := sqlmock.NewRows([]string{"id", "name", "domain", "user_id", "created_at", "updated_at"}).
		AddRow(namespaceID, "my-namespace", "example.com", userID, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(namespaceID, userID).
		WillReturnRows(namespaceRows)

	// Mock update
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "short_urls"`).
		WithArgs(namespaceID, sqlmock.AnyArg(), shortURLID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Mock reload
	updatedRows := sqlmock.NewRows([]string{"id", "domain", "slug", "url", "user_id", "namespace_id", "created_at", "updated_at"}).
		AddRow(shortURLID, "example.com", "slug1", "https://example.com", userID, namespaceID, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs(shortURLID, shortURLID).
		WillReturnRows(updatedRows)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: shortURLID}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/short-urls/"+shortURLID, strings.NewReader(`{"namespace_id": "`+namespaceID+`"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ShortURL
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response.NamespaceID)
	assert.Equal(t, namespaceID, *response.NamespaceID)
}

func TestShortURLsHandler_Update_RemoveNamespace(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortURLsHandler(db, cfg)
	userID := uuid.New().String()
	shortURLID := uuid.New().String()
	namespaceID := uuid.New().String()
	now := time.Now()

	// Mock find short URL (with namespace)
	rows := sqlmock.NewRows([]string{"id", "domain", "slug", "url", "user_id", "namespace_id", "created_at", "updated_at"}).
		AddRow(shortURLID, "example.com", "slug1", "https://example.com", userID, namespaceID, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs(shortURLID, userID).
		WillReturnRows(rows)

	// Mock update (set namespace_id to NULL)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "short_urls"`).
		WithArgs(nil, sqlmock.AnyArg(), shortURLID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Mock reload
	updatedRows := sqlmock.NewRows([]string{"id", "domain", "slug", "url", "user_id", "namespace_id", "created_at", "updated_at"}).
		AddRow(shortURLID, "example.com", "slug1", "https://example.com", userID, nil, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs(shortURLID, shortURLID).
		WillReturnRows(updatedRows)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: shortURLID}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/short-urls/"+shortURLID, strings.NewReader(`{"namespace_id": ""}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ShortURL
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Nil(t, response.NamespaceID)
}

func TestShortURLsHandler_Update_WithNamespace_NotFound(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.Config{
		AvailableShortDomains: []string{"example.com", "localhost:3000"},
	}

	handler := NewShortURLsHandler(db, cfg)
	userID := uuid.New().String()
	shortURLID := uuid.New().String()
	namespaceID := uuid.New().String()
	now := time.Now()

	// Mock find short URL
	rows := sqlmock.NewRows([]string{"id", "domain", "slug", "url", "user_id", "namespace_id", "created_at", "updated_at"}).
		AddRow(shortURLID, "example.com", "slug1", "https://example.com", userID, nil, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "short_urls"`).
		WithArgs(shortURLID, userID).
		WillReturnRows(rows)

	// Mock namespace ownership check - not found
	mock.ExpectQuery(`SELECT (.+) FROM "namespaces"`).
		WithArgs(namespaceID, userID).
		WillReturnError(gorm.ErrRecordNotFound)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: shortURLID}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/short-urls/"+shortURLID, strings.NewReader(`{"namespace_id": "`+namespaceID+`"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Update(c)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "not found or you do not have permission")
}
