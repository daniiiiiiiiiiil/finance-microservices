DROP INDEX IF EXISTS users.idx_users_full_name_email;
DROP INDEX IF EXISTS users.idx_users_full_name;
DROP INDEX IF EXISTS users.idx_users_phone_number;

DROP TABLE IF EXISTS users.users;

DROP SCHEMA IF EXISTS users;