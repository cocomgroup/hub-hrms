package service

import (
	"context"
	"errors"
	"log"
	"fmt"
	"time"

	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
)

var (
	ErrAlreadyClockedIn  = errors.New("already clocked in")
	ErrNotClockedIn      = errors.New("not clocked in")
	ErrInvalidTimeRange  = errors.New("invalid time range")
	ErrTimesheetNotDraft = errors.New("timesheet must be in draft status")
)

// TimesheetService interface defines timesheet operations
type TimesheetService interface {
	ClockIn(ctx context.Context, employeeID uuid.UUID, notes string) (*models.Timesheet, error)
	ClockOut(ctx context.Context, employeeID uuid.UUID, breakMinutes int, notes string) (*models.Timesheet, error)
	GetActiveClockIn(ctx context.Context, employeeID uuid.UUID) (*models.Timesheet, error)
	CreateTimeEntry(ctx context.Context, req *models.TimesheetCreateRequest) (*models.Timesheet, error)
	UpdateTimeEntry(ctx context.Context, id uuid.UUID, req *models.TimesheetUpdateRequest) (*models.Timesheet, error)
	DeleteTimeEntry(ctx context.Context, id uuid.UUID) error
	GetTimeEntries(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]*models.Timesheet, error)
	GetTimesheetsByStatus(ctx context.Context, status string) ([]*models.Timesheet, error)
	SubmitTimesheet(ctx context.Context, id uuid.UUID, employeeID uuid.UUID) (*models.Timesheet, error)
	ApproveTimesheet(ctx context.Context, id uuid.UUID, req *models.TimesheetApprovalRequest) (*models.Timesheet, error)
	GetPendingApprovals(ctx context.Context) ([]*models.Timesheet, error)
	GetEmployeeSummary(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) (*models.TimesheetSummary, error)
	GetProjects(ctx context.Context) ([]*models.Project, error)
	CreateProject(ctx context.Context, req *models.ProjectCreateRequest) (*models.Project, error)
}

// timesheetService implements TimesheetService interface
type timesheetService struct {
	repo repository.TimesheetRepository
}

// NewTimesheetService creates a new timesheet service
func NewTimesheetService(repos *repository.Repositories) TimesheetService {
	return &timesheetService{repo: repos.Timesheet}
}

// ClockIn creates a clock-in entry for the employee
func (s *timesheetService) ClockIn(ctx context.Context, employeeID uuid.UUID, notes string) (*models.Timesheet, error) {
	// Check if already clocked in
	active, err := s.repo.GetActiveTimesheet(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check active clock-in: %w", err)
	}
	if active != nil {
		return nil, ErrAlreadyClockedIn
	}

	// Create new clock-in entry
	now := time.Now()
	entry := &models.Timesheet{
		EmployeeID: employeeID,
		Date:       time.Now().Truncate(24 * time.Hour),
		ClockIn:    &now,
		Notes:      notes,
		Type:       "regular",
		Status:     "draft",
	}

	if err := s.repo.Create(ctx, entry); err != nil {
		log.Printf("Timesheet params: %s", employeeID)
		return nil, fmt.Errorf("failed to create clock-in: %w", err)
	}

	return entry, nil
}

// ClockOut updates the active clock-in with clock-out time
func (s *timesheetService) ClockOut(ctx context.Context, employeeID uuid.UUID, breakMinutes int, notes string) (*models.Timesheet, error) {
	// Get active clock-in
	entry, err := s.repo.GetActiveTimesheet(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active clock-in: %w", err)
	}
	if entry == nil {
		return nil, ErrNotClockedIn
	}

	// Update with clock-out
	now := time.Now()
	entry.ClockOut = &now
	entry.BreakMinutes = breakMinutes
	if notes != "" {
		entry.Notes = notes
	}

	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to clock out: %w", err)
	}

	// Get updated entry with calculated hours
	return s.repo.GetByID(ctx, entry.ID)
}

// GetActiveClockIn gets the employee's active clock-in
func (s *timesheetService) GetActiveClockIn(ctx context.Context, employeeID uuid.UUID) (*models.Timesheet, error) {
	return s.repo.GetActiveTimesheet(ctx, employeeID)
}

// CreateTimeEntry creates a manual time entry
func (s *timesheetService) CreateTimeEntry(ctx context.Context, req *models.TimesheetCreateRequest) (*models.Timesheet, error) {
	// Validate times
	if req.ClockIn != nil && req.ClockOut != nil && req.ClockOut.Before(*req.ClockIn) {
		return nil, ErrInvalidTimeRange
	}

	// Create entry
	entry := &models.Timesheet{
		EmployeeID:   req.EmployeeID,
		Date:         req.Date.Truncate(24 * time.Hour),
		ClockIn:      req.ClockIn,
		ClockOut:     req.ClockOut,
		BreakMinutes: req.BreakMinutes,
		Notes:        req.Notes,
		Type:         req.Type,
		Status:       "draft",
	}

	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create time entry: %w", err)
	}

	return s.repo.GetByID(ctx, entry.ID)
}

// UpdateTimeEntry updates an existing time entry
func (s *timesheetService) UpdateTimeEntry(ctx context.Context, id uuid.UUID, req *models.TimesheetUpdateRequest) (*models.Timesheet, error) {
	// Get existing entry
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get time entry: %w", err)
	}

	// Check if entry can be modified (must be draft)
	if entry.Status != "draft" {
		return nil, ErrTimesheetNotDraft
	}

	// Update fields
	if req.ClockIn != nil {
		entry.ClockIn = req.ClockIn
	}
	if req.ClockOut != nil {
		entry.ClockOut = req.ClockOut
	}
	if req.BreakMinutes != nil {
		entry.BreakMinutes = *req.BreakMinutes
	}
	if req.Notes != nil {
		entry.Notes = *req.Notes
	}
	if req.Type != nil {
		entry.Type = *req.Type
	}

	// Validate times
	if entry.ClockIn != nil && entry.ClockOut != nil && entry.ClockOut.Before(*entry.ClockIn) {
		return nil, ErrInvalidTimeRange
	}

	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to update time entry: %w", err)
	}

	return s.repo.GetByID(ctx, id)
}

// DeleteTimeEntry deletes a time entry
func (s *timesheetService) DeleteTimeEntry(ctx context.Context, id uuid.UUID) error {
	// Get entry
	entry, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get time entry: %w", err)
	}

	// Check if entry can be deleted (must be draft)
	if entry.Status != "draft" {
		return ErrTimesheetNotDraft
	}

	// Use type assertion to access Delete if available
	if deleter, ok := s.repo.(interface {
		Delete(ctx context.Context, id uuid.UUID) error
	}); ok {
		return deleter.Delete(ctx, id)
	}

	return fmt.Errorf("delete not supported")
}

// GetTimeEntries gets time entries for an employee
func (s *timesheetService) GetTimeEntries(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]*models.Timesheet, error) {
	filters := map[string]interface{}{
		"start_date": startDate,
		"end_date":   endDate,
	}
	return s.repo.GetByEmployee(ctx, employeeID, filters)
}

// GetTimesheetsByStatus gets timesheets by status (for manager approval)
func (s *timesheetService) GetTimesheetsByStatus(ctx context.Context, status string) ([]*models.Timesheet, error) {
	filters := map[string]interface{}{
		"status": status,
	}
	return s.repo.List(ctx, filters)
}

// SubmitTimesheet submits a timesheet for approval
func (s *timesheetService) SubmitTimesheet(ctx context.Context, id uuid.UUID, employeeID uuid.UUID) (*models.Timesheet, error) {
	// Get timesheet
	timesheet, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get timesheet: %w", err)
	}

	// Verify employee owns this timesheet
	if timesheet.EmployeeID != employeeID {
		return nil, ErrUnauthorized
	}

	// Check status
	if timesheet.Status != "draft" {
		return nil, fmt.Errorf("timesheet already submitted")
	}

	// Update status
	timesheet.Status = "submitted"

	if err := s.repo.Update(ctx, timesheet); err != nil {
		return nil, fmt.Errorf("failed to submit timesheet: %w", err)
	}

	return timesheet, nil
}

// ApproveTimesheet approves or rejects a timesheet
func (s *timesheetService) ApproveTimesheet(ctx context.Context, id uuid.UUID, req *models.TimesheetApprovalRequest) (*models.Timesheet, error) {
	// Get timesheet
	timesheet, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get timesheet: %w", err)
	}

	// Check status
	if timesheet.Status != "submitted" {
		return nil, fmt.Errorf("timesheet not submitted for approval")
	}

	// Validate approval status
	if req.Status != "approved" && req.Status != "rejected" {
		return nil, fmt.Errorf("invalid approval status")
	}

	// Update timesheet
	timesheet.Status = req.Status
	if req.Status == "rejected" {
		timesheet.Notes = req.RejectionNotes
	}

	if err := s.repo.Update(ctx, timesheet); err != nil {
		return nil, fmt.Errorf("failed to approve timesheet: %w", err)
	}

	return timesheet, nil
}

// GetPendingApprovals gets all timesheets pending approval
func (s *timesheetService) GetPendingApprovals(ctx context.Context) ([]*models.Timesheet, error) {
	// Use type assertion to access GetPendingApprovals if available
	if getter, ok := s.repo.(interface {
		GetPendingApprovals(ctx context.Context) ([]*models.Timesheet, error)
	}); ok {
		return getter.GetPendingApprovals(ctx)
	}

	// Fallback to filter by status
	filters := map[string]interface{}{
		"status": "submitted",
	}
	return s.repo.List(ctx, filters)
}

// GetEmployeeSummary gets hours summary for an employee
func (s *timesheetService) GetEmployeeSummary(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) (*models.TimesheetSummary, error) {
	// Use type assertion to access summary method if available
	if summarizer, ok := s.repo.(interface {
		GetEmployeeHoursSummary(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) (*models.TimesheetSummary, error)
	}); ok {
		return summarizer.GetEmployeeHoursSummary(ctx, employeeID, startDate, endDate)
	}

	return nil, fmt.Errorf("summary not supported")
}

// GetProjects gets all active projects
func (s *timesheetService) GetProjects(ctx context.Context) ([]*models.Project, error) {
	// Use type assertion to access projects method if available
	if projectGetter, ok := s.repo.(interface {
		GetProjects(ctx context.Context) ([]*models.Project, error)
	}); ok {
		return projectGetter.GetProjects(ctx)
	}

	return nil, fmt.Errorf("projects not supported")
}

// CreateProject creates a new project
func (s *timesheetService) CreateProject(ctx context.Context, req *models.ProjectCreateRequest) (*models.Project, error) {
	project := &models.Project{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		ClientName:  req.ClientName,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		BudgetHours: req.BudgetHours,
		Status:      "active",
	}

	// Use type assertion to access project creation if available
	if projectCreator, ok := s.repo.(interface {
		CreateProject(ctx context.Context, project *models.Project) error
	}); ok {
		if err := projectCreator.CreateProject(ctx, project); err != nil {
			return nil, err
		}
		return project, nil
	}

	return nil, fmt.Errorf("project creation not supported")
}
