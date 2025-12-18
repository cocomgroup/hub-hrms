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


type payrollRepository struct {
	db *pgxpool.Pool
}

func NewPayrollRepository(db *pgxpool.Pool) PayrollRepository {
	return &payrollRepository{db: db}
}

// ===============================
// Compensation Methods
// ===============================

func (r *payrollRepository) CreateCompensation(ctx context.Context, comp *models.EmployeeCompensation) error {
	query := `
		INSERT INTO employee_compensation (
			id, employee_id, employment_type, pay_type, hourly_rate, annual_salary,
			pay_frequency, effective_date, end_date, overtime_eligible, 
			standard_hours_per_week, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	comp.ID = uuid.New()
	now := time.Now()
	comp.CreatedAt = now
	comp.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		comp.ID, comp.EmployeeID, comp.EmploymentType, comp.PayType, comp.HourlyRate,
		comp.AnnualSalary, comp.PayFrequency, comp.EffectiveDate, comp.EndDate,
		comp.OvertimeEligible, comp.StandardHoursPerWeek, comp.CreatedAt, comp.UpdatedAt,
	)

	return err
}

func (r *payrollRepository) GetCompensationByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.EmployeeCompensation, error) {
	query := `
		SELECT id, employee_id, employment_type, pay_type, hourly_rate, annual_salary,
		       pay_frequency, effective_date, end_date, overtime_eligible,
		       standard_hours_per_week, created_at, updated_at
		FROM employee_compensation
		WHERE employee_id = $1 AND (end_date IS NULL OR end_date > NOW())
		ORDER BY effective_date DESC
		LIMIT 1
	`

	var comp models.EmployeeCompensation
	err := r.db.QueryRow(ctx, query, employeeID).Scan(
		&comp.ID, &comp.EmployeeID, &comp.EmploymentType, &comp.PayType,
		&comp.HourlyRate, &comp.AnnualSalary, &comp.PayFrequency, &comp.EffectiveDate,
		&comp.EndDate, &comp.OvertimeEligible, &comp.StandardHoursPerWeek,
		&comp.CreatedAt, &comp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("compensation not found for employee")
	}

	return &comp, err
}

func (r *payrollRepository) UpdateCompensation(ctx context.Context, comp *models.EmployeeCompensation) error {
	query := `
		UPDATE employee_compensation
		SET employment_type = $1, pay_type = $2, hourly_rate = $3, annual_salary = $4,
		    pay_frequency = $5, overtime_eligible = $6, standard_hours_per_week = $7,
		    updated_at = $8
		WHERE id = $9
	`

	comp.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		comp.EmploymentType, comp.PayType, comp.HourlyRate, comp.AnnualSalary,
		comp.PayFrequency, comp.OvertimeEligible, comp.StandardHoursPerWeek,
		comp.UpdatedAt, comp.ID,
	)

	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("compensation not found")
	}

	return nil
}

// ===============================
// Tax Withholding Methods (W2)
// ===============================

func (r *payrollRepository) CreateTaxWithholding(ctx context.Context, tax *models.W2TaxWithholding) error {
	query := `
		INSERT INTO w2_tax_withholding (
			id, employee_id, filing_status, federal_allowances, state_allowances,
			additional_withholding, exempt_federal, exempt_state, exempt_fica,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	tax.ID = uuid.New()
	now := time.Now()
	tax.CreatedAt = now
	tax.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		tax.ID, tax.EmployeeID, tax.FilingStatus, tax.FederalAllowances,
		tax.StateAllowances, tax.AdditionalWithholding, tax.ExemptFederal,
		tax.ExemptState, tax.ExemptFICA, tax.CreatedAt, tax.UpdatedAt,
	)

	return err
}

func (r *payrollRepository) GetTaxWithholdingByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.W2TaxWithholding, error) {
	query := `
		SELECT id, employee_id, filing_status, federal_allowances, state_allowances,
		       additional_withholding, exempt_federal, exempt_state, exempt_fica,
		       created_at, updated_at
		FROM w2_tax_withholding
		WHERE employee_id = $1
	`

	var tax models.W2TaxWithholding
	err := r.db.QueryRow(ctx, query, employeeID).Scan(
		&tax.ID, &tax.EmployeeID, &tax.FilingStatus, &tax.FederalAllowances,
		&tax.StateAllowances, &tax.AdditionalWithholding, &tax.ExemptFederal,
		&tax.ExemptState, &tax.ExemptFICA, &tax.CreatedAt, &tax.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tax withholding not found")
	}

	return &tax, err
}

func (r *payrollRepository) UpdateTaxWithholding(ctx context.Context, tax *models.W2TaxWithholding) error {
	query := `
		UPDATE w2_tax_withholding
		SET filing_status = $1, federal_allowances = $2, state_allowances = $3,
		    additional_withholding = $4, exempt_federal = $5, exempt_state = $6,
		    exempt_fica = $7, updated_at = $8
		WHERE id = $9
	`

	tax.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		tax.FilingStatus, tax.FederalAllowances, tax.StateAllowances,
		tax.AdditionalWithholding, tax.ExemptFederal, tax.ExemptState,
		tax.ExemptFICA, tax.UpdatedAt, tax.ID,
	)

	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tax withholding not found")
	}

	return nil
}

// ===============================
// Payroll Period Methods
// ===============================

func (r *payrollRepository) CreatePeriod(ctx context.Context, period *models.PayrollPeriod) error {
	query := `
		INSERT INTO payroll_periods (
			id, start_date, end_date, pay_date, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	period.ID = uuid.New()
	now := time.Now()
	period.CreatedAt = now
	period.UpdatedAt = now
	period.Status = "draft"

	_, err := r.db.Exec(ctx, query,
		period.ID, period.StartDate, period.EndDate, period.PayDate,
		period.Status, period.CreatedAt, period.UpdatedAt,
	)

	return err
}

func (r *payrollRepository) GetPeriodByID(ctx context.Context, id uuid.UUID) (*models.PayrollPeriod, error) {
	query := `
		SELECT id, start_date, end_date, pay_date, status, processed_by,
		       processed_at, created_at, updated_at
		FROM payroll_periods
		WHERE id = $1
	`

	var period models.PayrollPeriod
	err := r.db.QueryRow(ctx, query, id).Scan(
		&period.ID, &period.StartDate, &period.EndDate, &period.PayDate,
		&period.Status, &period.ProcessedBy, &period.ProcessedAt,
		&period.CreatedAt, &period.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payroll period not found")
	}

	return &period, err
}

func (r *payrollRepository) ListPeriods(ctx context.Context, filters map[string]interface{}) ([]*models.PayrollPeriod, error) {
	query := `
		SELECT id, start_date, end_date, pay_date, status, processed_by,
		       processed_at, created_at, updated_at
		FROM payroll_periods
		ORDER BY start_date DESC
		LIMIT 50
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var periods []*models.PayrollPeriod
	for rows.Next() {
		var period models.PayrollPeriod
		err := rows.Scan(
			&period.ID, &period.StartDate, &period.EndDate, &period.PayDate,
			&period.Status, &period.ProcessedBy, &period.ProcessedAt,
			&period.CreatedAt, &period.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		periods = append(periods, &period)
	}

	return periods, rows.Err()
}

func (r *payrollRepository) UpdatePeriod(ctx context.Context, period *models.PayrollPeriod) error {
	query := `
		UPDATE payroll_periods
		SET status = $1, processed_by = $2, processed_at = $3, updated_at = $4
		WHERE id = $5
	`

	period.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		period.Status, period.ProcessedBy, period.ProcessedAt,
		period.UpdatedAt, period.ID,
	)

	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("payroll period not found")
	}

	return nil
}

// ===============================
// Pay Stub Methods
// ===============================

func (r *payrollRepository) CreatePayStub(ctx context.Context, stub *models.PayStub) error {
	query := `
		INSERT INTO pay_stubs (
			id, employee_id, payroll_period_id, gross_pay, federal_tax, state_tax,
			social_security, medicare, other_deductions, net_pay, hours_worked,
			overtime_hours, hourly_rate, benefits_deductions, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	stub.ID = uuid.New()
	now := time.Now()
	stub.CreatedAt = now
	stub.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		stub.ID, stub.EmployeeID, stub.PayrollPeriodID, stub.GrossPay,
		stub.FederalTax, stub.StateTax, stub.SocialSecurity, stub.Medicare,
		stub.OtherDeductions, stub.NetPay, stub.HoursWorked, stub.OvertimeHours,
		stub.HourlyRate, stub.BenefitsDeductions, stub.CreatedAt, stub.UpdatedAt,
	)

	return err
}

func (r *payrollRepository) GetPayStubByID(ctx context.Context, id uuid.UUID) (*models.PayStub, error) {
	query := `
		SELECT id, employee_id, payroll_period_id, gross_pay, federal_tax, state_tax,
		       social_security, medicare, other_deductions, net_pay, hours_worked,
		       overtime_hours, hourly_rate, benefits_deductions, created_at, updated_at
		FROM pay_stubs
		WHERE id = $1
	`

	var stub models.PayStub
	err := r.db.QueryRow(ctx, query, id).Scan(
		&stub.ID, &stub.EmployeeID, &stub.PayrollPeriodID, &stub.GrossPay,
		&stub.FederalTax, &stub.StateTax, &stub.SocialSecurity, &stub.Medicare,
		&stub.OtherDeductions, &stub.NetPay, &stub.HoursWorked, &stub.OvertimeHours,
		&stub.HourlyRate, &stub.BenefitsDeductions, &stub.CreatedAt, &stub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("pay stub not found")
	}

	return &stub, err
}

func (r *payrollRepository) ListPayStubsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.PayStub, error) {
	query := `
		SELECT ps.id, ps.employee_id, ps.payroll_period_id, ps.gross_pay, ps.federal_tax,
		       ps.state_tax, ps.social_security, ps.medicare, ps.other_deductions,
		       ps.net_pay, ps.hours_worked, ps.overtime_hours, ps.hourly_rate,
		       ps.benefits_deductions, ps.created_at, ps.updated_at
		FROM pay_stubs ps
		JOIN payroll_periods pp ON ps.payroll_period_id = pp.id
		WHERE ps.employee_id = $1
		ORDER BY pp.start_date DESC
		LIMIT 12
	`

	rows, err := r.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stubs []*models.PayStub
	for rows.Next() {
		var stub models.PayStub
		err := rows.Scan(
			&stub.ID, &stub.EmployeeID, &stub.PayrollPeriodID, &stub.GrossPay,
			&stub.FederalTax, &stub.StateTax, &stub.SocialSecurity, &stub.Medicare,
			&stub.OtherDeductions, &stub.NetPay, &stub.HoursWorked, &stub.OvertimeHours,
			&stub.HourlyRate, &stub.BenefitsDeductions, &stub.CreatedAt, &stub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stubs = append(stubs, &stub)
	}

	return stubs, rows.Err()
}

func (r *payrollRepository) ListPayStubsByPeriod(ctx context.Context, periodID uuid.UUID) ([]*models.PayStub, error) {
	query := `
		SELECT id, employee_id, payroll_period_id, gross_pay, federal_tax, state_tax,
		       social_security, medicare, other_deductions, net_pay, hours_worked,
		       overtime_hours, hourly_rate, benefits_deductions, created_at, updated_at
		FROM pay_stubs
		WHERE payroll_period_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, periodID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stubs []*models.PayStub
	for rows.Next() {
		var stub models.PayStub
		err := rows.Scan(
			&stub.ID, &stub.EmployeeID, &stub.PayrollPeriodID, &stub.GrossPay,
			&stub.FederalTax, &stub.StateTax, &stub.SocialSecurity, &stub.Medicare,
			&stub.OtherDeductions, &stub.NetPay, &stub.HoursWorked, &stub.OvertimeHours,
			&stub.HourlyRate, &stub.BenefitsDeductions, &stub.CreatedAt, &stub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stubs = append(stubs, &stub)
	}

	return stubs, rows.Err()
}

// ===============================
// Pay Stub Details Methods
// ===============================

func (r *payrollRepository) CreatePayStubEarning(ctx context.Context, earning *models.PayStubEarning) error {
	query := `
		INSERT INTO pay_stub_earnings (
			id, pay_stub_id, earning_type, description, hours, rate, amount, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	earning.ID = uuid.New()
	earning.CreatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		earning.ID, earning.PayStubID, earning.EarningType, earning.Description,
		earning.Hours, earning.Rate, earning.Amount, earning.CreatedAt,
	)

	return err
}

func (r *payrollRepository) CreatePayStubDeduction(ctx context.Context, deduction *models.PayStubDeduction) error {
	query := `
		INSERT INTO pay_stub_deductions (
			id, pay_stub_id, deduction_type, description, amount, employer_match, pre_tax, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	deduction.ID = uuid.New()
	deduction.CreatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		deduction.ID, deduction.PayStubID, deduction.DeductionType, deduction.Description,
		deduction.Amount, deduction.EmployerMatch, deduction.PreTax, deduction.CreatedAt,
	)

	return err
}

func (r *payrollRepository) CreatePayStubTax(ctx context.Context, tax *models.PayStubTax) error {
	query := `
		INSERT INTO pay_stub_taxes (
			id, pay_stub_id, tax_type, description, amount, taxable_wage, tax_rate, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	tax.ID = uuid.New()
	tax.CreatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		tax.ID, tax.PayStubID, tax.TaxType, tax.Description,
		tax.Amount, tax.TaxableWage, tax.TaxRate, tax.CreatedAt,
	)

	return err
}

func (r *payrollRepository) GetPayStubEarnings(ctx context.Context, payStubID uuid.UUID) ([]models.PayStubEarning, error) {
	query := `
		SELECT id, pay_stub_id, earning_type, description, hours, rate, amount, created_at
		FROM pay_stub_earnings
		WHERE pay_stub_id = $1
		ORDER BY earning_type
	`

	rows, err := r.db.Query(ctx, query, payStubID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var earnings []models.PayStubEarning
	for rows.Next() {
		var earning models.PayStubEarning
		err := rows.Scan(
			&earning.ID, &earning.PayStubID, &earning.EarningType, &earning.Description,
			&earning.Hours, &earning.Rate, &earning.Amount, &earning.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		earnings = append(earnings, earning)
	}

	return earnings, rows.Err()
}

func (r *payrollRepository) GetPayStubDeductions(ctx context.Context, payStubID uuid.UUID) ([]models.PayStubDeduction, error) {
	query := `
		SELECT id, pay_stub_id, deduction_type, description, amount, employer_match, pre_tax, created_at
		FROM pay_stub_deductions
		WHERE pay_stub_id = $1
		ORDER BY deduction_type
	`

	rows, err := r.db.Query(ctx, query, payStubID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deductions []models.PayStubDeduction
	for rows.Next() {
		var deduction models.PayStubDeduction
		err := rows.Scan(
			&deduction.ID, &deduction.PayStubID, &deduction.DeductionType, &deduction.Description,
			&deduction.Amount, &deduction.EmployerMatch, &deduction.PreTax, &deduction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		deductions = append(deductions, deduction)
	}

	return deductions, rows.Err()
}

func (r *payrollRepository) GetPayStubTaxes(ctx context.Context, payStubID uuid.UUID) ([]models.PayStubTax, error) {
	query := `
		SELECT id, pay_stub_id, tax_type, description, amount, taxable_wage, tax_rate, created_at
		FROM pay_stub_taxes
		WHERE pay_stub_id = $1
		ORDER BY tax_type
	`

	rows, err := r.db.Query(ctx, query, payStubID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taxes []models.PayStubTax
	for rows.Next() {
		var tax models.PayStubTax
		err := rows.Scan(
			&tax.ID, &tax.PayStubID, &tax.TaxType, &tax.Description,
			&tax.Amount, &tax.TaxableWage, &tax.TaxRate, &tax.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		taxes = append(taxes, tax)
	}

	return taxes, rows.Err()
}

// ===============================
// 1099 Form Methods
// ===============================

func (r *payrollRepository) Create1099(ctx context.Context, form *models.Form1099) error {
	query := `
		INSERT INTO form_1099 (
			id, employee_id, tax_year, total_payments, federal_tax_withheld,
			state_tax_withheld, status, filed_date, corrected_form_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	form.ID = uuid.New()
	now := time.Now()
	form.CreatedAt = now
	form.UpdatedAt = now
	form.Status = "draft"

	_, err := r.db.Exec(ctx, query,
		form.ID, form.EmployeeID, form.TaxYear, form.TotalPayments,
		form.FederalTaxWithheld, form.StateTaxWithheld, form.Status,
		form.FiledDate, form.CorrectedFormID, form.CreatedAt, form.UpdatedAt,
	)

	return err
}

func (r *payrollRepository) Get1099ByEmployeeAndYear(ctx context.Context, employeeID uuid.UUID, year int) (*models.Form1099, error) {
	query := `
		SELECT id, employee_id, tax_year, total_payments, federal_tax_withheld,
		       state_tax_withheld, status, filed_date, corrected_form_id, created_at, updated_at
		FROM form_1099
		WHERE employee_id = $1 AND tax_year = $2
	`

	var form models.Form1099
	err := r.db.QueryRow(ctx, query, employeeID, year).Scan(
		&form.ID, &form.EmployeeID, &form.TaxYear, &form.TotalPayments,
		&form.FederalTaxWithheld, &form.StateTaxWithheld, &form.Status,
		&form.FiledDate, &form.CorrectedFormID, &form.CreatedAt, &form.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("1099 form not found")
	}

	return &form, err
}

func (r *payrollRepository) List1099ByYear(ctx context.Context, year int) ([]*models.Form1099, error) {
	query := `
		SELECT id, employee_id, tax_year, total_payments, federal_tax_withheld,
		       state_tax_withheld, status, filed_date, corrected_form_id, created_at, updated_at
		FROM form_1099
		WHERE tax_year = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var forms []*models.Form1099
	for rows.Next() {
		var form models.Form1099
		err := rows.Scan(
			&form.ID, &form.EmployeeID, &form.TaxYear, &form.TotalPayments,
			&form.FederalTaxWithheld, &form.StateTaxWithheld, &form.Status,
			&form.FiledDate, &form.CorrectedFormID, &form.CreatedAt, &form.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		forms = append(forms, &form)
	}

	return forms, rows.Err()
}

func (r *payrollRepository) Update1099(ctx context.Context, form *models.Form1099) error {
	query := `
		UPDATE form_1099
		SET total_payments = $1, federal_tax_withheld = $2, state_tax_withheld = $3,
		    status = $4, filed_date = $5, updated_at = $6
		WHERE id = $7
	`

	form.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		form.TotalPayments, form.FederalTaxWithheld, form.StateTaxWithheld,
		form.Status, form.FiledDate, form.UpdatedAt, form.ID,
	)

	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("1099 form not found")
	}

	return nil
}

// ===============================
// Year-to-Date Calculations
// ===============================

func (r *payrollRepository) GetYTDEarnings(ctx context.Context, employeeID uuid.UUID, year int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(gross_pay), 0)
		FROM pay_stubs ps
		JOIN payroll_periods pp ON ps.payroll_period_id = pp.id
		WHERE ps.employee_id = $1
		  AND EXTRACT(YEAR FROM pp.end_date) = $2
	`

	var total float64
	err := r.db.QueryRow(ctx, query, employeeID, year).Scan(&total)
	return total, err
}

func (r *payrollRepository) GetYTDTaxes(ctx context.Context, employeeID uuid.UUID, year int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(federal_tax + state_tax + social_security + medicare), 0)
		FROM pay_stubs ps
		JOIN payroll_periods pp ON ps.payroll_period_id = pp.id
		WHERE ps.employee_id = $1
		  AND EXTRACT(YEAR FROM pp.end_date) = $2
	`

	var total float64
	err := r.db.QueryRow(ctx, query, employeeID, year).Scan(&total)
	return total, err
}