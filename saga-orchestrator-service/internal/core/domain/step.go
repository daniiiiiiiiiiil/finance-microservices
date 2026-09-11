package domain

import "time"

type Step struct {
	ID          int
	SagaID      int
	Name        string
	Order       int
	Status      StepStatus
	Error       string
	Metadata    map[string]interface{}
	StartedAt   *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewStep(sagaID int, name string, order int) *Step {
	return &Step{
		SagaID:    sagaID,
		Name:      name,
		Order:     order,
		Status:    StepStatusPending,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *Step) IsCompleted() bool {
	return s.Status == StepStatusCompleted
}

func (s *Step) IsPending() bool {
	return s.Status == StepStatusPending
}

func (s *Step) MarkStarted() {
	now := time.Now()
	s.StartedAt = &now
	s.Status = StepStatusInProgress
	s.UpdatedAt = now
}

func (s *Step) MarkCompleted() {
	now := time.Now()
	s.CompletedAt = &now
	s.Status = StepStatusCompleted
	s.UpdatedAt = now
}

func (s *Step) MarkFailed(err string) {
	now := time.Now()
	s.CompletedAt = &now
	s.Status = StepStatusFailed
	s.Error = err
	s.UpdatedAt = now
}

func (s *Step) MarkCompensated() {
	now := time.Now()
	s.CompletedAt = &now
	s.Status = StepStatusCompensated
	s.UpdatedAt = now
}

func (s *Step) MarkSkipped() {
	now := time.Now()
	s.CompletedAt = &now
	s.Status = StepStatusSkipped
	s.UpdatedAt = now
}

func (s *Step) SetMetadata(key string, value interface{}) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}
	s.Metadata[key] = value
}

func (s *Step) GetMetadata(key string) (interface{}, bool) {
	if s.Metadata == nil {
		return nil, false
	}
	v, ok := s.Metadata[key]
	return v, ok
}
