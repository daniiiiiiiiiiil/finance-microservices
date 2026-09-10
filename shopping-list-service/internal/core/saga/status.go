package saga

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
	//StepStatusSkipped     StepStatus = "skipped"
)

func (s Status) IsFinal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusCompensated
}

func (s Status) IsSuccess() bool {
	return s == StatusCompleted
}

func (s Status) String() string {
	return string(s)
}
