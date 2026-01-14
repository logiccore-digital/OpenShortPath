package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"openshortpath/server/config"
	"openshortpath/server/models"
	"openshortpath/server/services"
	"openshortpath/server/utils"
)

type SignupHandler struct {
	db       *gorm.DB
	jwtConfig *config.JWT
}

type SignupRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignupResponse struct {
	Token string `json:"token"`
}

func NewSignupHandler(db *gorm.DB, jwtConfig *config.JWT) *SignupHandler {
	return &SignupHandler{
		db:        db,
		jwtConfig: jwtConfig,
	}
}

func (h *SignupHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Check if username already exists
	var existingUser models.User
	result := h.db.Where("username = ?", req.Username).First(&existingUser)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Username already exists",
		})
		return
	}
	if result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
		return
	}

	// Generate unique user ID
	userID := uuid.New().String()

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to hash password",
			"details": err.Error(),
		})
		return
	}

	// Create user with Active: true by default
	user := models.User{
		UserID:         userID,
		Username:       &req.Username,
		HashedPassword: &hashedPassword,
		Active:         true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	// Generate JWT token
	token, err := services.SignToken(user.UserID, h.jwtConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate token",
			"details": err.Error(),
		})
		return
	}

	// Return token
	c.JSON(http.StatusCreated, SignupResponse{
		Token: token,
	})
}
