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
	"time"

	"openshortpath/server/config"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	cfg         *config.Config
	dashboardFS fs.FS
}

func NewDashboardHandler(cfg *config.Config, dashboardFS fs.FS) *DashboardHandler {
	return &DashboardHandler{
		cfg:         cfg,
		dashboardFS: dashboardFS,
	}
}

// ServeDashboard handles dashboard requests
// If DashboardDevServerURL is set, it proxies to the dev server
// Otherwise, it serves embedded static files
func (h *DashboardHandler) ServeDashboard(c *gin.Context) {
	// If dev server URL is configured, proxy to it
	if h.cfg.DashboardDevServerURL != "" {
		h.proxyToDevServer(c)
		return
	}

	// Otherwise, serve embedded files
	h.serveEmbeddedFiles(c)
}

// proxyToDevServer proxies requests to the development server
func (h *DashboardHandler) proxyToDevServer(c *gin.Context) {
	targetURL, err := url.Parse(h.cfg.DashboardDevServerURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid dashboard dev server URL",
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

		// Vite dev server with base: '/dashboard' expects paths to start with /dashboard
		// So we keep the /dashboard prefix when proxying
		// This way /dashboard -> Vite's /dashboard, /dashboard/@vite/client -> Vite's /dashboard/@vite/client
		req.URL.Path = originalPath
		req.URL.RawPath = originalPath

		// Set host to target host
		req.Host = targetURL.Host
	}

	// Modify the response to rewrite redirects
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Rewrite Location header in redirects
		// Since Vite already uses /dashboard as base, redirects from Vite will already have /dashboard
		// We just need to ensure they're relative to our server, not Vite's
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			if location := resp.Header.Get("Location"); location != "" {
				if parsed, err := url.Parse(location); err == nil {
					// Only rewrite if it's pointing to the same host (Vite dev server)
					if parsed.Host == "" || parsed.Host == targetURL.Host {
						// If the path doesn't already start with /dashboard, add it
						// But if Vite redirects /dashboard -> /dashboard/, we should keep it as is
						if !strings.HasPrefix(parsed.Path, "/dashboard") {
							if strings.HasPrefix(parsed.Path, "/") {
								parsed.Path = "/dashboard" + parsed.Path
							} else if parsed.Path != "" {
								parsed.Path = "/dashboard/" + parsed.Path
							} else {
								parsed.Path = "/dashboard"
							}
						}
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
		rw.Write([]byte("Bad Gateway: Unable to connect to dashboard dev server"))
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// serveEmbeddedFiles serves the embedded dashboard files
func (h *DashboardHandler) serveEmbeddedFiles(c *gin.Context) {
	// Get the file system from embed
	fsys, err := fs.Sub(h.dashboardFS, "dashboard-dist")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to access embedded files",
		})
		return
	}

	// Remove /dashboard prefix from path
	path := strings.TrimPrefix(c.Request.URL.Path, "/dashboard")
	if path == "" || path == "/" {
		path = "/index.html"
	}

	// Clean the path
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "index.html"
	}

	// Try to serve the file
	file, err := fsys.Open(path)
	if err != nil {
		// If file doesn't exist, try to serve index.html for SPA routing
		indexFile, indexErr := fsys.Open("index.html")
		if indexErr != nil {
			// If index.html doesn't exist either, check if we only have placeholder files
			// In that case, embedded files aren't available (dev mode should use proxy)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Dashboard files not embedded. Use dashboard_dev_server_url for development or build with 'make server'",
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

	// Set content type based on file extension
	ext := filepath.Ext(path)
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
	default:
		c.Header("Content-Type", "application/octet-stream")
	}

	// Serve the file
	c.Status(http.StatusOK)
	io.Copy(c.Writer, file)
}

// fileInfoModTime is a helper to get mod time from a file
func fileInfoModTime(file fs.File) time.Time {
	if info, err := file.Stat(); err == nil {
		return info.ModTime()
	}
	return time.Now()
}
