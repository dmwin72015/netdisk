-- Remove parent_slug column
DROP INDEX IF EXISTS idx_user_files_parent_slug;
ALTER TABLE user_files DROP COLUMN IF EXISTS parent_slug;
