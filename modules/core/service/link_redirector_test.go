package service

import (
	"context"
	"testing"
	"time"

	"github.com/nanda/doit/modules/core/internal/test/mocks"
	"github.com/nanda/doit/modules/core/lib"
	"go.uber.org/mock/gomock"
)

func TestLinkRedirectorService(t *testing.T) {
	tests := []struct {
		name             string
		setupURL         *string
		setupTTL         *int64
		redirectCount    int
		inputShortCode   *string
		expectError      error
		expectLongURL    *string
		expectClickCount *int64
	}{
		{
			name:          "valid_short_code_returns_long_url",
			setupURL:      ptr("https://example.com"),
			redirectCount: 1,
			expectLongURL: ptr("https://example.com"),
		},
		{
			name:             "redirect_updates_click_count",
			setupURL:         ptr("https://click-test.com"),
			redirectCount:    5,
			expectClickCount: ptr(int64(5)),
		},
		{
			name:           "nonexistent_short_code_returns_error",
			inputShortCode: ptr("fffff"),
			redirectCount:  1,
			expectError:    ErrNotFound,
		},
		{
			name:           "invalid_short_code_returns_error",
			inputShortCode: ptr("invalid!"),
			redirectCount:  1,
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
				ttl := DefaultTTL
				if tt.setupTTL != nil {
					ttl = time.Duration(*tt.setupTTL) * time.Second
				}

				cacheRepo.EXPECT().
					Create(gomock.Any(), *tt.setupURL, ttl).
					Return(urlID, nil)

				analyticRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)

				creatorSvc := NewLinkCreatorService(cacheRepo, analyticRepo)
				var err error
				shortCode, err = creatorSvc.Create(ctx, *tt.setupURL, tt.setupTTL)
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			if tt.inputShortCode != nil {
				shortCode = *tt.inputShortCode
			}

			// Setup expectations for redirect
			redirectorSvc := NewLinkRedirectorService(cacheRepo, analyticRepo)

			var longURL string
			var err error

			if tt.expectError == nil && tt.setupURL != nil {
				// Expect Get calls
				cacheRepo.EXPECT().
					Get(gomock.Any(), urlID).
					Return(*tt.setupURL, nil).
					Times(tt.redirectCount)

				// Expect UpdateStat calls (async, may complete after function returns)
				analyticRepo.EXPECT().
					UpdateStat(gomock.Any(), urlID, gomock.Any()).
					Return(nil).
					Times(tt.redirectCount)
			} else if tt.expectError == ErrNotFound && tt.inputShortCode != nil {
				// Check if this is a decode error (invalid characters) or a Get error
				_, decodeErr := lib.HexDecode(*tt.inputShortCode)
				if decodeErr == nil {
					// Valid short code format, will reach Get
					cacheRepo.EXPECT().
						Get(gomock.Any(), gomock.Any()).
						Return("", ErrNotFound).
						Times(tt.redirectCount)
				}
				// If decode fails, no Get call will be made
			}

			for i := 0; i < tt.redirectCount; i++ {
				longURL, err = redirectorSvc.Redirect(ctx, shortCode)
			}

			// Wait briefly for async operations (gomock will verify they were called)
			time.Sleep(10 * time.Millisecond)

			if tt.expectError != nil {
				if err != tt.expectError {
					t.Errorf("expected error %v, got %v", tt.expectError, err)
				}
			}

			if tt.expectLongURL != nil {
				if longURL != *tt.expectLongURL {
					t.Errorf("expected long URL %s, got %s", *tt.expectLongURL, longURL)
				}
			}
		})
	}
}
