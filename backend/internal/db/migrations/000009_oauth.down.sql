-- Drop OAuth tables and columns
DROP TABLE IF EXISTS user_oauth_accounts;
ALTER TABLE users DROP COLUMN IF EXISTS register_method;
