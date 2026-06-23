-- Add status column to users table (missing from initial schema)

ALTER TABLE users ADD COLUMN IF NOT EXISTS status SMALLINT NOT NULL DEFAULT 1;
