package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthzHandler struct {
	checks []HealthCheck
}

type HealthCheck func() error

type HealthResponse struct {
	Status string `json:"status"`
}

func NewHealthzHandler(checks ...HealthCheck) *HealthzHandler {
	return &HealthzHandler{checks: checks}
}

func (h *HealthzHandler) Handle(c echo.Context) error {
	for _, check := range h.checks {
		if err := check(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, HealthResponse{Status: "unhealthy"})
		}
	}

	return c.JSON(http.StatusOK, HealthResponse{Status: "healthy"})
}
