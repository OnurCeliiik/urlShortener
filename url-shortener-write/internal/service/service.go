package service

import (
	"OnurCeliiik/urlShortener-write/internal/apperr"
	"OnurCeliiik/urlShortener-write/internal/dto"
	"OnurCeliiik/urlShortener-write/internal/model"
	"OnurCeliiik/urlShortener-write/internal/repository"
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
	"regexp"
	"strings"
)

const (
	codeAlphabet      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	generatedCodeLen  = 7
	maxInsertAttempts = 5
	maxOriginalURLLen = 2048
)

var shortCodePattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

type ShortenRepository interface {
	InsertIfAbsent(ctx context.Context, code, originalURL string) (*model.URL, bool, error)
	FindByCode(ctx context.Context, code string) (*model.URL, error)
	FindByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error)
}

type ShortenService struct {
	repo    ShortenRepository
	baseURL string
}

func NewShortenService(repo ShortenRepository, baseURL string) *ShortenService {
	return &ShortenService{
		repo:    repo,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (s *ShortenService) Shorten(ctx context.Context, in dto.ShortenInput) (dto.ShortenURLResponse, bool, error) {
	in.OriginalURL = strings.TrimSpace(in.OriginalURL)
	in.ShortCode = strings.TrimSpace(in.ShortCode)

	if err := validateOriginalURL(in.OriginalURL); err != nil {
		return dto.ShortenURLResponse{}, false, err
	}

	existing, err := s.repo.FindByOriginalURL(ctx, in.OriginalURL)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return dto.ShortenURLResponse{}, false, err
	}
	if existing != nil {
		if in.ShortCode != "" && in.ShortCode != existing.ShortCode {
			return dto.ShortenURLResponse{}, false, apperr.ErrURLAlreadyMapped
		}
		return s.toResponse(existing), false, nil
	}

	if in.ShortCode != "" {
		if err := validateShortCode(in.ShortCode); err != nil {
			return dto.ShortenURLResponse{}, false, err
		}
		created, inserted, err := s.repo.InsertIfAbsent(ctx, in.ShortCode, in.OriginalURL)
		if err != nil {
			return dto.ShortenURLResponse{}, false, err
		}
		return s.toResponse(created), inserted, nil
	}

	for i := 0; i < maxInsertAttempts; i++ {
		code, err := generateShortCode()
		if err != nil {
			return dto.ShortenURLResponse{}, false, err
		}

		created, inserted, err := s.repo.InsertIfAbsent(ctx, code, in.OriginalURL)
		if err != nil {
			if errors.Is(err, apperr.ErrCodeConflict) {
				continue
			}
			return dto.ShortenURLResponse{}, false, err
		}
		return s.toResponse(created), inserted, nil
	}

	return dto.ShortenURLResponse{}, false, apperr.ErrCodeUnavailable
}

func (s *ShortenService) toResponse(u *model.URL) dto.ShortenURLResponse {
	return dto.ShortenURLResponse{
		ShortCode:   u.ShortCode,
		ShortURL:    s.baseURL + "/" + u.ShortCode,
		OriginalURL: u.OriginalURL,
	}
}

func validateOriginalURL(raw string) error {
	if raw == "" || len(raw) > maxOriginalURLLen {
		return apperr.ErrInvalidURL
	}

	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return apperr.ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return apperr.ErrInvalidURL
	}
	if parsed.Host == "" {
		return apperr.ErrInvalidURL
	}

	return nil
}

func validateShortCode(code string) error {
	if len(code) < 4 || len(code) > 10 {
		return apperr.ErrInvalidShortCode
	}
	if !shortCodePattern.MatchString(code) {
		return apperr.ErrInvalidShortCode
	}
	return nil
}

func generateShortCode() (string, error) {
	max := big.NewInt(int64(len(codeAlphabet)))
	buf := make([]byte, generatedCodeLen)
	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = codeAlphabet[n.Int64()]
	}
	return string(buf), nil
}
