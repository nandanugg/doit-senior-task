package service

import (
	"context"
	"time"

	"github.com/nanda/doit/modules/core/entity"
)

// URLCacheRepo interface for URL caching operations (Redis).
type URLCacheRepo interface {
	// Create generates a new ID and stores the URL mapping with the specified TTL.
	Create(ctx context.Context, longURL string, ttl time.Duration) (int64, error)

	// Get retrieves the long URL for the given ID.
	Get(ctx context.Context, id int64) (string, error)

	// Set stores a URL mapping with the specified TTL.
	Set(ctx context.Context, id int64, longURL string, ttl time.Duration) error

	// Delete removes a URL mapping from the cache.
	Delete(ctx context.Context, id int64) error
}

// URLAnalyticRepo interface for URL analytics repository operations (PostgreSQL).
type URLAnalyticRepo interface {
	Create(ctx context.Context, analytic *entity.URLAnalytic) (int64, error)
	GetByURLID(ctx context.Context, urlID int64) (*entity.URLAnalytic, error)
	UpdateStat(ctx context.Context, urlID int64, now time.Time) error
}
