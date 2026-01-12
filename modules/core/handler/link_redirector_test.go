package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nanda/doit/modules/core/internal/test/mocks"
	"github.com/nanda/doit/modules/core/service"
	"go.uber.org/mock/gomock"
)

func TestLinkRedirectorHandler(t *testing.T) {
	tests := []struct {
		name           string
		shortCode      string
		mockReturn     string
		mockError      error
		expectStatus   *int
		expectLocation *string
	}{
		{
			name:         "successful_redirect_returns_302",
			shortCode:    "abc123",
			mockReturn:   "https://example.com",
			mockError:    nil,
			expectStatus: ptr(http.StatusFound),
		},
		{
			name:           "successful_redirect_sets_location_header",
			shortCode:      "abc123",
			mockReturn:     "https://example.com",
			mockError:      nil,
			expectLocation: ptr("https://example.com"),
		},
		{
			name:         "not_found_returns_404",
			shortCode:    "notfound",
			mockReturn:   "",
			mockError:    service.ErrNotFound,
			expectStatus: ptr(http.StatusNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			e := echo.New()
			mockService := mocks.NewMockLinkRedirector(ctrl)

			// Setup expectations
			mockService.EXPECT().
				Redirect(gomock.Any(), tt.shortCode).
				Return(tt.mockReturn, tt.mockError).
				Times(1)

			handler := NewLinkRedirectorHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/s/"+tt.shortCode, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/s/:short_code")
			c.SetParamNames("short_code")
			c.SetParamValues(tt.shortCode)

			_ = handler.Handle(c)

			if tt.expectStatus != nil {
				if rec.Code != *tt.expectStatus {
					t.Errorf("expected status %d, got %d", *tt.expectStatus, rec.Code)
				}
			}

			if tt.expectLocation != nil {
				location := rec.Header().Get("Location")
				if location != *tt.expectLocation {
					t.Errorf("expected Location %s, got %s", *tt.expectLocation, location)
				}
			}
		})
	}
}
