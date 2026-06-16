ALTER TABLE users ALTER COLUMN email SET NOT NULL;
ALTER TABLE user_oauth_accounts DROP COLUMN oauth_email;
