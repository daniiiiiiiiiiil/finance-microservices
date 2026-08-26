ALTER TABLE users.users ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';

CREATE INDEX idx_users_status ON users.users(status);

UPDATE users.users SET status = 'active' WHERE status IS NULL;