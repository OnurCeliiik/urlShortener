package repository

import (
	"OnurCeliiik/urlShortener-read/internal/model"
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("not found")

const selectByCode = `
	SELECT short_code, original_url, created_at
	FROM urls
	WHERE short_code = $1
`

type ResolveRepository struct {
	db *sql.DB
}

func NewResolveRepository(db *sql.DB) *ResolveRepository {
	return &ResolveRepository{db: db}
}

func (r *ResolveRepository) FindByCode(ctx context.Context, code string) (*model.URL, error) {
	var url model.URL
	err := r.db.QueryRowContext(ctx, selectByCode, code).Scan(&url.ShortCode, &url.OriginalURL, &url.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &url, nil
}
