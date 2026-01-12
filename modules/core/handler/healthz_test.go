package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestHealthzHandler(t *testing.T) {
	tests := []struct {
		name         string
		checkError   error
		expectStatus *int
	}{
		{
			name:         "healthy_returns_200",
			checkError:   nil,
			expectStatus: ptr(http.StatusOK),
		},
		{
			name:         "unhealthy_returns_503",
			checkError:   errors.New("db connection failed"),
			expectStatus: ptr(http.StatusServiceUnavailable),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			handler := NewHealthzHandler(func() error {
				return tt.checkError
			})

			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = handler.Handle(c)

			if tt.expectStatus != nil {
				if rec.Code != *tt.expectStatus {
					t.Errorf("expected status %d, got %d", *tt.expectStatus, rec.Code)
				}
			}
		})
	}
}
