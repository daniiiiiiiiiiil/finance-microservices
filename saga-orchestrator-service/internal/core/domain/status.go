package domain

type Status string

const (
	StatusPending      Status = "pending"
	StatusInProgress   Status = "in_progress"
	StatusCompleted    Status = "completed"
	StatusFailed       Status = "failed"
	StatusCompensating Status = "compensating"
	StatusCompensated  Status = "compensated"
)

type StepStatus string

const (
	StepStatusPending     StepStatus = "pending"
	StepStatusInProgress  StepStatus = "in_progress"
	StepStatusCompleted   StepStatus = "completed"
	StepStatusFailed      StepStatus = "failed"
	StepStatusCompensated StepStatus = "compensated"
	StepStatusSkipped     StepStatus = "skipped"
)

type CompensationStatus string

const (
	CompensationStatusPending    CompensationStatus = "pending"
	CompensationStatusInProgress CompensationStatus = "in_progress"
	CompensationStatusCompleted  CompensationStatus = "completed"
	CompensationStatusFailed     CompensationStatus = "failed"
)

func (s Status) IsFinal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusCompensated
}

func (s Status) IsSuccess() bool {
	return s == StatusCompleted
}

func (s Status) IsActive() bool {
	return s == StatusPending || s == StatusInProgress || s == StatusCompensating
}

func (s Status) String() string {
	return string(s)
}

func (s StepStatus) String() string {
	return string(s)
}

func (s CompensationStatus) String() string {
	return string(s)
}
