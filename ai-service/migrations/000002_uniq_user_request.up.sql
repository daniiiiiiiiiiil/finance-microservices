ALTER TABLE ai.shopping_plans DROP CONSTRAINT uq_shopping_plans_request_id;
ALTER TABLE ai.shopping_plans ADD CONSTRAINT uq_shopping_plans_user_request UNIQUE (user_id, request_id);