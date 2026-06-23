-- Rollback: remove register_method column from users table

ALTER TABLE users DROP COLUMN IF EXISTS register_method;
