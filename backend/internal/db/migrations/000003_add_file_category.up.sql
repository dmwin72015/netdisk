-- Add file_category column to user_files
ALTER TABLE user_files
ADD COLUMN file_category VARCHAR(20) NOT NULL DEFAULT 'other';

-- Backfill existing rows based on mime_type
UPDATE user_files SET file_category = 'folder' WHERE is_dir = TRUE;
UPDATE user_files SET file_category = 'video'  WHERE mime_type LIKE 'video/%';
UPDATE user_files SET file_category = 'audio'  WHERE mime_type LIKE 'audio/%';
UPDATE user_files SET file_category = 'image'  WHERE mime_type LIKE 'image/%';
UPDATE user_files SET file_category = 'document' WHERE mime_type LIKE 'text/%'
    OR mime_type LIKE 'application/pdf'
    OR mime_type LIKE 'application/msword'
    OR mime_type LIKE 'application/vnd.openxmlformats-officedocument.%'
    OR mime_type LIKE 'application/vnd.oasis.opendocument.%'
    OR mime_type LIKE 'application/vnd.ms-%';
UPDATE user_files SET file_category = 'archive' WHERE mime_type LIKE 'application/zip'
    OR mime_type LIKE 'application/x-rar%'
    OR mime_type LIKE 'application/x-tar%'
    OR mime_type LIKE 'application/gzip'
    OR mime_type LIKE 'application/x-7z%'
    OR mime_type LIKE 'application/x-bzip%';

-- CHECK constraint
ALTER TABLE user_files
ADD CONSTRAINT chk_file_category
CHECK (file_category IN ('folder','video','audio','image','document','archive','other'));

-- Index for efficient category filtering
CREATE INDEX idx_user_files_category ON user_files (user_id, file_category)
WHERE is_trashed = FALSE;
