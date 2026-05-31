DROP INDEX IF EXISTS uq_user_files_system_kind_child;
DROP INDEX IF EXISTS uq_user_files_system_kind_root;
DROP INDEX IF EXISTS uq_user_files_system_name_child;
DROP INDEX IF EXISTS uq_user_files_system_name_root;
DROP INDEX IF EXISTS uq_user_files_normal_name_child;
DROP INDEX IF EXISTS uq_user_files_normal_name_root;

ALTER TABLE user_files
    DROP CONSTRAINT IF EXISTS chk_user_files_system_kind;

ALTER TABLE user_files
    DROP COLUMN IF EXISTS system_kind,
    DROP COLUMN IF EXISTS is_system;

CREATE UNIQUE INDEX uq_user_files_name_root
ON user_files (user_id, file_name)
WHERE parent_id IS NULL AND is_trashed = FALSE;

CREATE UNIQUE INDEX uq_user_files_name_child
ON user_files (user_id, parent_id, file_name)
WHERE parent_id IS NOT NULL AND is_trashed = FALSE;
