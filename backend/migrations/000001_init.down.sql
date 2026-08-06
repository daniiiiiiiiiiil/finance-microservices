DROP INDEX IF EXISTS finance.idx_transactions_category;
DROP INDEX IF EXISTS finance.idx_transactions_type;
DROP INDEX IF EXISTS finance.idx_transactions_created_at;
DROP INDEX IF EXISTS finance.idx_transactions_user_id;
DROP INDEX IF EXISTS finance.idx_users_full_name_email;
DROP INDEX IF EXISTS finance.idx_users_full_name;
DROP INDEX IF EXISTS finance.idx_users_phone_number;

DROP TABLE IF EXISTS finance.transactions;
DROP TABLE IF EXISTS finance.users;

DROP SCHEMA IF EXISTS finance;