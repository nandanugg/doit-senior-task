package service

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/nanda/doit/modules/core/entity"
	"github.com/nanda/doit/modules/core/lib"
)

const (
	DefaultTTL = 24 * time.Hour
	MinTTL     = 1 * time.Hour
	MaxTTL     = 7 * 24 * time.Hour
	MaxURLLen  = 2048
)

var (
	ErrInvalidURL = errors.New("invalid URL: must be a valid HTTP or HTTPS URL")
	ErrURLTooLong = errors.New("URL too long: maximum length is 2048 characters")
	ErrInvalidTTL = errors.New("invalid TTL: must be between 1 hour and 1 week")
)

type LinkCreator interface {
	Create(ctx context.Context, longURL string, ttlSeconds *int64) (string, error)
}

type LinkCreatorService struct {
	cacheRepo    URLCacheRepo
	analyticRepo URLAnalyticRepo
}

func NewLinkCreatorService(
	cacheRepo URLCacheRepo,
	analyticRepo URLAnalyticRepo,
) *LinkCreatorService {
	return &LinkCreatorService{
		cacheRepo:    cacheRepo,
		analyticRepo: analyticRepo,
	}
}

func (s *LinkCreatorService) Create(ctx context.Context, longURL string, ttlSeconds *int64) (string, error) {
	if err := validateURL(longURL); err != nil {
		return "", err
	}

	ttl := DefaultTTL
	if ttlSeconds != nil {
		ttl = time.Duration(*ttlSeconds) * time.Second
		if ttl < MinTTL || ttl > MaxTTL {
			return "", ErrInvalidTTL
		}
	}

	now := time.Now()
	expiresAt := now.Add(ttl)

	// Create URL in Redis cache with TTL
	id, err := s.cacheRepo.Create(ctx, longURL, ttl)
	if err != nil {
		return "", err
	}

	// Create analytics record in PostgreSQL
	analyticEntity := &entity.URLAnalytic{
		URLID:      id,
		LongURL:    longURL,
		CreatedAt:  now,
		ExpiresAt:  expiresAt,
		ClickCount: 0,
	}

	_, err = s.analyticRepo.Create(ctx, analyticEntity)
	if err != nil {
		return "", err
	}

	return lib.HexEncode(id), nil
}

func validateURL(rawURL string) error {
	if len(rawURL) > MaxURLLen {
		return ErrURLTooLong
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidURL
	}

	if parsed.Host == "" {
		return ErrInvalidURL
	}

	return nil
}
