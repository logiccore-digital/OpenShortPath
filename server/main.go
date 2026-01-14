package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"openshortpath/server/config"
	"openshortpath/server/handlers"
	"openshortpath/server/middleware"
	"openshortpath/server/models"
)

//go:embed dashboard-dist
var dashboardFS embed.FS

//go:embed landing-dist/*
var landingFS embed.FS

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to configuration file (YAML)")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	var db *gorm.DB
	if cfg.PostgresURI != "" {
		// Use Postgres if postgres_uri is provided
		db, err = gorm.Open(postgres.Open(cfg.PostgresURI), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to Postgres database: %v", err)
		}
		log.Printf("Connected to Postgres database")
	} else {
		// Use SQLite
		sqlitePath := cfg.SQLitePath
		if sqlitePath == "" {
			sqlitePath = "db.sqlite"
		}
		db, err = gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to SQLite database: %v", err)
		}
		log.Printf("Connected to SQLite database: %s", sqlitePath)
	}

	// Auto-migrate database models
	if err := db.AutoMigrate(&models.ShortURL{}, &models.User{}, &models.APIKey{}, &models.Namespace{}, &models.RateLimit{}, &models.MonthlyLinkLimit{}); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	r := gin.Default()

	// Initialize JWT middleware if JWT config is provided
	var jwtMiddleware *middleware.JWTMiddleware
	var apiKeyMiddleware *middleware.APIKeyMiddleware
	if cfg.JWT != nil {
		jwtMiddleware = middleware.NewJWTMiddleware(cfg.JWT)
		apiKeyMiddleware = middleware.NewAPIKeyMiddleware(db)
		jwtMiddleware.SetAPIKeyMiddleware(apiKeyMiddleware)
		r.Use(jwtMiddleware.OptionalAuth())
		log.Printf("JWT authentication enabled (algorithm: %s)", cfg.JWT.Algorithm)
		log.Printf("API key authentication enabled")
	}

	// Create API v1 route group
	// Rate limiting middleware runs after OptionalAuth so user context is available
	apiV1 := r.Group("/api/v1")

	// Initialize handlers with database
	shortenHandler := handlers.NewShortenHandler(db, cfg)
	redirectHandler := handlers.NewRedirectHandler(db, cfg)
	authProviderHandler := handlers.NewAuthProviderHandler(cfg)
	domainsHandler := handlers.NewDomainsHandler(cfg)

	// Register API routes first (highest priority)
	// Shorten endpoint - authentication is optional (handled by OptionalAuth middleware)
	// Rate limiting is applied only to the shorten endpoint per IP for anonymous users, per user for authenticated users
	apiV1.POST("/shorten", middleware.RateLimitMiddleware(db), shortenHandler.Shorten)
	log.Printf("Rate limiting enabled for /api/v1/shorten endpoint")

	// Public endpoints without rate limiting
	apiV1.GET("/auth-provider", authProviderHandler.GetAuthProvider)
	apiV1.GET("/domains", domainsHandler.GetDomains)

	// Register login endpoint only if auth_provider is "local"
	if cfg.AuthProvider == "local" {
		loginHandler := handlers.NewLoginHandler(db, cfg.JWT)
		apiV1.POST("/login", loginHandler.Login)
		log.Printf("Login endpoint enabled at /api/v1/login")

		// Register signup endpoint only if signup is enabled
		if cfg.EnableSignup {
			signupHandler := handlers.NewSignupHandler(db, cfg.JWT)
			apiV1.POST("/signup", signupHandler.Signup)
			log.Printf("Signup endpoint enabled at /api/v1/signup")
		}
	}

	// Register admin endpoints if admin password is configured
	if cfg.AdminPassword != "" {
		adminMiddleware := middleware.NewAdminMiddleware(cfg.AdminPassword)
		adminUsersHandler := handlers.NewAdminUsersHandler(db)

		// Create admin route group with authentication middleware
		// Note: Admin routes are under /api/v1 so they inherit rate limiting
		adminRoutes := apiV1.Group("/__admin")
		adminRoutes.Use(adminMiddleware.RequireAdmin())

		// Register admin user management routes
		adminRoutes.POST("/users", adminUsersHandler.CreateUser)
		adminRoutes.GET("/users", adminUsersHandler.ListUsers)
		adminRoutes.PUT("/users/:user_id", adminUsersHandler.UpdateUser)
		adminRoutes.DELETE("/users/:user_id", adminUsersHandler.DeleteUser)

		log.Printf("Admin endpoints enabled at /api/v1/__admin/*")
	}

	// Register short URL management endpoints if JWT config is provided
	if cfg.JWT != nil {
		shortURLsHandler := handlers.NewShortURLsHandler(db, cfg)

		// Create route group with required authentication middleware
		shortURLsRoutes := apiV1.Group("/short-urls")
		shortURLsRoutes.Use(jwtMiddleware.RequireAuth())

		// Register short URL management routes with scope checks
		shortURLsRoutes.GET("", middleware.RequireScope("read_urls"), shortURLsHandler.List)
		shortURLsRoutes.GET("/:id", middleware.RequireScope("read_urls"), shortURLsHandler.Get)
		shortURLsRoutes.PUT("/:id", middleware.RequireScope("write_urls"), shortURLsHandler.Update)
		shortURLsRoutes.DELETE("/:id", middleware.RequireScope("write_urls"), shortURLsHandler.Delete)

		log.Printf("Short URL management endpoints enabled at /api/v1/short-urls/*")

		// Register namespace management endpoints with JWT authentication
		namespacesHandler := handlers.NewNamespacesHandler(db, cfg)
		namespacesRoutes := apiV1.Group("/namespaces")
		namespacesRoutes.Use(jwtMiddleware.RequireAuth())

		// Register namespace management routes with scope checks
		namespacesRoutes.POST("", middleware.RequireScope("write_urls"), namespacesHandler.CreateNamespace)
		namespacesRoutes.GET("", middleware.RequireScope("read_urls"), namespacesHandler.ListNamespaces)
		namespacesRoutes.GET("/:id", middleware.RequireScope("read_urls"), namespacesHandler.GetNamespace)
		namespacesRoutes.PUT("/:id", middleware.RequireScope("write_urls"), namespacesHandler.UpdateNamespace)
		namespacesRoutes.DELETE("/:id", middleware.RequireScope("write_urls"), namespacesHandler.DeleteNamespace)

		log.Printf("Namespace management endpoints enabled at /api/v1/namespaces/*")

		// Register user endpoints with required authentication middleware
		meHandler := handlers.NewMeHandler(db)
		apiV1.GET("/me", jwtMiddleware.RequireAuth(), meHandler.GetMe)

		log.Printf("User endpoints enabled at /api/v1/me")

		// Register API key management endpoints with JWT authentication
		apiKeysHandler := handlers.NewAPIKeysHandler(db)
		apiKeysRoutes := apiV1.Group("/api-keys")
		apiKeysRoutes.Use(jwtMiddleware.RequireAuth())
		apiKeysRoutes.POST("", apiKeysHandler.CreateAPIKey)
		apiKeysRoutes.GET("", apiKeysHandler.ListAPIKeys)
		apiKeysRoutes.DELETE("/:id", apiKeysHandler.DeleteAPIKey)

		log.Printf("API key management endpoints enabled at /api/v1/api-keys/*")
	}

	// Register dashboard route (must be before landing routes to avoid conflicts)
	dashboardHandler := handlers.NewDashboardHandler(cfg, dashboardFS)
	r.Any("/dashboard", dashboardHandler.ServeDashboard)       // Match /dashboard exactly
	r.Any("/dashboard/*path", dashboardHandler.ServeDashboard) // Match /dashboard/*
	log.Printf("Dashboard enabled at /dashboard/*")

	// Register landing page route for root
	landingHandler := handlers.NewLandingHandler(cfg, landingFS)
	r.Any("/", landingHandler.ServeLanding)
	log.Printf("Landing page enabled at /")

	// NoRoute handler: try redirect first (for short URLs), then fall back to landing page
	// This allows short URLs to work while also supporting landing page routes like /docs
	r.NoRoute(func(c *gin.Context) {
		// Skip if it's an API or dashboard route (shouldn't happen, but safety check)
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/dashboard/") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Route not found",
			})
			return
		}

		// Reserved paths that should always be served by landing page (Next.js routes and static assets)
		reservedPaths := []string{
			"/_next",      // Next.js static assets and internal routes
			"/docs",       // Landing page docs routes
			"/favicon.ico", // Favicon
			"/robots.txt",  // Robots.txt
			"/sitemap.xml", // Sitemap
		}

		// Check if path is a reserved landing page path
		isReservedPath := false
		for _, reserved := range reservedPaths {
			if path == reserved || strings.HasPrefix(path, reserved+"/") {
				isReservedPath = true
				break
			}
		}

		// If it's a reserved path, serve landing page directly
		if isReservedPath {
			landingHandler.ServeLanding(c)
			return
		}

		// Check if path looks like a short URL (1-2 path segments, not reserved)
		// If it does, try redirect first; otherwise serve landing page directly
		pathParts := strings.Split(strings.Trim(path, "/"), "/")
		isPotentialShortURL := len(pathParts) <= 2 && len(pathParts) > 0

		// Also check if first segment is a reserved namespace name
		if isPotentialShortURL && len(pathParts) > 0 {
			firstSegment := strings.ToLower(pathParts[0])
			if handlers.IsReservedNamespaceName(firstSegment) {
				// First segment is reserved, serve landing page
				landingHandler.ServeLanding(c)
				return
			}
		}

		if isPotentialShortURL {
			// Use response recorder to buffer redirect handler's response
			// This allows us to check if it found a short URL before committing the response
			w := httptest.NewRecorder()
			newContext, _ := gin.CreateTestContext(w)
			newContext.Request = c.Request
			newContext.Params = c.Params
			
			// Try redirect handler with test context
			redirectHandler.Redirect(newContext)
			
			// Check if redirect found a short URL (status 301)
			if w.Code == http.StatusMovedPermanently {
				// Copy the redirect response to actual response
				for k, v := range w.Header() {
					for _, val := range v {
						c.Writer.Header().Add(k, val)
					}
				}
				c.Writer.WriteHeader(w.Code)
				c.Writer.Write(w.Body.Bytes())
				return
			}
			// Redirect didn't find a match (returned 404), fall through to serve landing page
		}

		// Serve landing page (either because it's not a short URL pattern, or redirect didn't find a match)
		landingHandler.ServeLanding(c)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%d", cfg.Port)
	}

	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
