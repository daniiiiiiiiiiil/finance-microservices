package postgres

import (
	"fmt"

	"context"
)

func (r *SagaRepository) Delete(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	_, err := r.pool.Exec(ctx, `DELETE FROM saga.sagas WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete saga: %w", err)
	}
	return nil
}
