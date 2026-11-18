ALTER TABLE finance.transactions
    DROP CONSTRAINT IF EXISTS chk_transactions_description_length;

ALTER TABLE finance.transactions
    DROP COLUMN IF EXISTS description;
