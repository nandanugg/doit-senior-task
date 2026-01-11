package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nanda/doit/config"
	"github.com/nanda/doit/modules/core"
	"github.com/nanda/doit/modules/core/handler"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := config.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize Redis
	redisClient, err := config.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		}
	}()

	// Build application dependencies
	builder := core.NewBuilder(db, redisClient)

	// Initialize Echo server
	e := echo.New()

	// Register middleware
	e.Use(middleware.Recover())
	e.Use(handler.ProcessingTimeMiddleware())
	e.Use(handler.PrometheusMiddleware())

	// Register routes
	e.GET("/healthz", builder.HealthzHandler.Handle)
	e.GET("/metrics", echo.WrapHandler(config.NewPrometheusHandler()))

	e.POST("/s", builder.LinkCreatorHandler.Handle)
	e.GET("/s/:short_code", builder.LinkRedirectorHandler.Handle)
	e.GET("/stats/:short_code", builder.LinkAnalyzerHandler.Handle)

	// Start server
	log.Printf("Starting server on :%s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
