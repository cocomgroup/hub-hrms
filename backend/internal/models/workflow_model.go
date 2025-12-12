package models

import (
	"time"

	"github.com/google/uuid"
)

// Workflow models

// OnboardingWorkflow represents the main workflow for an employee
type OnboardingWorkflow struct {
	ID                   uuid.UUID  `json:"id"`
	EmployeeID           uuid.UUID  `json:"employee_id"`
	TemplateName         string     `json:"template_name"`
	Status               string     `json:"status"`
	CurrentStage         string     `json:"current_stage"`
	ProgressPercentage   int        `json:"progress_percentage"`
	StartedAt            time.Time  `json:"started_at"`
	ExpectedCompletion   *time.Time `json:"expected_completion,omitempty"`
	ActualCompletion     *time.Time `json:"actual_completion,omitempty"`
	CreatedBy            *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// WorkflowStep represents a single step in the workflow
type WorkflowStep struct {
	ID                uuid.UUID  `json:"id"`
	WorkflowID        uuid.UUID  `json:"workflow_id"`
	StepOrder         int        `json:"step_order"`
	StepName          string     `json:"step_name"`
	StepType          string     `json:"step_type"`
	Stage             string     `json:"stage"`
	Status            string     `json:"status"`
	Description       string     `json:"description,omitempty"`
	Dependencies      []uuid.UUID `json:"dependencies,omitempty"`
	AssignedTo        *uuid.UUID `json:"assigned_to,omitempty"`
	IntegrationType   string     `json:"integration_type,omitempty"`
	IntegrationConfig map[string]interface{} `json:"integration_config,omitempty"`
	DueDate           *time.Time `json:"due_date,omitempty"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	CompletedBy       *uuid.UUID `json:"completed_by,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// WorkflowIntegration represents an external integration call
type WorkflowIntegration struct {
	ID              uuid.UUID              `json:"id"`
	WorkflowID      uuid.UUID              `json:"workflow_id"`
	StepID          uuid.UUID              `json:"step_id"`
	IntegrationType string                 `json:"integration_type"`
	ExternalID      string                 `json:"external_id,omitempty"`
	Status          string                 `json:"status"`
	RequestPayload  map[string]interface{} `json:"request_payload,omitempty"`
	ResponsePayload map[string]interface{} `json:"response_payload,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	RetryCount      int                    `json:"retry_count"`
	MaxRetries      int                    `json:"max_retries"`
	LastAttemptAt   *time.Time             `json:"last_attempt_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// WorkflowException represents an exception or issue in the workflow
type WorkflowException struct {
	ID               uuid.UUID  `json:"id"`
	WorkflowID       uuid.UUID  `json:"workflow_id"`
	StepID           *uuid.UUID `json:"step_id,omitempty"`
	ExceptionType    string     `json:"exception_type"`
	Severity         string     `json:"severity"`
	Title            string     `json:"title"`
	Description      string     `json:"description,omitempty"`
	ResolutionStatus string     `json:"resolution_status"`
	AssignedTo       *uuid.UUID `json:"assigned_to,omitempty"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy       *uuid.UUID `json:"resolved_by,omitempty"`
	ResolutionNotes  string     `json:"resolution_notes,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// WorkflowDocument represents a document in the workflow
type WorkflowDocument struct {
	ID           uuid.UUID              `json:"id"`
	WorkflowID   uuid.UUID              `json:"workflow_id"`
	StepID       *uuid.UUID             `json:"step_id,omitempty"`
	DocumentName string                 `json:"document_name"`
	DocumentType string                 `json:"document_type"`
	S3Key        string                 `json:"s3_key,omitempty"`
	FileType     string                 `json:"file_type"`
	FileSize     int                    `json:"file_size,omitempty"`
	Status       string                 `json:"status"`
	UploadedBy   *uuid.UUID             `json:"uploaded_by,omitempty"`
	UploadedAt   *time.Time             `json:"uploaded_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// WorkflowWithDetails includes workflow and related data
type WorkflowWithDetails struct {
	Workflow   OnboardingWorkflow  `json:"workflow"`
	Steps      []WorkflowStep      `json:"steps"`
	Exceptions []WorkflowException `json:"exceptions"`
	Documents  []WorkflowDocument  `json:"documents"`
	Employee   Employee            `json:"employee"`
}

// Mock integration responses
type DocuSignMockResponse struct {
	EnvelopeID string    `json:"envelope_id"`
	Status     string    `json:"status"`
	SentAt     time.Time `json:"sent_at"`
	SignedAt   *time.Time `json:"signed_at,omitempty"`
	SignerEmail string   `json:"signer_email"`
}

type BackgroundCheckMockResponse struct {
	CheckID     string    `json:"check_id"`
	Status      string    `json:"status"`
	Candidate   string    `json:"candidate"`
	CheckTypes  []string  `json:"check_types"`
	InitiatedAt time.Time `json:"initiated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Result      string    `json:"result,omitempty"` // 'clear', 'review', 'failed'
}

type DocSearchMockResponse struct {
	Documents []DocSearchDocument `json:"documents"`
	TotalCount int                `json:"total_count"`
}

type DocSearchDocument struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	DocumentType string                 `json:"document_type"`
	S3Key        string                 `json:"s3_key"`
	FileType     string                 `json:"file_type"`
	FileSize     int                    `json:"file_size"`
	UploadedAt   time.Time              `json:"uploaded_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}