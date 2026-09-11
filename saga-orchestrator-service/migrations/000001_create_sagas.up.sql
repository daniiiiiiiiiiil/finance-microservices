CREATE SCHEMA IF NOT EXISTS saga;

CREATE TABLE IF NOT EXISTS saga.sagas (
    id              INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    saga_id         VARCHAR(255) UNIQUE NOT NULL,
    saga_type       VARCHAR(100) NOT NULL,
    user_id         INT NOT NULL,
    status          VARCHAR(50) NOT NULL,
    current_step    INT NOT NULL DEFAULT 0,
    total_steps     INT NOT NULL,
    error           TEXT,
    metadata        JSONB DEFAULT '{}'::jsonb,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMP,

    CONSTRAINT pk_sagas_id PRIMARY KEY (id),
    CONSTRAINT chk_saga_status CHECK (
       status IN ('pending', 'in_progress', 'completed', 'failed', 'compensating', 'compensated')
    )
);

CREATE INDEX IF NOT EXISTS idx_sagas_saga_id ON saga.sagas(saga_id);
CREATE INDEX IF NOT EXISTS idx_sagas_user_id ON saga.sagas(user_id);
CREATE INDEX IF NOT EXISTS idx_sagas_status ON saga.sagas(status);
CREATE INDEX IF NOT EXISTS idx_sagas_type ON saga.sagas(saga_type);
CREATE INDEX IF NOT EXISTS idx_sagas_created_at ON saga.sagas(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_sagas_active ON saga.sagas(status, created_at DESC)
    WHERE status IN ('pending', 'in_progress', 'compensating');