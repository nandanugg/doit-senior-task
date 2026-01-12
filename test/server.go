package test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/nanda/doit/modules/core"
	"github.com/nanda/doit/modules/core/handler"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// TestServer holds the test HTTP server and its dependencies.
type TestServer struct {
	server *http.Server
	url    string
}

// NewTestServer creates a new test HTTP server with the given database and Redis connections.
func NewTestServer(db *sql.DB, redisClient *redis.Client) *TestServer {
	// Build application dependencies
	builder := core.NewBuilder(db, redisClient)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Register middleware
	e.Use(handler.ProcessingTimeMiddleware())

	// Register routes
	e.GET("/healthz", builder.HealthzHandler.Handle)
	e.POST("/s", builder.LinkCreatorHandler.Handle)
	e.GET("/s/:short_code", builder.LinkRedirectorHandler.Handle)
	e.GET("/stats/:short_code", builder.LinkAnalyzerHandler.Handle)

	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("failed to find available port: %v", err))
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: e,
	}

	// Start server in background
	go func() {
		_ = server.ListenAndServe()
	}()

	// Wait for server to be ready
	baseURL := "http://" + addr
	maxRetries := 50
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(baseURL + "/healthz")
		if err == nil {
			_ = resp.Body.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
		if i == maxRetries-1 {
			panic("server failed to start in time")
		}
	}

	return &TestServer{
		server: server,
		url:    baseURL,
	}
}

// Close shuts down the test server.
func (ts *TestServer) Close() {
	_ = ts.server.Shutdown(context.Background())
}

// URL returns the base URL of the test server.
func (ts *TestServer) URL() string {
	return ts.url
}
