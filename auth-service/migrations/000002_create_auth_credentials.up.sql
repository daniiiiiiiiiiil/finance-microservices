CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE auth.credentials
(
    id            INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    email         VARCHAR(100) UNIQUE              NOT NULL,
    password_hash VARCHAR(255)                     NOT NULL,
    status        VARCHAR(20)                      NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ                      NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ                      NOT NULL DEFAULT NOW(),
    CONSTRAINT PK_credentials_id PRIMARY KEY (id)
);

CREATE INDEX idx_credentials_email ON auth.credentials (email);
CREATE INDEX idx_credentials_status ON auth.credentials (status);