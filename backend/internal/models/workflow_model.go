package models

import (
	"time"

	"github.com/google/uuid"
)

// Workflow models - Generic workflow engine for ALL employee lifecycle events

// EmployeeWorkflow represents an active workflow instance for any lifecycle event
// This replaces the old "OnboardingWorkflow" to avoid confusion
type EmployeeWorkflow struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	EmployeeID           uuid.UUID  `json:"employee_id" db:"employee_id"`
	TemplateID           *uuid.UUID `json:"template_id,omitempty" db:"template_id"` // Link to WorkflowTemplate
	TemplateName         string     `json:"template_name" db:"template_name"`
	WorkflowType         string     `json:"workflow_type" db:"workflow_type"` // onboarding, offboarding, performance, leave, vendor
	Status               string     `json:"status" db:"status"` // active, completed, cancelled, on-hold
	CurrentStage         string     `json:"current_stage" db:"current_stage"`
	ProgressPercentage   int        `json:"progress_percentage" db:"progress_percentage"`
	StartedAt            time.Time  `json:"started_at" db:"started_at"`
	ExpectedCompletion   *time.Time `json:"expected_completion,omitempty" db:"expected_completion"`
	ActualCompletion     *time.Time `json:"actual_completion,omitempty" db:"actual_completion"`
	CreatedBy            *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// WorkflowStep represents a single step in a workflow (template or instance)
type WorkflowStep struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	WorkflowID        uuid.UUID  `json:"workflow_id" db:"workflow_id"` // References WorkflowTemplate OR EmployeeWorkflow
	StepOrder         int        `json:"step_order" db:"step_order"`
	StepName          string     `json:"step_name" db:"step_name"`
	StepType          string     `json:"step_type" db:"step_type"` // manual, integration, agent, approval, document
	Stage             string     `json:"stage,omitempty" db:"stage"` // Only for instances: pre-boarding, day-1, week-1, etc.
	Status            string     `json:"status" db:"status"` // pending, in-progress, completed, failed, skipped, blocked
	Description       string     `json:"description,omitempty" db:"description"`
	Dependencies      []uuid.UUID `json:"dependencies,omitempty" db:"dependencies"`
	AssignedTo        *uuid.UUID `json:"assigned_to,omitempty" db:"assigned_to"`
	IntegrationType   string     `json:"integration_type,omitempty" db:"integration_type"`
	IntegrationConfig map[string]interface{} `json:"integration_config,omitempty" db:"integration_config"`
	DueDate           *time.Time `json:"due_date,omitempty" db:"due_date"`
	StartedAt         *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CompletedBy       *uuid.UUID `json:"completed_by,omitempty" db:"completed_by"`
	Metadata          map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	// Template-specific fields
	Required      bool       `json:"required" db:"required"`
	AutoTrigger   bool       `json:"auto_trigger" db:"auto_trigger"`
	AssignedRole  string     `json:"assigned_role,omitempty" db:"assigned_role"` // hr, manager, it, employee
	DueDays       *int       `json:"due_days,omitempty" db:"due_days"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// WorkflowIntegration represents an external integration call
type WorkflowIntegration struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	WorkflowID      uuid.UUID              `json:"workflow_id" db:"workflow_id"`
	StepID          uuid.UUID              `json:"step_id" db:"step_id"`
	IntegrationType string                 `json:"integration_type" db:"integration_type"`
	ExternalID      string                 `json:"external_id,omitempty" db:"external_id"`
	Status          string                 `json:"status" db:"status"`
	RequestPayload  map[string]interface{} `json:"request_payload,omitempty" db:"request_payload"`
	ResponsePayload map[string]interface{} `json:"response_payload,omitempty" db:"response_payload"`
	ErrorMessage    string                 `json:"error_message,omitempty" db:"error_message"`
	RetryCount      int                    `json:"retry_count" db:"retry_count"`
	MaxRetries      int                    `json:"max_retries" db:"max_retries"`
	LastAttemptAt   *time.Time             `json:"last_attempt_at,omitempty" db:"last_attempt_at"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// WorkflowException represents an exception or issue in the workflow
type WorkflowException struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	WorkflowID       uuid.UUID  `json:"workflow_id" db:"workflow_id"`
	StepID           *uuid.UUID `json:"step_id,omitempty" db:"step_id"`
	ExceptionType    string     `json:"exception_type" db:"exception_type"`
	Severity         string     `json:"severity" db:"severity"`
	Title            string     `json:"title" db:"title"`
	Description      string     `json:"description,omitempty" db:"description"`
	ResolutionStatus string     `json:"resolution_status" db:"resolution_status"`
	AssignedTo       *uuid.UUID `json:"assigned_to,omitempty" db:"assigned_to"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy       *uuid.UUID `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolutionNotes  string     `json:"resolution_notes,omitempty" db:"resolution_notes"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// WorkflowDocument represents a document in the workflow
type WorkflowDocument struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	WorkflowID   uuid.UUID              `json:"workflow_id" db:"workflow_id"`
	StepID       *uuid.UUID             `json:"step_id,omitempty" db:"step_id"`
	DocumentName string                 `json:"document_name" db:"document_name"`
	DocumentType string                 `json:"document_type" db:"document_type"`
	S3Key        string                 `json:"s3_key,omitempty" db:"s3_key"`
	FileType     string                 `json:"file_type" db:"file_type"`
	FileSize     int                    `json:"file_size,omitempty" db:"file_size"`
	Status       string                 `json:"status" db:"status"`
	UploadedBy   *uuid.UUID             `json:"uploaded_by,omitempty" db:"uploaded_by"`
	UploadedAt   *time.Time             `json:"uploaded_at,omitempty" db:"uploaded_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// WorkflowWithDetails includes workflow and related data
type WorkflowWithDetails struct {
	Workflow   EmployeeWorkflow    `json:"workflow"`
	Steps      []WorkflowStep      `json:"steps"`
	Exceptions []WorkflowException `json:"exceptions"`
	Documents  []WorkflowDocument  `json:"documents"`
	Employee   Employee            `json:"employee"`
}

// Mock integration responses
type DocuSignMockResponse struct {
	EnvelopeID  string     `json:"envelope_id"`
	Status      string     `json:"status"`
	SentAt      time.Time  `json:"sent_at"`
	SignedAt    *time.Time `json:"signed_at,omitempty"`
	SignerEmail string     `json:"signer_email"`
}

type BackgroundCheckMockResponse struct {
	CheckID     string     `json:"check_id"`
	Status      string     `json:"status"`
	Candidate   string     `json:"candidate"`
	CheckTypes  []string   `json:"check_types"`
	InitiatedAt time.Time  `json:"initiated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Result      string     `json:"result,omitempty"` // 'clear', 'review', 'failed'
}

type DocSearchMockResponse struct {
	Documents  []DocSearchDocument `json:"documents"`
	TotalCount int                 `json:"total_count"`
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

// WorkflowTemplate represents a reusable workflow definition/template
// Templates define the structure for lifecycle events (onboarding, offboarding, etc.)
type WorkflowTemplate struct {
	ID           uuid.UUID         `json:"id" db:"id"`
	Name         string            `json:"name" db:"name"`
	Description  string            `json:"description" db:"description"`
	WorkflowType string            `json:"workflow_type" db:"workflow_type"` // onboarding, offboarding, performance, leave, vendor
	Status       string            `json:"status" db:"status"`               // active, inactive, draft
	CreatedBy    uuid.UUID         `json:"created_by" db:"created_by"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" db:"updated_at"`
	Steps        []WorkflowStepDef `json:"steps,omitempty"` // Not in DB, loaded separately
}

// WorkflowStepDef represents a step definition in a workflow template
// This is essentially the same as WorkflowStep but used in template context
type WorkflowStepDef struct {
	ID           uuid.UUID `json:"id" db:"id"`
	WorkflowID   uuid.UUID `json:"workflow_id" db:"workflow_id"` // References workflow_templates.id
	StepOrder    int       `json:"step_order" db:"step_order"`
	StepType     string    `json:"step_type" db:"step_type"` // document, approval, background_check, equipment, training
	StepName     string    `json:"step_name" db:"step_name"`
	Description  string    `json:"description" db:"description"`
	Required     bool      `json:"required" db:"required"`
	AutoTrigger  bool      `json:"auto_trigger" db:"auto_trigger"`
	AssignedRole string    `json:"assigned_role" db:"assigned_role"` // hr, manager, it, employee
	DueDays      *int      `json:"due_days" db:"due_days"`           // Days to complete from workflow start
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// WorkflowTemplateWithSteps includes template and its steps
type WorkflowTemplateWithSteps struct {
	Template WorkflowTemplate  `json:"template"`
	Steps    []WorkflowStepDef `json:"steps"`
}

// OnboardingWithDetails includes onboarding workflow and related data
// This is separate from WorkflowWithDetails which uses EmployeeWorkflow
type OnboardingWithDetails struct {
	Workflow   OnboardingWorkflow    `json:"workflow"`
	Tasks      []OnboardingTask      `json:"tasks"`
	Milestones []OnboardingMilestone `json:"milestones"`
	Employee   Employee              `json:"employee"`
}