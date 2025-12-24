package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"openshortpath/server/constants"
	"openshortpath/server/models"
	"openshortpath/server/services"
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

	// Get user plan (default to hobbyist if not set)
	plan := user.Plan
	if plan == "" {
		plan = constants.PlanHobbyist
	}

	// Get monthly link limit for the plan
	monthlyLinkLimit := services.GetMonthlyLinkLimitForPlan(plan)

	// Get current monthly link usage
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	var monthlyLimit models.MonthlyLinkLimit
	monthlyLinksUsed := 0
	var monthlyLinkReset time.Time

	// Calculate reset time (start of next month)
	if now.Month() == 12 {
		monthlyLinkReset = time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		monthlyLinkReset = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	}

	// Query monthly link limit record for current month
	monthlyLimitResult := h.db.Where("identifier = ? AND type = ? AND month_start = ?", userID, constants.RateLimitTypeUser, monthStart).First(&monthlyLimit)
	if monthlyLimitResult.Error == nil {
		monthlyLinksUsed = monthlyLimit.LinkCount
	}

	// Get rate limit information for the user's plan
	rateLimitPerHour := services.GetRateLimitForPlan(plan)
	rateLimitInfo, err := services.GetRateLimitInfo(h.db, userID, constants.RateLimitTypeUser, rateLimitPerHour)
	if err != nil {
		// Log error but don't fail the request - rate limit info is optional
		// In production, you might want to log this properly
		rateLimitInfo = &services.RateLimitInfo{
			Limit:     rateLimitPerHour,
			Remaining: -1,
			Reset:     time.Time{},
			Exceeded:  false,
		}
	}

	// Build response (without password hash)
	response := UserResponse{
		UserID:             user.UserID,
		Active:             user.Active,
		Plan:               plan,
		MonthlyLinkLimit:   monthlyLinkLimit,
		MonthlyLinksUsed:   monthlyLinksUsed,
		RateLimitPerHour:   rateLimitInfo.Limit,
		RateLimitRemaining: rateLimitInfo.Remaining,
		CreatedAt:          user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if user.Username != nil {
		response.Username = *user.Username
	}
	if !monthlyLinkReset.IsZero() {
		response.MonthlyLinkReset = monthlyLinkReset.Format("2006-01-02T15:04:05Z07:00")
	}
	if !rateLimitInfo.Reset.IsZero() {
		response.RateLimitReset = rateLimitInfo.Reset.Format("2006-01-02T15:04:05Z07:00")
	}

	c.JSON(http.StatusOK, response)
}
