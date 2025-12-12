package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"hub-hrms/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BenefitsRepository interface defines repository operations
type BenefitsRepository interface {
	GetAllBenefitPlans(ctx context.Context, activeOnly bool) ([]models.BenefitPlan, error)
	GetBenefitPlanByID(ctx context.Context, id uuid.UUID) (*models.BenefitPlan, error)
	CreateBenefitPlan(ctx context.Context, plan *models.BenefitPlan) error
	UpdateBenefitPlan(ctx context.Context, id uuid.UUID, plan *models.BenefitPlan) error
	GetEmployeeEnrollments(ctx context.Context, employeeID uuid.UUID) ([]models.BenefitEnrollment, error)
	CreateEnrollment(ctx context.Context, enrollment *models.BenefitEnrollment) error
	GetEnrollmentByID(ctx context.Context, id uuid.UUID) (*models.BenefitEnrollment, error)
	GetEnrollmentDependents(ctx context.Context, enrollmentID uuid.UUID) ([]models.Dependent, error)
	CancelEnrollment(ctx context.Context, id uuid.UUID, terminationDate time.Time) error
	GetAllEnrollments(ctx context.Context, status *models.EnrollmentStatus) ([]models.BenefitEnrollment, error)
}

// benefitsRepository implements BenefitsRepository interface
type benefitsRepository struct {
	db *pgxpool.Pool
}

// NewBenefitsRepository creates a new benefits repository
func NewBenefitsRepository(db *pgxpool.Pool) BenefitsRepository {
	return &benefitsRepository{db: db}
}

// GetAllBenefitPlans retrieves all benefit plans
func (r *benefitsRepository) GetAllBenefitPlans(ctx context.Context, activeOnly bool) ([]models.BenefitPlan, error) {
	query := `
		SELECT id, name, category, plan_type, provider, description,
		       employee_cost, employer_cost, deductible_single, deductible_family,
		       out_of_pocket_max_single, out_of_pocket_max_family,
		       copay_primary_care, copay_specialist, copay_emergency, coinsurance_rate,
		       active, enrollment_start_date, enrollment_end_date, effective_date,
		       termination_date, created_at, updated_at
		FROM benefit_plans
	`

	if activeOnly {
		query += ` WHERE active = true`
	}

	query += ` ORDER BY category, name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []models.BenefitPlan
	for rows.Next() {
		var plan models.BenefitPlan
		err := rows.Scan(
			&plan.ID, &plan.Name, &plan.Category, &plan.PlanType, &plan.Provider,
			&plan.Description, &plan.EmployeeCost, &plan.EmployerCost,
			&plan.DeductibleSingle, &plan.DeductibleFamily,
			&plan.OutOfPocketMaxSingle, &plan.OutOfPocketMaxFamily,
			&plan.CopayPrimaryCare, &plan.CopaySpecialist, &plan.CopayEmergency,
			&plan.CoinsuranceRate, &plan.Active, &plan.EnrollmentStartDate,
			&plan.EnrollmentEndDate, &plan.EffectiveDate, &plan.TerminationDate,
			&plan.CreatedAt, &plan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, rows.Err()
}

// GetBenefitPlanByID retrieves a specific benefit plan
func (r *benefitsRepository) GetBenefitPlanByID(ctx context.Context, id uuid.UUID) (*models.BenefitPlan, error) {
	query := `
		SELECT id, name, category, plan_type, provider, description,
		       employee_cost, employer_cost, deductible_single, deductible_family,
		       out_of_pocket_max_single, out_of_pocket_max_family,
		       copay_primary_care, copay_specialist, copay_emergency, coinsurance_rate,
		       active, enrollment_start_date, enrollment_end_date, effective_date,
		       termination_date, created_at, updated_at
		FROM benefit_plans
		WHERE id = $1
	`

	var plan models.BenefitPlan
	err := r.db.QueryRow(ctx, query, id).Scan(
		&plan.ID, &plan.Name, &plan.Category, &plan.PlanType, &plan.Provider,
		&plan.Description, &plan.EmployeeCost, &plan.EmployerCost,
		&plan.DeductibleSingle, &plan.DeductibleFamily,
		&plan.OutOfPocketMaxSingle, &plan.OutOfPocketMaxFamily,
		&plan.CopayPrimaryCare, &plan.CopaySpecialist, &plan.CopayEmergency,
		&plan.CoinsuranceRate, &plan.Active, &plan.EnrollmentStartDate,
		&plan.EnrollmentEndDate, &plan.EffectiveDate, &plan.TerminationDate,
		&plan.CreatedAt, &plan.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("benefit plan not found")
	}
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

// CreateBenefitPlan creates a new benefit plan
func (r *benefitsRepository) CreateBenefitPlan(ctx context.Context, plan *models.BenefitPlan) error {
	query := `
		INSERT INTO benefit_plans (
			id, name, category, plan_type, provider, description,
			employee_cost, employer_cost, deductible_single, deductible_family,
			out_of_pocket_max_single, out_of_pocket_max_family,
			copay_primary_care, copay_specialist, copay_emergency, coinsurance_rate,
			active, enrollment_start_date, enrollment_end_date, effective_date,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22
		)
	`

	now := time.Now()
	plan.ID = uuid.New()
	plan.CreatedAt = now
	plan.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		plan.ID, plan.Name, plan.Category, plan.PlanType, plan.Provider,
		plan.Description, plan.EmployeeCost, plan.EmployerCost,
		plan.DeductibleSingle, plan.DeductibleFamily,
		plan.OutOfPocketMaxSingle, plan.OutOfPocketMaxFamily,
		plan.CopayPrimaryCare, plan.CopaySpecialist, plan.CopayEmergency,
		plan.CoinsuranceRate, plan.Active, plan.EnrollmentStartDate,
		plan.EnrollmentEndDate, plan.EffectiveDate, plan.CreatedAt, plan.UpdatedAt,
	)

	return err
}

// UpdateBenefitPlan updates a benefit plan
func (r *benefitsRepository) UpdateBenefitPlan(ctx context.Context, id uuid.UUID, plan *models.BenefitPlan) error {
	query := `
		UPDATE benefit_plans
		SET name = $1, category = $2, plan_type = $3, provider = $4,
		    description = $5, employee_cost = $6, employer_cost = $7,
		    deductible_single = $8, deductible_family = $9,
		    out_of_pocket_max_single = $10, out_of_pocket_max_family = $11,
		    copay_primary_care = $12, copay_specialist = $13, copay_emergency = $14,
		    coinsurance_rate = $15, active = $16, updated_at = $17
		WHERE id = $18
	`

	result, err := r.db.Exec(ctx, query,
		plan.Name, plan.Category, plan.PlanType, plan.Provider, plan.Description,
		plan.EmployeeCost, plan.EmployerCost, plan.DeductibleSingle, plan.DeductibleFamily,
		plan.OutOfPocketMaxSingle, plan.OutOfPocketMaxFamily,
		plan.CopayPrimaryCare, plan.CopaySpecialist, plan.CopayEmergency,
		plan.CoinsuranceRate, plan.Active, time.Now(), id,
	)

	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("benefit plan not found")
	}

	return nil
}

// GetEmployeeEnrollments retrieves all enrollments for an employee
func (r *benefitsRepository) GetEmployeeEnrollments(ctx context.Context, employeeID uuid.UUID) ([]models.BenefitEnrollment, error) {
	query := `
		SELECT e.id, e.employee_id, e.plan_id, e.coverage_level, e.status,
		       e.enrollment_date, e.effective_date, e.termination_date,
		       e.employee_cost, e.employer_cost, e.total_cost, e.payroll_deduction,
		       e.created_at, e.updated_at,
		       p.name as plan_name, p.category as plan_category,
		       emp.first_name || ' ' || emp.last_name as employee_name
		FROM benefit_enrollments e
		INNER JOIN benefit_plans p ON e.plan_id = p.id
		INNER JOIN employees emp ON e.employee_id = emp.id
		WHERE e.employee_id = $1
		ORDER BY e.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []models.BenefitEnrollment
	for rows.Next() {
		var enrollment models.BenefitEnrollment
		err := rows.Scan(
			&enrollment.ID, &enrollment.EmployeeID, &enrollment.PlanID,
			&enrollment.CoverageLevel, &enrollment.Status,
			&enrollment.EnrollmentDate, &enrollment.EffectiveDate, &enrollment.TerminationDate,
			&enrollment.EmployeeCost, &enrollment.EmployerCost, &enrollment.TotalCost,
			&enrollment.PayrollDeduction, &enrollment.CreatedAt, &enrollment.UpdatedAt,
			&enrollment.PlanName, &enrollment.PlanCategory, &enrollment.EmployeeName,
		)
		if err != nil {
			return nil, err
		}

		// Load dependents
		dependents, err := r.GetEnrollmentDependents(ctx, enrollment.ID)
		if err == nil {
			enrollment.Dependents = dependents
		}

		enrollments = append(enrollments, enrollment)
	}

	return enrollments, rows.Err()
}

// CreateEnrollment creates a new benefit enrollment
func (r *benefitsRepository) CreateEnrollment(ctx context.Context, enrollment *models.BenefitEnrollment) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO benefit_enrollments (
			id, employee_id, plan_id, coverage_level, status,
			enrollment_date, effective_date, employee_cost, employer_cost,
			total_cost, payroll_deduction, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	now := time.Now()
	enrollment.ID = uuid.New()
	enrollment.CreatedAt = now
	enrollment.UpdatedAt = now
	enrollment.EnrollmentDate = now
	enrollment.Status = models.EnrollmentStatusActive

	_, err = tx.Exec(ctx, query,
		enrollment.ID, enrollment.EmployeeID, enrollment.PlanID,
		enrollment.CoverageLevel, enrollment.Status, enrollment.EnrollmentDate,
		enrollment.EffectiveDate, enrollment.EmployeeCost, enrollment.EmployerCost,
		enrollment.TotalCost, enrollment.PayrollDeduction, enrollment.CreatedAt,
		enrollment.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// Add dependents if any
	for _, dep := range enrollment.Dependents {
		dep.EnrollmentID = enrollment.ID
		if err := r.createDependentInTx(ctx, tx, &dep); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// createDependentInTx creates a dependent within a transaction
func (r *benefitsRepository) createDependentInTx(ctx context.Context, tx interface{}, dependent *models.Dependent) error {
	query := `
		INSERT INTO benefit_dependents (
			id, enrollment_id, first_name, last_name, relationship,
			date_of_birth, ssn, active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	dependent.ID = uuid.New()
	dependent.Active = true
	dependent.CreatedAt = now
	dependent.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		dependent.ID, dependent.EnrollmentID, dependent.FirstName, dependent.LastName,
		dependent.Relationship, dependent.DateOfBirth, dependent.SSN, dependent.Active,
		dependent.CreatedAt, dependent.UpdatedAt,
	)

	return err
}

// GetEnrollmentDependents retrieves dependents for an enrollment
func (r *benefitsRepository) GetEnrollmentDependents(ctx context.Context, enrollmentID uuid.UUID) ([]models.Dependent, error) {
	query := `
		SELECT id, enrollment_id, first_name, last_name, relationship,
		       date_of_birth, active, created_at, updated_at
		FROM benefit_dependents
		WHERE enrollment_id = $1 AND active = true
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, enrollmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dependents []models.Dependent
	for rows.Next() {
		var dep models.Dependent
		err := rows.Scan(
			&dep.ID, &dep.EnrollmentID, &dep.FirstName, &dep.LastName,
			&dep.Relationship, &dep.DateOfBirth, &dep.Active,
			&dep.CreatedAt, &dep.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		dependents = append(dependents, dep)
	}

	return dependents, rows.Err()
}

// CancelEnrollment cancels a benefit enrollment
func (r *benefitsRepository) CancelEnrollment(ctx context.Context, id uuid.UUID, terminationDate time.Time) error {
	query := `
		UPDATE benefit_enrollments
		SET status = $1, termination_date = $2, updated_at = $3
		WHERE id = $4 AND status = $5
	`

	result, err := r.db.Exec(ctx, query,
		models.EnrollmentStatusCancelled, terminationDate, time.Now(), id,
		models.EnrollmentStatusActive,
	)

	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("enrollment not found or already cancelled")
	}

	return nil
}

// GetAllEnrollments retrieves all enrollments (admin view)
func (r *benefitsRepository) GetAllEnrollments(ctx context.Context, status *models.EnrollmentStatus) ([]models.BenefitEnrollment, error) {
	query := `
		SELECT e.id, e.employee_id, e.plan_id, e.coverage_level, e.status,
		       e.enrollment_date, e.effective_date, e.termination_date,
		       e.employee_cost, e.employer_cost, e.total_cost, e.payroll_deduction,
		       e.created_at, e.updated_at,
		       p.name as plan_name, p.category as plan_category,
		       emp.first_name || ' ' || emp.last_name as employee_name
		FROM benefit_enrollments e
		INNER JOIN benefit_plans p ON e.plan_id = p.id
		INNER JOIN employees emp ON e.employee_id = emp.id
	`

	args := []interface{}{}
	if status != nil {
		query += ` WHERE e.status = $1`
		args = append(args, *status)
	}

	query += ` ORDER BY e.created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []models.BenefitEnrollment
	for rows.Next() {
		var enrollment models.BenefitEnrollment
		err := rows.Scan(
			&enrollment.ID, &enrollment.EmployeeID, &enrollment.PlanID,
			&enrollment.CoverageLevel, &enrollment.Status,
			&enrollment.EnrollmentDate, &enrollment.EffectiveDate, &enrollment.TerminationDate,
			&enrollment.EmployeeCost, &enrollment.EmployerCost, &enrollment.TotalCost,
			&enrollment.PayrollDeduction, &enrollment.CreatedAt, &enrollment.UpdatedAt,
			&enrollment.PlanName, &enrollment.PlanCategory, &enrollment.EmployeeName,
		)
		if err != nil {
			return nil, err
		}
		enrollments = append(enrollments, enrollment)
	}

	return enrollments, rows.Err()
}

// GetEnrollmentByID retrieves a specific enrollment
func (r *benefitsRepository) GetEnrollmentByID(ctx context.Context, id uuid.UUID) (*models.BenefitEnrollment, error) {
	query := `
		SELECT e.id, e.employee_id, e.plan_id, e.coverage_level, e.status,
		       e.enrollment_date, e.effective_date, e.termination_date,
		       e.employee_cost, e.employer_cost, e.total_cost, e.payroll_deduction,
		       e.created_at, e.updated_at,
		       p.name as plan_name, p.category as plan_category,
		       emp.first_name || ' ' || emp.last_name as employee_name
		FROM benefit_enrollments e
		INNER JOIN benefit_plans p ON e.plan_id = p.id
		INNER JOIN employees emp ON e.employee_id = emp.id
		WHERE e.id = $1
	`

	var enrollment models.BenefitEnrollment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&enrollment.ID, &enrollment.EmployeeID, &enrollment.PlanID,
		&enrollment.CoverageLevel, &enrollment.Status,
		&enrollment.EnrollmentDate, &enrollment.EffectiveDate, &enrollment.TerminationDate,
		&enrollment.EmployeeCost, &enrollment.EmployerCost, &enrollment.TotalCost,
		&enrollment.PayrollDeduction, &enrollment.CreatedAt, &enrollment.UpdatedAt,
		&enrollment.PlanName, &enrollment.PlanCategory, &enrollment.EmployeeName,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("enrollment not found")
	}
	if err != nil {
		return nil, err
	}

	// Load dependents
	dependents, err := r.GetEnrollmentDependents(ctx, enrollment.ID)
	if err == nil {
		enrollment.Dependents = dependents
	}

	return &enrollment, nil
}