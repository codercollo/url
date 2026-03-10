DROP INDEX IF EXISTS idx_admins_reset_token;

ALTER TABLE admins
    DROP COLUMN IF EXISTS reset_token,
    DROP COLUMN IF EXISTS reset_token_expires_at;