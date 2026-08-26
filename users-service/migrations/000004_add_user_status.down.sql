DROP INDEX IF EXISTS users.idx_users_status;

ALTER TABLE users.users DROP COLUMN IF EXISTS status;