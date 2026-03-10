ALTER TABLE admins
    DROP COLUMN IF EXISTS activation_token,
    DROP COLUMN IF EXISTS token_expires_at,
    DROP COLUMN IF EXISTS is_active;