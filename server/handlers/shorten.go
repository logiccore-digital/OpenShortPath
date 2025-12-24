package handlers

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"openshortpath/server/config"
	"openshortpath/server/constants"
	"openshortpath/server/models"
	"openshortpath/server/services"
)

type ShortenHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

type ShortenRequest struct {
	Domain      string  `json:"domain" binding:"required"`
	URL         string  `json:"url" binding:"required"`
	Slug        string  `json:"slug,omitempty"`
	NamespaceID *string `json:"namespace_id,omitempty"`
}

func NewShortenHandler(db *gorm.DB, cfg *config.Config) *ShortenHandler {
	return &ShortenHandler{
		db:  db,
		cfg: cfg,
	}
}

// generateRandomSlug generates a random 5-character alphanumeric string
func generateRandomSlug() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 5

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[bytes[i]%byte(len(charset))]
	}

	return string(result), nil
}

// isValidDomain checks if the domain exists in the available short domains list
func isValidDomain(domain string, availableDomains []string) bool {
	for _, availableDomain := range availableDomains {
		if domain == availableDomain {
			return true
		}
	}
	return false
}

func (h *ShortenHandler) Shorten(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate domain
	if !isValidDomain(req.Domain, h.cfg.AvailableShortDomains) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Domain '%s' is not in the list of available short domains", req.Domain),
		})
		return
	}

	// Generate slug if not provided
	slug := req.Slug
	if slug == "" {
		var err error
		slug, err = generateRandomSlug()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate random slug",
			})
			return
		}
	}

	// Check for duplicate (domain, slug) combination
	var existing models.ShortURL
	result := h.db.Where("domain = ? AND slug = ?", req.Domain, slug).First(&existing)
	if result.Error == nil {
		// Record exists
		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("Short URL with domain '%s' and slug '%s' already exists", req.Domain, slug),
		})
		return
	}
	if result.Error != gorm.ErrRecordNotFound {
		// Database error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database error",
			"details": result.Error.Error(),
		})
		return
	}

	// Get user ID from context if available (from JWT token)
	userID := ""
	if userIDValue, exists := c.Get(constants.ContextKeyUserID); exists {
		if userIDStr, ok := userIDValue.(string); ok {
			userID = userIDStr
		}
	}

	// Check monthly link limit before creating the link
	clientIP := services.GetClientIP(c)
	var monthlyLimitInfo *services.MonthlyLinkLimitInfo
	var err error
	var limitType string
	var identifier string
	var limitPerMonth int

	if userID != "" {
		// Authenticated user - check user-level monthly limit
		plan, planErr := services.GetUserPlan(h.db, userID)
		if planErr != nil {
			// If we can't get the plan, default to hobbyist limit
			plan = constants.PlanHobbyist
		}

		limitPerMonth = services.GetMonthlyLinkLimitForPlan(plan)
		limitType = constants.RateLimitTypeUser
		identifier = userID

		// Check monthly link limit
		monthlyLimitInfo, err = services.CheckMonthlyLinkLimit(h.db, identifier, limitType, limitPerMonth)
	} else {
		// Anonymous user - check IP-level monthly limit
		limitPerMonth = 1000 // Anonymous users: 1,000 links per month per IP
		limitType = constants.RateLimitTypeIP
		identifier = clientIP

		// Check monthly link limit
		monthlyLimitInfo, err = services.CheckMonthlyLinkLimit(h.db, identifier, limitType, limitPerMonth)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Monthly link limit check failed",
			"details": err.Error(),
		})
		return
	}

	// Set monthly link limit headers
	if monthlyLimitInfo.Limit > 0 {
		c.Header("X-Monthly-Link-Limit", strconv.Itoa(monthlyLimitInfo.Limit))
		if monthlyLimitInfo.Remaining >= 0 {
			c.Header("X-Monthly-Link-Remaining", strconv.Itoa(monthlyLimitInfo.Remaining))
		}
		if !monthlyLimitInfo.Reset.IsZero() {
			c.Header("X-Monthly-Link-Reset", strconv.FormatInt(monthlyLimitInfo.Reset.Unix(), 10))
		}
	}

	if monthlyLimitInfo.Exceeded {
		resetTimeStr := "the start of next month"
		if !monthlyLimitInfo.Reset.IsZero() {
			resetTimeStr = monthlyLimitInfo.Reset.Format("2006-01-02 15:04:05 UTC")
		}
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": fmt.Sprintf("Monthly link limit exceeded. Limit: %d links per month. Reset time: %s", monthlyLimitInfo.Limit, resetTimeStr),
		})
		return
	}

	// Validate namespace ownership if namespace_id is provided
	if req.NamespaceID != nil && *req.NamespaceID != "" {
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required to use namespace",
			})
			return
		}

		var namespace models.Namespace
		result := h.db.Where("id = ? AND user_id = ?", *req.NamespaceID, userID).First(&namespace)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Namespace not found or you do not have permission to use it",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Database error",
				"details": result.Error.Error(),
			})
			return
		}
	}

	// Generate UUID for ID field
	id := uuid.New().String()

	// Create new ShortURL record
	shortURL := models.ShortURL{
		ID:          id,
		Domain:      req.Domain,
		Slug:        slug,
		URL:         req.URL,
		UserID:      userID,
		NamespaceID: req.NamespaceID,
	}

	if err := h.db.Create(&shortURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create short URL",
			"details": err.Error(),
		})
		return
	}

	// Return the full ShortURL object
	c.JSON(http.StatusCreated, shortURL)
}
