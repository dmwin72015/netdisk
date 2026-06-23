-- Add register_method column to users table

ALTER TABLE users ADD COLUMN IF NOT EXISTS register_method VARCHAR(20) NOT NULL DEFAULT 'email';
