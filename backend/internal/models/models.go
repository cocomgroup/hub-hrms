package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents an authenticated user
type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	EmployeeID   *uuid.UUID `json:"employee_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Employee represents an employee record
type Employee struct {
	ID                     uuid.UUID  `json:"id"`
	FirstName              string     `json:"first_name"`
	LastName               string     `json:"last_name"`
	Email                  string     `json:"email"`
	Phone                  string     `json:"phone"`
	DateOfBirth            *time.Time `json:"date_of_birth,omitempty"`
	HireDate               time.Time  `json:"hire_date"`
	Department             string     `json:"department"`
	Position               string     `json:"position"`
	ManagerID              *uuid.UUID `json:"manager_id,omitempty"`
	EmploymentType         string     `json:"employment_type"`
	Status                 string     `json:"status"`
	StreetAddress          string     `json:"street_address"`
	City                   string     `json:"city"`
	State                  string     `json:"state"`
	ZipCode                string     `json:"zip_code"`
	Country                string     `json:"country"`
	EmergencyContactName   string     `json:"emergency_contact_name"`
	EmergencyContactPhone  string     `json:"emergency_contact_phone"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// OnboardingTask represents a task in the onboarding process
type OnboardingTask struct {
	ID                uuid.UUID  `json:"id"`
	EmployeeID        uuid.UUID  `json:"employee_id"`
	TaskName          string     `json:"task_name"`
	Description       *string    `json:"description,omitempty"`
	Category          *string    `json:"category,omitempty"`
	Status            string     `json:"status"`
	DueDate           *time.Time `json:"due_date,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	AssignedTo        *uuid.UUID `json:"assigned_to,omitempty"`
	DocumentsRequired bool       `json:"documents_required"`
	DocumentURL       *string    `json:"document_url,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Timesheet represents a time tracking entry
type Timesheet struct {
	ID           uuid.UUID  `json:"id"`
	EmployeeID   uuid.UUID  `json:"employee_id"`
	ClockIn      time.Time  `json:"clock_in"`
	ClockOut     *time.Time `json:"clock_out,omitempty"`
	BreakMinutes int        `json:"break_minutes"`
	TotalHours   *float64   `json:"total_hours,omitempty"`
	ProjectCode  *string    `json:"project_code,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	Status       string     `json:"status"`
	ApprovedBy   *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// PTOBalance represents an employee's PTO balance
type PTOBalance struct {
	ID                   uuid.UUID  `json:"id"`
	EmployeeID           uuid.UUID  `json:"employee_id"`
	VacationDays         float64    `json:"vacation_days"`
	SickDays             float64    `json:"sick_days"`
	PersonalDays         float64    `json:"personal_days"`
	AccrualRateVacation  float64    `json:"accrual_rate_vacation"`
	AccrualRateSick      float64    `json:"accrual_rate_sick"`
	LastAccrualDate      *time.Time `json:"last_accrual_date,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// PTORequest represents a PTO request
type PTORequest struct {
	ID             uuid.UUID  `json:"id"`
	EmployeeID     uuid.UUID  `json:"employee_id"`
	PTOType        string     `json:"pto_type"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        time.Time  `json:"end_date"`
	DaysRequested  float64    `json:"days_requested"`
	Reason         *string    `json:"reason,omitempty"`
	Status         string     `json:"status"`
	ReviewedBy     *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	ReviewNotes    *string    `json:"review_notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// BenefitPlan represents a benefit plan offered by the company
type BenefitPlan struct {
	ID              uuid.UUID              `json:"id"`
	PlanName        string                 `json:"plan_name"`
	PlanType        string                 `json:"plan_type"`
	Provider        *string                `json:"provider,omitempty"`
	Description     *string                `json:"description,omitempty"`
	EmployeeCost    *float64               `json:"employee_cost,omitempty"`
	EmployerCost    *float64               `json:"employer_cost,omitempty"`
	CoverageDetails map[string]interface{} `json:"coverage_details,omitempty"`
	Active          bool                   `json:"active"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// BenefitEnrollment represents an employee's enrollment in a benefit plan
type BenefitEnrollment struct {
	ID              uuid.UUID              `json:"id"`
	EmployeeID      uuid.UUID              `json:"employee_id"`
	PlanID          uuid.UUID              `json:"plan_id"`
	EnrollmentDate  time.Time              `json:"enrollment_date"`
	EffectiveDate   time.Time              `json:"effective_date"`
	TerminationDate *time.Time             `json:"termination_date,omitempty"`
	Status          string                 `json:"status"`
	Dependents      map[string]interface{} `json:"dependents,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// PayrollPeriod represents a payroll period
type PayrollPeriod struct {
	ID          uuid.UUID  `json:"id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	PayDate     time.Time  `json:"pay_date"`
	Status      string     `json:"status"`
	ProcessedBy *uuid.UUID `json:"processed_by,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PayStub represents an employee's pay stub
type PayStub struct {
	ID                  uuid.UUID `json:"id"`
	EmployeeID          uuid.UUID `json:"employee_id"`
	PayrollPeriodID     uuid.UUID `json:"payroll_period_id"`
	GrossPay            float64   `json:"gross_pay"`
	FederalTax          float64   `json:"federal_tax"`
	StateTax            float64   `json:"state_tax"`
	SocialSecurity      float64   `json:"social_security"`
	Medicare            float64   `json:"medicare"`
	OtherDeductions     float64   `json:"other_deductions"`
	NetPay              float64   `json:"net_pay"`
	HoursWorked         *float64  `json:"hours_worked,omitempty"`
	OvertimeHours       *float64  `json:"overtime_hours,omitempty"`
	HourlyRate          *float64  `json:"hourly_rate,omitempty"`
	BenefitsDeductions  float64   `json:"benefits_deductions"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Request/Response DTOs

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	User      User      `json:"user"`
	Employee  *Employee `json:"employee,omitempty"`
}

type ClockInRequest struct {
	EmployeeID  uuid.UUID `json:"employee_id"`
	ProjectCode *string   `json:"project_code,omitempty"`
}

type ClockOutRequest struct {
	TimesheetID  uuid.UUID `json:"timesheet_id"`
	BreakMinutes int       `json:"break_minutes"`
	Notes        *string   `json:"notes,omitempty"`
}

type PTORequestCreate struct {
	PTOType       string    `json:"pto_type"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	DaysRequested float64   `json:"days_requested"`
	Reason        *string   `json:"reason,omitempty"`
}

type PTORequestReview struct {
	Status      string  `json:"status"`
	ReviewNotes *string `json:"review_notes,omitempty"`
}

type EnrollmentCreate struct {
	PlanID         uuid.UUID              `json:"plan_id"`
	EffectiveDate  time.Time              `json:"effective_date"`
	Dependents     map[string]interface{} `json:"dependents,omitempty"`
}

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
