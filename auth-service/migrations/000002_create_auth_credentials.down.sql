DROP INDEX IF EXISTS auth.idx_credentials_email;
DROP INDEX IF EXISTS auth.idx_credentials_status;

DROP TABLE IF EXISTS auth.credentials;

DROP SCHEMA IF EXISTS auth;