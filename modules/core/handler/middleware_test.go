package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestProcessingTimeMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		expectHeaderExists *bool
		expectHeaderValid  *bool
	}{
		{
			name:               "header_is_set",
			expectHeaderExists: ptr(true),
		},
		{
			name:              "header_is_valid_integer",
			expectHeaderValid: ptr(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Use(ProcessingTimeMiddleware())

			e.GET("/test", func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			header := rec.Header().Get("X-Processing-Time-Micros")

			if tt.expectHeaderExists != nil {
				if (header != "") != *tt.expectHeaderExists {
					t.Error("expected X-Processing-Time-Micros header to be set")
				}
			}

			if tt.expectHeaderValid != nil {
				_, err := strconv.ParseInt(header, 10, 64)
				if (err == nil) != *tt.expectHeaderValid {
					t.Errorf("X-Processing-Time-Micros should be a valid integer: %v", err)
				}
			}
		})
	}
}
