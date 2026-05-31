ALTER TABLE user_files
    ADD COLUMN is_system BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN system_kind VARCHAR(64);

ALTER TABLE user_files
    ADD CONSTRAINT chk_user_files_system_kind
    CHECK (
        (is_system = FALSE AND system_kind IS NULL)
        OR
        (is_system = TRUE AND system_kind IS NOT NULL AND is_dir = TRUE)
    );

DROP INDEX IF EXISTS uq_user_files_name_root;
DROP INDEX IF EXISTS uq_user_files_name_child;

CREATE UNIQUE INDEX uq_user_files_normal_name_root
ON user_files (user_id, file_name)
WHERE parent_id IS NULL AND is_trashed = FALSE AND is_system = FALSE;

CREATE UNIQUE INDEX uq_user_files_normal_name_child
ON user_files (user_id, parent_id, file_name)
WHERE parent_id IS NOT NULL AND is_trashed = FALSE AND is_system = FALSE;

CREATE UNIQUE INDEX uq_user_files_system_name_root
ON user_files (user_id, file_name)
WHERE parent_id IS NULL AND is_trashed = FALSE AND is_system = TRUE;

CREATE UNIQUE INDEX uq_user_files_system_name_child
ON user_files (user_id, parent_id, file_name)
WHERE parent_id IS NOT NULL AND is_trashed = FALSE AND is_system = TRUE;

CREATE UNIQUE INDEX uq_user_files_system_kind_root
ON user_files (user_id, system_kind)
WHERE parent_id IS NULL AND is_trashed = FALSE AND is_system = TRUE;

CREATE UNIQUE INDEX uq_user_files_system_kind_child
ON user_files (user_id, parent_id, system_kind)
WHERE parent_id IS NOT NULL AND is_trashed = FALSE AND is_system = TRUE;
