CREATE SCHEMA IF NOT EXISTS shopping;

CREATE TABLE IF NOT EXISTS shopping.shopping(
    id int GENERATED ALWAYS AS IDENTITY NOT NULL,
    version INT NOT NULL DEFAULT 1,
    title VARCHAR(200) NOT NULL CHECK(CHAR_LENGTH(title) BETWEEN 1 AND 200),
    description VARCHAR(1000) CHECK(CHAR_LENGTH(description) BETWEEN 1 AND 1000),
    amount_now DECIMAL(10,2) NOT NULL DEFAULT 0,
    amount_finish DECIMAL(10,2) NOT NULL,
    image_key VARCHAR(500) CHECK(CHAR_LENGTH(title) BETWEEN 1 AND 500),
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    create_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    update_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    completion_date TIMESTAMPTZ,

    CONSTRAINT PK_shopping_id PRIMARY KEY(id)
);

CREATE INDEX idx_shopping_title ON shopping.shopping(title);
CREATE INDEX idx_shopping_amount_finish ON shopping.shopping(amount_finish);
CREATE INDEX idx_shopping_amount_now ON shopping.shopping(amount_now);
CREATE INDEX idx_shopping_completed ON shopping.shopping(completed);