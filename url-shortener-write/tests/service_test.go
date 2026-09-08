package tests

import (
	"OnurCeliiik/urlShortener-write/internal/apperr"
	"OnurCeliiik/urlShortener-write/internal/dto"
	"OnurCeliiik/urlShortener-write/internal/model"
	"OnurCeliiik/urlShortener-write/internal/repository"
	"OnurCeliiik/urlShortener-write/internal/service"
	"context"
	"sync"
	"testing"
	"time"
)

type fakeRepo struct {
	mu     sync.Mutex
	byCode map[string]*model.URL
	byURL  map[string]*model.URL
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		byCode: make(map[string]*model.URL),
		byURL:  make(map[string]*model.URL),
	}
}

func (f *fakeRepo) InsertIfAbsent(_ context.Context, code, originalURL string) (*model.URL, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if existing, ok := f.byCode[code]; ok {
		if existing.OriginalURL == originalURL {
			copy := *existing
			return &copy, false, nil
		}
		return nil, false, apperr.ErrCodeConflict
	}

	if existing, ok := f.byURL[originalURL]; ok {
		copy := *existing
		return &copy, false, nil
	}

	created := &model.URL{
		ShortCode:   code,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}
	f.byCode[code] = created
	f.byURL[originalURL] = created

	copy := *created
	return &copy, true, nil
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

func (f *fakeRepo) FindByOriginalURL(_ context.Context, originalURL string) (*model.URL, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	existing, ok := f.byURL[originalURL]
	if !ok {
		return nil, repository.ErrNotFound
	}
	copy := *existing
	return &copy, nil
}

func TestShortenGeneratesCode(t *testing.T) {
	svc := service.NewShortenService(newFakeRepo(), "http://localhost:8080")

	resp, created, err := svc.Shorten(context.Background(), dto.ShortenInput{
		OriginalURL: "https://example.com/a",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected a new mapping to be created")
	}
	if resp.ShortCode == "" {
		t.Fatal("expected a generated short code")
	}
	if resp.ShortURL != "http://localhost:8080/"+resp.ShortCode {
		t.Fatalf("unexpected short url: %s", resp.ShortURL)
	}
}

func TestShortenCustomCode(t *testing.T) {
	svc := service.NewShortenService(newFakeRepo(), "http://localhost:8080")

	resp, created, err := svc.Shorten(context.Background(), dto.ShortenInput{
		OriginalURL: "https://example.com/b",
		ShortCode:   "docs42",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected a new mapping to be created")
	}
	if resp.ShortCode != "docs42" {
		t.Fatalf("got short code %s", resp.ShortCode)
	}
}

func TestShortenIdempotentForSameURL(t *testing.T) {
	svc := service.NewShortenService(newFakeRepo(), "http://localhost:8080")
	in := dto.ShortenInput{OriginalURL: "https://example.com/c"}

	first, _, err := svc.Shorten(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, created, err := svc.Shorten(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created {
		t.Fatal("expected existing mapping to be reused")
	}
	if first.ShortCode != second.ShortCode {
		t.Fatalf("expected the same short code, got %s and %s", first.ShortCode, second.ShortCode)
	}
}

func TestShortenRejectsCodeConflict(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewShortenService(repo, "http://localhost:8080")

	if _, _, err := svc.Shorten(context.Background(), dto.ShortenInput{
		OriginalURL: "https://example.com/d",
		ShortCode:   "taken1",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _, err := svc.Shorten(context.Background(), dto.ShortenInput{
		OriginalURL: "https://example.com/e",
		ShortCode:   "taken1",
	})
	if err != apperr.ErrCodeConflict {
		t.Fatalf("expected code conflict, got %v", err)
	}
}

func TestShortenRejectsURLAlreadyMapped(t *testing.T) {
	svc := service.NewShortenService(newFakeRepo(), "http://localhost:8080")

	if _, _, err := svc.Shorten(context.Background(), dto.ShortenInput{
		OriginalURL: "https://example.com/f",
		ShortCode:   "first1",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _, err := svc.Shorten(context.Background(), dto.ShortenInput{
		OriginalURL: "https://example.com/f",
		ShortCode:   "other1",
	})
	if err != apperr.ErrURLAlreadyMapped {
		t.Fatalf("expected url already mapped, got %v", err)
	}
}

func TestShortenValidatesInput(t *testing.T) {
	svc := service.NewShortenService(newFakeRepo(), "http://localhost:8080")

	if _, _, err := svc.Shorten(context.Background(), dto.ShortenInput{OriginalURL: "not-a-url"}); err != apperr.ErrInvalidURL {
		t.Fatalf("expected invalid url, got %v", err)
	}

	if _, _, err := svc.Shorten(context.Background(), dto.ShortenInput{
		OriginalURL: "https://example.com/g",
		ShortCode:   "ab",
	}); err != apperr.ErrInvalidShortCode {
		t.Fatalf("expected invalid short code, got %v", err)
	}
}
