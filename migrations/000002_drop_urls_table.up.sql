-- Drop URLs table as we're moving to Redis for URL storage
-- Analytics table remains in PostgreSQL
DROP INDEX IF EXISTS idx_urls_expires_at_brin;
DROP INDEX IF EXISTS idx_urls_id_brin;
DROP TABLE IF EXISTS urls;
