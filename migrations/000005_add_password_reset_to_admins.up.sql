ALTER TABLE admins
    ADD COLUMN reset_token TEXT,
    ADD COLUMN reset_token_expires_at TIMESTAMPTZ;

CREATE INDEX idx_admins_reset_token ON admins(reset_token);