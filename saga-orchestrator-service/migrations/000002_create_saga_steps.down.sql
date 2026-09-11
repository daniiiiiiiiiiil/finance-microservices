DROP INDEX IF EXISTS saga.idx_saga_steps_saga_id;
DROP INDEX IF EXISTS saga.idx_saga_steps_status;
DROP INDEX IF EXISTS saga.idx_saga_steps_order;

DROP TABLE IF EXISTS saga.saga_steps CASCADE;