CREATE SCHEMA IF NOT EXISTS finance;

CREATE TABLE finance.users
(
    id INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    version       BIGINT              NOT NULL DEFAULT 1,
    full_name     VARCHAR(100)        NOT NULL CHECK (CHAR_LENGTH(full_name) BETWEEN 3 AND 100),
    email         VARCHAR(100) UNIQUE NOT NULL CHECK ( CHAR_LENGTH(email) BETWEEN 3 AND 100),
    password_hash VARCHAR(255)        NOT NULL CHECK ( CHAR_LENGTH(password_hash) BETWEEN 8 AND 255),
    phone_number  VARCHAR(15) CHECK (
        phone_number ~'^\+[0-9]+$'
        AND
        CHAR_LENGTH(phone_number) BETWEEN 10 AND 15
    ),
    is_admin      BOOLEAN             NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ         NOT NULL DEFAULT NOW(), 
    updated_at    TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    CONSTRAINT PK_users_id PRIMARY KEY(id)
);


CREATE TABLE finance.transactions
(
    id INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    type_transaction VARCHAR(60) NOT NULL CHECK(CHAR_LENGTH(type_transaction) BETWEEN 1 AND 60),
    amount DECIMAL(10,2) NOT NULL CHECK ( amount > 0 ),
    category VARCHAR(60) NOT NULL CHECK ( CHAR_LENGTH(category) BETWEEN 1 AND 60),
    created_at TIMESTAMPTZ NOT NULL,

    user_id INTEGER NOT NULL,

    CONSTRAINT PK_transactions_id PRIMARY KEY(id),
    CONSTRAINT FK_transactions_user_user_id
    FOREIGN KEY (user_id)
    REFERENCES finance.users(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE
)

CREATE INDEX idx_users_phone_number ON finance.users(phone_number);
CREATE INDEX idx_users_full_name ON finance.users(full_name);
CREATE INDEX idx_users_full_name_email ON finance.users(full_name, email);
CREATE INDEX idx_transactions_user_id ON finance.transactions(user_id);
CREATE INDEX idx_transactions_created_at ON finance.transactions(created_at DESC);
CREATE INDEX idx_transactions_type ON finance.transactions(type_transaction);
CREATE INDEX idx_transactions_category ON finance.transactions(category);