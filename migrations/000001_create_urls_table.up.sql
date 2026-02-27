CREATE TABLE IF NOT EXISTS urls (
    id           SERIAL PRIMARY KEY,
    original_url TEXT NOT NULL,
    short_code   VARCHAR(20) UNIQUE NOT NULL,
    created_by   VARCHAR(100),
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMP,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);