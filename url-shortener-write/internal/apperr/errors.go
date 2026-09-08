package apperr

import "errors"

var (
	ErrInvalidURL       = errors.New("original_url must be a valid http or https URL")
	ErrInvalidShortCode = errors.New("short_code must be 4-10 alphanumeric characters")
	ErrCodeConflict     = errors.New("short code already mapped to a different URL")
	ErrURLAlreadyMapped = errors.New("original URL is already mapped to a different short code")
	ErrCodeUnavailable  = errors.New("unable to allocate a unique short code")
)
