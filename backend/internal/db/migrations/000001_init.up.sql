-- Users: authentication identity and account status only.
CREATE TABLE users (
    id            BIGSERIAL    PRIMARY KEY,
    slug          VARCHAR(21)  NOT NULL UNIQUE,
    username      VARCHAR(64)  NOT NULL UNIQUE,
    email         VARCHAR(256) NOT NULL UNIQUE,
    password_hash VARCHAR(256) NOT NULL,
    status        SMALLINT     NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_slug ON users (slug);
CREATE INDEX idx_users_email ON users (email);

-- User profiles: display info, avatar, bio.
CREATE TABLE user_profiles (
    id           BIGSERIAL    PRIMARY KEY,
    user_id      BIGINT       NOT NULL UNIQUE,
    display_name VARCHAR(64),
    avatar_path  TEXT,
    bio          TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_profiles_user ON user_profiles (user_id);

-- User storage stats: quota and usage, atomic updates.
CREATE TABLE user_storage_stats (
    id            BIGSERIAL    PRIMARY KEY,
    user_id       BIGINT       NOT NULL UNIQUE,
    storage_used  BIGINT       NOT NULL DEFAULT 0,
    storage_quota BIGINT       NOT NULL DEFAULT 10737418240,
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CHECK (storage_used >= 0),
    CHECK (storage_quota >= 0)
);

CREATE INDEX idx_user_storage_stats_user ON user_storage_stats (user_id);

-- User levels: membership tier.
CREATE TABLE user_levels (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     BIGINT       NOT NULL UNIQUE,
    level_code  VARCHAR(32)  NOT NULL DEFAULT 'free',
    level_name  VARCHAR(64)  NOT NULL DEFAULT '免费用户',
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_levels_user ON user_levels (user_id);
CREATE INDEX idx_user_levels_code ON user_levels (level_code);

-- User transactions: payment and adjustment records.
CREATE TABLE user_transactions (
    id              BIGSERIAL    PRIMARY KEY,
    slug            VARCHAR(21)  NOT NULL UNIQUE,
    user_id         BIGINT       NOT NULL,
    transaction_no  VARCHAR(64)  NOT NULL UNIQUE,
    type            VARCHAR(32)  NOT NULL,
    amount_cents    BIGINT       NOT NULL DEFAULT 0,
    currency        VARCHAR(8)   NOT NULL DEFAULT 'CNY',
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
    metadata        JSONB        NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_transactions_user ON user_transactions (user_id, created_at DESC);
CREATE INDEX idx_user_transactions_status ON user_transactions (status);

-- Physical files: deduplicated storage objects.
CREATE TABLE physical_files (
    id           BIGSERIAL    PRIMARY KEY,
    slug         VARCHAR(21)  NOT NULL UNIQUE,
    hash_algo    VARCHAR(16)  NOT NULL DEFAULT 'sha256',
    file_hash    VARCHAR(64)  NOT NULL UNIQUE,
    pre_hash     VARCHAR(64)  NOT NULL,
    file_size    BIGINT       NOT NULL,
    mime_type    VARCHAR(128) NOT NULL DEFAULT 'application/octet-stream',
    storage_path TEXT         NOT NULL,
    status       VARCHAR(20)  NOT NULL DEFAULT 'completed',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_physical_files_hash ON physical_files (hash_algo, file_hash);
CREATE INDEX idx_physical_files_pre ON physical_files (file_size, pre_hash);

-- User files: virtual filesystem entries (files and directories).
CREATE TABLE user_files (
    id               BIGSERIAL    PRIMARY KEY,
    slug             VARCHAR(21)  NOT NULL UNIQUE,
    user_id          BIGINT       NOT NULL,
    physical_file_id BIGINT,
    parent_id        BIGINT,
    file_name        VARCHAR(512) NOT NULL,
    is_dir           BOOLEAN      NOT NULL DEFAULT FALSE,
    file_size        BIGINT       NOT NULL DEFAULT 0,
    mime_type        VARCHAR(128),
    is_starred       BOOLEAN      NOT NULL DEFAULT FALSE,
    is_trashed       BOOLEAN      NOT NULL DEFAULT FALSE,
    trashed_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CHECK (
        (is_dir = TRUE AND physical_file_id IS NULL)
        OR
        (is_dir = FALSE AND physical_file_id IS NOT NULL)
    )
);

CREATE INDEX idx_user_files_user_parent ON user_files (user_id, parent_id)
    WHERE is_trashed = FALSE;
CREATE INDEX idx_user_files_user_starred ON user_files (user_id)
    WHERE is_starred = TRUE AND is_trashed = FALSE;
CREATE INDEX idx_user_files_slug ON user_files (slug);

CREATE UNIQUE INDEX uq_user_files_name_root
ON user_files (user_id, file_name)
WHERE parent_id IS NULL AND is_trashed = FALSE;

CREATE UNIQUE INDEX uq_user_files_name_child
ON user_files (user_id, parent_id, file_name)
WHERE parent_id IS NOT NULL AND is_trashed = FALSE;

-- Upload tasks: chunked upload sessions for physical files.
CREATE TABLE upload_tasks (
    id               BIGSERIAL    PRIMARY KEY,
    slug             VARCHAR(21)  NOT NULL UNIQUE,
    owner_user_id    BIGINT       NOT NULL,
    hash_algo        VARCHAR(16)  NOT NULL DEFAULT 'sha256',
    file_hash        VARCHAR(64)  NOT NULL,
    pre_hash         VARCHAR(64)  NOT NULL,
    file_size        BIGINT       NOT NULL,
    mime_type        VARCHAR(128) NOT NULL DEFAULT 'application/octet-stream',
    total_chunks     INT          NOT NULL,
    chunk_size       INT          NOT NULL DEFAULT 4194304,
    status           VARCHAR(20)  NOT NULL DEFAULT 'created',
    physical_file_id BIGINT,
    error_msg        TEXT,
    expires_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW() + INTERVAL '7 days',
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_upload_tasks_owner ON upload_tasks (owner_user_id, status);
CREATE INDEX idx_upload_tasks_hash ON upload_tasks (hash_algo, file_hash);

-- Refresh tokens: DB-backed token storage with revocation.
CREATE TABLE refresh_tokens (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    BIGINT       NOT NULL,
    token_hash VARCHAR(64)  NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ  NOT NULL,
    revoked    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id)
    WHERE revoked = FALSE;

-- Media transcodes: shared HLS transcoding products.
CREATE TABLE media_transcodes (
    id               BIGSERIAL    PRIMARY KEY,
    slug             VARCHAR(21)  NOT NULL UNIQUE,
    physical_file_id BIGINT       NOT NULL,
    profile          VARCHAR(32)  NOT NULL DEFAULT 'hls_1080p',
    status           VARCHAR(20)  NOT NULL DEFAULT 'pending',
    hls_dir          TEXT,
    duration_sec     INT,
    error_msg        TEXT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_media_transcodes_file_profile
ON media_transcodes (physical_file_id, profile);

CREATE INDEX idx_media_transcodes_status ON media_transcodes (status)
    WHERE status IN ('pending', 'processing');

-- Media items: user's media library entries.
CREATE TABLE media_items (
    id               BIGSERIAL    PRIMARY KEY,
    slug             VARCHAR(21)  NOT NULL UNIQUE,
    user_id          BIGINT       NOT NULL,
    user_file_id     BIGINT       NOT NULL,
    physical_file_id BIGINT       NOT NULL,
    transcode_id     BIGINT,
    title            VARCHAR(512) NOT NULL,
    poster_path      TEXT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_media_items_user_file ON media_items (user_id, user_file_id);
CREATE INDEX idx_media_items_user ON media_items (user_id, created_at DESC);
CREATE INDEX idx_media_items_transcode ON media_items (transcode_id);

-- Media jobs: transcoding task queue.
CREATE TABLE media_jobs (
    id            BIGSERIAL    PRIMARY KEY,
    slug          VARCHAR(21)  NOT NULL UNIQUE,
    transcode_id  BIGINT       NOT NULL,
    status        VARCHAR(20)  NOT NULL DEFAULT 'pending',
    error_msg     TEXT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_media_jobs_transcode_active
ON media_jobs (transcode_id)
WHERE status IN ('pending', 'processing');

CREATE INDEX idx_media_jobs_status ON media_jobs (status)
    WHERE status IN ('pending', 'processing');
