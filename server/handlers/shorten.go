package handlers

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"openshortpath/server/config"
	"openshortpath/server/constants"
	"openshortpath/server/models"
)

type ShortenHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

type ShortenRequest struct {
	Domain string `json:"domain" binding:"required"`
	URL    string `json:"url" binding:"required"`
	Slug   string `json:"slug,omitempty"`
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

	// Generate UUID for ID field
	id := uuid.New().String()

	// Create new ShortURL record
	shortURL := models.ShortURL{
		ID:     id,
		Domain: req.Domain,
		Slug:   slug,
		URL:    req.URL,
		UserID: userID,
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
