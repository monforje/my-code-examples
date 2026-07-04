// Package domain
package domain

import "time"

// CIReportStatus — статус CI-прогона.
type CIReportStatus string

const (
	StatusPending CIReportStatus = "pending"
	StatusPassed  CIReportStatus = "passed"
	StatusFailed  CIReportStatus = "failed"
	StatusCrashed CIReportStatus = "crashed"
	StatusTimeout CIReportStatus = "timeout"
)

// CIStepStatus — статус отдельного шага пайплайна.
type CIStepStatus string

const (
	StepSuccess CIStepStatus = "success"
	StepFailure CIStepStatus = "failure"
	StepSkipped CIStepStatus = "skipped"
)

// CIReportStep — результат одного шага пайплайна.
type CIReportStep struct {
	Name       string
	Status     CIStepStatus
	DurationMs int32
}

// CIReport — распарсенный CI-отчёт, приведённый к нормализованному виду.
type CIReport struct {
	UID       string
	Commit    string
	RunID     string
	Status    CIReportStatus
	ExitCode  int32
	Stage     string
	Steps     []CIReportStep
	Stdout    string
	Stderr    string
	CreatedAt time.Time
}

// WebhookEvent — тип события от task_runner.
type WebhookEvent string

const (
	EventCIStarted  WebhookEvent = "ci_started"
	EventCIFinished WebhookEvent = "ci_finished"
)

// WebhookPayload — сырой входящий вебхук от task_runner.
type WebhookPayload struct {
	Event    WebhookEvent `json:"event"`
	UID      string       `json:"uid"`
	Commit   string       `json:"commit"`
	VMID     string       `json:"vm_id,omitempty"`
	ExitCode *int32       `json:"exit_code,omitempty"`
	Stdout   string       `json:"stdout,omitempty"`
	Stderr   string       `json:"stderr,omitempty"`
	Stage    string       `json:"stage,omitempty"`
}
