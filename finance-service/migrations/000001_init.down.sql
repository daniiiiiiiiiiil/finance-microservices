DROP INDEX IF EXISTS finance.idx_transactions_category;
DROP INDEX IF EXISTS finance.idx_transactions_type;
DROP INDEX IF EXISTS finance.idx_transactions_created_at;
DROP INDEX IF EXISTS finance.idx_transactions_user_id;

DROP TABLE IF EXISTS finance.transactions;

DROP SCHEMA IF EXISTS finance;