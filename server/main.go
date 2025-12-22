package main

import (
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
	"openshortpath/server/models"
)

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
	if err := db.AutoMigrate(&models.ShortURL{}); err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	r := gin.Default()

	// Initialize handlers with database
	helloHandler := handlers.NewHelloHandler(db)
	shortenHandler := handlers.NewShortenHandler(db, cfg)

	// Routes
	r.GET("/", helloHandler.HelloWorld)
	r.POST("/api/v1/shorten", shortenHandler.Shorten)

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

