package orchestrator

import (
	"fmt"
	"sync"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/ports"
)

type SagaRegistryImpl struct {
	mtx         sync.RWMutex
	definitions map[domain.SagaType]ports.SagaDefinition
}

func NewSagaRegistryImpl() *SagaRegistryImpl {
	return &SagaRegistryImpl{
		definitions: make(map[domain.SagaType]ports.SagaDefinition),
	}
}

func (r *SagaRegistryImpl) Register(sagaType domain.SagaType, definition ports.SagaDefinition) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if !sagaType.IsValid() {
		return fmt.Errorf("Invalid saga type: %s", sagaType)
	}

	if _, exists := r.definitions[sagaType]; exists {
		return fmt.Errorf("Saga type %s is already registered", sagaType)
	}

	if definition.Build == nil {
		return fmt.Errorf("Saga type %s is missing a build definition", sagaType)
	}

	if len(definition.Steps) == 0 {
		return fmt.Errorf("Saga type %s has no steps", sagaType)
	}

	names := make(map[string]bool)
	for _, step := range definition.Steps {
		if step.Name == "" {
			return fmt.Errorf("Saga type %s has no step name", sagaType)
		}
		if names[step.Name] {
			return fmt.Errorf("Saga type %s has a duplicate step name", sagaType)
		}
		names[step.Name] = true
		if step.Action == nil {
			return fmt.Errorf("Saga type %s has no action definition", sagaType)
		}
	}
	definition.Type = sagaType
	r.definitions[sagaType] = definition

	return nil
}

func (r *SagaRegistryImpl) Get(sagaType domain.SagaType) (ports.SagaDefinition, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	definition, ok := r.definitions[sagaType]
	if !ok {
		return ports.SagaDefinition{}, fmt.Errorf("saga type %s not registered", sagaType)
	}

	return definition, nil
}

func (r *SagaRegistryImpl) GetAll() []ports.SagaDefinition {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	definitions := make([]ports.SagaDefinition, 0, len(r.definitions))
	for _, definition := range r.definitions {
		definitions = append(definitions, definition)
	}

	return definitions
}

func (r *SagaRegistryImpl) IsRegistered(sagaType domain.SagaType) bool {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	_, ok := r.definitions[sagaType]
	return ok
}

func (r *SagaRegistryImpl) Count() int {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return len(r.definitions)
}
