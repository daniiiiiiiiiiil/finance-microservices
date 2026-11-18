CREATE SCHEMA IF NOT EXISTS telegram;

CREATE TABLE IF NOT EXISTS telegram.accounts
(
    telegram_user_id INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    chat_id          BIGINT      NOT NULL,
    user_id          INTEGER     NOT NULL,
    username         VARCHAR(64),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_telegram_accounts PRIMARY KEY (telegram_user_id),
    CONSTRAINT uq_telegram_accounts_user UNIQUE (user_id)
);

CREATE INDEX IF NOT EXISTS idx_telegram_accounts_user_id ON telegram.accounts (user_id);

CREATE TABLE IF NOT EXISTS telegram.link_codes
(
    id         INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    user_id    INTEGER     NOT NULL,
    code_hash  VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    attempts   INTEGER     NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_telegram_link_codes PRIMARY KEY (id),
    CONSTRAINT uq_telegram_link_codes_hash UNIQUE (code_hash)
);

CREATE INDEX IF NOT EXISTS idx_telegram_link_codes_user_id ON telegram.link_codes (user_id);
CREATE INDEX IF NOT EXISTS idx_telegram_link_codes_expires_at ON telegram.link_codes (expires_at);

CREATE TABLE IF NOT EXISTS telegram.notification_settings
(
    user_id        INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    budget_alerts  BOOLEAN     NOT NULL DEFAULT TRUE,
    plan_alerts    BOOLEAN     NOT NULL DEFAULT TRUE,
    finance_alerts BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_telegram_notification_settings PRIMARY KEY (user_id)
);
