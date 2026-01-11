package core

import (
	"context"
	"database/sql"

	"github.com/nanda/doit/modules/core/handler"
	"github.com/nanda/doit/modules/core/internal/repo/cache"
	"github.com/nanda/doit/modules/core/internal/repo/db"
	"github.com/nanda/doit/modules/core/service"
	"github.com/redis/go-redis/v9"
)

// Builder holds all the dependencies for the application.
type Builder struct {
	// Repositories
	URLCacheRepo    service.URLCacheRepo
	URLAnalyticRepo service.URLAnalyticRepo

	// Services
	LinkCreatorService    *service.LinkCreatorService
	LinkRedirectorService *service.LinkRedirectorService
	LinkAnalyzerService   *service.LinkAnalyzerService

	// Handlers
	LinkCreatorHandler    *handler.LinkCreatorHandler
	LinkRedirectorHandler *handler.LinkRedirectorHandler
	LinkAnalyzerHandler   *handler.LinkAnalyzerHandler
	HealthzHandler        *handler.HealthzHandler
}

// NewBuilder creates a new Builder with all dependencies initialized.
func NewBuilder(database *sql.DB, redisClient *redis.Client) *Builder {
	// Initialize repositories
	cacheRepo := cache.NewRedisURLCacheRepo(redisClient)
	analyticRepo := db.NewPostgresURLAnalyticRepo(database)

	// Initialize services
	creatorSvc := service.NewLinkCreatorService(cacheRepo, analyticRepo)
	redirectorSvc := service.NewLinkRedirectorService(cacheRepo, analyticRepo)
	analyzerSvc := service.NewLinkAnalyzerService(analyticRepo)

	// Initialize handlers
	creatorHandler := handler.NewLinkCreatorHandler(creatorSvc)
	redirectorHandler := handler.NewLinkRedirectorHandler(redirectorSvc)
	analyzerHandler := handler.NewLinkAnalyzerHandler(analyzerSvc)
	healthzHandler := handler.NewHealthzHandler(func() error {
		// Check both database and Redis health
		if err := database.Ping(); err != nil {
			return err
		}
		return redisClient.Ping(context.Background()).Err()
	})

	return &Builder{
		URLCacheRepo:          cacheRepo,
		URLAnalyticRepo:       analyticRepo,
		LinkCreatorService:    creatorSvc,
		LinkRedirectorService: redirectorSvc,
		LinkAnalyzerService:   analyzerSvc,
		LinkCreatorHandler:    creatorHandler,
		LinkRedirectorHandler: redirectorHandler,
		LinkAnalyzerHandler:   analyzerHandler,
		HealthzHandler:        healthzHandler,
	}
}
