package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"openshortpath/server/config"
	"openshortpath/server/models"
)

// ReservedNamespaceNames contains namespace names that cannot be used
var ReservedNamespaceNames = map[string]bool{
	"api":        true,
	"dashboard":  true,
	"admin":      true,
	"www":        true,
	"mail":       true,
	"ftp":        true,
	"localhost":  true,
	"test":       true,
	"dev":        true,
	"staging":    true,
	"prod":       true,
	"production": true,
}

// IsReservedNamespaceName checks if a namespace name is in the reserved list
func IsReservedNamespaceName(name string) bool {
	return ReservedNamespaceNames[strings.ToLower(name)]
}

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

// Redirect handles redirects for both namespace and non-namespace URLs
// It checks the path to determine if it's /:slug or /:namespace/:slug
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

	// Parse the path to determine if we have a namespace or not
	path := c.Request.URL.Path
	
	// Skip dashboard and API routes - these should be handled by other routes
	if strings.HasPrefix(path, "/dashboard") || strings.HasPrefix(path, "/api") {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
		})
		return
	}
	
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	// If we have 2 parts, it's namespace/slug
	// If we have 1 part, it's just slug
	if len(pathParts) == 2 {
		// Handle namespace/slug pattern
		namespaceName := pathParts[0]
		slug := pathParts[1]

		if namespaceName == "" || slug == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Namespace or slug not found",
			})
			return
		}

		// First, find the namespace by name and domain
		var namespace models.Namespace
		namespaceResult := h.db.Where("domain = ? AND name = ?", hostname, namespaceName).First(&namespace)
		if namespaceResult.Error != nil {
			if namespaceResult.Error == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Namespace not found",
				})
				return
			}
			// Database error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
				"details": namespaceResult.Error.Error(),
			})
			return
		}

		// Query database for ShortURL matching domain, namespace_id, and slug
		var shortURL models.ShortURL
		result := h.db.Where("domain = ? AND namespace_id = ? AND slug = ?", hostname, namespace.ID, slug).First(&shortURL)
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
		return
	} else if len(pathParts) == 1 {
		// Handle single slug pattern (no namespace)
		slug := pathParts[0]
		if slug == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Slug not found",
			})
			return
		}

		// Query database for ShortURL matching domain and slug (without namespace)
		var shortURL models.ShortURL
		result := h.db.Where("domain = ? AND slug = ? AND namespace_id IS NULL", hostname, slug).First(&shortURL)
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
		return
	}

	// Invalid path format
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Invalid path format",
	})
}

