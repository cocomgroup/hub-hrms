package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"hub-hrms/backend/internal/models"
)

// ============================================================================
// SIMPLIFIED TIMESHEET REPOSITORY
// ============================================================================
// NOTE: This repository ONLY manages time entries and timesheets.
// Projects are managed via ProjectRepository (project_repository.go).
// ============================================================================

// TimesheetRepository interface defines timesheet and time entry operations
type TimesheetRepository interface {
	// Time Entry operations (daily)
	CreateTimeEntry(ctx context.Context, entry *models.TimeEntry) error
	GetTimeEntry(ctx context.Context, id uuid.UUID) (*models.TimeEntry, error)
	UpdateTimeEntry(ctx context.Context, entry *models.TimeEntry) error
	DeleteTimeEntry(ctx context.Context, id uuid.UUID) error
	GetTimeEntriesByEmployee(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]*models.TimeEntry, error)
	GetTimeEntriesByTimesheet(ctx context.Context, timesheetID uuid.UUID) ([]*models.TimeEntry, error)
	
	// Timesheet operations (weekly)
	CreateTimesheet(ctx context.Context, timesheet *models.Timesheet) error
	GetTimesheet(ctx context.Context, id uuid.UUID) (*models.Timesheet, error)
	GetTimesheetByWeek(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) (*models.Timesheet, error)
	UpdateTimesheet(ctx context.Context, timesheet *models.Timesheet) error
	GetTimesheetsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.Timesheet, error)
	GetTimesheetsByStatus(ctx context.Context, status string, managerID *uuid.UUID) ([]*models.Timesheet, error)
}

type timesheetRepository struct {
	db *pgxpool.Pool
}

func NewTimesheetRepository(db *pgxpool.Pool) TimesheetRepository {
	return &timesheetRepository{db: db}
}

// ============================================================================
// TIME ENTRY OPERATIONS (Daily)
// ============================================================================

func (r *timesheetRepository) CreateTimeEntry(ctx context.Context, entry *models.TimeEntry) error {
	query := `
		INSERT INTO time_entries (
			id, employee_id, entry_date, hours, project_id, entry_type, notes, timesheet_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		RETURNING created_at, updated_at
	`
	
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	
	return r.db.QueryRow(
		ctx, query,
		entry.ID, entry.EmployeeID, entry.Date, entry.Hours, entry.ProjectID,
		entry.Type, entry.Notes, entry.TimesheetID,
	).Scan(&entry.CreatedAt, &entry.UpdatedAt)
}

func (r *timesheetRepository) GetTimeEntry(ctx context.Context, id uuid.UUID) (*models.TimeEntry, error) {
	query := `
		SELECT 
			te.id, te.employee_id, te.entry_date, te.hours, te.project_id,
			te.entry_type, te.notes, te.timesheet_id,
			te.created_at, te.updated_at,
			p.name as project_name,
			p.code as project_code
		FROM time_entries te
		LEFT JOIN projects p ON p.id = te.project_id
		WHERE te.id = $1
	`
	
	entry := &models.TimeEntry{}
	var projectName, projectCode sql.NullString
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&entry.ID, &entry.EmployeeID, &entry.Date, &entry.Hours, &entry.ProjectID,
		&entry.Type, &entry.Notes, &entry.TimesheetID,
		&entry.CreatedAt, &entry.UpdatedAt,
		&projectName, &projectCode,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("time entry not found")
	}
	if err != nil {
		return nil, err
	}
	
	if projectName.Valid {
		entry.ProjectName = projectName.String
	}
	if projectCode.Valid {
		entry.ProjectCode = projectCode.String
	}
	
	return entry, nil
}

func (r *timesheetRepository) UpdateTimeEntry(ctx context.Context, entry *models.TimeEntry) error {
	query := `
		UPDATE time_entries
		SET hours = $1,
		    project_id = $2,
		    entry_type = $3,
		    notes = $4,
		    updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`
	
	return r.db.QueryRow(
		ctx, query,
		entry.Hours, entry.ProjectID, entry.Type, entry.Notes, entry.ID,
	).Scan(&entry.UpdatedAt)
}

func (r *timesheetRepository) DeleteTimeEntry(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM time_entries 
		WHERE id = $1 
		AND (timesheet_id IS NULL OR timesheet_id IN (
			SELECT id FROM timesheets WHERE status = 'draft'
		))
	`
	
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("time entry not found or cannot be deleted")
	}
	
	return nil
}

func (r *timesheetRepository) GetTimeEntriesByEmployee(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]*models.TimeEntry, error) {
	query := `
		SELECT 
			te.id, te.employee_id, te.entry_date, te.hours, te.project_id,
			te.entry_type, te.notes, te.timesheet_id,
			te.created_at, te.updated_at,
			p.name as project_name,
			p.code as project_code
		FROM time_entries te
		LEFT JOIN projects p ON p.id = te.project_id
		WHERE te.employee_id = $1
		  AND te.entry_date >= $2
		  AND te.entry_date <= $3
		ORDER BY te.entry_date DESC
	`
	
	rows, err := r.db.Query(ctx, query, employeeID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var entries []*models.TimeEntry
	for rows.Next() {
		entry := &models.TimeEntry{}
		var projectName, projectCode sql.NullString
		
		err := rows.Scan(
			&entry.ID, &entry.EmployeeID, &entry.Date, &entry.Hours, &entry.ProjectID,
			&entry.Type, &entry.Notes, &entry.TimesheetID,
			&entry.CreatedAt, &entry.UpdatedAt,
			&projectName, &projectCode,
		)
		if err != nil {
			return nil, err
		}
		
		if projectName.Valid {
			entry.ProjectName = projectName.String
		}
		if projectCode.Valid {
			entry.ProjectCode = projectCode.String
		}
		
		entries = append(entries, entry)
	}
	
	return entries, rows.Err()
}

func (r *timesheetRepository) GetTimeEntriesByTimesheet(ctx context.Context, timesheetID uuid.UUID) ([]*models.TimeEntry, error) {
	query := `
		SELECT 
			te.id, te.employee_id, te.entry_date, te.hours, te.project_id,
			te.entry_type, te.notes, te.timesheet_id,
			te.created_at, te.updated_at,
			p.name as project_name,
			p.code as project_code
		FROM time_entries te
		LEFT JOIN projects p ON p.id = te.project_id
		WHERE te.timesheet_id = $1
		ORDER BY te.entry_date ASC
	`
	
	rows, err := r.db.Query(ctx, query, timesheetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var entries []*models.TimeEntry
	for rows.Next() {
		entry := &models.TimeEntry{}
		var projectName, projectCode sql.NullString
		
		err := rows.Scan(
			&entry.ID, &entry.EmployeeID, &entry.Date, &entry.Hours, &entry.ProjectID,
			&entry.Type, &entry.Notes, &entry.TimesheetID,
			&entry.CreatedAt, &entry.UpdatedAt,
			&projectName, &projectCode,
		)
		if err != nil {
			return nil, err
		}
		
		if projectName.Valid {
			entry.ProjectName = projectName.String
		}
		if projectCode.Valid {
			entry.ProjectCode = projectCode.String
		}
		
		entries = append(entries, entry)
	}
	
	return entries, rows.Err()
}

// ============================================================================
// TIMESHEET OPERATIONS (Weekly)
// ============================================================================

func (r *timesheetRepository) CreateTimesheet(ctx context.Context, timesheet *models.Timesheet) error {
	query := `
		INSERT INTO timesheets (
			id, employee_id, start_date, end_date, status, 
			total_hours, regular_hours, pto_hours, holiday_hours
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		RETURNING created_at, updated_at
	`
	
	if timesheet.ID == uuid.Nil {
		timesheet.ID = uuid.New()
	}
	
	return r.db.QueryRow(
		ctx, query,
		timesheet.ID, timesheet.EmployeeID, timesheet.StartDate, timesheet.EndDate,
		timesheet.Status, timesheet.TotalHours, timesheet.RegularHours,
		timesheet.PTOHours, timesheet.HolidayHours,
	).Scan(&timesheet.CreatedAt, &timesheet.UpdatedAt)
}

func (r *timesheetRepository) GetTimesheet(ctx context.Context, id uuid.UUID) (*models.Timesheet, error) {
	query := `
		SELECT 
			t.id, t.employee_id, t.start_date, t.end_date, t.status,
			t.total_hours, t.regular_hours, t.pto_hours, t.holiday_hours,
			t.submitted_at, t.approved_at, t.approved_by, t.rejection_reason,
			t.created_at, t.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name,
			CONCAT(a.first_name, ' ', a.last_name) as approver_name
		FROM timesheets t
		JOIN employees e ON e.id = t.employee_id
		LEFT JOIN employees a ON a.id = t.approved_by
		WHERE t.id = $1
	`
	
	timesheet := &models.Timesheet{}
	var approverName sql.NullString
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&timesheet.ID, &timesheet.EmployeeID, &timesheet.StartDate, &timesheet.EndDate,
		&timesheet.Status, &timesheet.TotalHours, &timesheet.RegularHours,
		&timesheet.PTOHours, &timesheet.HolidayHours,
		&timesheet.SubmittedAt, &timesheet.ApprovedAt, &timesheet.ApprovedBy,
		&timesheet.RejectionReason, &timesheet.CreatedAt, &timesheet.UpdatedAt,
		&timesheet.EmployeeName, &approverName,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("timesheet not found")
	}
	if err != nil {
		return nil, err
	}
	
	if approverName.Valid {
		timesheet.ApproverName = approverName.String
	}
	
	// Load time entries
	entries, err := r.GetTimeEntriesByTimesheet(ctx, id)
	if err != nil {
		return nil, err
	}
	timesheet.TimeEntries = entries
	
	return timesheet, nil
}

func (r *timesheetRepository) GetTimesheetByWeek(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) (*models.Timesheet, error) {
	query := `
		SELECT 
			t.id, t.employee_id, t.start_date, t.end_date, t.status,
			t.total_hours, t.regular_hours, t.pto_hours, t.holiday_hours,
			t.submitted_at, t.approved_at, t.approved_by, t.rejection_reason,
			t.created_at, t.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name,
			CONCAT(a.first_name, ' ', a.last_name) as approver_name
		FROM timesheets t
		JOIN employees e ON e.id = t.employee_id
		LEFT JOIN employees a ON a.id = t.approved_by
		WHERE t.employee_id = $1
		  AND t.start_date = $2
		  AND t.end_date = $3
		LIMIT 1
	`
	
	timesheet := &models.Timesheet{}
	var approverName sql.NullString
	
	err := r.db.QueryRow(ctx, query, employeeID, startDate, endDate).Scan(
		&timesheet.ID, &timesheet.EmployeeID, &timesheet.StartDate, &timesheet.EndDate,
		&timesheet.Status, &timesheet.TotalHours, &timesheet.RegularHours,
		&timesheet.PTOHours, &timesheet.HolidayHours,
		&timesheet.SubmittedAt, &timesheet.ApprovedAt, &timesheet.ApprovedBy,
		&timesheet.RejectionReason, &timesheet.CreatedAt, &timesheet.UpdatedAt,
		&timesheet.EmployeeName, &approverName,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil // No timesheet for this week is not an error
	}
	if err != nil {
		return nil, err
	}
	
	if approverName.Valid {
		timesheet.ApproverName = approverName.String
	}
	
	// Load time entries
	entries, err := r.GetTimeEntriesByTimesheet(ctx, timesheet.ID)
	if err != nil {
		return nil, err
	}
	timesheet.TimeEntries = entries
	
	return timesheet, nil
}

func (r *timesheetRepository) UpdateTimesheet(ctx context.Context, timesheet *models.Timesheet) error {
	query := `
		UPDATE timesheets
		SET status = $1,
		    total_hours = $2,
		    regular_hours = $3,
		    pto_hours = $4,
		    holiday_hours = $5,
		    submitted_at = $6,
		    approved_at = $7,
		    approved_by = $8,
		    rejection_reason = $9,
		    updated_at = NOW()
		WHERE id = $10
		RETURNING updated_at
	`
	
	return r.db.QueryRow(
		ctx, query,
		timesheet.Status, timesheet.TotalHours, timesheet.RegularHours,
		timesheet.PTOHours, timesheet.HolidayHours,
		timesheet.SubmittedAt, timesheet.ApprovedAt, timesheet.ApprovedBy,
		timesheet.RejectionReason, timesheet.ID,
	).Scan(&timesheet.UpdatedAt)
}

func (r *timesheetRepository) GetTimesheetsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.Timesheet, error) {
	query := `
		SELECT 
			t.id, t.employee_id, t.start_date, t.end_date, t.status,
			t.total_hours, t.regular_hours, t.pto_hours, t.holiday_hours,
			t.submitted_at, t.approved_at, t.approved_by, t.rejection_reason,
			t.created_at, t.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name,
			CONCAT(a.first_name, ' ', a.last_name) as approver_name
		FROM timesheets t
		JOIN employees e ON e.id = t.employee_id
		LEFT JOIN employees a ON a.id = t.approved_by
		WHERE t.employee_id = $1
		ORDER BY t.start_date DESC
		LIMIT 100
	`
	
	rows, err := r.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var timesheets []*models.Timesheet
	for rows.Next() {
		t := &models.Timesheet{}
		var approverName sql.NullString
		
		err := rows.Scan(
			&t.ID, &t.EmployeeID, &t.StartDate, &t.EndDate, &t.Status,
			&t.TotalHours, &t.RegularHours, &t.PTOHours, &t.HolidayHours,
			&t.SubmittedAt, &t.ApprovedAt, &t.ApprovedBy, &t.RejectionReason,
			&t.CreatedAt, &t.UpdatedAt,
			&t.EmployeeName, &approverName,
		)
		if err != nil {
			return nil, err
		}
		
		if approverName.Valid {
			t.ApproverName = approverName.String
		}
		
		timesheets = append(timesheets, t)
	}
	
	return timesheets, rows.Err()
}

func (r *timesheetRepository) GetTimesheetsByStatus(ctx context.Context, status string, managerID *uuid.UUID) ([]*models.Timesheet, error) {
	query := `
		SELECT 
			t.id, t.employee_id, t.start_date, t.end_date, t.status,
			t.total_hours, t.regular_hours, t.pto_hours, t.holiday_hours,
			t.submitted_at, t.approved_at, t.approved_by, t.rejection_reason,
			t.created_at, t.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name,
			CONCAT(a.first_name, ' ', a.last_name) as approver_name
		FROM timesheets t
		JOIN employees e ON e.id = t.employee_id
		LEFT JOIN employees a ON a.id = t.approved_by
		WHERE t.status = $1
	`
	
	args := []interface{}{status}
	
	// If manager ID provided, filter to their reports
	if managerID != nil {
		query += ` AND e.manager_id = $2`
		args = append(args, *managerID)
	}
	
	query += ` ORDER BY t.submitted_at DESC LIMIT 100`
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var timesheets []*models.Timesheet
	for rows.Next() {
		t := &models.Timesheet{}
		var approverName sql.NullString
		
		err := rows.Scan(
			&t.ID, &t.EmployeeID, &t.StartDate, &t.EndDate, &t.Status,
			&t.TotalHours, &t.RegularHours, &t.PTOHours, &t.HolidayHours,
			&t.SubmittedAt, &t.ApprovedAt, &t.ApprovedBy, &t.RejectionReason,
			&t.CreatedAt, &t.UpdatedAt,
			&t.EmployeeName, &approverName,
		)
		if err != nil {
			return nil, err
		}
		
		if approverName.Valid {
			t.ApproverName = approverName.String
		}
		
		timesheets = append(timesheets, t)
	}
	
	return timesheets, rows.Err()
}
