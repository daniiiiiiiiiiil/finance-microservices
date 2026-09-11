package postgres

import (
	"encoding/json"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
)

type SagaModel struct {
	ID          int
	SagaID      string
	SagaType    string
	UserID      int
	Status      string
	CurrentStep int
	TotalSteps  int
	Error       *string
	Metadata    []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

type StepModel struct {
	ID          int
	SagaID      int
	StepName    string
	StepOrder   int
	Status      string
	Error       *string
	Metadata    []byte
	StartedAt   *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CompensationModel struct {
	ID               int
	SagaID           int
	StepName         string
	CompensationData []byte
	Status           string
	Error            *string
	StartedAt        *time.Time
	CompletedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func sagaToModel(saga *domain.Saga) *SagaModel {
	metadata, _ := json.Marshal(saga.Metadata)

	var err *string
	if saga.Error != "" {
		err = &saga.Error
	}

	return &SagaModel{
		ID:          saga.ID,
		SagaID:      saga.SagaID,
		SagaType:    saga.Type.String(),
		UserID:      saga.UserID,
		Status:      saga.Status.String(),
		CurrentStep: saga.CurrentStep,
		TotalSteps:  saga.TotalSteps,
		Error:       err,
		Metadata:    metadata,
		CreatedAt:   saga.CreatedAt,
		UpdatedAt:   saga.UpdatedAt,
		CompletedAt: saga.CompletedAt,
	}
}

func stepToModel(step *domain.Step) *StepModel {
	metadata, _ := json.Marshal(step.Metadata)

	var err *string
	if step.Error != "" {
		err = &step.Error
	}

	return &StepModel{
		ID:          step.ID,
		SagaID:      step.SagaID,
		StepName:    step.Name,
		StepOrder:   step.Order,
		Status:      step.Status.String(),
		Error:       err,
		Metadata:    metadata,
		StartedAt:   step.StartedAt,
		CompletedAt: step.CompletedAt,
		CreatedAt:   step.CreatedAt,
		UpdatedAt:   step.UpdatedAt,
	}
}

func compensationToModel(comp *domain.Compensation) *CompensationModel {
	data, _ := json.Marshal(comp.Data)

	var err *string
	if comp.Error != "" {
		err = &comp.Error
	}

	return &CompensationModel{
		ID:               comp.ID,
		SagaID:           comp.SagaID,
		StepName:         comp.StepName,
		CompensationData: data,
		Status:           comp.Status.String(),
		Error:            err,
		StartedAt:        comp.StartedAt,
		CompletedAt:      comp.CompletedAt,
		CreatedAt:        comp.CreatedAt,
		UpdatedAt:        comp.UpdatedAt,
	}
}

func sagaFromModel(model *SagaModel) *domain.Saga {
	var metadata map[string]interface{}
	if len(model.Metadata) > 0 {
		_ = json.Unmarshal(model.Metadata, &metadata)
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	var errMsg string
	if model.Error != nil {
		errMsg = *model.Error
	}

	return &domain.Saga{
		ID:          model.ID,
		SagaID:      model.SagaID,
		Type:        domain.SagaType(model.SagaType),
		UserID:      model.UserID,
		Status:      domain.Status(model.Status),
		CurrentStep: model.CurrentStep,
		TotalSteps:  model.TotalSteps,
		Error:       errMsg,
		Metadata:    metadata,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		CompletedAt: model.CompletedAt,
	}
}

func stepFromModel(model *StepModel) *domain.Step {
	var metadata map[string]interface{}
	if len(model.Metadata) > 0 {
		_ = json.Unmarshal(model.Metadata, &metadata)
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	var errMsg string
	if model.Error != nil {
		errMsg = *model.Error
	}

	return &domain.Step{
		ID:          model.ID,
		SagaID:      model.SagaID,
		Name:        model.StepName,
		Order:       model.StepOrder,
		Status:      domain.StepStatus(model.Status),
		Error:       errMsg,
		Metadata:    metadata,
		StartedAt:   model.StartedAt,
		CompletedAt: model.CompletedAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func compensationFromModel(model *CompensationModel) *domain.Compensation {
	var data map[string]interface{}
	if len(model.CompensationData) > 0 {
		_ = json.Unmarshal(model.CompensationData, &data)
	}

	var errMsg string
	if model.Error != nil {
		errMsg = *model.Error
	}

	return &domain.Compensation{
		ID:          model.ID,
		SagaID:      model.SagaID,
		StepName:    model.StepName,
		Data:        data,
		Status:      domain.CompensationStatus(model.Status),
		Error:       errMsg,
		StartedAt:   model.StartedAt,
		CompletedAt: model.CompletedAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}
