package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/nanda/doit/config"
	"github.com/nanda/doit/modules/core/entity"
	"github.com/nanda/doit/modules/core/internal/repo/db"
)

func TestPostgresURLAnalyticRepo_Create(t *testing.T) {
	testDB := config.SetupTestDB(t)
	defer testDB.Cleanup()

	analyticRepo := db.NewPostgresURLAnalyticRepo(testDB.DB)
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name      string
		analytic  *entity.URLAnalytic
		expectID  bool
		expectErr bool
	}{
		{
			name: "create_valid_analytic_returns_id",
			analytic: &entity.URLAnalytic{
				URLID:      1,
				LongURL:    "https://example.com",
				CreatedAt:  now,
				ExpiresAt:  now.Add(24 * time.Hour),
				ClickCount: 0,
			},
			expectID:  true,
			expectErr: false,
		},
		{
			name: "create_analytic_with_last_accessed_at",
			analytic: &entity.URLAnalytic{
				URLID:          2,
				LongURL:        "https://example.com/page",
				CreatedAt:      now,
				ExpiresAt:      now.Add(48 * time.Hour),
				ClickCount:     5,
				LastAccessedAt: &now,
			},
			expectID:  true,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := analyticRepo.Create(ctx, tt.analytic)

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

func TestPostgresURLAnalyticRepo_GetByURLID(t *testing.T) {
	testDB := config.SetupTestDB(t)
	defer testDB.Cleanup()

	analyticRepo := db.NewPostgresURLAnalyticRepo(testDB.DB)
	ctx := context.Background()
	now := time.Now()

	// Create an analytic first
	createdAnalytic := &entity.URLAnalytic{
		URLID:      100,
		LongURL:    "https://example.com/gettest",
		CreatedAt:  now,
		ExpiresAt:  now.Add(24 * time.Hour),
		ClickCount: 10,
	}
	_, err := analyticRepo.Create(ctx, createdAnalytic)
	if err != nil {
		t.Fatalf("failed to create analytic: %v", err)
	}

	tests := []struct {
		name             string
		urlID            int64
		expectAnalytic   bool
		expectErr        bool
		expectLongURL    string
		expectClickCount int64
	}{
		{
			name:             "get_existing_analytic_returns_analytic",
			urlID:            100,
			expectAnalytic:   true,
			expectErr:        false,
			expectLongURL:    "https://example.com/gettest",
			expectClickCount: 10,
		},
		{
			name:           "get_nonexistent_analytic_returns_error",
			urlID:          99999,
			expectAnalytic: false,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analytic, err := analyticRepo.GetByURLID(ctx, tt.urlID)

			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.expectAnalytic {
				if analytic == nil {
					t.Error("expected analytic, got nil")
				} else {
					if analytic.LongURL != tt.expectLongURL {
						t.Errorf("expected long url %s, got %s", tt.expectLongURL, analytic.LongURL)
					}
					if analytic.ClickCount != tt.expectClickCount {
						t.Errorf("expected click count %d, got %d", tt.expectClickCount, analytic.ClickCount)
					}
				}
			}
		})
	}
}

func TestPostgresURLAnalyticRepo_UpdateStat(t *testing.T) {
	testDB := config.SetupTestDB(t)
	defer testDB.Cleanup()

	analyticRepo := db.NewPostgresURLAnalyticRepo(testDB.DB)
	ctx := context.Background()
	now := time.Now()

	// Create an analytic first
	createdAnalytic := &entity.URLAnalytic{
		URLID:      200,
		LongURL:    "https://example.com/updatetest",
		CreatedAt:  now,
		ExpiresAt:  now.Add(24 * time.Hour),
		ClickCount: 0,
	}
	_, err := analyticRepo.Create(ctx, createdAnalytic)
	if err != nil {
		t.Fatalf("failed to create analytic: %v", err)
	}

	tests := []struct {
		name             string
		urlID            int64
		updateCount      int
		expectClickCount int64
	}{
		{
			name:             "single_update_increments_click_count",
			urlID:            200,
			updateCount:      1,
			expectClickCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.updateCount; i++ {
				err := analyticRepo.UpdateStat(ctx, tt.urlID, time.Now())
				if err != nil {
					t.Errorf("failed to update stat: %v", err)
				}
			}

			analytic, err := analyticRepo.GetByURLID(ctx, tt.urlID)
			if err != nil {
				t.Fatalf("failed to get analytic: %v", err)
			}

			if analytic.ClickCount != tt.expectClickCount {
				t.Errorf("expected click count %d, got %d", tt.expectClickCount, analytic.ClickCount)
			}
			if analytic.LastAccessedAt == nil {
				t.Error("expected last_accessed_at to be set")
			}
		})
	}
}

func TestPostgresURLAnalyticRepo_UpdateStatMultipleTimes(t *testing.T) {
	testDB := config.SetupTestDB(t)
	defer testDB.Cleanup()

	analyticRepo := db.NewPostgresURLAnalyticRepo(testDB.DB)
	ctx := context.Background()
	now := time.Now()

	// Create an analytic
	createdAnalytic := &entity.URLAnalytic{
		URLID:      300,
		LongURL:    "https://example.com/multiupdate",
		CreatedAt:  now,
		ExpiresAt:  now.Add(24 * time.Hour),
		ClickCount: 0,
	}
	_, err := analyticRepo.Create(ctx, createdAnalytic)
	if err != nil {
		t.Fatalf("failed to create analytic: %v", err)
	}

	// Update 100 times
	for i := 0; i < 100; i++ {
		err := analyticRepo.UpdateStat(ctx, 300, time.Now())
		if err != nil {
			t.Fatalf("failed to update stat on iteration %d: %v", i, err)
		}
	}

	// Verify click count
	analytic, err := analyticRepo.GetByURLID(ctx, 300)
	if err != nil {
		t.Fatalf("failed to get analytic: %v", err)
	}

	if analytic.ClickCount != 100 {
		t.Errorf("expected click count 100, got %d", analytic.ClickCount)
	}
}
