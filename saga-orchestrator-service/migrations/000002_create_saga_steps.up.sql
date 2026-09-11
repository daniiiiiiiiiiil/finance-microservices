CREATE TABLE IF NOT EXISTS saga.saga_steps (
    id              INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    saga_id         INT NOT NULL REFERENCES saga.sagas(id) ON DELETE CASCADE,
    step_name       VARCHAR(255) NOT NULL,
    step_order      INT NOT NULL,
    status          VARCHAR(50) NOT NULL,
    error           TEXT,
    metadata        JSONB DEFAULT '{}'::jsonb,
    started_at      TIMESTAMP,
    completed_at    TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_saga_steps_id PRIMARY KEY (id),
    CONSTRAINT chk_step_status CHECK (
        status IN ('pending', 'in_progress', 'completed', 'failed', 'compensated', 'skipped')
    ),
    CONSTRAINT uq_saga_step UNIQUE (saga_id, step_name)
    );

CREATE INDEX IF NOT EXISTS idx_saga_steps_saga_id ON saga.saga_steps(saga_id);
CREATE INDEX IF NOT EXISTS idx_saga_steps_status ON saga.saga_steps(status);
CREATE INDEX IF NOT EXISTS idx_saga_steps_order ON saga.saga_steps(saga_id, step_order);