ALTER TABLE finance.transactions
    ADD COLUMN IF NOT EXISTS description VARCHAR(500) NOT NULL DEFAULT '';

ALTER TABLE finance.transactions
    ADD CONSTRAINT chk_transactions_description_length CHECK (CHAR_LENGTH(description) <= 500);
