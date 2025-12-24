package services

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

	"openshortpath/server/constants"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
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

func TestGetClientIP_XForwardedFor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
	c.Request = req

	ip := GetClientIP(c)
	assert.Equal(t, "192.168.1.1", ip)
}

func TestGetClientIP_XRealIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "10.0.0.2")
	c.Request = req

	ip := GetClientIP(c)
	assert.Equal(t, "10.0.0.2", ip)
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:54321"
	c.Request = req

	ip := GetClientIP(c)
	assert.Equal(t, "192.168.1.100", ip)
}

func TestGetClientIP_XForwardedForTakesPrecedence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	req.Header.Set("X-Real-IP", "10.0.0.2")
	req.RemoteAddr = "172.16.0.1:12345"
	c.Request = req

	ip := GetClientIP(c)
	assert.Equal(t, "192.168.1.1", ip)
}

func TestGetClientIP_XForwardedForMultipleIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", " 192.168.1.1 , 10.0.0.1, 172.16.0.1 ")
	c.Request = req

	ip := GetClientIP(c)
	assert.Equal(t, "192.168.1.1", ip) // Should take first IP after trimming
}

func TestGetClientIP_RemoteAddrNoPort(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100" // No port
	c.Request = req

	ip := GetClientIP(c)
	assert.Equal(t, "192.168.1.100", ip)
}

func TestGetUserPlan_ExistingUser(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	userID := uuid.New().String()
	plan := constants.PlanPro

	rows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "plan", "created_at", "updated_at"}).
		AddRow(userID, "testuser", nil, true, plan, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs(userID).
		WillReturnRows(rows)

	resultPlan, err := GetUserPlan(db, userID)
	assert.NoError(t, err)
	assert.Equal(t, plan, resultPlan)
}

func TestGetUserPlan_UserNotFound(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	userID := uuid.New().String()

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs(userID).
		WillReturnError(gorm.ErrRecordNotFound)

	resultPlan, err := GetUserPlan(db, userID)
	assert.NoError(t, err)
	assert.Equal(t, constants.PlanHobbyist, resultPlan) // Should default to hobbyist
}

func TestGetUserPlan_EmptyPlan(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	userID := uuid.New().String()

	rows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "plan", "created_at", "updated_at"}).
		AddRow(userID, "testuser", nil, true, "", time.Now(), time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs(userID).
		WillReturnRows(rows)

	resultPlan, err := GetUserPlan(db, userID)
	assert.NoError(t, err)
	assert.Equal(t, constants.PlanHobbyist, resultPlan) // Should default to hobbyist
}

func TestGetUserPlan_EmptyUserID(t *testing.T) {
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	_, err := GetUserPlan(db, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is empty")
}

func TestGetRateLimitForPlan_Hobbyist(t *testing.T) {
	limit := GetRateLimitForPlan(constants.PlanHobbyist)
	assert.Equal(t, 5, limit)
}

func TestGetRateLimitForPlan_EmptyPlan(t *testing.T) {
	limit := GetRateLimitForPlan("")
	assert.Equal(t, 5, limit) // Should default to hobbyist limit
}

func TestGetRateLimitForPlan_VerifiedAccess(t *testing.T) {
	limit := GetRateLimitForPlan(constants.PlanVerifiedAccess)
	assert.Equal(t, 0, limit) // Unlimited
}

func TestGetRateLimitForPlan_Pro(t *testing.T) {
	limit := GetRateLimitForPlan(constants.PlanPro)
	assert.Equal(t, 0, limit) // Unlimited
}

func TestGetRateLimitForPlan_UnknownPlan(t *testing.T) {
	limit := GetRateLimitForPlan("unknown_plan")
	assert.Equal(t, 5, limit) // Should default to hobbyist limit
}

func TestCheckRateLimit_Unlimited(t *testing.T) {
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	// Unlimited plans should not check database
	info, err := CheckRateLimit(db, "test-id", constants.RateLimitTypeUser, 0)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.False(t, info.Exceeded)
	assert.Equal(t, 0, info.Limit)
	assert.Equal(t, -1, info.Remaining) // -1 indicates unlimited
}

func TestCheckRateLimit_NewRecord(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	identifier := "test-identifier"
	limitType := constants.RateLimitTypeIP
	limitPerHour := 5
	windowStart := time.Now().UTC().Truncate(time.Hour)

	// Transaction begins
	mock.ExpectBegin()
	// First query - no existing record
	mock.ExpectQuery(`SELECT (.+) FROM "rate_limits"`).
		WithArgs(identifier, limitType, windowStart).
		WillReturnError(gorm.ErrRecordNotFound)

	// Insert new record - GORM uses Exec for INSERT, not Query
	mock.ExpectExec(`INSERT INTO "rate_limits"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	info, err := CheckRateLimit(db, identifier, limitType, limitPerHour)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.False(t, info.Exceeded) // First request should not exceed limit
	assert.Equal(t, limitPerHour, info.Limit)
	assert.Equal(t, limitPerHour-1, info.Remaining) // 1 request made, so 4 remaining
	assert.False(t, info.Reset.IsZero())
}

func TestCheckRateLimit_ExistingRecord_UnderLimit(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	identifier := "test-identifier"
	limitType := constants.RateLimitTypeUser
	limitPerHour := 5
	windowStart := time.Now().UTC().Truncate(time.Hour)
	rateLimitID := uuid.New().String()

	// Transaction begins
	mock.ExpectBegin()
	// Query existing record
	rows := sqlmock.NewRows([]string{"id", "identifier", "type", "request_count", "window_start", "created_at", "updated_at"}).
		AddRow(rateLimitID, identifier, limitType, 3, windowStart, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM "rate_limits"`).
		WithArgs(identifier, limitType, windowStart).
		WillReturnRows(rows)

	// Update record
	mock.ExpectExec(`UPDATE "rate_limits"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	info, err := CheckRateLimit(db, identifier, limitType, limitPerHour)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.False(t, info.Exceeded) // 4 requests < 5 limit
	assert.Equal(t, limitPerHour, info.Limit)
	assert.Equal(t, 1, info.Remaining) // 4 requests made, so 1 remaining
	assert.False(t, info.Reset.IsZero())
}

func TestCheckRateLimit_ExistingRecord_ExceedsLimit(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	identifier := "test-identifier"
	limitType := constants.RateLimitTypeIP
	limitPerHour := 5
	windowStart := time.Now().UTC().Truncate(time.Hour)
	rateLimitID := uuid.New().String()

	// Transaction begins
	mock.ExpectBegin()
	// Query existing record with count = 5 (at limit)
	rows := sqlmock.NewRows([]string{"id", "identifier", "type", "request_count", "window_start", "created_at", "updated_at"}).
		AddRow(rateLimitID, identifier, limitType, 5, windowStart, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM "rate_limits"`).
		WithArgs(identifier, limitType, windowStart).
		WillReturnRows(rows)

	// Update record (increments to 6)
	mock.ExpectExec(`UPDATE "rate_limits"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	info, err := CheckRateLimit(db, identifier, limitType, limitPerHour)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.True(t, info.Exceeded) // 6 requests > 5 limit
	assert.Equal(t, limitPerHour, info.Limit)
	assert.Equal(t, 0, info.Remaining) // Limit exceeded, so 0 remaining
	assert.False(t, info.Reset.IsZero())
}

func TestCleanupOldRateLimits(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	// GORM wraps Delete in a transaction
	mock.ExpectBegin()
	// Use sqlmock.AnyArg() to handle time precision differences
	mock.ExpectExec(`DELETE FROM "rate_limits" WHERE window_start < \$1`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 10)) // 10 rows deleted
	mock.ExpectCommit()

	err := CleanupOldRateLimits(db, 2*time.Hour)
	assert.NoError(t, err)
}
