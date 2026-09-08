CREATE TABLE IF NOT EXISTS urls (
    short_code VARCHAR(10) PRIMARY KEY,
    original_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT urls_original_url_unique UNIQUE (original_url)
);

CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls (created_at);
