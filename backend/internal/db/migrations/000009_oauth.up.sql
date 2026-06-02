-- Add register_method to users table
ALTER TABLE users ADD COLUMN register_method VARCHAR(16) NOT NULL DEFAULT 'email';

-- OAuth-linked accounts
CREATE TABLE user_oauth_accounts (
    id                   BIGSERIAL    PRIMARY KEY,
    user_id              BIGINT       NOT NULL,
    provider             VARCHAR(32)  NOT NULL,
    provider_account_id  VARCHAR(256) NOT NULL,
    access_token         TEXT         NOT NULL DEFAULT '',
    refresh_token        TEXT         NOT NULL DEFAULT '',
    token_expires_at     TIMESTAMPTZ,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_account_id)
);

CREATE INDEX idx_user_oauth_accounts_user_id ON user_oauth_accounts(user_id);
CREATE INDEX idx_user_oauth_accounts_provider ON user_oauth_accounts(provider, provider_account_id);
