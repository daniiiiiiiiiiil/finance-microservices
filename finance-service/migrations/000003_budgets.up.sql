CREATE TABLE IF NOT EXISTS finance.budgets (
                                               id           INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                                               user_id      INTEGER      NOT NULL,
                                               category     VARCHAR(60)  NOT NULL,
    amount_limit DECIMAL(10,2) NOT NULL CHECK (amount_limit > 0),
    period       VARCHAR(20)  NOT NULL,  -- 'day' | 'week' | 'month' | 'year'
    period_start TIMESTAMPTZ  NOT NULL,
    period_end   TIMESTAMPTZ  NOT NULL,
    version      BIGINT       NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_budgets_user_category_period
    UNIQUE (user_id, category, period, period_start)
    );

CREATE INDEX IF NOT EXISTS idx_budgets_user_period
    ON finance.budgets(user_id, period, period_start);