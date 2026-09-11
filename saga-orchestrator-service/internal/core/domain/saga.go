package domain

import (
	"fmt"
	"time"
)

type Saga struct {
	ID          int
	SagaID      string
	Type        SagaType
	UserID      int
	Status      Status
	CurrentStep int
	TotalSteps  int
	Error       string
	Metadata    map[string]interface{}
	Steps       []*Step
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

func NewSaga(sagaID string, sagaType SagaType, userID int, totalSteps int) *Saga {
	now := time.Now()
	return &Saga{
		SagaID:      sagaID,
		Type:        sagaType,
		UserID:      userID,
		Status:      StatusPending,
		CurrentStep: 0,
		TotalSteps:  totalSteps,
		Metadata:    make(map[string]interface{}),
		Steps:       make([]*Step, 0, totalSteps),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (s *Saga) AddStep(step *Step) {
	step.SagaID = s.ID
	s.Steps = append(s.Steps, step)
}

func (s *Saga) GetCurrentStep() (*Step, bool) {
	if s.CurrentStep < 0 || s.CurrentStep > len(s.Steps) {
		return nil, false
	}
	return s.Steps[s.CurrentStep], true
}

func (s *Saga) GetStepByName(name string) (*Step, bool) {
	for _, step := range s.Steps {
		if step.Name == name {
			return step, true
		}
	}
	return nil, false
}

func (s *Saga) MarkInProgress() {
	s.Status = StatusInProgress
	s.UpdatedAt = time.Now()
}

func (s *Saga) MarkCompleted() {
	now := time.Now()
	s.Status = StatusCompleted
	s.CompletedAt = &now
	s.UpdatedAt = now
}

func (s *Saga) MarkFailed(err string) {
	now := time.Now()
	s.Status = StatusFailed
	s.Error = err
	s.CompletedAt = &now
	s.UpdatedAt = now
}

func (s *Saga) MarkCompensating() {
	s.Status = StatusCompensating
	s.UpdatedAt = time.Now()
}

func (s *Saga) MarkCompensated() {
	now := time.Now()
	s.Status = StatusCompensated
	s.CompletedAt = &now
	s.UpdatedAt = now
}

func (s *Saga) SetCurrentStep(index int) {
	s.CurrentStep = index
	s.UpdatedAt = time.Now()
}

func (s *Saga) IsCompleted() bool {
	return s.Status == StatusCompleted
}

func (s *Saga) IsFailed() bool {
	return s.Status == StatusFailed
}

func (s *Saga) IsActive() bool {
	return s.Status.IsActive()
}

func (s *Saga) CanCompensate() bool {
	return s.Status == StatusFailed || s.Status == StatusInProgress
}

func (s *Saga) SetMetadata(key string, value interface{}) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}
	s.Metadata[key] = value
}

func (s *Saga) GetMetadata(key string) (interface{}, bool) {
	if s.Metadata == nil {
		return nil, false
	}
	v, ok := s.Metadata[key]
	return v, ok
}

func (s *Saga) Validate() error {
	if s.SagaID == "" {
		return fmt.Errorf("saga_id is required")
	}
	if !s.Type.IsValid() {
		return fmt.Errorf("invalid saga_type: %s", s.Type)
	}
	if s.UserID <= 0 {
		return fmt.Errorf("user_id must be positive")
	}
	if s.TotalSteps <= 0 {
		return fmt.Errorf("total_steps must be positive")
	}
	return nil
}

func (s *Saga) String() string {
	return fmt.Sprintf("Saga{ID=%d, SagaID=%s, Type=%s, UserID=%d, Status=%s, Step=%d/%d}",
		s.ID, s.SagaID, s.Type, s.UserID, s.Status, s.CurrentStep+1, s.TotalSteps)
}
