package handlers

import (
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"

	"openshortpath/server/config"

	"github.com/gin-gonic/gin"
)

type LandingHandler struct {
	cfg       *config.Config
	landingFS fs.FS
}

func NewLandingHandler(cfg *config.Config, landingFS fs.FS) *LandingHandler {
	return &LandingHandler{
		cfg:       cfg,
		landingFS: landingFS,
	}
}

// ServeLanding handles landing page requests
// If LandingDevServerURL is set, it proxies to the dev server
// Otherwise, it serves embedded static files
func (h *LandingHandler) ServeLanding(c *gin.Context) {
	// If dev server URL is configured, proxy to it
	if h.cfg.LandingDevServerURL != "" {
		h.proxyToDevServer(c)
		return
	}

	// Otherwise, serve embedded files
	h.serveEmbeddedFiles(c)
}

// proxyToDevServer proxies requests to the development server
func (h *LandingHandler) proxyToDevServer(c *gin.Context) {
	targetURL, err := url.Parse(h.cfg.LandingDevServerURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid landing dev server URL",
		})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Get the original path from the request
		originalPath := req.URL.Path

		// Next.js dev server expects paths as-is
		req.URL.Path = originalPath
		req.URL.RawPath = originalPath

		// Set host to target host
		req.Host = targetURL.Host
	}

	// Modify the response to rewrite redirects
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Rewrite Location header in redirects
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			if location := resp.Header.Get("Location"); location != "" {
				if parsed, err := url.Parse(location); err == nil {
					// Only rewrite if it's pointing to the same host (Next.js dev server)
					if parsed.Host == "" || parsed.Host == targetURL.Host {
						// Preserve the path as-is for Next.js
						// Preserve query and fragment
						resp.Header.Set("Location", parsed.String())
					}
				}
			}
		}

		return nil
	}

	// Handle errors
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte("Bad Gateway: Unable to connect to landing dev server"))
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// serveEmbeddedFiles serves the embedded landing page files
func (h *LandingHandler) serveEmbeddedFiles(c *gin.Context) {
	// Safety check: don't serve landing page for API or dashboard routes
	requestPath := c.Request.URL.Path
	if strings.HasPrefix(requestPath, "/api/") || strings.HasPrefix(requestPath, "/dashboard/") {
		c.Status(http.StatusNotFound)
		return
	}

	// Get the file system from embed
	fsys, err := fs.Sub(h.landingFS, "landing-dist")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to access embedded files",
		})
		return
	}

	// Get the path from the request
	path := requestPath
	if path == "" || path == "/" {
		path = "/index.html"
	}

	// Clean the path - remove leading slash for filesystem access
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "index.html"
	}

	// Try to serve the file - Next.js static export creates files with .html extension
	// Try multiple variations:
	// 1. Exact path (for static assets like _next/static/*)
	// 2. Path with .html extension (for routes like docs/privacy-policy.html)
	// 3. Path/index.html (for routes like docs/privacy-policy/index.html)
	var file fs.File
	var actualPath string // Track which path actually worked
	
	// Check if this is a static asset (has a file extension)
	hasExtension := filepath.Ext(path) != ""
	
	// First try exact path (for static assets and HTML files)
	file, err = fsys.Open(path)
	if err == nil {
		// Check if it's actually a directory (directories can be "opened" but aren't readable as files)
		fileInfo, statErr := file.Stat()
		if statErr == nil && fileInfo.IsDir() {
			// It's a directory, close it and try other paths
			file.Close()
			file = nil
			err = fs.ErrNotExist // Treat as not found so we try other paths
		} else {
			actualPath = path
		}
	}
	
	if err != nil {
		// For non-HTML paths without extension, try with .html extension (Next.js static export format)
		if !hasExtension && !strings.HasSuffix(path, ".html") {
			htmlPath := path + ".html"
			file, err = fsys.Open(htmlPath)
			if err == nil {
				// Double-check it's not a directory
				fileInfo, statErr := file.Stat()
				if statErr == nil && fileInfo.IsDir() {
					file.Close()
					file = nil
					err = fs.ErrNotExist
				} else {
					actualPath = htmlPath
				}
			}
		}
	}
	
	
	if err != nil {
		// Try path/index.html (Next.js might create directories for routes)
		// Only for paths without extensions (HTML routes)
		if !hasExtension {
			indexPath := strings.TrimSuffix(path, "/") + "/index.html"
			file, err = fsys.Open(indexPath)
			if err == nil {
				actualPath = indexPath
			}
		}
	}
	
	if err != nil {
		// If it's a static asset (has extension), return 404 - don't fall back to index.html
		if hasExtension {
			c.Status(http.StatusNotFound)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			return
		}
		
		// For HTML routes without extension, try to serve index.html for SPA routing
		indexFile, indexErr := fsys.Open("index.html")
		if indexErr != nil {
			// If index.html doesn't exist either, check if we only have placeholder files
			// In that case, embedded files aren't available (dev mode should use proxy)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Landing page files not embedded. Use landing_dev_server_url for development or build with 'make server'",
			})
			return
		}
		defer indexFile.Close()

		// Set content type
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusOK)
		io.Copy(c.Writer, indexFile)
		return
	}
	defer file.Close()

	// Set content type based on the actual file path that was found, not the original request path
	// If actualPath is empty (shouldn't happen), fall back to original path
	contentTypePath := actualPath
	if contentTypePath == "" {
		contentTypePath = path
	}
	ext := filepath.Ext(contentTypePath)
	
	// If no extension found and it's not a static asset request, assume HTML
	if ext == "" && !hasExtension {
		ext = ".html"
	}
	
	switch ext {
	case ".html":
		c.Header("Content-Type", "text/html; charset=utf-8")
	case ".js":
		c.Header("Content-Type", "application/javascript; charset=utf-8")
	case ".css":
		c.Header("Content-Type", "text/css; charset=utf-8")
	case ".json":
		c.Header("Content-Type", "application/json")
	case ".png":
		c.Header("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		c.Header("Content-Type", "image/jpeg")
	case ".svg":
		c.Header("Content-Type", "image/svg+xml")
	case ".woff":
		c.Header("Content-Type", "font/woff")
	case ".woff2":
		c.Header("Content-Type", "font/woff2")
	case ".ico":
		c.Header("Content-Type", "image/x-icon")
	case ".webp":
		c.Header("Content-Type", "image/webp")
	default:
		c.Header("Content-Type", "application/octet-stream")
	}

	// Serve the file
	c.Status(http.StatusOK)
	io.Copy(c.Writer, file)
}

