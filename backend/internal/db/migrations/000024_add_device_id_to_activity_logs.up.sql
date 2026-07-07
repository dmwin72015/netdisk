-- Add a stable device identifier to activity logs so the "login devices"
-- view can distinguish real devices instead of grouping by IP only.
ALTER TABLE user_activity_logs ADD COLUMN device_id VARCHAR(64);

CREATE INDEX idx_user_activity_logs_device_id ON user_activity_logs (device_id);
