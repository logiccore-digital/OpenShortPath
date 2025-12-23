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

	"openshortpath/server/constants"
)

func TestAPIKeysHandler_CreateAPIKey_Success(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	handler := NewAPIKeysHandler(db)
	userID := uuid.New().String()

	// Mock database insert - GORM uses Exec for INSERT
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "api_keys"`).
		WithArgs(sqlmock.AnyArg(), userID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)

	reqBody := `{"scopes": ["shorten_url", "read_urls"]}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/api-keys", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateAPIKey(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response CreateAPIKeyResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	assert.NotEmpty(t, response.Key)
	assert.True(t, strings.HasPrefix(response.Key, "osp_sk_"))
	assert.Equal(t, []string{"shorten_url", "read_urls"}, response.Scopes)
	assert.NotEmpty(t, response.CreatedAt)
}

func TestAPIKeysHandler_CreateAPIKey_InvalidScopes(t *testing.T) {
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	handler := NewAPIKeysHandler(db)
	userID := uuid.New().String()

	testCases := []struct {
		name      string
		reqBody   string
		expectErr bool
	}{
		{"Invalid scope", `{"scopes": ["invalid_scope"]}`, true},
		{"Empty scopes", `{"scopes": []}`, false}, // Empty scopes should be allowed
		{"Missing scopes", `{}`, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set(constants.ContextKeyUserID, userID)

			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/api-keys", strings.NewReader(tc.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateAPIKey(c)

			if tc.expectErr {
				assert.NotEqual(t, http.StatusCreated, w.Code)
			}
		})
	}
}

func TestAPIKeysHandler_CreateAPIKey_NoUserID(t *testing.T) {
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	handler := NewAPIKeysHandler(db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Don't set user_id in context

	reqBody := `{"scopes": ["shorten_url"]}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/api-keys", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateAPIKey(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeysHandler_ListAPIKeys_Success(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	handler := NewAPIKeysHandler(db)
	userID := uuid.New().String()

	// Mock database query - Scopes need to be JSON bytes for proper scanning
	scopes1JSON, _ := json.Marshal([]string{"shorten_url"})
	scopes2JSON, _ := json.Marshal([]string{"read_urls", "write_urls"})
	now1 := time.Now()
	now2 := time.Now().Add(time.Hour)
	rows := sqlmock.NewRows([]string{"id", "user_id", "hashed_key", "scopes", "created_at", "updated_at"}).
		AddRow(uuid.New().String(), userID, "hashed1", scopes1JSON, now1, now1).
		AddRow(uuid.New().String(), userID, "hashed2", scopes2JSON, now2, now2)

	mock.ExpectQuery(`SELECT (.+) FROM "api_keys"`).
		WithArgs(userID).
		WillReturnRows(rows)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/api-keys", nil)

	handler.ListAPIKeys(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListAPIKeysResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Keys, 2)
	assert.Equal(t, []string{"shorten_url"}, response.Keys[0].Scopes)
	assert.Equal(t, []string{"read_urls", "write_urls"}, response.Keys[1].Scopes)
	// Keys should not contain the actual key values (APIKeyListItem doesn't have Key field)
}

func TestAPIKeysHandler_ListAPIKeys_NoUserID(t *testing.T) {
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	handler := NewAPIKeysHandler(db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Don't set user_id in context
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/api-keys", nil)

	handler.ListAPIKeys(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeysHandler_ListAPIKeys_EmptyList(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	handler := NewAPIKeysHandler(db)
	userID := uuid.New().String()

	// Mock database query - return empty result
	rows := sqlmock.NewRows([]string{"id", "user_id", "hashed_key", "scopes", "created_at", "updated_at"})
	mock.ExpectQuery(`SELECT (.+) FROM "api_keys"`).
		WithArgs(userID).
		WillReturnRows(rows)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/api-keys", nil)

	handler.ListAPIKeys(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListAPIKeysResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Keys, 0)
}

func TestAPIKeysHandler_DeleteAPIKey_Success(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	handler := NewAPIKeysHandler(db)
	userID := uuid.New().String()
	apiKeyID := uuid.New().String()

	// Mock database query - find API key
	scopesJSON, _ := json.Marshal([]string{"shorten_url"})
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "hashed_key", "scopes", "created_at", "updated_at"}).
		AddRow(apiKeyID, userID, "hashed", scopesJSON, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "api_keys"`).
		WithArgs(apiKeyID, userID).
		WillReturnRows(rows)

	// Mock database delete
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "api_keys"`).
		WithArgs(apiKeyID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: apiKeyID}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/api-keys/"+apiKeyID, nil)

	handler.DeleteAPIKey(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAPIKeysHandler_DeleteAPIKey_NotFound(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	handler := NewAPIKeysHandler(db)
	userID := uuid.New().String()
	apiKeyID := uuid.New().String()

	// Mock database query - return no rows
	rows := sqlmock.NewRows([]string{"id", "user_id", "hashed_key", "scopes", "created_at", "updated_at"})
	mock.ExpectQuery(`SELECT (.+) FROM "api_keys"`).
		WithArgs(apiKeyID, userID).
		WillReturnRows(rows)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: apiKeyID}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/api-keys/"+apiKeyID, nil)

	handler.DeleteAPIKey(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPIKeysHandler_DeleteAPIKey_WrongUser(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	handler := NewAPIKeysHandler(db)
	userID := uuid.New().String()
	apiKeyID := uuid.New().String()

	// Mock database query - return no rows (key belongs to different user)
	rows := sqlmock.NewRows([]string{"id", "user_id", "hashed_key", "scopes", "created_at", "updated_at"})
	mock.ExpectQuery(`SELECT (.+) FROM "api_keys"`).
		WithArgs(apiKeyID, userID).
		WillReturnRows(rows)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: apiKeyID}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/api-keys/"+apiKeyID, nil)

	handler.DeleteAPIKey(c)

	// Should return 404 (not found) since user doesn't own the key
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPIKeysHandler_DeleteAPIKey_NoUserID(t *testing.T) {
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	handler := NewAPIKeysHandler(db)
	apiKeyID := uuid.New().String()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Don't set user_id in context
	c.Params = gin.Params{gin.Param{Key: "id", Value: apiKeyID}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/api-keys/"+apiKeyID, nil)

	handler.DeleteAPIKey(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeysHandler_CreateAPIKey_AllScopes(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	handler := NewAPIKeysHandler(db)
	userID := uuid.New().String()

	// Mock database insert - GORM uses Exec for INSERT
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "api_keys"`).
		WithArgs(sqlmock.AnyArg(), userID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(constants.ContextKeyUserID, userID)

	// Test with all valid scopes
	reqBody := `{"scopes": ["shorten_url", "read_urls", "write_urls"]}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/api-keys", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateAPIKey(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response CreateAPIKeyResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, []string{"shorten_url", "read_urls", "write_urls"}, response.Scopes)
	
	// Verify the key format
	assert.True(t, strings.HasPrefix(response.Key, "osp_sk_"))
}

