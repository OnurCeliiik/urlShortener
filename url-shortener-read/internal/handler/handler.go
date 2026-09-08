package handler

import (
	"OnurCeliiik/urlShortener-read/internal/apperr"
	"OnurCeliiik/urlShortener-read/internal/dto"
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResolveService interface {
	Resolve(ctx context.Context, in dto.ResolveInput) (string, error)
}

type ResolveHandler struct {
	service ResolveService
}

func NewResolveHandler(service ResolveService) *ResolveHandler {
	return &ResolveHandler{service: service}
}

func (h *ResolveHandler) Redirect(c *gin.Context) {
	originalURL, err := h.service.Resolve(c.Request.Context(), dto.ResolveInput{
		ShortCode: c.Param("code"),
	})
	if err != nil {
		status, msg := mapError(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, apperr.ErrInvalidShortCode):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, apperr.ErrNotFound):
		return http.StatusNotFound, err.Error()
	default:
		return http.StatusInternalServerError, "failed to resolve url"
	}
}
