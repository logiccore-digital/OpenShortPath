package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"openshortpath/server/config"
	"openshortpath/server/models"
	"openshortpath/server/services"
	"openshortpath/server/utils"
)

type LoginHandler struct {
	db       *gorm.DB
	jwtConfig *config.JWT
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func NewLoginHandler(db *gorm.DB, jwtConfig *config.JWT) *LoginHandler {
	return &LoginHandler{
		db:        db,
		jwtConfig: jwtConfig,
	}
}

func (h *LoginHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Find user by username
	var user models.User
	result := h.db.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid credentials",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
		return
	}

	// Check if user is active
	if !user.Active {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Account is inactive",
		})
		return
	}

	// Check if user has a password (required for local auth)
	if user.HashedPassword == nil || *user.HashedPassword == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Verify password
	valid, err := utils.VerifyPassword(req.Password, *user.HashedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to verify password",
			"details": err.Error(),
		})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
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
	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
	})
}

