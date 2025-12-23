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
	if err := db.AutoMigrate(&models.ShortURL{}, &models.User{}); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	r := gin.Default()

	// Initialize JWT middleware if JWT config is provided
	if cfg.JWT != nil {
		jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWT)
		r.Use(jwtMiddleware.OptionalAuth())
		log.Printf("JWT authentication enabled (algorithm: %s)", cfg.JWT.Algorithm)
	}

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
	r.GET("/:slug", redirectHandler.Redirect)
	r.POST("/api/v1/shorten", shortenHandler.Shorten)
	r.GET("/api/v1/auth-provider", authProviderHandler.GetAuthProvider)
	r.GET("/api/v1/domains", domainsHandler.GetDomains)

	// Register login endpoint only if auth_provider is "local"
	if cfg.AuthProvider == "local" {
		loginHandler := handlers.NewLoginHandler(db, cfg.JWT)
		r.POST("/api/v1/login", loginHandler.Login)
		log.Printf("Login endpoint enabled at /api/v1/login")
	}

	// Register admin endpoints if admin password is configured
	if cfg.AdminPassword != "" {
		adminMiddleware := middleware.NewAdminMiddleware(cfg.AdminPassword)
		adminUsersHandler := handlers.NewAdminUsersHandler(db)

		// Create admin route group with authentication middleware
		adminRoutes := r.Group("/api/v1/__admin")
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
		jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWT)
		shortURLsHandler := handlers.NewShortURLsHandler(db, cfg)

		// Create route group with required authentication middleware
		shortURLsRoutes := r.Group("/api/v1/short-urls")
		shortURLsRoutes.Use(jwtMiddleware.RequireAuth())

		// Register short URL management routes
		shortURLsRoutes.GET("", shortURLsHandler.List)
		shortURLsRoutes.GET("/:id", shortURLsHandler.Get)
		shortURLsRoutes.PUT("/:id", shortURLsHandler.Update)
		shortURLsRoutes.DELETE("/:id", shortURLsHandler.Delete)

		log.Printf("Short URL management endpoints enabled at /api/v1/short-urls/*")

		// Register user endpoints with required authentication middleware
		meHandler := handlers.NewMeHandler(db)
		meRoutes := r.Group("/api/v1")
		meRoutes.Use(jwtMiddleware.RequireAuth())
		meRoutes.GET("/me", meHandler.GetMe)

		log.Printf("User endpoints enabled at /api/v1/me")
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
