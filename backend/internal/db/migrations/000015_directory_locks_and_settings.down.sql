DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS user_directory_unlocks;
DROP INDEX IF EXISTS idx_user_files_locked;

ALTER TABLE user_files
DROP COLUMN IF EXISTS locked_at,
DROP COLUMN IF EXISTS lock_password_hash;
