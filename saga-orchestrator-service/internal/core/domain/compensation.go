package domain

import "time"

type Compensation struct {
	ID          int
	SagaID      int
	StepName    string
	Data        map[string]interface{}
	Status      CompensationStatus
	Error       string
	StartedAt   *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewCompensation(sagaID int, stepName string, data map[string]interface{}) *Compensation {
	now := time.Now()
	return &Compensation{
		SagaID:    sagaID,
		StepName:  stepName,
		Data:      data,
		Status:    CompensationStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Compensation) MarkInProgress() {
	now := time.Now()
	c.StartedAt = &now
	c.Status = CompensationStatusInProgress
	c.UpdatedAt = now
}

func (c *Compensation) MarkCompleted() {
	now := time.Now()
	c.CompletedAt = &now
	c.Status = CompensationStatusCompleted
	c.UpdatedAt = now
}

func (c *Compensation) MarkFailed(err string) {
	now := time.Now()
	c.CompletedAt = &now
	c.Status = CompensationStatusFailed
	c.Error = err
	c.UpdatedAt = now
}

func (c *Compensation) IsCompleted() bool {
	return c.Status == CompensationStatusCompleted
}

func (c *Compensation) IsFailed() bool {
	return c.Status == CompensationStatusFailed
}
