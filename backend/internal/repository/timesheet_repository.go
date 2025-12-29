package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"hub-hrms/backend/internal/models"
)

// TimesheetRepository interface
type TimesheetRepository interface {
	Create(ctx context.Context, timesheet *models.Timesheet) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Timesheet, error)
	GetByEmployee(ctx context.Context, employeeID uuid.UUID, filters map[string]interface{}) ([]*models.Timesheet, error)
	GetActiveTimesheet(ctx context.Context, employeeID uuid.UUID) (*models.Timesheet, error)
	Update(ctx context.Context, timesheet *models.Timesheet) error
	List(ctx context.Context, filters map[string]interface{}) ([]*models.Timesheet, error)
}


type timesheetRepository struct{ db *pgxpool.Pool }

func NewTimesheetRepository(db *pgxpool.Pool) TimesheetRepository { return &timesheetRepository{db: db} }

func (r *timesheetRepository) Create(ctx context.Context, timesheet *models.Timesheet) error {
	query := `
		INSERT INTO time_entries (
			id, employee_id, entry_date, clock_in, clock_out, break_duration,
			notes, entry_type, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		RETURNING created_at, updated_at, total_hours
	`
	
	if timesheet.ID == uuid.Nil {
		timesheet.ID = uuid.New()
	}
	
	log.Printf("TimesheetRepo params employee_id %s clockin %s ", timesheet.EmployeeID, timesheet.ClockIn )
	return r.db.QueryRow(
		ctx, query,
		timesheet.ID, timesheet.EmployeeID, timesheet.Date, timesheet.ClockIn, timesheet.ClockOut,
		timesheet.BreakMinutes, timesheet.Notes, timesheet.Type, timesheet.Status,
	).Scan(&timesheet.CreatedAt, &timesheet.UpdatedAt, &timesheet.TotalHours)
}

func (r *timesheetRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Timesheet, error) {
	query := `
		SELECT 
			te.id, te.employee_id, te.entry_date, te.clock_in, te.clock_out,
			te.break_duration, te.notes, te.entry_type, te.status,
			te.total_hours, te.created_at, te.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM time_entries te
		JOIN employees e ON e.id = te.employee_id
		WHERE te.id = $1
	`
	
	timesheet := &models.Timesheet{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&timesheet.ID, &timesheet.EmployeeID, &timesheet.Date, &timesheet.ClockIn, &timesheet.ClockOut,
		&timesheet.BreakMinutes, &timesheet.Notes, &timesheet.Type, &timesheet.Status,
		&timesheet.TotalHours, &timesheet.CreatedAt, &timesheet.UpdatedAt, &timesheet.EmployeeName,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("timesheet not found")
	}
	if err != nil {
		return nil, err
	}
	
	return timesheet, nil
}

func (r *timesheetRepository) GetByEmployee(ctx context.Context, employeeID uuid.UUID, filters map[string]interface{}) ([]*models.Timesheet, error) {
	query := `
		SELECT 
			te.id, te.employee_id, te.entry_date, te.clock_in, te.clock_out,
			te.break_duration, te.notes, te.entry_type, te.status,
			te.total_hours, te.created_at, te.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM time_entries te
		JOIN employees e ON e.id = te.employee_id
		WHERE te.employee_id = $1
	`
	
	args := []interface{}{employeeID}
	argCount := 1
	
	// Add filters if provided
	if startDate, ok := filters["start_date"].(time.Time); ok {
		argCount++
		query += fmt.Sprintf(" AND te.entry_date >= $%d::date", argCount)
		args = append(args, startDate.Format("2006-01-02"))
	}
	
	if endDate, ok := filters["end_date"].(time.Time); ok {
		argCount++
		query += fmt.Sprintf(" AND te.entry_date <= $%d::date", argCount)
		args = append(args, endDate.Format("2006-01-02"))
	}
	
	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		query += fmt.Sprintf(" AND te.status = $%d", argCount)
		args = append(args, status)
	}
	
	query += " ORDER BY te.entry_date DESC, te.clock_in DESC"
	log.Printf("Timesheet GetByEmployee args %s ", args)
	log.Printf("Timesheet query %s ", query)
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var timesheets []*models.Timesheet
	for rows.Next() {
		ts := &models.Timesheet{}
		err := rows.Scan(
			&ts.ID, &ts.EmployeeID, &ts.Date, &ts.ClockIn, &ts.ClockOut,
			&ts.BreakMinutes, &ts.Notes, &ts.Type, &ts.Status,
			&ts.TotalHours, &ts.CreatedAt, &ts.UpdatedAt, &ts.EmployeeName,
		)
		if err != nil {
			return nil, err
		}
		timesheets = append(timesheets, ts)
	}
	
	return timesheets, rows.Err()
}

func (r *timesheetRepository) GetActiveTimesheet(ctx context.Context, employeeID uuid.UUID) (*models.Timesheet, error) {
	query := `
		SELECT 
			te.id, te.employee_id, te.entry_date, te.clock_in, te.clock_out,
			te.break_duration, te.notes, te.entry_type, te.status,
			te.total_hours, te.created_at, te.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM time_entries te
		JOIN employees e ON e.id = te.employee_id
		WHERE te.employee_id = $1
		  AND te.clock_in IS NOT NULL
		  AND te.clock_out IS NULL
		  AND te.entry_date = CURRENT_DATE
		ORDER BY te.clock_in DESC
		LIMIT 1
	`
	
	timesheet := &models.Timesheet{}
	err := r.db.QueryRow(ctx, query, employeeID).Scan(
		&timesheet.ID, &timesheet.EmployeeID, &timesheet.Date, &timesheet.ClockIn, &timesheet.ClockOut,
		&timesheet.BreakMinutes, &timesheet.Notes, &timesheet.Type, &timesheet.Status,
		&timesheet.TotalHours, &timesheet.CreatedAt, &timesheet.UpdatedAt, &timesheet.EmployeeName,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil // No active timesheet is not an error
	}
	if err != nil {
		return nil, err
	}
	
	return timesheet, nil
}

func (r *timesheetRepository) Update(ctx context.Context, timesheet *models.Timesheet) error {
	query := `
		UPDATE time_entries
		SET clock_in = $1,
		    clock_out = $2,
		    break_duration = $3,
		    notes = $4,
		    entry_type = $5,
		    status = $6,
		    updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at, total_hours
	`
	
	return r.db.QueryRow(
		ctx, query,
		timesheet.ClockIn, timesheet.ClockOut, timesheet.BreakMinutes,
		timesheet.Notes, timesheet.Type, timesheet.Status, timesheet.ID,
	).Scan(&timesheet.UpdatedAt, &timesheet.TotalHours)
}

func (r *timesheetRepository) List(ctx context.Context, filters map[string]interface{}) ([]*models.Timesheet, error) {
	query := `
		SELECT 
			te.id, te.employee_id, te.entry_date, te.clock_in, te.clock_out,
			te.break_duration, te.notes, te.entry_type, te.status,
			te.total_hours, te.created_at, te.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM time_entries te
		JOIN employees e ON e.id = te.employee_id
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 0
	
	// Add filters
	if startDate, ok := filters["start_date"].(time.Time); ok {
		argCount++
		query += fmt.Sprintf(" AND te.entry_date >= $%d::date", argCount)
		args = append(args, startDate.Format("2006-01-02"))
	}
	
	if endDate, ok := filters["end_date"].(time.Time); ok {
		argCount++
		query += fmt.Sprintf(" AND te.entry_date <= $%d::date", argCount)
		args = append(args, endDate.Format("2006-01-02"))
	}
	
	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		query += fmt.Sprintf(" AND te.status = $%d", argCount)
		args = append(args, status)
	}
	
	if department, ok := filters["department"].(string); ok && department != "" {
		argCount++
		query += fmt.Sprintf(" AND e.department = $%d", argCount)
		args = append(args, department)
	}
	
	query += " ORDER BY te.entry_date DESC, te.clock_in DESC LIMIT 1000"
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var timesheets []*models.Timesheet
	for rows.Next() {
		ts := &models.Timesheet{}
		err := rows.Scan(
			&ts.ID, &ts.EmployeeID, &ts.Date, &ts.ClockIn, &ts.ClockOut,
			&ts.BreakMinutes, &ts.Notes, &ts.Type, &ts.Status,
			&ts.TotalHours, &ts.CreatedAt, &ts.UpdatedAt, &ts.EmployeeName,
		)
		if err != nil {
			return nil, err
		}
		timesheets = append(timesheets, ts)
	}
	
	return timesheets, rows.Err()
}

// ============================================================================
// Additional Helper Methods (not in interface but useful)
// ============================================================================

// Delete removes a timesheet entry (only if status is draft)
func (r *timesheetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM time_entries WHERE id = $1 AND status = 'draft'`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("timesheet not found or cannot be deleted")
	}
	
	return nil
}

// ============================================================================
// Project Tracking Methods
// ============================================================================

// GetProjects gets all active projects
func (r *timesheetRepository) GetProjects(ctx context.Context) ([]*models.Project, error) {
	query := `
		SELECT id, name, code, description, client_name, status,
		       start_date, end_date, budget_hours,
		       created_at, updated_at
		FROM projects
		WHERE status = 'active'
		ORDER BY name ASC
	`
	
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var projects []*models.Project
	for rows.Next() {
		p := &models.Project{}
		err := rows.Scan(
			&p.ID, &p.Name, &p.Code, &p.Description, &p.ClientName, &p.Status,
			&p.StartDate, &p.EndDate, &p.BudgetHours,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	
	return projects, rows.Err()
}

// CreateProject creates a new project
func (r *timesheetRepository) CreateProject(ctx context.Context, project *models.Project) error {
	query := `
		INSERT INTO projects (
			id, name, code, description, client_name, status,
			start_date, end_date, budget_hours, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`
	
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}
	
	return r.db.QueryRow(
		ctx, query,
		project.ID, project.Name, project.Code, project.Description, project.ClientName,
		project.Status, project.StartDate, project.EndDate, project.BudgetHours, project.CreatedBy,
	).Scan(&project.CreatedAt, &project.UpdatedAt)
}

// ============================================================================
// Reporting Methods
// ============================================================================

// GetEmployeeHoursSummary gets hours summary for an employee in a date range
func (r *timesheetRepository) GetEmployeeHoursSummary(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) (*models.TimesheetSummary, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN entry_type = 'regular' THEN total_hours ELSE 0 END), 0) as regular_hours,
			COALESCE(SUM(CASE WHEN entry_type = 'overtime' THEN total_hours ELSE 0 END), 0) as overtime_hours,
			COALESCE(SUM(CASE WHEN entry_type IN ('pto', 'sick', 'holiday') THEN total_hours ELSE 0 END), 0) as pto_hours,
			COALESCE(SUM(total_hours), 0) as total_hours,
			COUNT(*) as entry_count
		FROM time_entries
		WHERE employee_id = $1
		  AND entry_date BETWEEN $2 AND $3
		  AND status != 'rejected'
	`
	
	summary := &models.TimesheetSummary{
		EmployeeID: employeeID,
		StartDate:  startDate,
		EndDate:    endDate,
	}
	
	err := r.db.QueryRow(ctx, query, employeeID, startDate, endDate).Scan(
		&summary.RegularHours, &summary.OvertimeHours, &summary.PTOHours,
		&summary.TotalHours, &summary.EntryCount,
	)
	
	return summary, err
}

// GetPendingApprovals gets all timesheets pending approval
func (r *timesheetRepository) GetPendingApprovals(ctx context.Context) ([]*models.Timesheet, error) {
	query := `
		SELECT 
			te.id, te.employee_id, te.entry_date, te.clock_in, te.clock_out,
			te.break_duration, te.notes, te.entry_type, te.status,
			te.total_hours, te.created_at, te.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM time_entries te
		JOIN employees e ON e.id = te.employee_id
		WHERE te.status = 'submitted'
		ORDER BY te.entry_date DESC
	`
	
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var timesheets []*models.Timesheet
	for rows.Next() {
		ts := &models.Timesheet{}
		err := rows.Scan(
			&ts.ID, &ts.EmployeeID, &ts.Date, &ts.ClockIn, &ts.ClockOut,
			&ts.BreakMinutes, &ts.Notes, &ts.Type, &ts.Status,
			&ts.TotalHours, &ts.CreatedAt, &ts.UpdatedAt, &ts.EmployeeName,
		)
		if err != nil {
			return nil, err
		}
		timesheets = append(timesheets, ts)
	}
	
	return timesheets, rows.Err()
}