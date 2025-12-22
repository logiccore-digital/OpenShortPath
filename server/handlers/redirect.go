package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"openshortpath/server/config"
	"openshortpath/server/models"
)

type RedirectHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewRedirectHandler(db *gorm.DB, cfg *config.Config) *RedirectHandler {
	return &RedirectHandler{
		db:  db,
		cfg: cfg,
	}
}

func (h *RedirectHandler) Redirect(c *gin.Context) {
	// Extract hostname from request
	hostname := c.Request.Host

	// Validate domain
	if !isValidDomain(hostname, h.cfg.AvailableShortDomains) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Domain not found",
		})
		return
	}

	// Extract slug from path parameter
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Slug not found",
		})
		return
	}

	// Query database for ShortURL matching domain and slug
	var shortURL models.ShortURL
	result := h.db.Where("domain = ? AND slug = ?", hostname, slug).First(&shortURL)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Short URL not found",
			})
			return
		}
		// Database error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
			"details": result.Error.Error(),
		})
		return
	}

	// Return 301 redirect to target URL
	c.Redirect(http.StatusMovedPermanently, shortURL.URL)
}

