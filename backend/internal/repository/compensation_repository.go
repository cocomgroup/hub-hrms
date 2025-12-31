package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"hub-hrms/backend/internal/models"
)

// CompensationRepository interface
type CompensationRepository interface {
	// Compensation Plans
	CreatePlan(ctx context.Context, plan *models.CompensationPlan) error
	GetPlan(ctx context.Context, id uuid.UUID) (*models.CompensationPlan, error)
	GetPlansByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.CompensationPlan, error)
	GetActivePlan(ctx context.Context, employeeID uuid.UUID) (*models.CompensationPlan, error)
	GetAllPlans(ctx context.Context) ([]*models.CompensationPlan, error)
	UpdatePlan(ctx context.Context, plan *models.CompensationPlan) error
	DeletePlan(ctx context.Context, id uuid.UUID) error
	
	// Bonuses
	CreateBonus(ctx context.Context, bonus *models.Bonus) error
	GetBonus(ctx context.Context, id uuid.UUID) (*models.Bonus, error)
	GetBonusesByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.Bonus, error)
	GetAllBonuses(ctx context.Context) ([]*models.Bonus, error)
	GetBonusesByStatus(ctx context.Context, status string) ([]*models.Bonus, error)
	UpdateBonus(ctx context.Context, bonus *models.Bonus) error
	ApproveBonus(ctx context.Context, id uuid.UUID, approverID uuid.UUID) error
	MarkBonusPaid(ctx context.Context, id uuid.UUID) error
	DeleteBonus(ctx context.Context, id uuid.UUID) error
}

type compensationRepository struct {
	db *pgxpool.Pool
}

// NewCompensationRepository creates a new compensation repository
func NewCompensationRepository(db *pgxpool.Pool) CompensationRepository {
	return &compensationRepository{db: db}
}

// === COMPENSATION PLANS ===

// CreatePlan creates a new compensation plan
func (r *compensationRepository) CreatePlan(ctx context.Context, plan *models.CompensationPlan) error {
	query := `
		INSERT INTO compensation_plans (
			id, employee_id, compensation_type, base_amount, currency,
			pay_frequency, effective_date, end_date, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		RETURNING created_at, updated_at
	`

	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}
	if plan.Currency == "" {
		plan.Currency = "USD"
	}
	if plan.Status == "" {
		plan.Status = "active"
	}

	return r.db.QueryRow(
		ctx, query,
		plan.ID, plan.EmployeeID, plan.CompensationType, plan.BaseAmount,
		plan.Currency, plan.PayFrequency, plan.EffectiveDate, plan.EndDate, plan.Status,
	).Scan(&plan.CreatedAt, &plan.UpdatedAt)
}

// GetPlan retrieves a compensation plan by ID
func (r *compensationRepository) GetPlan(ctx context.Context, id uuid.UUID) (*models.CompensationPlan, error) {
	query := `
		SELECT 
			cp.id, cp.employee_id, cp.compensation_type, cp.base_amount, cp.currency,
			cp.pay_frequency, cp.effective_date, cp.end_date, cp.status,
			cp.created_at, cp.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM compensation_plans cp
		LEFT JOIN employees e ON cp.employee_id = e.id
		WHERE cp.id = $1
	`

	plan := &models.CompensationPlan{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&plan.ID, &plan.EmployeeID, &plan.CompensationType, &plan.BaseAmount, &plan.Currency,
		&plan.PayFrequency, &plan.EffectiveDate, &plan.EndDate, &plan.Status,
		&plan.CreatedAt, &plan.UpdatedAt, &plan.EmployeeName,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("compensation plan not found")
	}
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// GetPlansByEmployee retrieves all compensation plans for an employee
func (r *compensationRepository) GetPlansByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.CompensationPlan, error) {
	query := `
		SELECT 
			cp.id, cp.employee_id, cp.compensation_type, cp.base_amount, cp.currency,
			cp.pay_frequency, cp.effective_date, cp.end_date, cp.status,
			cp.created_at, cp.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM compensation_plans cp
		LEFT JOIN employees e ON cp.employee_id = e.id
		WHERE cp.employee_id = $1
		ORDER BY cp.effective_date DESC
	`

	rows, err := r.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*models.CompensationPlan
	for rows.Next() {
		plan := &models.CompensationPlan{}
		err := rows.Scan(
			&plan.ID, &plan.EmployeeID, &plan.CompensationType, &plan.BaseAmount, &plan.Currency,
			&plan.PayFrequency, &plan.EffectiveDate, &plan.EndDate, &plan.Status,
			&plan.CreatedAt, &plan.UpdatedAt, &plan.EmployeeName,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, rows.Err()
}

// GetActivePlan retrieves the active compensation plan for an employee
func (r *compensationRepository) GetActivePlan(ctx context.Context, employeeID uuid.UUID) (*models.CompensationPlan, error) {
	query := `
		SELECT 
			cp.id, cp.employee_id, cp.compensation_type, cp.base_amount, cp.currency,
			cp.pay_frequency, cp.effective_date, cp.end_date, cp.status,
			cp.created_at, cp.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM compensation_plans cp
		LEFT JOIN employees e ON cp.employee_id = e.id
		WHERE cp.employee_id = $1 
		  AND cp.status = 'active'
		  AND cp.effective_date <= CURRENT_DATE
		  AND (cp.end_date IS NULL OR cp.end_date >= CURRENT_DATE)
		ORDER BY cp.effective_date DESC
		LIMIT 1
	`

	plan := &models.CompensationPlan{}
	err := r.db.QueryRow(ctx, query, employeeID).Scan(
		&plan.ID, &plan.EmployeeID, &plan.CompensationType, &plan.BaseAmount, &plan.Currency,
		&plan.PayFrequency, &plan.EffectiveDate, &plan.EndDate, &plan.Status,
		&plan.CreatedAt, &plan.UpdatedAt, &plan.EmployeeName,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no active compensation plan found")
	}
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// GetAllPlans retrieves all compensation plans
func (r *compensationRepository) GetAllPlans(ctx context.Context) ([]*models.CompensationPlan, error) {
	query := `
		SELECT 
			cp.id, cp.employee_id, cp.compensation_type, cp.base_amount, cp.currency,
			cp.pay_frequency, cp.effective_date, cp.end_date, cp.status,
			cp.created_at, cp.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM compensation_plans cp
		LEFT JOIN employees e ON cp.employee_id = e.id
		ORDER BY cp.created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*models.CompensationPlan
	for rows.Next() {
		plan := &models.CompensationPlan{}
		err := rows.Scan(
			&plan.ID, &plan.EmployeeID, &plan.CompensationType, &plan.BaseAmount, &plan.Currency,
			&plan.PayFrequency, &plan.EffectiveDate, &plan.EndDate, &plan.Status,
			&plan.CreatedAt, &plan.UpdatedAt, &plan.EmployeeName,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, rows.Err()
}

// UpdatePlan updates a compensation plan
func (r *compensationRepository) UpdatePlan(ctx context.Context, plan *models.CompensationPlan) error {
	query := `
		UPDATE compensation_plans
		SET compensation_type = $1,
		    base_amount = $2,
		    currency = $3,
		    pay_frequency = $4,
		    effective_date = $5,
		    end_date = $6,
		    status = $7,
		    updated_at = NOW()
		WHERE id = $8
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		ctx, query,
		plan.CompensationType, plan.BaseAmount, plan.Currency,
		plan.PayFrequency, plan.EffectiveDate, plan.EndDate,
		plan.Status, plan.ID,
	).Scan(&plan.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("compensation plan not found")
	}
	return err
}

// DeletePlan deletes a compensation plan
func (r *compensationRepository) DeletePlan(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM compensation_plans WHERE id = $1"
	
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("compensation plan not found")
	}

	return nil
}

// === BONUSES ===

// CreateBonus creates a new bonus
func (r *compensationRepository) CreateBonus(ctx context.Context, bonus *models.Bonus) error {
	query := `
		INSERT INTO bonuses (
			id, employee_id, bonus_type, amount, currency,
			description, payment_date, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		RETURNING created_at, updated_at
	`

	if bonus.ID == uuid.Nil {
		bonus.ID = uuid.New()
	}
	if bonus.Currency == "" {
		bonus.Currency = "USD"
	}
	if bonus.Status == "" {
		bonus.Status = "pending"
	}

	return r.db.QueryRow(
		ctx, query,
		bonus.ID, bonus.EmployeeID, bonus.BonusType, bonus.Amount,
		bonus.Currency, bonus.Description, bonus.PaymentDate, bonus.Status,
	).Scan(&bonus.CreatedAt, &bonus.UpdatedAt)
}

// GetBonus retrieves a bonus by ID
func (r *compensationRepository) GetBonus(ctx context.Context, id uuid.UUID) (*models.Bonus, error) {
	query := `
		SELECT 
			b.id, b.employee_id, b.bonus_type, b.amount, b.currency,
			b.description, b.payment_date, b.status,
			b.approved_by, b.approved_at, b.paid_at,
			b.created_at, b.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM bonuses b
		LEFT JOIN employees e ON b.employee_id = e.id
		WHERE b.id = $1
	`

	bonus := &models.Bonus{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&bonus.ID, &bonus.EmployeeID, &bonus.BonusType, &bonus.Amount, &bonus.Currency,
		&bonus.Description, &bonus.PaymentDate, &bonus.Status,
		&bonus.ApprovedBy, &bonus.ApprovedAt, &bonus.PaidAt,
		&bonus.CreatedAt, &bonus.UpdatedAt, &bonus.EmployeeName,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bonus not found")
	}
	if err != nil {
		return nil, err
	}

	return bonus, nil
}

// GetBonusesByEmployee retrieves all bonuses for an employee
func (r *compensationRepository) GetBonusesByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.Bonus, error) {
	query := `
		SELECT 
			b.id, b.employee_id, b.bonus_type, b.amount, b.currency,
			b.description, b.payment_date, b.status,
			b.approved_by, b.approved_at, b.paid_at,
			b.created_at, b.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM bonuses b
		LEFT JOIN employees e ON b.employee_id = e.id
		WHERE b.employee_id = $1
		ORDER BY b.payment_date DESC
	`

	rows, err := r.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bonuses []*models.Bonus
	for rows.Next() {
		bonus := &models.Bonus{}
		err := rows.Scan(
			&bonus.ID, &bonus.EmployeeID, &bonus.BonusType, &bonus.Amount, &bonus.Currency,
			&bonus.Description, &bonus.PaymentDate, &bonus.Status,
			&bonus.ApprovedBy, &bonus.ApprovedAt, &bonus.PaidAt,
			&bonus.CreatedAt, &bonus.UpdatedAt, &bonus.EmployeeName,
		)
		if err != nil {
			return nil, err
		}
		bonuses = append(bonuses, bonus)
	}

	return bonuses, rows.Err()
}

// GetAllBonuses retrieves all bonuses
func (r *compensationRepository) GetAllBonuses(ctx context.Context) ([]*models.Bonus, error) {
	query := `
		SELECT 
			b.id, b.employee_id, b.bonus_type, b.amount, b.currency,
			b.description, b.payment_date, b.status,
			b.approved_by, b.approved_at, b.paid_at,
			b.created_at, b.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM bonuses b
		LEFT JOIN employees e ON b.employee_id = e.id
		ORDER BY b.created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bonuses []*models.Bonus
	for rows.Next() {
		bonus := &models.Bonus{}
		err := rows.Scan(
			&bonus.ID, &bonus.EmployeeID, &bonus.BonusType, &bonus.Amount, &bonus.Currency,
			&bonus.Description, &bonus.PaymentDate, &bonus.Status,
			&bonus.ApprovedBy, &bonus.ApprovedAt, &bonus.PaidAt,
			&bonus.CreatedAt, &bonus.UpdatedAt, &bonus.EmployeeName,
		)
		if err != nil {
			return nil, err
		}
		bonuses = append(bonuses, bonus)
	}

	return bonuses, rows.Err()
}

// GetBonusesByStatus retrieves bonuses by status
func (r *compensationRepository) GetBonusesByStatus(ctx context.Context, status string) ([]*models.Bonus, error) {
	query := `
		SELECT 
			b.id, b.employee_id, b.bonus_type, b.amount, b.currency,
			b.description, b.payment_date, b.status,
			b.approved_by, b.approved_at, b.paid_at,
			b.created_at, b.updated_at,
			CONCAT(e.first_name, ' ', e.last_name) as employee_name
		FROM bonuses b
		LEFT JOIN employees e ON b.employee_id = e.id
		WHERE b.status = $1
		ORDER BY b.payment_date DESC
	`

	rows, err := r.db.Query(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bonuses []*models.Bonus
	for rows.Next() {
		bonus := &models.Bonus{}
		err := rows.Scan(
			&bonus.ID, &bonus.EmployeeID, &bonus.BonusType, &bonus.Amount, &bonus.Currency,
			&bonus.Description, &bonus.PaymentDate, &bonus.Status,
			&bonus.ApprovedBy, &bonus.ApprovedAt, &bonus.PaidAt,
			&bonus.CreatedAt, &bonus.UpdatedAt, &bonus.EmployeeName,
		)
		if err != nil {
			return nil, err
		}
		bonuses = append(bonuses, bonus)
	}

	return bonuses, rows.Err()
}

// UpdateBonus updates a bonus
func (r *compensationRepository) UpdateBonus(ctx context.Context, bonus *models.Bonus) error {
	query := `
		UPDATE bonuses
		SET bonus_type = $1,
		    amount = $2,
		    currency = $3,
		    description = $4,
		    payment_date = $5,
		    status = $6,
		    updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		ctx, query,
		bonus.BonusType, bonus.Amount, bonus.Currency,
		bonus.Description, bonus.PaymentDate, bonus.Status,
		bonus.ID,
	).Scan(&bonus.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("bonus not found")
	}
	return err
}

// ApproveBonus approves a bonus
func (r *compensationRepository) ApproveBonus(ctx context.Context, id uuid.UUID, approverID uuid.UUID) error {
	query := `
		UPDATE bonuses 
		SET status = 'approved', 
		    approved_by = $1, 
		    approved_at = NOW(), 
		    updated_at = NOW()
		WHERE id = $2 AND status = 'pending'
	`

	result, err := r.db.Exec(ctx, query, approverID, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("bonus not found or not in pending status")
	}

	return nil
}

// MarkBonusPaid marks a bonus as paid
func (r *compensationRepository) MarkBonusPaid(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE bonuses 
		SET status = 'paid', 
		    paid_at = NOW(), 
		    updated_at = NOW()
		WHERE id = $1 AND status = 'approved'
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("bonus not found or not in approved status")
	}

	return nil
}

// DeleteBonus deletes a bonus
func (r *compensationRepository) DeleteBonus(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM bonuses WHERE id = $1"
	
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("bonus not found")
	}

	return nil
}