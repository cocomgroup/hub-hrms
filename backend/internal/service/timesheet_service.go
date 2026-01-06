package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
)

// ============================================================================
// SIMPLIFIED TIMESHEET SERVICE
// ============================================================================
// NOTE: This service integrates with ProjectRepository to:
// 1. Validate that projects exist and are active
// 2. Verify employees are assigned to projects they're logging time against
// 3. Fetch project lists for employee selection
// ============================================================================

// TimesheetService  interface defines simplified timesheet operations
type TimesheetService  interface {
	// Time Entry operations (daily)
	CreateTimeEntry(ctx context.Context, employeeID uuid.UUID, req *models.TimeEntryCreateRequest) (*models.TimeEntry, error)
	UpdateTimeEntry(ctx context.Context, entryID uuid.UUID, employeeID uuid.UUID, req *models.TimeEntryUpdateRequest) (*models.TimeEntry, error)
	DeleteTimeEntry(ctx context.Context, entryID uuid.UUID, employeeID uuid.UUID) error
	GetTimeEntry(ctx context.Context, entryID uuid.UUID) (*models.TimeEntry, error)
	GetTimeEntries(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]*models.TimeEntry, error)
	BulkCreateTimeEntries(ctx context.Context, employeeID uuid.UUID, req *models.TimeEntryBulkCreateRequest) ([]*models.TimeEntry, error)
	
	// Timesheet operations (weekly)
	GetWeeklySummary(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) (*models.TimesheetSummary, error)
	SubmitTimesheet(ctx context.Context, employeeID uuid.UUID, req *models.TimesheetSubmitRequest) (*models.Timesheet, error)
	GetTimesheet(ctx context.Context, timesheetID uuid.UUID) (*models.Timesheet, error)
	GetTimesheetsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.Timesheet, error)
	
	// Manager operations
	GetPendingTimesheets(ctx context.Context, managerID uuid.UUID) ([]*models.Timesheet, error)
	ApproveTimesheet(ctx context.Context, timesheetID uuid.UUID, managerID uuid.UUID, req *models.TimesheetApprovalRequest) (*models.Timesheet, error)
	
	// Project operations (read-only, delegates to ProjectRepository)
	GetAvailableProjects(ctx context.Context, employeeID uuid.UUID) ([]*models.Project, error)
}

type timesheetService struct {
	timesheetRepo repository.TimesheetRepository 
	projectRepo   repository.ProjectRepository
}

func NewTimesheetService(
	timesheetRepo repository.TimesheetRepository ,
	projectRepo repository.ProjectRepository,
) TimesheetService  {
	return &timesheetService {
		timesheetRepo: timesheetRepo,
		projectRepo:   projectRepo,
	}
}

// ============================================================================
// TIME ENTRY OPERATIONS (Daily)
// ============================================================================

func (s *timesheetService ) CreateTimeEntry(ctx context.Context, employeeID uuid.UUID, req *models.TimeEntryCreateRequest) (*models.TimeEntry, error) {
	// Validate hours
	if req.Hours < 0 || req.Hours > 24 {
		return nil, fmt.Errorf("hours must be between 0 and 24")
	}
	
	// Validate type
	if req.Type != models.TimeEntryTypeRegular && 
	   req.Type != models.TimeEntryTypePTO && 
	   req.Type != models.TimeEntryTypeHoliday {
		return nil, fmt.Errorf("invalid entry type")
	}
	
	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format")
	}
	
	// If project is specified, validate it
	if req.ProjectID != nil {
		if err := s.validateProjectAccess(ctx, employeeID, *req.ProjectID); err != nil {
			return nil, err
		}
	}
	
	// Check if entry already exists for this date AND project
	existing, _ := s.timesheetRepo.GetTimeEntriesByEmployee(ctx, employeeID, date, date)
	for _, entry := range existing {
		// Check if same date and same project (or both nil)
		sameDate := entry.Date.Format("2006-01-02") == req.Date
		sameProject := false
		
		if entry.ProjectID == nil && req.ProjectID == nil {
			sameProject = true
		} else if entry.ProjectID != nil && req.ProjectID != nil && *entry.ProjectID == *req.ProjectID {
			sameProject = true
		}
		
		if sameDate && sameProject {
			return nil, fmt.Errorf("time entry already exists for this date and project")
		}
	}
	
	// Create entry
	entry := &models.TimeEntry{
		EmployeeID: employeeID,
		Date:       date,
		Hours:      req.Hours,
		ProjectID:  req.ProjectID,
		Type:       req.Type,
		Notes:      req.Notes,
	}
	
	if err := s.timesheetRepo.CreateTimeEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create time entry: %w", err)
	}
	
	return s.timesheetRepo.GetTimeEntry(ctx, entry.ID)
}

func (s *timesheetService ) UpdateTimeEntry(ctx context.Context, entryID uuid.UUID, employeeID uuid.UUID, req *models.TimeEntryUpdateRequest) (*models.TimeEntry, error) {
	// Get existing entry
	entry, err := s.timesheetRepo.GetTimeEntry(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get time entry: %w", err)
	}
	
	// Verify ownership
	if entry.EmployeeID != employeeID {
		return nil, fmt.Errorf("unauthorized")
	}
	
	// Check if timesheet is submitted
	if entry.TimesheetID != nil {
		timesheet, err := s.timesheetRepo.GetTimesheet(ctx, *entry.TimesheetID)
		if err == nil && timesheet.Status != models.TimesheetStatusDraft {
			return nil, fmt.Errorf("cannot modify entry in %s timesheet", timesheet.Status)
		}
	}
	
	// Update fields
	if req.Hours != nil {
		if *req.Hours < 0 || *req.Hours > 24 {
			return nil, fmt.Errorf("hours must be between 0 and 24")
		}
		entry.Hours = *req.Hours
	}
	
	if req.ProjectID != nil {
		// Validate project access
		if err := s.validateProjectAccess(ctx, employeeID, *req.ProjectID); err != nil {
			return nil, err
		}
		entry.ProjectID = req.ProjectID
	}
	
	if req.Type != nil {
		if *req.Type != models.TimeEntryTypeRegular && 
		   *req.Type != models.TimeEntryTypePTO && 
		   *req.Type != models.TimeEntryTypeHoliday {
			return nil, fmt.Errorf("invalid entry type")
		}
		entry.Type = *req.Type
	}
	
	if req.Notes != nil {
		entry.Notes = *req.Notes
	}
	
	if err := s.timesheetRepo.UpdateTimeEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to update time entry: %w", err)
	}
	
	return s.timesheetRepo.GetTimeEntry(ctx, entryID)
}

func (s *timesheetService ) DeleteTimeEntry(ctx context.Context, entryID uuid.UUID, employeeID uuid.UUID) error {
	// Get entry
	entry, err := s.timesheetRepo.GetTimeEntry(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to get time entry: %w", err)
	}
	
	// Verify ownership
	if entry.EmployeeID != employeeID {
		return fmt.Errorf("unauthorized")
	}
	
	// Check if timesheet is submitted
	if entry.TimesheetID != nil {
		timesheet, err := s.timesheetRepo.GetTimesheet(ctx, *entry.TimesheetID)
		if err == nil && timesheet.Status != models.TimesheetStatusDraft {
			return fmt.Errorf("cannot delete entry in %s timesheet", timesheet.Status)
		}
	}
	
	return s.timesheetRepo.DeleteTimeEntry(ctx, entryID)
}

func (s *timesheetService ) GetTimeEntry(ctx context.Context, entryID uuid.UUID) (*models.TimeEntry, error) {
	return s.timesheetRepo.GetTimeEntry(ctx, entryID)
}

func (s *timesheetService ) GetTimeEntries(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]*models.TimeEntry, error) {
	return s.timesheetRepo.GetTimeEntriesByEmployee(ctx, employeeID, startDate, endDate)
}

func (s *timesheetService ) BulkCreateTimeEntries(ctx context.Context, employeeID uuid.UUID, req *models.TimeEntryBulkCreateRequest) ([]*models.TimeEntry, error) {
	var processedEntries []*models.TimeEntry
	var errors []string
	
	// Get date range from request to check existing entries
	if len(req.Entries) == 0 {
		return nil, fmt.Errorf("no entries provided")
	}
	
	// Parse all dates to find min/max range
	var minDate, maxDate time.Time
	for i, entryReq := range req.Entries {
		date, err := time.Parse("2006-01-02", entryReq.Date)
		if err != nil {
			errors = append(errors, fmt.Sprintf("invalid date %s: %v", entryReq.Date, err))
			continue
		}
		if i == 0 || date.Before(minDate) {
			minDate = date
		}
		if i == 0 || date.After(maxDate) {
			maxDate = date
		}
	}
	
	// Get existing entries for this date range
	existingEntries, err := s.timesheetRepo.GetTimeEntriesByEmployee(ctx, employeeID, minDate, maxDate)
	if err != nil {
		// If error, just proceed with creates
		existingEntries = []*models.TimeEntry{}
	}
	
	// Build a map of existing entries: key = "date-projectID"
	existingMap := make(map[string]*models.TimeEntry)
	for _, entry := range existingEntries {
		projectIDStr := "nil"
		if entry.ProjectID != nil {
			projectIDStr = entry.ProjectID.String()
		}
		key := fmt.Sprintf("%s-%s", entry.Date.Format("2006-01-02"), projectIDStr)
		existingMap[key] = entry
	}
	
	// Process each entry request
	for _, entryReq := range req.Entries {
		// Build key for this entry
		projectIDStr := "nil"
		if entryReq.ProjectID != nil {
			projectIDStr = entryReq.ProjectID.String()
		}
		key := fmt.Sprintf("%s-%s", entryReq.Date, projectIDStr)
		
		var entry *models.TimeEntry
		
		// Check if entry exists
		if existingEntry, exists := existingMap[key]; exists {
			// UPDATE existing entry
			updateReq := &models.TimeEntryUpdateRequest{
				Hours:     &entryReq.Hours,
				ProjectID: entryReq.ProjectID,
				Type:      &entryReq.Type,
				Notes:     &entryReq.Notes,
			}
			
			updatedEntry, err := s.UpdateTimeEntry(ctx, existingEntry.ID, employeeID, updateReq)
			if err != nil {
				errors = append(errors, fmt.Sprintf("failed to update entry for %s: %v", entryReq.Date, err))
				continue
			}
			entry = updatedEntry
		} else {
			// CREATE new entry
			createdEntry, err := s.CreateTimeEntry(ctx, employeeID, &entryReq)
			if err != nil {
				errors = append(errors, fmt.Sprintf("failed to create entry for %s: %v", entryReq.Date, err))
				continue
			}
			entry = createdEntry
		}
		
		processedEntries = append(processedEntries, entry)
	}
	
	if len(processedEntries) == 0 {
		errorMsg := "no entries were processed successfully"
		if len(errors) > 0 {
			errorMsg = fmt.Sprintf("%s. Errors: %v", errorMsg, errors)
		}
		return nil, fmt.Errorf("description: %s",errorMsg)
	}
	
	// Return success even if some entries failed
	return processedEntries, nil
}

// ============================================================================
// TIMESHEET OPERATIONS (Weekly)
// ============================================================================

func (s *timesheetService ) GetWeeklySummary(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) (*models.TimesheetSummary, error) {
	entries, err := s.timesheetRepo.GetTimeEntriesByEmployee(ctx, employeeID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	
	summary := &models.TimesheetSummary{
		EmployeeID: employeeID,
		StartDate:  startDate,
		EndDate:    endDate,
		EntryCount: len(entries),
	}
	
	for _, entry := range entries {
		summary.TotalHours += entry.Hours
		
		switch entry.Type {
		case models.TimeEntryTypeRegular:
			summary.RegularHours += entry.Hours
		case models.TimeEntryTypePTO:
			summary.PTOHours += entry.Hours
		case models.TimeEntryTypeHoliday:
			summary.HolidayHours += entry.Hours
		}
	}
	
	// Check if timesheet exists
	timesheet, _ := s.timesheetRepo.GetTimesheetByWeek(ctx, employeeID, startDate, endDate)
	if timesheet != nil {
		summary.Status = timesheet.Status
	} else {
		summary.Status = models.TimesheetStatusDraft
	}
	
	return summary, nil
}

func (s *timesheetService ) SubmitTimesheet(ctx context.Context, employeeID uuid.UUID, req *models.TimesheetSubmitRequest) (*models.Timesheet, error) {
	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format")
	}
	
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format")
	}
	
	// Validate date range (must be a week)
	if endDate.Sub(startDate).Hours() != 6*24 {
		return nil, fmt.Errorf("timesheet must span exactly 7 days (one week)")
	}
	
	// Get time entries for this week
	entries, err := s.timesheetRepo.GetTimeEntriesByEmployee(ctx, employeeID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	
	if len(entries) == 0 {
		return nil, fmt.Errorf("no time entries found for this week")
	}
	
	// Calculate totals
	var totalHours, regularHours, ptoHours, holidayHours float64
	for _, entry := range entries {
		totalHours += entry.Hours
		switch entry.Type {
		case models.TimeEntryTypeRegular:
			regularHours += entry.Hours
		case models.TimeEntryTypePTO:
			ptoHours += entry.Hours
		case models.TimeEntryTypeHoliday:
			holidayHours += entry.Hours
		}
	}
	
	// Check if timesheet already exists
	existingTimesheet, _ := s.timesheetRepo.GetTimesheetByWeek(ctx, employeeID, startDate, endDate)
	
	now := time.Now()
	
	if existingTimesheet != nil {
		// Update existing timesheet
		if existingTimesheet.Status != models.TimesheetStatusDraft && 
		   existingTimesheet.Status != models.TimesheetStatusRejected {
			return nil, fmt.Errorf("timesheet already %s", existingTimesheet.Status)
		}
		
		existingTimesheet.Status = models.TimesheetStatusSubmitted
		existingTimesheet.TotalHours = totalHours
		existingTimesheet.RegularHours = regularHours
		existingTimesheet.PTOHours = ptoHours
		existingTimesheet.HolidayHours = holidayHours
		existingTimesheet.SubmittedAt = &now
		
		if err := s.timesheetRepo.UpdateTimesheet(ctx, existingTimesheet); err != nil {
			return nil, fmt.Errorf("failed to update timesheet: %w", err)
		}
		
		// Link entries to timesheet
		for _, entry := range entries {
			entry.TimesheetID = &existingTimesheet.ID
			s.timesheetRepo.UpdateTimeEntry(ctx, entry)
		}
		
		return s.timesheetRepo.GetTimesheet(ctx, existingTimesheet.ID)
	}
	
	// Create new timesheet
	timesheet := &models.Timesheet{
		EmployeeID:   employeeID,
		StartDate:    startDate,
		EndDate:      endDate,
		Status:       models.TimesheetStatusSubmitted,
		TotalHours:   totalHours,
		RegularHours: regularHours,
		PTOHours:     ptoHours,
		HolidayHours: holidayHours,
		SubmittedAt:  &now,
	}
	
	if err := s.timesheetRepo.CreateTimesheet(ctx, timesheet); err != nil {
		return nil, fmt.Errorf("failed to create timesheet: %w", err)
	}
	
	// Link entries to timesheet
	for _, entry := range entries {
		entry.TimesheetID = &timesheet.ID
		s.timesheetRepo.UpdateTimeEntry(ctx, entry)
	}
	
	return s.timesheetRepo.GetTimesheet(ctx, timesheet.ID)
}

func (s *timesheetService ) GetTimesheet(ctx context.Context, timesheetID uuid.UUID) (*models.Timesheet, error) {
	return s.timesheetRepo.GetTimesheet(ctx, timesheetID)
}

func (s *timesheetService ) GetTimesheetsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.Timesheet, error) {
	return s.timesheetRepo.GetTimesheetsByEmployee(ctx, employeeID)
}

// ============================================================================
// MANAGER OPERATIONS
// ============================================================================

func (s *timesheetService ) GetPendingTimesheets(ctx context.Context, managerID uuid.UUID) ([]*models.Timesheet, error) {
	return s.timesheetRepo.GetTimesheetsByStatus(ctx, models.TimesheetStatusSubmitted, &managerID)
}

func (s *timesheetService ) ApproveTimesheet(ctx context.Context, timesheetID uuid.UUID, managerID uuid.UUID, req *models.TimesheetApprovalRequest) (*models.Timesheet, error) {
	// Get timesheet
	timesheet, err := s.timesheetRepo.GetTimesheet(ctx, timesheetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get timesheet: %w", err)
	}
	
	// Check status
	if timesheet.Status != models.TimesheetStatusSubmitted {
		return nil, fmt.Errorf("timesheet is not submitted for approval")
	}
	
	// Validate action
	if req.Action != "approve" && req.Action != "reject" {
		return nil, fmt.Errorf("invalid action: must be 'approve' or 'reject'")
	}
	
	now := time.Now()
	
	if req.Action == "approve" {
		timesheet.Status = models.TimesheetStatusApproved
		timesheet.ApprovedAt = &now
		timesheet.ApprovedBy = &managerID
		timesheet.RejectionReason = nil
	} else {
		if req.RejectionReason == "" {
			return nil, fmt.Errorf("rejection reason is required")
		}
		timesheet.Status = models.TimesheetStatusRejected
		timesheet.RejectionReason = &req.RejectionReason
		timesheet.ApprovedAt = nil
		timesheet.ApprovedBy = nil
	}
	
	if err := s.timesheetRepo.UpdateTimesheet(ctx, timesheet); err != nil {
		return nil, fmt.Errorf("failed to update timesheet: %w", err)
	}
	
	return s.timesheetRepo.GetTimesheet(ctx, timesheetID)
}

// ============================================================================
// PROJECT OPERATIONS (Read-Only Integration with ProjectRepository)
// ============================================================================

// GetAvailableProjects returns projects the employee is assigned to
func (s *timesheetService ) GetAvailableProjects(ctx context.Context, employeeID uuid.UUID) ([]*models.Project, error) {
	// Get projects the employee is assigned to
	projects, err := s.projectRepo.GetEmployeeProjects(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee projects: %w", err)
	}
	
	// Filter to only active projects
	activeProjects := []*models.Project{}
	for _, project := range projects {
		if project.Status == "active" {
			activeProjects = append(activeProjects, project)
		}
	}
	
	return activeProjects, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// validateProjectAccess verifies that:
// 1. The project exists
// 2. The project is active
// 3. The employee is assigned to the project
func (s *timesheetService ) validateProjectAccess(ctx context.Context, employeeID uuid.UUID, projectID uuid.UUID) error {
	// Get project details
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found")
	}
	
	// Check if project is active
	if project.Status != "active" {
		return fmt.Errorf("project is not active")
	}
	
	// Verify employee is assigned to the project
	employeeProjects, err := s.projectRepo.GetEmployeeProjects(ctx, employeeID)
	if err != nil {
		return fmt.Errorf("failed to verify project access")
	}
	
	isAssigned := false
	for _, p := range employeeProjects {
		if p.ID == projectID {
			isAssigned = true
			break
		}
	}
	
	if !isAssigned {
		return fmt.Errorf("employee is not assigned to this project")
	}
	
	return nil
}