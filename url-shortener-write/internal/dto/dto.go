package dto

type ShortenURLRequest struct {
	OriginalURL string `json:"original_url"`
	ShortCode   string `json:"short_code"`
}

type ShortenInput struct {
	OriginalURL string
	ShortCode   string
}

type ShortenURLResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
