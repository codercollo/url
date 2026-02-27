CREATE TABLE IF NOT EXISTS clicks (
    id          SERIAL PRIMARY KEY,
    short_code  VARCHAR(20) NOT NULL,
    clicked_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    ip_address  VARCHAR(50),
    referrer    TEXT,
    user_agent  TEXT,
    CONSTRAINT fk_clicks_short_code
        FOREIGN KEY (short_code) REFERENCES urls(short_code)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_clicks_short_code ON clicks(short_code);
CREATE INDEX IF NOT EXISTS idx_clicks_clicked_at ON clicks(clicked_at);