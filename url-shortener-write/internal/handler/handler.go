package handler

import (
	"OnurCeliiik/urlShortener-write/internal/apperr"
	"OnurCeliiik/urlShortener-write/internal/audit"
	"OnurCeliiik/urlShortener-write/internal/dto"
	"OnurCeliiik/urlShortener-write/internal/obs"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ShortenService interface {
	Shorten(ctx context.Context, in dto.ShortenInput) (dto.ShortenURLResponse, bool, error)
}

type ShortenHandler struct {
	service ShortenService
	audit   *audit.Emitter
}

func NewShortenHandler(service ShortenService, auditor *audit.Emitter) *ShortenHandler {
	return &ShortenHandler{
		service: service,
		audit:   auditor,
	}
}

func (h *ShortenHandler) Shorten(c *gin.Context) {
	var req dto.ShortenURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.emit(c, http.StatusBadRequest, req.ShortCode, req.OriginalURL, "invalid JSON body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	resp, created, err := h.service.Shorten(c.Request.Context(), dto.ShortenInput{
		OriginalURL: req.OriginalURL,
		ShortCode:   req.ShortCode,
	})
	if err != nil {
		status, msg := mapError(err)
		h.emit(c, status, req.ShortCode, req.OriginalURL, msg)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	h.emit(c, status, resp.ShortCode, resp.OriginalURL, "")
	c.JSON(status, resp)
}

func (h *ShortenHandler) emit(c *gin.Context, status int, shortCode, originalURL, errMsg string) {
	h.audit.Emit(audit.Event{
		Timestamp:   time.Now().UTC(),
		Service:     "write",
		Action:      "shorten",
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		Status:      status,
		LatencyMS:   obs.ElapsedMS(c),
		RequestID:   obs.RequestIDFrom(c),
		Method:      c.Request.Method,
		Path:        c.Request.URL.Path,
		Error:       errMsg,
	})
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
