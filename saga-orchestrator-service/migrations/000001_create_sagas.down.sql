DROP INDEX IF EXISTS saga.idx_sagas_saga_id;
DROP INDEX IF EXISTS saga.idx_sagas_user_id;
DROP INDEX IF EXISTS saga.idx_sagas_status;
DROP INDEX IF EXISTS saga.idx_sagas_type;
DROP INDEX IF EXISTS saga.idx_sagas_created_at;
DROP INDEX IF EXISTS saga.idx_sagas_active;

DROP TABLE IF EXISTS saga.sagas CASCADE;
DROP SCHEMA IF EXISTS saga CASCADE;