CREATE DATABASE IF NOT EXISTS finance;

CREATE TABLE IF NOT EXISTS finance.transactions_analytics (
    transaction_id  UInt64,
    user_id         UInt64,
    type            LowCardinality(String),
    amount          Decimal(18, 2),
    category        LowCardinality(String),
    created_at      DateTime
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(created_at)
ORDER BY (user_id, created_at, transaction_id);