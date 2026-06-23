ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
UPDATE users SET role = lower(role) WHERE role <> lower(role);
