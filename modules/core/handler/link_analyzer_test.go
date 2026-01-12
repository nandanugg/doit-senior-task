package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nanda/doit/modules/core/entity"
	"github.com/nanda/doit/modules/core/internal/test/mocks"
	"github.com/nanda/doit/modules/core/service"
	"go.uber.org/mock/gomock"
)

func TestLinkAnalyzerHandler(t *testing.T) {
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	lastAccessed := fixedTime.Add(time.Hour)

	tests := []struct {
		name           string
		shortCode      string
		mockReturn     *entity.URLAnalytic
		mockError      error
		expectStatus   *int
		expectContains *string
	}{
		{
			name:      "successful_analysis_returns_200",
			shortCode: "abc123",
			mockReturn: &entity.URLAnalytic{
				LongURL:    "https://example.com",
				CreatedAt:  fixedTime,
				ExpiresAt:  fixedTime.Add(24 * time.Hour),
				ClickCount: 42,
			},
			mockError:    nil,
			expectStatus: ptr(http.StatusOK),
		},
		{
			name:      "successful_analysis_returns_long_url",
			shortCode: "abc123",
			mockReturn: &entity.URLAnalytic{
				LongURL:    "https://example.com",
				CreatedAt:  fixedTime,
				ExpiresAt:  fixedTime.Add(24 * time.Hour),
				ClickCount: 42,
			},
			mockError:      nil,
			expectContains: ptr("https://example.com"),
		},
		{
			name:      "successful_analysis_returns_click_count",
			shortCode: "abc123",
			mockReturn: &entity.URLAnalytic{
				LongURL:    "https://example.com",
				CreatedAt:  fixedTime,
				ExpiresAt:  fixedTime.Add(24 * time.Hour),
				ClickCount: 42,
			},
			mockError:      nil,
			expectContains: ptr("42"),
		},
		{
			name:         "not_found_returns_404",
			shortCode:    "notfound",
			mockReturn:   nil,
			mockError:    service.ErrNotFound,
			expectStatus: ptr(http.StatusNotFound),
		},
		{
			name:      "with_last_accessed_at_returns_field",
			shortCode: "abc",
			mockReturn: &entity.URLAnalytic{
				LongURL:        "https://example.com",
				CreatedAt:      fixedTime,
				ExpiresAt:      fixedTime.Add(24 * time.Hour),
				ClickCount:     1,
				LastAccessedAt: &lastAccessed,
			},
			mockError:      nil,
			expectContains: ptr("last_accessed_at"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			e := echo.New()
			mockService := mocks.NewMockLinkAnalyzer(ctrl)

			// Setup expectations
			mockService.EXPECT().
				Analyze(gomock.Any(), tt.shortCode).
				Return(tt.mockReturn, tt.mockError).
				Times(1)

			handler := NewLinkAnalyzerHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/stats/"+tt.shortCode, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/stats/:short_code")
			c.SetParamNames("short_code")
			c.SetParamValues(tt.shortCode)

			_ = handler.Handle(c)

			if tt.expectStatus != nil {
				if rec.Code != *tt.expectStatus {
					t.Errorf("expected status %d, got %d", *tt.expectStatus, rec.Code)
				}
			}

			if tt.expectContains != nil {
				if !strings.Contains(rec.Body.String(), *tt.expectContains) {
					t.Errorf("expected body to contain %s", *tt.expectContains)
				}
			}
		})
	}
}
