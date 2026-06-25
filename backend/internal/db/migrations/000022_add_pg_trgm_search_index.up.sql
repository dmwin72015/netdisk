CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_user_files_name_trgm ON user_files USING gin (file_name gin_trgm_ops);
