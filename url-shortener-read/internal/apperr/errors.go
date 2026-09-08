package apperr

import "errors"

var (
	ErrInvalidShortCode = errors.New("short_code must be 4-10 alphanumeric characters")
	ErrNotFound         = errors.New("short code not found")
)
