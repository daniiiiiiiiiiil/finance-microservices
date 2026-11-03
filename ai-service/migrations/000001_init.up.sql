CREATE SCHEMA IF NOT EXISTS ai;

CREATE TABLE IF NOT EXISTS ai.shopping_plans (
    id          INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    user_id     INT NOT NULL,
    request_id  VARCHAR(64) NOT NULL,
    text        TEXT NOT NULL,
    address     TEXT NOT NULL,
    plan        JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_shopping_plans_id PRIMARY KEY (id),
    CONSTRAINT uq_shopping_plans_request_id UNIQUE (request_id)
);

CREATE INDEX IF NOT EXISTS idx_shopping_plans_user_created
    ON ai.shopping_plans (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_shopping_plans_request_id
    ON ai.shopping_plans (request_id);