package core

import (
	"context"

	"github.com/nanda/doit/modules/core/handler"
	"github.com/nanda/doit/modules/core/internal/repo"
	"github.com/nanda/doit/modules/core/internal/repo/cache"
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
func NewBuilder(redisClient *redis.Client) *Builder {
	// Initialize repositories
	cacheRepo := cache.NewRedisURLCacheRepo(redisClient)
	// TODO: Replace stub with real PostgreSQL repo in next step
	analyticRepo := repo.NewStubURLAnalyticRepo()

	// Initialize services
	creatorSvc := service.NewLinkCreatorService(cacheRepo, analyticRepo)
	redirectorSvc := service.NewLinkRedirectorService(cacheRepo, analyticRepo)
	analyzerSvc := service.NewLinkAnalyzerService(analyticRepo)

	// Initialize handlers
	creatorHandler := handler.NewLinkCreatorHandler(creatorSvc)
	redirectorHandler := handler.NewLinkRedirectorHandler(redirectorSvc)
	analyzerHandler := handler.NewLinkAnalyzerHandler(analyzerSvc)
	healthzHandler := handler.NewHealthzHandler(func() error {
		// Check Redis health
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
