CREATE TABLE user_activity_logs (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL,
    action        VARCHAR(32) NOT NULL,
    resource_type VARCHAR(32),
    resource_name VARCHAR(255),
    ip            VARCHAR(64),
    ip_region     VARCHAR(128),
    user_agent    VARCHAR(512),
    os            VARCHAR(64),
    browser       VARCHAR(64),
    extra         JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_logs_user_id    ON user_activity_logs(user_id);
CREATE INDEX idx_activity_logs_created_at ON user_activity_logs(created_at DESC);
CREATE INDEX idx_activity_logs_action     ON user_activity_logs(action);
