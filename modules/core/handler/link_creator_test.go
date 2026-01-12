package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nanda/doit/modules/core/internal/test/mocks"
	"github.com/nanda/doit/modules/core/service"
	"go.uber.org/mock/gomock"
)

func ptr[T any](v T) *T {
	return &v
}

func TestLinkCreatorHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockReturn     string
		mockError      error
		expectStatus   *int
		expectContains *string
		expectTTL      *int64
	}{
		{
			name:         "successful_creation_returns_200",
			requestBody:  `{"long_url":"https://example.com"}`,
			mockReturn:   "abc123",
			mockError:    nil,
			expectStatus: ptr(http.StatusOK),
		},
		{
			name:           "successful_creation_returns_short_code",
			requestBody:    `{"long_url":"https://example.com"}`,
			mockReturn:     "abc123",
			mockError:      nil,
			expectContains: ptr("abc123"),
		},
		{
			name:         "missing_long_url_returns_400",
			requestBody:  `{}`,
			mockReturn:   "",
			mockError:    nil,
			expectStatus: ptr(http.StatusBadRequest),
		},
		{
			name:         "invalid_url_error_returns_400",
			requestBody:  `{"long_url":"not-a-url"}`,
			mockReturn:   "",
			mockError:    service.ErrInvalidURL,
			expectStatus: ptr(http.StatusBadRequest),
		},
		{
			name:         "url_too_long_error_returns_400",
			requestBody:  `{"long_url":"https://example.com"}`,
			mockReturn:   "",
			mockError:    service.ErrURLTooLong,
			expectStatus: ptr(http.StatusBadRequest),
		},
		{
			name:         "invalid_ttl_error_returns_400",
			requestBody:  `{"long_url":"https://example.com","ttl_seconds":60}`,
			mockReturn:   "",
			mockError:    service.ErrInvalidTTL,
			expectStatus: ptr(http.StatusBadRequest),
		},
		{
			name:        "custom_ttl_is_passed_to_service",
			requestBody: `{"long_url":"https://example.com","ttl_seconds":7200}`,
			mockReturn:  "def456",
			mockError:   nil,
			expectTTL:   ptr(int64(7200)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			e := echo.New()
			mockService := mocks.NewMockLinkCreator(ctrl)

			// Setup expectations based on test case
			if tt.mockReturn != "" || tt.mockError != nil {
				if tt.expectTTL != nil {
					mockService.EXPECT().
						Create(gomock.Any(), gomock.Any(), gomock.Eq(tt.expectTTL)).
						Return(tt.mockReturn, tt.mockError).
						Times(1)
				} else {
					mockService.EXPECT().
						Create(gomock.Any(), gomock.Any(), gomock.Any()).
						Return(tt.mockReturn, tt.mockError).
						MaxTimes(1)
				}
			}

			handler := NewLinkCreatorHandler(mockService)

			req := httptest.NewRequest(http.MethodPost, "/s", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

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
