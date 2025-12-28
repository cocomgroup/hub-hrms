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

// PTORepository interface
type PTORepository interface {
	GetBalance(ctx context.Context, employeeID uuid.UUID) (*models.PTOBalance, error)
	CreateBalance(ctx context.Context, balance *models.PTOBalance) error
	UpdateBalance(ctx context.Context, balance *models.PTOBalance) error
	CreateRequest(ctx context.Context, request *models.PTORequest) error
	GetRequestByID(ctx context.Context, id uuid.UUID) (*models.PTORequest, error)
	GetRequestsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.PTORequest, error)
	UpdateRequest(ctx context.Context, request *models.PTORequest) error
	ListRequests(ctx context.Context, filters map[string]interface{}) ([]*models.PTORequest, error)
}

type ptoRepository struct {
	db *pgxpool.Pool
}

func NewPTORepository(db *pgxpool.Pool) PTORepository {
	return &ptoRepository{db: db}
}

// ============================================================================
// PTO Balance Operations
// ============================================================================

func (r *ptoRepository) GetBalance(ctx context.Context, employeeID uuid.UUID) (*models.PTOBalance, error) {
	query := `
		SELECT id, employee_id, vacation_days, sick_days, personal_days, 
		       year, created_at, updated_at
		FROM pto_balances
		WHERE employee_id = $1 AND year = EXTRACT(YEAR FROM CURRENT_DATE)
	`

	balance := &models.PTOBalance{}
	err := r.db.QueryRow(ctx, query, employeeID).Scan(
		&balance.ID,
		&balance.EmployeeID,
		&balance.VacationDays,
		&balance.SickDays,
		&balance.PersonalDays,
		&balance.Year,
		&balance.CreatedAt,
		&balance.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create default balance if it doesn't exist
		return r.createDefaultBalance(ctx, employeeID)
	}

	return balance, err
}

func (r *ptoRepository) createDefaultBalance(ctx context.Context, employeeID uuid.UUID) (*models.PTOBalance, error) {
	balance := &models.PTOBalance{
		ID:           uuid.New(),
		EmployeeID:   employeeID,
		VacationDays: 15.0, // Default 15 vacation days
		SickDays:     10.0, // Default 10 sick days
		PersonalDays: 5.0,  // Default 5 personal days
		Year:         getCurrentYear(),
	}

	if err := r.CreateBalance(ctx, balance); err != nil {
		return nil, err
	}

	return balance, nil
}

func (r *ptoRepository) CreateBalance(ctx context.Context, balance *models.PTOBalance) error {
	query := `
		INSERT INTO pto_balances (
			id, employee_id, vacation_days, sick_days, personal_days, year
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`

	if balance.ID == uuid.Nil {
		balance.ID = uuid.New()
	}

	if balance.Year == 0 {
		balance.Year = getCurrentYear()
	}

	return r.db.QueryRow(
		ctx, query,
		balance.ID,
		balance.EmployeeID,
		balance.VacationDays,
		balance.SickDays,
		balance.PersonalDays,
		balance.Year,
	).Scan(&balance.CreatedAt, &balance.UpdatedAt)
}

func (r *ptoRepository) UpdateBalance(ctx context.Context, balance *models.PTOBalance) error {
	query := `
		UPDATE pto_balances
		SET vacation_days = $1,
		    sick_days = $2,
		    personal_days = $3,
		    updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	return r.db.QueryRow(
		ctx, query,
		balance.VacationDays,
		balance.SickDays,
		balance.PersonalDays,
		balance.ID,
	).Scan(&balance.UpdatedAt)
}

// ============================================================================
// PTO Request Operations
// ============================================================================

func (r *ptoRepository) CreateRequest(ctx context.Context, request *models.PTORequest) error {
	query := `
		INSERT INTO pto_requests (
			id, employee_id, pto_type, start_date, end_date,
			days_requested, reason, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	if request.ID == uuid.Nil {
		request.ID = uuid.New()
	}

	return r.db.QueryRow(
		ctx, query,
		request.ID,
		request.EmployeeID,
		request.PTOType,
		request.StartDate,
		request.EndDate,
		request.DaysRequested,
		request.Reason,
		request.Status,
	).Scan(&request.CreatedAt, &request.UpdatedAt)
}

func (r *ptoRepository) GetRequestByID(ctx context.Context, id uuid.UUID) (*models.PTORequest, error) {
	query := `
		SELECT pr.id, pr.employee_id, pr.pto_type, pr.start_date, pr.end_date,
		       pr.days_requested, pr.reason, pr.status, pr.reviewed_by,
		       pr.reviewed_at, pr.review_notes, pr.created_at, pr.updated_at,
		       CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM pto_requests pr
		JOIN employees e ON e.id = pr.employee_id
		WHERE pr.id = $1
	`

	request := &models.PTORequest{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&request.ID,
		&request.EmployeeID,
		&request.PTOType,
		&request.StartDate,
		&request.EndDate,
		&request.DaysRequested,
		&request.Reason,
		&request.Status,
		&request.ReviewedBy,
		&request.ReviewedAt,
		&request.ReviewNotes,
		&request.CreatedAt,
		&request.UpdatedAt,
		&request.EmployeeName,
	)

	return request, err
}

func (r *ptoRepository) GetRequestsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.PTORequest, error) {
	query := `
		SELECT pr.id, pr.employee_id, pr.pto_type, pr.start_date, pr.end_date,
		       pr.days_requested, pr.reason, pr.status, pr.reviewed_by,
		       pr.reviewed_at, pr.review_notes, pr.created_at, pr.updated_at,
		       CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM pto_requests pr
		JOIN employees e ON e.id = pr.employee_id
		WHERE pr.employee_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*models.PTORequest
	for rows.Next() {
		req := &models.PTORequest{}
		err := rows.Scan(
			&req.ID,
			&req.EmployeeID,
			&req.PTOType,
			&req.StartDate,
			&req.EndDate,
			&req.DaysRequested,
			&req.Reason,
			&req.Status,
			&req.ReviewedBy,
			&req.ReviewedAt,
			&req.ReviewNotes,
			&req.CreatedAt,
			&req.UpdatedAt,
			&req.EmployeeName,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, rows.Err()
}

func (r *ptoRepository) UpdateRequest(ctx context.Context, request *models.PTORequest) error {
	query := `
		UPDATE pto_requests
		SET status = $1,
		    reviewed_by = $2,
		    reviewed_at = $3,
		    review_notes = $4,
		    updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	return r.db.QueryRow(
		ctx, query,
		request.Status,
		request.ReviewedBy,
		request.ReviewedAt,
		request.ReviewNotes,
		request.ID,
	).Scan(&request.UpdatedAt)
}

func (r *ptoRepository) ListRequests(ctx context.Context, filters map[string]interface{}) ([]*models.PTORequest, error) {
	query := `
		SELECT pr.id, pr.employee_id, pr.pto_type, pr.start_date, pr.end_date,
		       pr.days_requested, pr.reason, pr.status, pr.reviewed_by,
		       pr.reviewed_at, pr.review_notes, pr.created_at, pr.updated_at,
		       CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM pto_requests pr
		JOIN employees e ON e.id = pr.employee_id
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		query += fmt.Sprintf(" AND pr.status = $%d", argCount)
		args = append(args, status)
	}

	if department, ok := filters["department"].(string); ok && department != "" {
		argCount++
		query += fmt.Sprintf(" AND e.department = $%d", argCount)
		args = append(args, department)
	}

	if ptoType, ok := filters["pto_type"].(string); ok && ptoType != "" {
		argCount++
		query += fmt.Sprintf(" AND pr.pto_type = $%d", argCount)
		args = append(args, ptoType)
	}

	query += " ORDER BY pr.created_at DESC LIMIT 100"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*models.PTORequest
	for rows.Next() {
		req := &models.PTORequest{}
		err := rows.Scan(
			&req.ID,
			&req.EmployeeID,
			&req.PTOType,
			&req.StartDate,
			&req.EndDate,
			&req.DaysRequested,
			&req.Reason,
			&req.Status,
			&req.ReviewedBy,
			&req.ReviewedAt,
			&req.ReviewNotes,
			&req.CreatedAt,
			&req.UpdatedAt,
			&req.EmployeeName,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, rows.Err()
}

// Helper function
func getCurrentYear() int {
	return time.Now().Year()
}
