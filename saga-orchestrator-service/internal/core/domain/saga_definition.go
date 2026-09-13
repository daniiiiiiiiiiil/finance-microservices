package domain

import "context"

type SagaStepDefinition struct {
	Name       string
	Order      int
	Action     func(ctx context.Context, saga *Saga) error
	Compensate func(ctx context.Context, saga *Saga) error
}

type SagaDefinition struct {
	Type  SagaType
	Steps []SagaStepDefinition
	Build func(ctx context.Context, saga *Saga) error
}
