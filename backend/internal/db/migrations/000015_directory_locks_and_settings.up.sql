ALTER TABLE user_files
ADD COLUMN lock_password_hash VARCHAR(256),
ADD COLUMN locked_at TIMESTAMPTZ;

CREATE INDEX idx_user_files_locked ON user_files (user_id, id)
WHERE lock_password_hash IS NOT NULL AND is_dir = TRUE AND is_trashed = FALSE;

CREATE TABLE user_directory_unlocks (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    directory_id    BIGINT NOT NULL,
    session_id      VARCHAR(128) NOT NULL,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, directory_id, session_id)
);

CREATE INDEX idx_user_directory_unlocks_lookup
ON user_directory_unlocks (user_id, directory_id, session_id);

CREATE INDEX idx_user_directory_unlocks_expires
ON user_directory_unlocks (expires_at)
WHERE expires_at IS NOT NULL;

CREATE TABLE user_settings (
    user_id    BIGINT PRIMARY KEY,
    settings   JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
