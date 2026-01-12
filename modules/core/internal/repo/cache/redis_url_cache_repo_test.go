package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/nanda/doit/config"
	"github.com/nanda/doit/modules/core/internal/repo/cache"
)

func TestRedisURLCacheRepo_Create(t *testing.T) {
	testRedis := config.SetupTestRedis(t)
	defer testRedis.Cleanup()

	repo := cache.NewRedisURLCacheRepo(testRedis.Client)
	ctx := context.Background()

	tests := []struct {
		name      string
		longURL   string
		ttl       time.Duration
		expectID  bool
		expectErr bool
	}{
		{
			name:      "create_url_with_valid_ttl_returns_id",
			longURL:   "https://example.com",
			ttl:       1 * time.Hour,
			expectID:  true,
			expectErr: false,
		},
		{
			name:      "create_multiple_urls_increments_id",
			longURL:   "https://test.com",
			ttl:       24 * time.Hour,
			expectID:  true,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := repo.Create(ctx, tt.longURL, tt.ttl)

			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.expectID && id <= 0 {
				t.Errorf("expected positive id, got %d", id)
			}
		})
	}
}

func TestRedisURLCacheRepo_Get(t *testing.T) {
	testRedis := config.SetupTestRedis(t)
	defer testRedis.Cleanup()

	repo := cache.NewRedisURLCacheRepo(testRedis.Client)
	ctx := context.Background()

	// Create a URL first
	longURL := "https://example.com/get-test"
	id, err := repo.Create(ctx, longURL, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to create URL: %v", err)
	}

	tests := []struct {
		name           string
		id             int64
		expectLongURL  string
		expectNotFound bool
	}{
		{
			name:          "get_existing_url_returns_long_url",
			id:            id,
			expectLongURL: longURL,
		},
		{
			name:           "get_nonexistent_url_returns_error",
			id:             99999,
			expectNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.Get(ctx, tt.id)

			if tt.expectNotFound {
				if err == nil {
					t.Error("expected error for nonexistent URL, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if result != tt.expectLongURL {
					t.Errorf("expected %s, got %s", tt.expectLongURL, result)
				}
			}
		})
	}
}

func TestRedisURLCacheRepo_Expiration(t *testing.T) {
	testRedis := config.SetupTestRedis(t)
	defer testRedis.Cleanup()

	repo := cache.NewRedisURLCacheRepo(testRedis.Client)
	ctx := context.Background()

	// Create a URL with very short TTL
	longURL := "https://example.com/expiration-test"
	id, err := repo.Create(ctx, longURL, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create URL: %v", err)
	}

	// Verify it exists immediately
	result, err := repo.Get(ctx, id)
	if err != nil {
		t.Errorf("expected URL to exist immediately, got error: %v", err)
	}
	if result != longURL {
		t.Errorf("expected %s, got %s", longURL, result)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Verify it's expired
	_, err = repo.Get(ctx, id)
	if err == nil {
		t.Error("expected error for expired URL, got nil")
	}
}

func TestRedisURLCacheRepo_Delete(t *testing.T) {
	testRedis := config.SetupTestRedis(t)
	defer testRedis.Cleanup()

	repo := cache.NewRedisURLCacheRepo(testRedis.Client)
	ctx := context.Background()

	// Create a URL
	longURL := "https://example.com/delete-test"
	id, err := repo.Create(ctx, longURL, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to create URL: %v", err)
	}

	// Verify it exists
	_, err = repo.Get(ctx, id)
	if err != nil {
		t.Errorf("expected URL to exist, got error: %v", err)
	}

	// Delete it
	err = repo.Delete(ctx, id)
	if err != nil {
		t.Errorf("expected no error on delete, got %v", err)
	}

	// Verify it's gone
	_, err = repo.Get(ctx, id)
	if err == nil {
		t.Error("expected error for deleted URL, got nil")
	}
}

func TestRedisURLCacheRepo_Set(t *testing.T) {
	testRedis := config.SetupTestRedis(t)
	defer testRedis.Cleanup()

	repo := cache.NewRedisURLCacheRepo(testRedis.Client)
	ctx := context.Background()

	// Set a URL directly
	id := int64(12345)
	longURL := "https://example.com/set-test"
	err := repo.Set(ctx, id, longURL, 1*time.Hour)
	if err != nil {
		t.Errorf("expected no error on set, got %v", err)
	}

	// Verify it can be retrieved
	result, err := repo.Get(ctx, id)
	if err != nil {
		t.Errorf("expected no error on get, got %v", err)
	}
	if result != longURL {
		t.Errorf("expected %s, got %s", longURL, result)
	}
}
