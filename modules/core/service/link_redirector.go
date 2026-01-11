package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/nanda/doit/modules/core/lib"
)

var ErrNotFound = errors.New("short code not found or expired")

type LinkRedirector interface {
	Redirect(ctx context.Context, shortCode string) (string, error)
}

type LinkRedirectorService struct {
	cacheRepo    URLCacheRepo
	analyticRepo URLAnalyticRepo
}

func NewLinkRedirectorService(
	cacheRepo URLCacheRepo,
	analyticRepo URLAnalyticRepo,
) *LinkRedirectorService {
	return &LinkRedirectorService{
		cacheRepo:    cacheRepo,
		analyticRepo: analyticRepo,
	}
}

func (s *LinkRedirectorService) Redirect(ctx context.Context, shortCode string) (string, error) {
	id, err := lib.HexDecode(shortCode)
	if err != nil {
		return "", ErrNotFound
	}

	// Get URL from Redis cache (Redis handles expiration via TTL)
	longURL, err := s.cacheRepo.Get(ctx, id)
	if err != nil {
		// If not found in cache, it's either expired or never existed
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "expired") {
			return "", ErrNotFound
		}
		return "", err
	}

	// Update analytics asynchronously (non-blocking)
	now := time.Now()
	go func() {
		_ = s.analyticRepo.UpdateStat(context.Background(), id, now)
	}()

	return longURL, nil
}
