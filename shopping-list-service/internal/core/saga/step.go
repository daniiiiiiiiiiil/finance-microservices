package saga

import (
	"context"
	"time"
)

type Step struct {
	Name        string
	Order       int
	Action      func(ctx context.Context) error
	Compensate  func(ctx context.Context) error
	Status      StepStatus
	Error       error
	StartedAt   *time.Time
	CompletedAt *time.Time
	Metadata    map[string]interface{}
}

func NewStep(name string, order int, action func(ctx context.Context) error, compensate func(ctx context.Context) error) *Step {
	return &Step{
		Name:       name,
		Order:      order,
		Action:     action,
		Compensate: compensate,
		Status:     StepStatusPending,
		Metadata:   make(map[string]interface{}),
	}
}

// проверка шага
func (s *Step) IsCompleted() bool {
	return s.Status == StepStatusCompleted
}

// проверка упал ли шаг
func (s *Step) IsFailed() bool {
	return s.Status == StepStatusFailed
}

// проверка можно ли откатить шаг
func (s *Step) CanCompensate() bool {
	return s.Compensate != nil && s.Status == StepStatusCompleted
}

// отмечаем начатй шаг
func (s *Step) MarkStarted() {
	now := time.Now()
	s.StartedAt = &now
	s.Status = StepStatusInProgress
}

// отмечаем выполненный шаг
func (s *Step) MarkCompleted() {
	now := time.Now()
	s.CompletedAt = &now
	s.Status = StepStatusCompleted
}

// отмечаем как упавший шаг
func (s *Step) MarkFailed(err error) {
	now := time.Now()
	s.CompletedAt = &now
	s.Status = StepStatusFailed
	s.Error = err
}

// отмечаем шаг как откаченный
func (s *Step) MarkCompensated() {
	now := time.Now()
	s.CompletedAt = &now
	s.Status = StepStatusCompensated
}

// устанавливаем метаданные шага
func (s *Step) SetMetadata(key string, value interface{}) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}
	s.Metadata[key] = value
}

// получаем метаданные шага
func (s *Step) GetMetadata(key string) (interface{}, bool) {
	if s.Metadata == nil {
		return nil, false
	}
	value, ok := s.Metadata[key]
	return value, ok
}
