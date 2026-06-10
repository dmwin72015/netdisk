CREATE TABLE file_shares (
    id            BIGSERIAL   PRIMARY KEY,
    slug          VARCHAR(21) NOT NULL UNIQUE,
    user_id       BIGINT      NOT NULL,
    file_id       BIGINT      NOT NULL,
    password_code VARCHAR(16),
    expires_at    TIMESTAMPTZ,
    disabled_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_file_shares_user_created ON file_shares (user_id, created_at DESC);
CREATE INDEX idx_file_shares_file ON file_shares (file_id);
CREATE INDEX idx_file_shares_active_slug ON file_shares (slug)
    WHERE disabled_at IS NULL;
