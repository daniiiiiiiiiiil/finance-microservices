CREATE SCHEMA IF NOT EXISTS finance;


CREATE TABLE finance.transactions
(
    id INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    type_transaction VARCHAR(60) NOT NULL CHECK(CHAR_LENGTH(type_transaction) BETWEEN 1 AND 60),
    amount DECIMAL(10,2) NOT NULL CHECK ( amount > 0 ),
    category VARCHAR(60) NOT NULL CHECK ( CHAR_LENGTH(category) BETWEEN 1 AND 60),
    created_at TIMESTAMPTZ NOT NULL,

    user_id INTEGER NOT NULL,

    CONSTRAINT PK_transactions_id PRIMARY KEY(id)
);

CREATE INDEX idx_transactions_user_id ON finance.transactions(user_id);
CREATE INDEX idx_transactions_created_at ON finance.transactions(created_at DESC);
CREATE INDEX idx_transactions_type ON finance.transactions(type_transaction);
CREATE INDEX idx_transactions_category ON finance.transactions(category);