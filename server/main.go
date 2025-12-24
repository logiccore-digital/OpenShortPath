package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"

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

	// Create API v1 route group with rate limiting middleware
	// Rate limiting middleware runs after OptionalAuth so user context is available
	apiV1 := r.Group("/api/v1")
	apiV1.Use(middleware.RateLimitMiddleware(db))
	log.Printf("Rate limiting enabled for /api/v1/* endpoints")

	// Initialize handlers with database
	helloHandler := handlers.NewHelloHandler(db)
	shortenHandler := handlers.NewShortenHandler(db, cfg)
	redirectHandler := handlers.NewRedirectHandler(db, cfg)
	authProviderHandler := handlers.NewAuthProviderHandler(cfg)
	domainsHandler := handlers.NewDomainsHandler(cfg)

	// Register dashboard route (must be before /:slug route to avoid conflicts)
	dashboardHandler := handlers.NewDashboardHandler(cfg, dashboardFS)
	r.Any("/dashboard", dashboardHandler.ServeDashboard)       // Match /dashboard exactly
	r.Any("/dashboard/*path", dashboardHandler.ServeDashboard) // Match /dashboard/*
	log.Printf("Dashboard enabled at /dashboard/*")

	// Routes
	r.GET("/", helloHandler.HelloWorld)
	// Note: Dashboard routes are registered above, so they take precedence
	// Use catch-all route to handle both /:slug and /:namespace/:slug patterns
	// This must come after dashboard routes to avoid conflicts
	r.NoRoute(redirectHandler.Redirect)

	// Shorten endpoint - authentication is optional (handled by OptionalAuth middleware)
	// Rate limiting is applied per IP for anonymous users, per user for authenticated users
	apiV1.POST("/shorten", shortenHandler.Shorten)

	apiV1.GET("/auth-provider", authProviderHandler.GetAuthProvider)
	apiV1.GET("/domains", domainsHandler.GetDomains)

	// Register login endpoint only if auth_provider is "local"
	if cfg.AuthProvider == "local" {
		loginHandler := handlers.NewLoginHandler(db, cfg.JWT)
		apiV1.POST("/login", loginHandler.Login)
		log.Printf("Login endpoint enabled at /api/v1/login")
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
