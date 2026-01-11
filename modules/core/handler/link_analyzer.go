package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nanda/doit/modules/core/service"
)

type AnalyzeResponse struct {
	LongURL        string  `json:"long_url"`
	CreatedAt      string  `json:"created_at"`
	ExpiresAt      string  `json:"expires_at"`
	ClickCount     int64   `json:"click_count"`
	LastAccessedAt *string `json:"last_accessed_at"`
}

type LinkAnalyzerHandler struct {
	service service.LinkAnalyzer
}

func NewLinkAnalyzerHandler(svc service.LinkAnalyzer) *LinkAnalyzerHandler {
	return &LinkAnalyzerHandler{service: svc}
}

func (h *LinkAnalyzerHandler) Handle(c echo.Context) error {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "short_code is required"})
	}

	analytic, err := h.service.Analyze(c.Request().Context(), shortCode)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	var lastAccessedAt *string
	if analytic.LastAccessedAt != nil {
		formatted := analytic.LastAccessedAt.Format(time.RFC3339)
		lastAccessedAt = &formatted
	}

	return c.JSON(http.StatusOK, AnalyzeResponse{
		LongURL:        analytic.LongURL,
		CreatedAt:      analytic.CreatedAt.Format(time.RFC3339),
		ExpiresAt:      analytic.ExpiresAt.Format(time.RFC3339),
		ClickCount:     analytic.ClickCount,
		LastAccessedAt: lastAccessedAt,
	})
}
