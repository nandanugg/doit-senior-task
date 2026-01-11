-- Recreate URLs table if rolling back
CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    long_url TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_urls_id_brin ON urls USING BRIN (id);
CREATE INDEX idx_urls_expires_at_brin ON urls USING BRIN (expires_at);
ALTER SEQUENCE urls_id_seq CACHE 100;
