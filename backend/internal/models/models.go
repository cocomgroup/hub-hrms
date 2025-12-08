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
	Phone                  *string    `json:"phone,omitempty"`
	DateOfBirth            *time.Time `json:"date_of_birth,omitempty"`
	HireDate               time.Time  `json:"hire_date"`
	Department             *string    `json:"department,omitempty"`
	Position               *string    `json:"position,omitempty"`
	ManagerID              *uuid.UUID `json:"manager_id,omitempty"`
	EmploymentType         *string    `json:"employment_type,omitempty"`
	Status                 string     `json:"status"`
	StreetAddress          *string    `json:"street_address,omitempty"`
	City                   *string    `json:"city,omitempty"`
	State                  *string    `json:"state,omitempty"`
	ZipCode                *string    `json:"zip_code,omitempty"`
	Country                *string    `json:"country,omitempty"`
	EmergencyContactName   *string    `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone  *string    `json:"emergency_contact_phone,omitempty"`
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
