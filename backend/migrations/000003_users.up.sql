CREATE SCHEMA IF NOT EXISTS users;

CREATE TABLE users.users
(
    id INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    version       BIGINT              NOT NULL DEFAULT 1,
    full_name     VARCHAR(100)        NOT NULL CHECK (CHAR_LENGTH(full_name) BETWEEN 3 AND 100),
    email         VARCHAR(100) UNIQUE NOT NULL CHECK ( CHAR_LENGTH(email) BETWEEN 3 AND 100),
    password_hash VARCHAR(255)        NOT NULL,
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

CREATE INDEX idx_users_phone_number ON users.users(phone_number);
CREATE INDEX idx_users_full_name ON users.users(full_name);
CREATE INDEX idx_users_full_name_email ON users.users(full_name, email);