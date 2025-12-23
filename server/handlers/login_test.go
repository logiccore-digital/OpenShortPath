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

func TestLoginHandler_Login_Success(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	// Hash password
	hashedPassword, err := utils.HashPassword("test-password")
	assert.NoError(t, err)

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewLoginHandler(db, cfg)

	// Mock database query
	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "created_at", "updated_at"}).
		AddRow("user123", "testuser", hashedPassword, true, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("testuser").
		WillReturnRows(rows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "testuser", "password": "test-password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Login(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["token"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginHandler_Login_InvalidCredentials_UserNotFound(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewLoginHandler(db, cfg)

	// Mock database query - user not found
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("nonexistent").
		WillReturnError(gorm.ErrRecordNotFound)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "nonexistent", "password": "password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Login(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginHandler_Login_InvalidCredentials_WrongPassword(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	// Hash password
	hashedPassword, err := utils.HashPassword("correct-password")
	assert.NoError(t, err)

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewLoginHandler(db, cfg)

	// Mock database query
	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "created_at", "updated_at"}).
		AddRow("user123", "testuser", hashedPassword, true, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("testuser").
		WillReturnRows(rows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "testuser", "password": "wrong-password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Login(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginHandler_Login_InactiveUser(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	hashedPassword, err := utils.HashPassword("test-password")
	assert.NoError(t, err)

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewLoginHandler(db, cfg)

	// Mock database query - inactive user
	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "created_at", "updated_at"}).
		AddRow("user123", "testuser", hashedPassword, false, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("testuser").
		WillReturnRows(rows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "testuser", "password": "test-password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Login(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Account is inactive", response["error"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginHandler_Login_NoPassword(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewLoginHandler(db, cfg)

	// Mock database query - user with no password (nil)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "created_at", "updated_at"}).
		AddRow("user123", "testuser", nil, true, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("testuser").
		WillReturnRows(rows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "testuser", "password": "test-password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Login(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginHandler_Login_EmptyPassword(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewLoginHandler(db, cfg)

	// Mock database query - user with empty password
	now := time.Now()
	emptyPassword := ""
	rows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "created_at", "updated_at"}).
		AddRow("user123", "testuser", emptyPassword, true, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("testuser").
		WillReturnRows(rows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "testuser", "password": "test-password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Login(c)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginHandler_Login_InvalidRequestBody(t *testing.T) {
	// Setup
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewLoginHandler(db, cfg)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"invalid": "json"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Login(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid request body")
}

func TestLoginHandler_Login_MissingFields(t *testing.T) {
	// Setup
	db, _, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewLoginHandler(db, cfg)

	// Test missing username
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"password": "password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test missing password
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)

	reqBody2 := `{"username": "user"}`
	c2.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(reqBody2))
	c2.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestLoginHandler_Login_DatabaseError(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	handler := NewLoginHandler(db, cfg)

	// Mock database query returning an error
	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("testuser").
		WillReturnError(sql.ErrConnDone)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "testuser", "password": "password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Login(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Database error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginHandler_Login_RS256(t *testing.T) {
	// Setup
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	hashedPassword, err := utils.HashPassword("test-password")
	assert.NoError(t, err)

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

	handler := NewLoginHandler(db, cfg)

	// Mock database query
	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "hashed_password", "active", "created_at", "updated_at"}).
		AddRow("user123", "testuser", hashedPassword, true, now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "users"`).
		WithArgs("testuser").
		WillReturnRows(rows)

	// Setup Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"username": "testuser", "password": "test-password"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.Login(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["token"])
	assert.NoError(t, mock.ExpectationsWereMet())
}
