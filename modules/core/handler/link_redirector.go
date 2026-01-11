package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nanda/doit/modules/core/service"
)

type LinkRedirectorHandler struct {
	service service.LinkRedirector
}

func NewLinkRedirectorHandler(svc service.LinkRedirector) *LinkRedirectorHandler {
	return &LinkRedirectorHandler{service: svc}
}

func (h *LinkRedirectorHandler) Handle(c echo.Context) error {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "short_code is required"})
	}

	longURL, err := h.service.Redirect(c.Request().Context(), shortCode)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}

	return c.Redirect(http.StatusFound, longURL)
}
