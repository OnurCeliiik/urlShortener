package repository

import (
	"OnurCeliiik/urlShortener-write/internal/apperr"
	"OnurCeliiik/urlShortener-write/internal/model"
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNotFound = errors.New("not found")

const (
	insertURL = `
		INSERT INTO urls (short_code, original_url)
		VALUES ($1, $2)
		RETURNING short_code, original_url, created_at
	`
	selectByCode = `
		SELECT short_code, original_url, created_at
		FROM urls
		WHERE short_code = $1
	`
	selectByOriginalURL = `
		SELECT short_code, original_url, created_at
		FROM urls
		WHERE original_url = $1
	`
)

type ShortenRepository struct {
	db *sql.DB
}

func NewShortenRepository(db *sql.DB) *ShortenRepository {
	return &ShortenRepository{db: db}
}

func (r *ShortenRepository) InsertIfAbsent(ctx context.Context, code, originalURL string) (*model.URL, bool, error) {
	created, err := scanURL(r.db.QueryRowContext(ctx, insertURL, code, originalURL))
	if err == nil {
		return created, true, nil
	}
	if !isUniqueViolation(err) {
		return nil, false, err
	}

	existing, findErr := r.FindByCode(ctx, code)
	if findErr == nil {
		if existing.OriginalURL == originalURL {
			return existing, false, nil
		}
		return nil, false, apperr.ErrCodeConflict
	}
	if !errors.Is(findErr, ErrNotFound) {
		return nil, false, findErr
	}

	byURL, findErr := r.FindByOriginalURL(ctx, originalURL)
	if findErr != nil {
		return nil, false, err
	}
	return byURL, false, nil
}

func (r *ShortenRepository) FindByCode(ctx context.Context, code string) (*model.URL, error) {
	url, err := scanURL(r.db.QueryRowContext(ctx, selectByCode, code))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return url, err
}

func (r *ShortenRepository) FindByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error) {
	url, err := scanURL(r.db.QueryRowContext(ctx, selectByOriginalURL, originalURL))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return url, err
}

func scanURL(row *sql.Row) (*model.URL, error) {
	var url model.URL
	if err := row.Scan(&url.ShortCode, &url.OriginalURL, &url.CreatedAt); err != nil {
		return nil, err
	}
	return &url, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
