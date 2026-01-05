package models

import (
	"time"
	"github.com/google/uuid"
)

// ============================================================================
// SIMPLIFIED TIMESHEET MODELS
// ============================================================================
// NOTE: This timesheet system integrates with the existing Project system.
// Projects are managed by HR/Project Managers via the project_handler.go
// Timesheet only allows employees to SELECT from projects they're assigned to.
// ============================================================================

// TimeEntry represents a daily time entry (hours worked per day)
type TimeEntry struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	EmployeeID  uuid.UUID  `json:"employee_id" db:"employee_id"`
	Date        time.Time  `json:"date" db:"entry_date"`
	Hours       float64    `json:"hours" db:"hours"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty" db:"project_id"`
	ProjectName string     `json:"project_name,omitempty" db:"project_name"`
	ProjectCode string     `json:"project_code,omitempty" db:"project_code"`
	Type        string     `json:"type" db:"entry_type"` // regular, pto, holiday
	Notes       string     `json:"notes,omitempty" db:"notes"`
	TimesheetID *uuid.UUID `json:"timesheet_id,omitempty" db:"timesheet_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Timesheet represents a weekly timesheet submission
type Timesheet struct {
	ID               uuid.UUID     `json:"id" db:"id"`
	EmployeeID       uuid.UUID     `json:"employee_id" db:"employee_id"`
	EmployeeName     string        `json:"employee_name,omitempty" db:"employee_name"`
	StartDate        time.Time     `json:"start_date" db:"start_date"`
	EndDate          time.Time     `json:"end_date" db:"end_date"`
	Status           string        `json:"status" db:"status"` // draft, submitted, approved, rejected
	TotalHours       float64       `json:"total_hours" db:"total_hours"`
	RegularHours     float64       `json:"regular_hours" db:"regular_hours"`
	PTOHours         float64       `json:"pto_hours" db:"pto_hours"`
	HolidayHours     float64       `json:"holiday_hours" db:"holiday_hours"`
	SubmittedAt      *time.Time    `json:"submitted_at,omitempty" db:"submitted_at"`
	ApprovedAt       *time.Time    `json:"approved_at,omitempty" db:"approved_at"`
	ApprovedBy       *uuid.UUID    `json:"approved_by,omitempty" db:"approved_by"`
	ApproverName     string        `json:"approver_name,omitempty" db:"approver_name"`
	RejectionReason  *string       `json:"rejection_reason,omitempty" db:"rejection_reason"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" db:"updated_at"`
	
	// Optional nested data - use pointer slice for consistency with repository
	TimeEntries      []*TimeEntry  `json:"time_entries,omitempty"`
}

// TimeEntryCreateRequest for creating/updating daily time entries
type TimeEntryCreateRequest struct {
	Date      string     `json:"date" binding:"required"` // "2025-01-06" format
	Hours     float64    `json:"hours" binding:"required,min=0,max=24"`
	ProjectID *uuid.UUID `json:"project_id,omitempty"`
	Type      string     `json:"type" binding:"required"` // regular, pto, holiday
	Notes     string     `json:"notes,omitempty"`
}

// TimeEntryUpdateRequest for updating time entries
type TimeEntryUpdateRequest struct {
	Hours     *float64   `json:"hours,omitempty" binding:"omitempty,min=0,max=24"`
	ProjectID *uuid.UUID `json:"project_id,omitempty"`
	Type      *string    `json:"type,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
}

// TimeEntryBulkCreateRequest for creating multiple entries at once
type TimeEntryBulkCreateRequest struct {
	Entries []TimeEntryCreateRequest `json:"entries" binding:"required,dive"`
}

// TimesheetSubmitRequest for submitting a weekly timesheet
type TimesheetSubmitRequest struct {
	StartDate string `json:"start_date" binding:"required"` // Monday of the week
	EndDate   string `json:"end_date" binding:"required"`   // Sunday of the week
}

// TimesheetApprovalRequest for manager approval/rejection
type TimesheetApprovalRequest struct {
	Action          string `json:"action" binding:"required"` // approve, reject
	RejectionReason string `json:"rejection_reason,omitempty"`
}

// TimesheetSummary provides aggregated hours information for a week
type TimesheetSummary struct {
	EmployeeID    uuid.UUID `json:"employee_id"`
	EmployeeName  string    `json:"employee_name,omitempty"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	RegularHours  float64   `json:"regular_hours"`
	PTOHours      float64   `json:"pto_hours"`
	HolidayHours  float64   `json:"holiday_hours"`
	TotalHours    float64   `json:"total_hours"`
	EntryCount    int       `json:"entry_count"`
	Status        string    `json:"status"`
}

// Status constants
const (
	TimesheetStatusDraft     = "draft"
	TimesheetStatusSubmitted = "submitted"
	TimesheetStatusApproved  = "approved"
	TimesheetStatusRejected  = "rejected"
)

// Type constants
const (
	TimeEntryTypeRegular = "regular"
	TimeEntryTypePTO     = "pto"
	TimeEntryTypeHoliday = "holiday"
)
