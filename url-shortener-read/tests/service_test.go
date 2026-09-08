package tests

import (
	"OnurCeliiik/urlShortener-read/internal/apperr"
	"OnurCeliiik/urlShortener-read/internal/dto"
	"OnurCeliiik/urlShortener-read/internal/model"
	"OnurCeliiik/urlShortener-read/internal/repository"
	"OnurCeliiik/urlShortener-read/internal/service"
	"context"
	"sync"
	"testing"
	"time"
)

type fakeRepo struct {
	mu     sync.Mutex
	byCode map[string]*model.URL
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byCode: make(map[string]*model.URL)}
}

func (f *fakeRepo) seed(code, originalURL string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byCode[code] = &model.URL{
		ShortCode:   code,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}
}

func (f *fakeRepo) FindByCode(_ context.Context, code string) (*model.URL, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	existing, ok := f.byCode[code]
	if !ok {
		return nil, repository.ErrNotFound
	}
	copy := *existing
	return &copy, nil
}

func TestResolveReturnsOriginalURL(t *testing.T) {
	repo := newFakeRepo()
	repo.seed("docs42", "https://example.com/docs")
	svc := service.NewResolveService(repo)

	got, err := svc.Resolve(context.Background(), dto.ResolveInput{ShortCode: "docs42"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com/docs" {
		t.Fatalf("got %s", got)
	}
}

func TestResolveNotFound(t *testing.T) {
	svc := service.NewResolveService(newFakeRepo())

	_, err := svc.Resolve(context.Background(), dto.ResolveInput{ShortCode: "missing1"})
	if err != apperr.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestResolveValidatesShortCode(t *testing.T) {
	svc := service.NewResolveService(newFakeRepo())

	if _, err := svc.Resolve(context.Background(), dto.ResolveInput{ShortCode: "ab"}); err != apperr.ErrInvalidShortCode {
		t.Fatalf("expected invalid short code, got %v", err)
	}

	if _, err := svc.Resolve(context.Background(), dto.ResolveInput{ShortCode: "bad_code!"}); err != apperr.ErrInvalidShortCode {
		t.Fatalf("expected invalid short code, got %v", err)
	}
}

func TestResolveTrimsShortCode(t *testing.T) {
	repo := newFakeRepo()
	repo.seed("docs42", "https://example.com/docs")
	svc := service.NewResolveService(repo)

	got, err := svc.Resolve(context.Background(), dto.ResolveInput{ShortCode: "  docs42  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com/docs" {
		t.Fatalf("got %s", got)
	}
}
