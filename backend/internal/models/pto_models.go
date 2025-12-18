package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// PTO Balance Models
// ============================================================================

// PTOBalance represents an employee's PTO balance for a given year
type PTOBalance struct {
	ID           uuid.UUID `json:"id" db:"id"`
	EmployeeID   uuid.UUID `json:"employee_id" db:"employee_id"`
	VacationDays float64   `json:"vacation_days" db:"vacation_days"`
	SickDays     float64   `json:"sick_days" db:"sick_days"`
	PersonalDays float64   `json:"personal_days" db:"personal_days"`
	Year         int       `json:"year" db:"year"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ============================================================================
// PTO Request Models
// ============================================================================

// PTORequest represents a PTO request from an employee
type PTORequest struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	EmployeeID    uuid.UUID  `json:"employee_id" db:"employee_id"`
	PTOType       string     `json:"pto_type" db:"pto_type"` // vacation, sick, personal
	StartDate     time.Time  `json:"start_date" db:"start_date"`
	EndDate       time.Time  `json:"end_date" db:"end_date"`
	DaysRequested float64    `json:"days_requested" db:"days_requested"`
	Reason        string     `json:"reason,omitempty" db:"reason"`
	Status        string     `json:"status" db:"status"` // pending, approved, denied, cancelled
	ReviewedBy    *uuid.UUID `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewNotes   string     `json:"review_notes,omitempty" db:"review_notes"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`

	// Joined fields
	EmployeeName string `json:"employee_name,omitempty" db:"employee_name"`
	Department   string `json:"department,omitempty" db:"department"`
	ReviewerName string `json:"reviewer_name,omitempty" db:"reviewer_name"`
}

// ============================================================================
// Request/Response Models
// ============================================================================

// PTORequestCreate represents a request to create a new PTO request
type PTORequestCreate struct {
	PTOType       string    `json:"pto_type" binding:"required"`       // vacation, sick, personal
	StartDate     time.Time `json:"start_date" binding:"required"`
	EndDate       time.Time `json:"end_date" binding:"required"`
	DaysRequested float64   `json:"days_requested" binding:"required"` // Can be decimal for half-days
	Reason        string    `json:"reason,omitempty"`
}

// PTORequestReview represents a manager's review of a PTO request
type PTORequestReview struct {
	Status      string `json:"status" binding:"required"` // approved, denied
	ReviewNotes string `json:"review_notes,omitempty"`
}

// PTORequestUpdate represents updates to a PTO request (for employee to cancel)
type PTORequestUpdate struct {
	Status string `json:"status"` // Only "cancelled" is allowed
}

// ============================================================================
// Response Models
// ============================================================================

// PTOBalanceResponse includes balance and usage information
type PTOBalanceResponse struct {
	Balance      PTOBalance `json:"balance"`
	Used         PTOUsage   `json:"used"`
	Available    PTOUsage   `json:"available"`
	PendingDays  float64    `json:"pending_days"`
}

// PTOUsage represents PTO days used or available
type PTOUsage struct {
	VacationDays float64 `json:"vacation_days"`
	SickDays     float64 `json:"sick_days"`
	PersonalDays float64 `json:"personal_days"`
}

// PTORequestSummary provides a summary of PTO requests
type PTORequestSummary struct {
	TotalRequests    int     `json:"total_requests"`
	PendingRequests  int     `json:"pending_requests"`
	ApprovedRequests int     `json:"approved_requests"`
	DeniedRequests   int     `json:"denied_requests"`
	TotalDaysUsed    float64 `json:"total_days_used"`
}

// ============================================================================
// Validation Methods
// ============================================================================

// Validate validates a PTORequestCreate
func (r *PTORequestCreate) Validate() error {
	if r.PTOType != "vacation" && r.PTOType != "sick" && r.PTOType != "personal" {
		return fmt.Errorf("invalid pto_type: must be vacation, sick, or personal")
	}

	if r.StartDate.IsZero() {
		return fmt.Errorf("start_date is required")
	}

	if r.EndDate.IsZero() {
		return fmt.Errorf("end_date is required")
	}

	if r.EndDate.Before(r.StartDate) {
		return fmt.Errorf("end_date must be after start_date")
	}

	if r.DaysRequested <= 0 {
		return fmt.Errorf("days_requested must be greater than 0")
	}

	// Calculate expected days (simple business day count)
	expectedDays := calculateBusinessDays(r.StartDate, r.EndDate)
	if r.DaysRequested > float64(expectedDays) {
		return fmt.Errorf("days_requested exceeds the number of business days in the date range")
	}

	return nil
}

// Validate validates a PTORequestReview
func (r *PTORequestReview) Validate() error {
	if r.Status != "approved" && r.Status != "denied" {
		return fmt.Errorf("status must be 'approved' or 'denied'")
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// calculateBusinessDays calculates the number of business days between two dates
func calculateBusinessDays(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}

	count := 0
	current := start

	for !current.After(end) {
		// Skip weekends (Saturday = 6, Sunday = 0)
		if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
			count++
		}
		current = current.AddDate(0, 0, 1)
	}

	return count
}

// IsPending checks if the request is pending
func (r *PTORequest) IsPending() bool {
	return r.Status == "pending"
}

// IsApproved checks if the request is approved
func (r *PTORequest) IsApproved() bool {
	return r.Status == "approved"
}

// IsDenied checks if the request is denied
func (r *PTORequest) IsDenied() bool {
	return r.Status == "denied"
}

// IsCancelled checks if the request is cancelled
func (r *PTORequest) IsCancelled() bool {
	return r.Status == "cancelled"
}

// CanCancel checks if the request can be cancelled by the employee
func (r *PTORequest) CanCancel() bool {
	return r.Status == "pending" || r.Status == "approved"
}

// CanReview checks if the request can be reviewed
func (r *PTORequest) CanReview() bool {
	return r.Status == "pending"
}
