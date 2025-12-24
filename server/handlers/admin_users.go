package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"openshortpath/server/models"
	"openshortpath/server/utils"
)

type AdminUsersHandler struct {
	db *gorm.DB
}

func NewAdminUsersHandler(db *gorm.DB) *AdminUsersHandler {
	return &AdminUsersHandler{
		db: db,
	}
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Active   *bool  `json:"active,omitempty"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	Active   *bool   `json:"active,omitempty"`
}

// UserResponse represents a user in API responses (without password hash)
type UserResponse struct {
	UserID            string `json:"user_id"`
	Username          string `json:"username,omitempty"`
	Active            bool   `json:"active"`
	Plan              string `json:"plan,omitempty"`
	MonthlyLinkLimit  int    `json:"monthly_link_limit,omitempty"`
	MonthlyLinksUsed  int    `json:"monthly_links_used,omitempty"`
	MonthlyLinkReset  string `json:"monthly_link_reset,omitempty"`
	RateLimitPerHour  int    `json:"rate_limit_per_hour,omitempty"`
	RateLimitRemaining int   `json:"rate_limit_remaining,omitempty"`
	RateLimitReset     string `json:"rate_limit_reset,omitempty"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// ListUsersResponse represents the paginated list response
type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// CreateUser handles POST /api/v1/__admin/users
func (h *AdminUsersHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
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

	// Set active status (default to true if not provided)
	active := true
	if req.Active != nil {
		active = *req.Active
	}

	// Create user
	user := models.User{
		UserID:         userID,
		Username:       &req.Username,
		HashedPassword: &hashedPassword,
		Active:         active,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	// Return created user (without password hash)
	response := UserResponse{
		UserID:    user.UserID,
		Username:  *user.Username,
		Active:    user.Active,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusCreated, response)
}

// ListUsers handles GET /api/v1/__admin/users
func (h *AdminUsersHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	limit := 50

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > 100 {
				limit = 100
			} else {
				limit = l
			}
		}
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get total count
	var total int64
	if err := h.db.Model(&models.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": err.Error(),
		})
		return
	}

	// Get paginated users
	var users []models.User
	if err := h.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": err.Error(),
		})
		return
	}

	// Convert to response format
	userResponses := make([]UserResponse, 0, len(users))
	for _, user := range users {
		response := UserResponse{
			UserID:    user.UserID,
			Active:    user.Active,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if user.Username != nil {
			response.Username = *user.Username
		}
		userResponses = append(userResponses, response)
	}

	response := ListUsersResponse{
		Users: userResponses,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUser handles PUT /api/v1/__admin/users/:user_id
func (h *AdminUsersHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Find user
	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": err.Error(),
		})
		return
	}

	// Update username if provided
	if req.Username != nil {
		// Check if new username already exists (excluding current user)
		var existingUser models.User
		result := h.db.Where("username = ? AND user_id != ?", *req.Username, userID).First(&existingUser)
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
		user.Username = req.Username
	}

	// Update password if provided
	if req.Password != nil {
		hashedPassword, err := utils.HashPassword(*req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to hash password",
				"details": err.Error(),
			})
			return
		}
		user.HashedPassword = &hashedPassword
	}

	// Update active status if provided
	if req.Active != nil {
		user.Active = *req.Active
	}

	// Save updates
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update user",
			"details": err.Error(),
		})
		return
	}

	// Return updated user (without password hash)
	response := UserResponse{
		UserID:    user.UserID,
		Active:    user.Active,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if user.Username != nil {
		response.Username = *user.Username
	}

	c.JSON(http.StatusOK, response)
}

// DeleteUser handles DELETE /api/v1/__admin/users/:user_id
func (h *AdminUsersHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	// Check if user exists
	var user models.User
	if err := h.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": err.Error(),
		})
		return
	}

	// Delete user (hard delete)
	if err := h.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete user",
			"details": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

