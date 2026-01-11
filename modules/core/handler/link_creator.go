package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nanda/doit/modules/core/service"
)

type CreateLinkRequest struct {
	LongURL    string `json:"long_url"`
	TTLSeconds *int64 `json:"ttl_seconds,omitempty"`
}

type CreateLinkResponse struct {
	ShortCode string `json:"short_code"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type LinkCreatorHandler struct {
	service service.LinkCreator
}

func NewLinkCreatorHandler(svc service.LinkCreator) *LinkCreatorHandler {
	return &LinkCreatorHandler{service: svc}
}

func (h *LinkCreatorHandler) Handle(c echo.Context) error {
	var req CreateLinkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.LongURL == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "long_url is required"})
	}

	shortCode, err := h.service.Create(c.Request().Context(), req.LongURL, req.TTLSeconds)
	if err != nil {
		return handleServiceError(c, err)
	}

	return c.JSON(http.StatusOK, CreateLinkResponse{ShortCode: shortCode})
}

func handleServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidURL):
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrURLTooLong):
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrInvalidTTL):
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
