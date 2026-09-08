package handler

import (
	"OnurCeliiik/urlShortener-read/internal/apperr"
	"OnurCeliiik/urlShortener-read/internal/audit"
	"OnurCeliiik/urlShortener-read/internal/dto"
	"OnurCeliiik/urlShortener-read/internal/obs"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ResolveService interface {
	Resolve(ctx context.Context, in dto.ResolveInput) (string, error)
}

type ResolveHandler struct {
	service ResolveService
	audit   *audit.Emitter
}

func NewResolveHandler(service ResolveService, auditor *audit.Emitter) *ResolveHandler {
	return &ResolveHandler{
		service: service,
		audit:   auditor,
	}
}

func (h *ResolveHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	originalURL, err := h.service.Resolve(c.Request.Context(), dto.ResolveInput{
		ShortCode: code,
	})
	if err != nil {
		status, msg := mapError(err)
		h.emit(c, status, code, "", msg)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	h.emit(c, http.StatusFound, code, originalURL, "")
	c.Redirect(http.StatusFound, originalURL)
}

func (h *ResolveHandler) emit(c *gin.Context, status int, shortCode, originalURL, errMsg string) {
	h.audit.Emit(audit.Event{
		Timestamp:   time.Now().UTC(),
		Service:     "read",
		Action:      "resolve",
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
	case errors.Is(err, apperr.ErrInvalidShortCode):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, apperr.ErrNotFound):
		return http.StatusNotFound, err.Error()
	default:
		return http.StatusInternalServerError, "failed to resolve url"
	}
}
