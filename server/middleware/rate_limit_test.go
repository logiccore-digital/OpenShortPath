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
)

func setupTestDBForRateLimit(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
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

func TestRateLimitMiddleware_AnonymousUser_UnderLimit(t *testing.T) {
	db, mock, sqlDB := setupTestDBForRateLimit(t)
	defer sqlDB.Close()

	clientIP := "192.168.1.1"
	windowStart := time.Now().UTC().Truncate(time.Hour)

	// Transaction begins
	mock.ExpectBegin()
	// No existing rate limit record
	mock.ExpectQuery(`SELECT (.+) FROM "rate_limits"`).
		WithArgs(clientIP, constants.RateLimitTypeIP, windowStart).
		WillReturnError(gorm.ErrRecordNotFound)

	// Create new record
	mock.ExpectExec(`INSERT INTO "rate_limits"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	middleware := RateLimitMiddleware(db)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.RemoteAddr = clientIP + ":54321"
	c.Request = req

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check rate limit headers are set even when not exceeded
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimitMiddleware_AnonymousUser_ExceedsLimit(t *testing.T) {
	db, mock, sqlDB := setupTestDBForRateLimit(t)
	defer sqlDB.Close()

	clientIP := "192.168.1.1"
	windowStart := time.Now().UTC().Truncate(time.Hour)
	rateLimitID := uuid.New().String()

	// Transaction begins
	mock.ExpectBegin()
	// Existing record with count = 5 (at limit)
	rows := sqlmock.NewRows([]string{"id", "identifier", "type", "request_count", "window_start", "created_at", "updated_at"}).
		AddRow(rateLimitID, clientIP, constants.RateLimitTypeIP, 5, windowStart, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM "rate_limits"`).
		WithArgs(clientIP, constants.RateLimitTypeIP, windowStart).
		WillReturnRows(rows)

	// Update record (increments to 6)
	mock.ExpectExec(`UPDATE "rate_limits"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	middleware := RateLimitMiddleware(db)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.RemoteAddr = clientIP + ":54321"
	c.Request = req

	middleware(c)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Check rate limit headers
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "Rate limit exceeded")
}

func TestRateLimitMiddleware_AuthenticatedUser_Hobbyist_UnderLimit(t *testing.T) {
	db, mock, sqlDB := setupTestDBForRateLimit(t)
	defer sqlDB.Close()

	userID := uuid.New().String()
	clientIP := "192.168.1.1"
	windowStart := time.Now().UTC().Truncate(time.Hour)

	// Get user plan
	userRows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "plan", "created_at", "updated_at"}).
		AddRow(userID, "testuser", nil, true, constants.PlanHobbyist, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs(userID).
		WillReturnRows(userRows)

	// Transaction begins
	mock.ExpectBegin()
	// No existing rate limit record
	mock.ExpectQuery(`SELECT (.+) FROM "rate_limits"`).
		WithArgs(userID, constants.RateLimitTypeUser, windowStart).
		WillReturnError(gorm.ErrRecordNotFound)

	// Create new record
	mock.ExpectExec(`INSERT INTO "rate_limits"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	middleware := RateLimitMiddleware(db)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.RemoteAddr = clientIP + ":54321"
	c.Request = req
	c.Set(constants.ContextKeyUserID, userID)

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check rate limit headers are set even when not exceeded
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimitMiddleware_AuthenticatedUser_Pro_Unlimited(t *testing.T) {
	db, mock, sqlDB := setupTestDBForRateLimit(t)
	defer sqlDB.Close()

	userID := uuid.New().String()
	clientIP := "192.168.1.1"

	// Get user plan
	userRows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "plan", "created_at", "updated_at"}).
		AddRow(userID, "testuser", nil, true, constants.PlanPro, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs(userID).
		WillReturnRows(userRows)

	// Pro plan has unlimited rate limit, so no database query for rate limits

	middleware := RateLimitMiddleware(db)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.RemoteAddr = clientIP + ":54321"
	c.Request = req
	c.Set(constants.ContextKeyUserID, userID)

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Pro plan has unlimited rate limit, so headers should not be set
	assert.Empty(t, w.Header().Get("X-RateLimit-Limit"))
}

func TestRateLimitMiddleware_AuthenticatedUser_VerifiedAccess_Unlimited(t *testing.T) {
	db, mock, sqlDB := setupTestDBForRateLimit(t)
	defer sqlDB.Close()

	userID := uuid.New().String()
	clientIP := "192.168.1.1"

	// Get user plan
	userRows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "plan", "created_at", "updated_at"}).
		AddRow(userID, "testuser", nil, true, constants.PlanVerifiedAccess, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs(userID).
		WillReturnRows(userRows)

	// Verified Access plan has unlimited rate limit, so no database query for rate limits

	middleware := RateLimitMiddleware(db)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.RemoteAddr = clientIP + ":54321"
	c.Request = req
	c.Set(constants.ContextKeyUserID, userID)

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verified Access plan has unlimited rate limit, so headers should not be set
	assert.Empty(t, w.Header().Get("X-RateLimit-Limit"))
}

func TestRateLimitMiddleware_AuthenticatedUser_UserNotFound(t *testing.T) {
	db, mock, sqlDB := setupTestDBForRateLimit(t)
	defer sqlDB.Close()

	userID := uuid.New().String()
	clientIP := "192.168.1.1"
	windowStart := time.Now().UTC().Truncate(time.Hour)

	// User not found - defaults to hobbyist
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs(userID).
		WillReturnError(gorm.ErrRecordNotFound)

	// Transaction begins
	mock.ExpectBegin()
	// No existing rate limit record
	mock.ExpectQuery(`SELECT (.+) FROM "rate_limits"`).
		WithArgs(userID, constants.RateLimitTypeUser, windowStart).
		WillReturnError(gorm.ErrRecordNotFound)

	// Create new record
	mock.ExpectExec(`INSERT INTO "rate_limits"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	middleware := RateLimitMiddleware(db)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.RemoteAddr = clientIP + ":54321"
	c.Request = req
	c.Set(constants.ContextKeyUserID, userID)

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check rate limit headers are set even when not exceeded
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimitMiddleware_XForwardedFor(t *testing.T) {
	db, mock, sqlDB := setupTestDBForRateLimit(t)
	defer sqlDB.Close()

	clientIP := "192.168.1.1"
	windowStart := time.Now().UTC().Truncate(time.Hour)

	// Transaction begins
	mock.ExpectBegin()
	// No existing rate limit record
	mock.ExpectQuery(`SELECT (.+) FROM "rate_limits"`).
		WithArgs(clientIP, constants.RateLimitTypeIP, windowStart).
		WillReturnError(gorm.ErrRecordNotFound)

	// Create new record
	mock.ExpectExec(`INSERT INTO "rate_limits"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	middleware := RateLimitMiddleware(db)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("X-Forwarded-For", clientIP)
	req.RemoteAddr = "10.0.0.1:54321"
	c.Request = req

	middleware(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check rate limit headers are set even when not exceeded
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimitMiddleware_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupTestDBForRateLimit(t)
	defer sqlDB.Close()

	clientIP := "192.168.1.1"
	windowStart := time.Now().UTC().Truncate(time.Hour)

	// Transaction begins
	mock.ExpectBegin()
	// Database error
	mock.ExpectQuery(`SELECT (.+) FROM "rate_limits"`).
		WithArgs(clientIP, constants.RateLimitTypeIP, windowStart).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	middleware := RateLimitMiddleware(db)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.RemoteAddr = clientIP + ":54321"
	c.Request = req

	middleware(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Rate limit check failed")
}
