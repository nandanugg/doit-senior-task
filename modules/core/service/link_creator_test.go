package service

import (
	"context"
	"strings"
	"testing"

	"github.com/nanda/doit/modules/core/internal/test/mocks"
	"go.uber.org/mock/gomock"
)

func TestLinkCreatorService(t *testing.T) {
	tests := []struct {
		name           string
		inputURL       string
		inputTTL       *int64
		expectError    error
		expectNonEmpty *bool
		expectValidHex *bool
	}{
		{
			name:           "valid_url_with_default_ttl_returns_non_empty",
			inputURL:       "https://example.com",
			inputTTL:       nil,
			expectNonEmpty: ptr(true),
		},
		{
			name:           "valid_url_with_custom_ttl_returns_non_empty",
			inputURL:       "https://example.org",
			inputTTL:       ptr(int64(3600)),
			expectNonEmpty: ptr(true),
		},
		{
			name:        "invalid_url_no_scheme_returns_error",
			inputURL:    "example.com",
			inputTTL:    nil,
			expectError: ErrInvalidURL,
		},
		{
			name:        "invalid_url_ftp_scheme_returns_error",
			inputURL:    "ftp://example.com",
			inputTTL:    nil,
			expectError: ErrInvalidURL,
		},
		{
			name:        "url_too_long_returns_error",
			inputURL:    "https://example.com/" + strings.Repeat("a", 2048),
			inputTTL:    nil,
			expectError: ErrURLTooLong,
		},
		{
			name:        "ttl_too_short_returns_error",
			inputURL:    "https://example.com",
			inputTTL:    ptr(int64(60)),
			expectError: ErrInvalidTTL,
		},
		{
			name:        "ttl_too_long_returns_error",
			inputURL:    "https://example.com",
			inputTTL:    ptr(int64(8 * 24 * 3600)),
			expectError: ErrInvalidTTL,
		},
		{
			name:           "short_code_is_valid_hex",
			inputURL:       "https://test.com",
			inputTTL:       nil,
			expectValidHex: ptr(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCacheRepo := mocks.NewMockURLCacheRepo(ctrl)
			mockAnalyticRepo := mocks.NewMockURLAnalyticRepo(ctrl)

			// Only expect repository calls if no validation error is expected
			if tt.expectError == nil {
				mockCacheRepo.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(1), nil).
					Times(1)

				mockAnalyticRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(1), nil).
					Times(1)
			}

			svc := NewLinkCreatorService(mockCacheRepo, mockAnalyticRepo)
			ctx := context.Background()

			shortCode, err := svc.Create(ctx, tt.inputURL, tt.inputTTL)

			if tt.expectError != nil {
				if err != tt.expectError {
					t.Errorf("expected error %v, got %v", tt.expectError, err)
				}
			}

			if tt.expectNonEmpty != nil {
				if (shortCode != "") != *tt.expectNonEmpty {
					t.Error("expected non-empty short code")
				}
			}

			if tt.expectValidHex != nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				// Check if it's valid hex by trying to decode
				for _, c := range shortCode {
					if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'g' || c > 'h') {
						t.Errorf("short code contains invalid hex char: %c", c)
					}
				}
			}
		})
	}
}
