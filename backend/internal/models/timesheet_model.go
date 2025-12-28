package models

import (
	"time"
	"github.com/google/uuid"
)

// ============================================================================
// Timesheet Models - ADD TO models.go
// ============================================================================

// Timesheet represents a time entry
type Timesheet struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	EmployeeID   uuid.UUID  `json:"employee_id" db:"employee_id"`
	Date         time.Time  `json:"date" db:"entry_date"`
	ClockIn      *time.Time `json:"clock_in,omitempty" db:"clock_in"`
	ClockOut     *time.Time `json:"clock_out,omitempty" db:"clock_out"`
	BreakMinutes int        `json:"break_minutes" db:"break_duration"`
	Notes        string     `json:"notes,omitempty" db:"notes"`
	Type         string     `json:"type" db:"entry_type"` // regular, overtime, pto, sick, holiday
	Status       string     `json:"status" db:"status"`   // draft, submitted, approved, rejected
	TotalHours   *float64   `json:"total_hours,omitempty" db:"total_hours"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	EmployeeName string     `json:"employee_name,omitempty" db:"employee_name"`
}


// TimesheetSummary provides aggregated hours information
type TimesheetSummary struct {
	EmployeeID    uuid.UUID `json:"employee_id"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	RegularHours  float64   `json:"regular_hours"`
	OvertimeHours float64   `json:"overtime_hours"`
	PTOHours      float64   `json:"pto_hours"`
	TotalHours    float64   `json:"total_hours"`
	EntryCount    int       `json:"entry_count"`
}

// TimesheetCreateRequest for creating manual time entries
type TimesheetCreateRequest struct {
	EmployeeID   uuid.UUID  `json:"employee_id"`
	Date         time.Time  `json:"date"`
	ClockIn      *time.Time `json:"clock_in,omitempty"`
	ClockOut     *time.Time `json:"clock_out,omitempty"`
	BreakMinutes int        `json:"break_minutes"`
	Notes        string     `json:"notes,omitempty"`
	Type         string     `json:"type"` // regular, overtime, pto, sick, holiday
}

// TimesheetUpdateRequest for updating time entries
type TimesheetUpdateRequest struct {
	ClockIn      *time.Time `json:"clock_in,omitempty"`
	ClockOut     *time.Time `json:"clock_out,omitempty"`
	BreakMinutes *int       `json:"break_minutes,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	Type         *string    `json:"type,omitempty"`
}

// TimesheetApprovalRequest for approving/rejecting timesheets
type TimesheetApprovalRequest struct {
	Status         string `json:"status"` // approved, rejected
	RejectionNotes string `json:"rejection_notes,omitempty"`
}

// ProjectCreateRequest for creating projects
type ProjectCreateRequest struct {
	Name        string     `json:"name" binding:"required"`
	Code        string     `json:"code" binding:"required"`
	Description string     `json:"description,omitempty"`
	ClientName  string     `json:"client_name,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	BudgetHours *float64   `json:"budget_hours,omitempty"`
}

// ProjectAllocationReq for allocating hours to projects
type ProjectAllocationReq struct {
	ProjectID uuid.UUID `json:"project_id"`
	Hours     float64   `json:"hours"`
	Notes     string    `json:"notes,omitempty"`
}

// TimeEntryProject links a time entry to a project
type TimeEntryProject struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TimeEntryID uuid.UUID `json:"time_entry_id" db:"time_entry_id"`
	ProjectID   uuid.UUID `json:"project_id" db:"project_id"`
	Hours       float64   `json:"hours" db:"hours"`
	Notes       string    `json:"notes,omitempty" db:"notes"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ClockInRequest represents a clock-in request
type ClockInRequest struct {
	Notes string `json:"notes,omitempty"`
}

// ClockOutRequest represents a clock-out request
type ClockOutRequest struct {
	BreakMinutes int    `json:"break_minutes"`
	Notes        string `json:"notes,omitempty"`
}
