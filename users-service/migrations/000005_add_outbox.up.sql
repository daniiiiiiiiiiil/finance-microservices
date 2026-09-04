CREATE TABLE IF NOT EXISTS outbox.outbox_messages (
    id              INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    aggregate_id    BIGINT      NOT NULL,
    aggregate_type  VARCHAR(50) NOT NULL,
    event_type      VARCHAR(50) NOT NULL,
    payload         JSONB       NOT NULL,
    headers         JSONB       DEFAULT '{}',
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count     INT         NOT NULL DEFAULT 0,
    max_retries     INT         NOT NULL DEFAULT 3,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ,
    last_error      TEXT,
    version         BIGINT      NOT NULL DEFAULT 1,

    CONSTRAINT PK_outbox_messages_id PRIMARY KEY (id)
    );

CREATE INDEX IF NOT EXISTS idx_outbox_status_created
    ON outbox.outbox_messages(status, created_at)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_outbox_aggregate
    ON outbox.outbox_messages(aggregate_type, aggregate_id);