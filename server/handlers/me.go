package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"openshortpath/server/constants"
	"openshortpath/server/models"
)

type MeHandler struct {
	db *gorm.DB
}

func NewMeHandler(db *gorm.DB) *MeHandler {
	return &MeHandler{
		db: db,
	}
}

// GetMe handles GET /api/v1/me
// Returns the current authenticated user's details
func (h *MeHandler) GetMe(c *gin.Context) {
	// Get user ID from context (set by RequireAuth middleware)
	userIDValue, exists := c.Get(constants.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user ID in context",
		})
		return
	}

	// Find the user by user_id
	var user models.User
	result := h.db.Where("user_id = ?", userID).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
		return
	}

	// Build response (without password hash)
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

