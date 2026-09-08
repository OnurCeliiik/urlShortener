package handler

import (
	"OnurCeliiik/urlShortener-write/internal/apperr"
	"OnurCeliiik/urlShortener-write/internal/dto"
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShortenService interface {
	Shorten(ctx context.Context, in dto.ShortenInput) (dto.ShortenURLResponse, bool, error)
}

type ShortenHandler struct {
	service ShortenService
}

func NewShortenHandler(service ShortenService) *ShortenHandler {
	return &ShortenHandler{service: service}
}

func (h *ShortenHandler) Shorten(c *gin.Context) {
	var req dto.ShortenURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	resp, created, err := h.service.Shorten(c.Request.Context(), dto.ShortenInput{
		OriginalURL: req.OriginalURL,
		ShortCode:   req.ShortCode,
	})
	if err != nil {
		status, msg := mapError(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	c.JSON(status, resp)
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, apperr.ErrInvalidURL), errors.Is(err, apperr.ErrInvalidShortCode):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, apperr.ErrCodeConflict), errors.Is(err, apperr.ErrURLAlreadyMapped):
		return http.StatusConflict, err.Error()
	case errors.Is(err, apperr.ErrCodeUnavailable):
		return http.StatusServiceUnavailable, err.Error()
	default:
		return http.StatusInternalServerError, "failed to shorten url"
	}
}
