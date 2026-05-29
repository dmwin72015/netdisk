DROP INDEX IF EXISTS idx_user_files_category;
ALTER TABLE user_files DROP CONSTRAINT IF EXISTS chk_file_category;
ALTER TABLE user_files DROP COLUMN IF EXISTS file_category;
