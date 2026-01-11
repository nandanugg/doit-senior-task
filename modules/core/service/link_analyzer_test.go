package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/nanda/doit/modules/core/entity"
	"github.com/nanda/doit/modules/core/internal/test/mocks"
	"github.com/nanda/doit/modules/core/lib"
	"go.uber.org/mock/gomock"
)

func TestLinkAnalyzerService(t *testing.T) {
	tests := []struct {
		name                  string
		setupURL              *string
		redirectCount         int
		inputShortCode        *string
		expectError           error
		expectLongURL         *string
		expectClickCount      *int64
		expectLastAccessedSet *bool
	}{
		{
			name:          "existing_short_code_returns_long_url",
			setupURL:      ptr("https://analyze-test.com"),
			expectLongURL: ptr("https://analyze-test.com"),
		},
		{
			name:             "new_url_has_zero_click_count",
			setupURL:         ptr("https://new-url.com"),
			expectClickCount: ptr(int64(0)),
		},
		{
			name:             "after_redirects_returns_correct_click_count",
			setupURL:         ptr("https://multi-click.com"),
			redirectCount:    3,
			expectClickCount: ptr(int64(3)),
		},
		{
			name:                  "after_redirect_last_accessed_at_is_set",
			setupURL:              ptr("https://accessed-url.com"),
			redirectCount:         1,
			expectLastAccessedSet: ptr(true),
		},
		{
			name:           "nonexistent_short_code_returns_error",
			inputShortCode: ptr("fffff"),
			expectError:    ErrNotFound,
		},
		{
			name:           "invalid_short_code_returns_error",
			inputShortCode: ptr("invalid!"),
			expectError:    ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cacheRepo := mocks.NewMockURLCacheRepo(ctrl)
			analyticRepo := mocks.NewMockURLAnalyticRepo(ctrl)
			ctx := context.Background()

			var shortCode string
			var urlID int64

			if tt.setupURL != nil {
				// Setup: Create URL
				urlID = int64(1)

				cacheRepo.EXPECT().
					Create(gomock.Any(), *tt.setupURL, DefaultTTL).
					Return(urlID, nil)

				analyticRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)

				creatorSvc := NewLinkCreatorService(cacheRepo, analyticRepo)
				var err error
				shortCode, err = creatorSvc.Create(ctx, *tt.setupURL, nil)
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			// Setup: Perform redirects
			if tt.redirectCount > 0 && tt.setupURL != nil {
				cacheRepo.EXPECT().
					Get(gomock.Any(), urlID).
					Return(*tt.setupURL, nil).
					Times(tt.redirectCount)

				analyticRepo.EXPECT().
					UpdateStat(gomock.Any(), urlID, gomock.Any()).
					Return(nil).
					Times(tt.redirectCount)

				redirectorSvc := NewLinkRedirectorService(cacheRepo, analyticRepo)
				for i := 0; i < tt.redirectCount; i++ {
					_, _ = redirectorSvc.Redirect(ctx, shortCode)
				}

				// Wait for async updates
				time.Sleep(50 * time.Millisecond)
			}

			if tt.inputShortCode != nil {
				shortCode = *tt.inputShortCode
			}

			// Setup expectations for analyzer
			if tt.expectError == nil && tt.setupURL != nil {
				lastAccessedAt := (*time.Time)(nil)
				if tt.expectLastAccessedSet != nil && *tt.expectLastAccessedSet {
					now := time.Now()
					lastAccessedAt = &now
				}

				clickCount := int64(0)
				if tt.expectClickCount != nil {
					clickCount = *tt.expectClickCount
				}

				analyticRepo.EXPECT().
					GetByURLID(gomock.Any(), urlID).
					Return(&entity.URLAnalytic{
						ID:             1,
						URLID:          urlID,
						LongURL:        *tt.setupURL,
						ClickCount:     clickCount,
						LastAccessedAt: lastAccessedAt,
					}, nil)
			} else if tt.expectError == ErrNotFound && tt.inputShortCode != nil {
				// Check if this is a decode error (invalid characters) or a GetByURLID error
				_, decodeErr := lib.HexDecode(*tt.inputShortCode)
				if decodeErr == nil {
					// Valid short code format, will reach GetByURLID
					analyticRepo.EXPECT().
						GetByURLID(gomock.Any(), gomock.Any()).
						Return(nil, sql.ErrNoRows)
				}
				// If decode fails, no GetByURLID call will be made
			}

			analyzerSvc := NewLinkAnalyzerService(analyticRepo)
			analytic, err := analyzerSvc.Analyze(ctx, shortCode)

			if tt.expectError != nil {
				if err != tt.expectError {
					t.Errorf("expected error %v, got %v", tt.expectError, err)
				}
			}

			if tt.expectLongURL != nil {
				if analytic.LongURL != *tt.expectLongURL {
					t.Errorf("expected long URL %s, got %s", *tt.expectLongURL, analytic.LongURL)
				}
			}

			if tt.expectClickCount != nil {
				if analytic.ClickCount != *tt.expectClickCount {
					t.Errorf("expected click count %d, got %d", *tt.expectClickCount, analytic.ClickCount)
				}
			}

			if tt.expectLastAccessedSet != nil {
				isSet := analytic.LastAccessedAt != nil
				if isSet != *tt.expectLastAccessedSet {
					t.Errorf("expected last_accessed_at set=%v, got set=%v", *tt.expectLastAccessedSet, isSet)
				}
			}
		})
	}
}
