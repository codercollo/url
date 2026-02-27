-- Rollback: drop urls table and its indexes
DROP INDEX IF EXISTS idx_urls_original_url;
DROP INDEX IF EXISTS idx_urls_short_code;
DROP TABLE IF EXISTS urls;