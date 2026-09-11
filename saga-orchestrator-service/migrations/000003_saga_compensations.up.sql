CREATE TABLE IF NOT EXISTS saga.saga_compensations (
    id                  INT GENERATED ALWAYS AS IDENTITY NOT NULL,
    saga_id             INT NOT NULL REFERENCES saga.sagas(id) ON DELETE CASCADE,
    step_name           VARCHAR(255) NOT NULL,
    compensation_data   JSONB,
    status              VARCHAR(50) NOT NULL,
    error               TEXT,
    started_at          TIMESTAMP,
    completed_at        TIMESTAMP,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_saga_compensations_id PRIMARY KEY (id),
    CONSTRAINT chk_compensation_status CHECK (
        status IN ('pending', 'in_progress', 'completed', 'failed')
    )
    );

CREATE INDEX IF NOT EXISTS idx_saga_compensations_saga_id ON saga.saga_compensations(saga_id);
CREATE INDEX IF NOT EXISTS idx_saga_compensations_status ON saga.saga_compensations(status);