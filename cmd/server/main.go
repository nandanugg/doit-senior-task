package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nanda/doit/modules/core/handler"
)

func main() {
	// Initialize Echo server
	e := echo.New()

	// Register middleware
	e.Use(middleware.Recover())
	e.Use(handler.ProcessingTimeMiddleware())
	e.Use(handler.PrometheusMiddleware())

	// Register routes
	// TODO: Wire up actual handlers with service dependencies (redis/db) in next steps
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})

	e.POST("/s", func(c echo.Context) error {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented yet"})
	})

	e.GET("/s/:short_code", func(c echo.Context) error {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented yet"})
	})

	e.GET("/stats/:short_code", func(c echo.Context) error {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "not implemented yet"})
	})

	// Start server
	log.Println("Starting server on :8080")
	if err := e.Start(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
