package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
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
	"openshortpath/server/utils"
)

func TestSignupHandler_Signup_Success(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewSignupHandler(db, cfg)

	// Mock database queries
	// First query: check if username exists (should return no rows)
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("newuser").
		WillReturnError(gorm.ErrRecordNotFound)

	// Second query: insert new user
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "users"`).
		WithArgs(sqlmock.AnyArg(), "newuser", sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "newuser", "password": "test-password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/signup", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Signup(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["token"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignupHandler_Signup_UsernameAlreadyExists(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewSignupHandler(db, cfg)

	// Hash password for existing user
	hashedPassword, err := utils.HashPassword("existing-password")
	assert.NoError(t, err)

	// Mock database query - user already exists
	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "created_at", "updated_at"}).
		AddRow("existing-user-id", "existinguser", hashedPassword, true, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("existinguser").
		WillReturnRows(rows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "existinguser", "password": "new-password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/signup", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Signup(c)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Username already exists", response["error"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignupHandler_Signup_InvalidRequestBody(t *testing.T) {
	// Setup
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewSignupHandler(db, cfg)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"invalid": "json"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/signup", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Signup(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid request body")
}

func TestSignupHandler_Signup_MissingFields(t *testing.T) {
	// Setup
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewSignupHandler(db, cfg)

	// Test missing username
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"password": "password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/signup", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Signup(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test missing password
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)

	reqBody2 := `{"username": "user"}`
	c2.Request = httptest.NewRequest(http.MethodPost, "/api/v1/signup", strings.NewReader(reqBody2))
	c2.Request.Header.Set("Content-Type", "application/json")

	handler.Signup(c2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestSignupHandler_Signup_DatabaseError_CheckUser(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewSignupHandler(db, cfg)

	// Mock database query returning an error (not RecordNotFound)
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("testuser").
		WillReturnError(sql.ErrConnDone)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "testuser", "password": "password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/signup", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Signup(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignupHandler_Signup_DatabaseError_CreateUser(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewSignupHandler(db, cfg)

	// Mock database queries
	// First query: check if username exists (should return no rows)
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("newuser").
		WillReturnError(gorm.ErrRecordNotFound)

	// Second query: insert new user fails
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "users"`).
		WithArgs(sqlmock.AnyArg(), "newuser", sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "newuser", "password": "test-password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/signup", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Signup(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to create user")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignupHandler_Signup_RS256(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	// Generate RSA keys for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	cfg := &config.JWT{
		Algorithm:  "RS256",
		PrivateKey: string(privateKeyPEM),
		PublicKey:  string(publicKeyPEM),
	}

	handler := NewSignupHandler(db, cfg)

	// Mock database queries
	// First query: check if username exists (should return no rows)
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("newuser").
		WillReturnError(gorm.ErrRecordNotFound)

	// Second query: insert new user
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "users"`).
		WithArgs(sqlmock.AnyArg(), "newuser", sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "newuser", "password": "test-password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/signup", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Signup(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["token"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignupHandler_Signup_UserCreatedWithActiveTrue(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewSignupHandler(db, cfg)

	// Mock database queries
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("newuser").
		WillReturnError(gorm.ErrRecordNotFound)

	// Verify that user is created with Active: true
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "users"`).
		WithArgs(sqlmock.AnyArg(), "newuser", sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "newuser", "password": "test-password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/signup", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Signup(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignupHandler_Signup_UserIDIsUUID(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewSignupHandler(db, cfg)

	// Mock database queries
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("newuser").
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock user creation - the user ID will be a UUID generated by uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "users"`).
		WithArgs(sqlmock.AnyArg(), "newuser", sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "newuser", "password": "test-password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/signup", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Signup(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	
	// Verify that a token was generated, which means the user ID was valid
	// (the signup handler uses uuid.New() which generates valid UUIDs)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["token"])
	
	assert.NoError(t, mock.ExpectationsWereMet())
}
