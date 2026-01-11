-- URLs table with BRIN index for sequential ID and expires_at
CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    long_url TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

-- BRIN index on id (sequential, ideal for BRIN)
CREATE INDEX idx_urls_id_brin ON urls USING BRIN (id);

-- BRIN index on expires_at for efficient expiration queries
CREATE INDEX idx_urls_expires_at_brin ON urls USING BRIN (expires_at);

-- URL Analytics table
CREATE TABLE url_analytics (
    id BIGSERIAL PRIMARY KEY,
    url_id BIGINT NOT NULL,
    long_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    click_count BIGINT NOT NULL DEFAULT 0,
    last_accessed_at TIMESTAMPTZ
);

-- BRIN index on id (sequential)
CREATE INDEX idx_url_analytics_id_brin ON url_analytics USING BRIN (id);

-- BRIN index on url_id for lookups
CREATE INDEX idx_url_analytics_url_id_brin ON url_analytics USING BRIN (url_id);

-- Sequence cache for better insert performance
ALTER SEQUENCE urls_id_seq CACHE 100;
ALTER SEQUENCE url_analytics_id_seq CACHE 100;
