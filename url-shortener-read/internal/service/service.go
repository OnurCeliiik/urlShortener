package service

import (
	"OnurCeliiik/urlShortener-read/internal/apperr"
	"OnurCeliiik/urlShortener-read/internal/dto"
	"OnurCeliiik/urlShortener-read/internal/model"
	"OnurCeliiik/urlShortener-read/internal/repository"
	"context"
	"errors"
	"regexp"
	"strings"
)

var shortCodePattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

type ResolveRepository interface {
	FindByCode(ctx context.Context, code string) (*model.URL, error)
}

type ResolveService struct {
	repo ResolveRepository
}

func NewResolveService(repo ResolveRepository) *ResolveService {
	return &ResolveService{repo: repo}
}

func (s *ResolveService) Resolve(ctx context.Context, in dto.ResolveInput) (string, error) {
	code := strings.TrimSpace(in.ShortCode)
	if err := validateShortCode(code); err != nil {
		return "", err
	}

	url, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", apperr.ErrNotFound
		}
		return "", err
	}

	return url.OriginalURL, nil
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
