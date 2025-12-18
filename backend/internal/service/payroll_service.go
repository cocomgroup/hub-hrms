package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
)

// PayrollService handles payroll operations
type PayrollService interface {
	CreateCompensation(ctx context.Context, req *models.CreateCompensationRequest) (*models.EmployeeCompensation, error)
	GetEmployeeCompensation(ctx context.Context, employeeID uuid.UUID) (*models.EmployeeCompensation, error)
	UpdateTaxWithholding(ctx context.Context, employeeID uuid.UUID, req *models.UpdateTaxWithholdingRequest) (*models.W2TaxWithholding, error)
	CreatePayrollPeriod(ctx context.Context, req *models.CreatePayrollPeriodRequest) (*models.PayrollPeriod, error)
	GetPayrollPeriod(ctx context.Context, periodID uuid.UUID) (*models.PayrollPeriod, error)
	ListPayrollPeriods(ctx context.Context) ([]*models.PayrollPeriod, error)
	ProcessPayroll(ctx context.Context, periodID, processedBy uuid.UUID) (*models.PayrollSummary, error)
	GetEmployeePayStubs(ctx context.Context, employeeID uuid.UUID) ([]*models.PayStub, error)
	GetPayStubDetail(ctx context.Context, payStubID uuid.UUID) (*models.PayStubDetail, error)
	Generate1099Forms(ctx context.Context, year int) ([]*models.Form1099, error)
}

type payrollService struct {
	repos *repository.Repositories
}

func NewPayrollService(repos *repository.Repositories) PayrollService {
	return &payrollService{repos: repos}
}

// ===============================
// Compensation Management
// ===============================

func (s *payrollService) CreateCompensation(ctx context.Context, req *models.CreateCompensationRequest) (*models.EmployeeCompensation, error) {
	// Validate employment type
	if req.EmploymentType != "W2" && req.EmploymentType != "1099" {
		return nil, fmt.Errorf("invalid employment type: must be W2 or 1099")
	}

	// Validate pay type
	validPayTypes := map[string]bool{"hourly": true, "salary": true, "commission": true}
	if !validPayTypes[req.PayType] {
		return nil, fmt.Errorf("invalid pay type: must be hourly, salary, or commission")
	}

	// Validate rate is provided
	if req.PayType == "hourly" && req.HourlyRate == nil {
		return nil, fmt.Errorf("hourly rate is required for hourly pay type")
	}
	if req.PayType == "salary" && req.AnnualSalary == nil {
		return nil, fmt.Errorf("annual salary is required for salary pay type")
	}

	comp := &models.EmployeeCompensation{
		EmployeeID:           req.EmployeeID,
		EmploymentType:       req.EmploymentType,
		PayType:              req.PayType,
		HourlyRate:           req.HourlyRate,
		AnnualSalary:         req.AnnualSalary,
		PayFrequency:         req.PayFrequency,
		EffectiveDate:        req.EffectiveDate,
		OvertimeEligible:     req.OvertimeEligible,
		StandardHoursPerWeek: req.StandardHoursPerWeek,
	}

	if err := s.repos.Payroll.CreateCompensation(ctx, comp); err != nil {
		return nil, fmt.Errorf("failed to create compensation: %w", err)
	}

	return comp, nil
}

func (s *payrollService) GetEmployeeCompensation(ctx context.Context, employeeID uuid.UUID) (*models.EmployeeCompensation, error) {
	return s.repos.Payroll.GetCompensationByEmployeeID(ctx, employeeID)
}

// ===============================
// Tax Withholding Management
// ===============================

func (s *payrollService) UpdateTaxWithholding(ctx context.Context, employeeID uuid.UUID, req *models.UpdateTaxWithholdingRequest) (*models.W2TaxWithholding, error) {
	// Check if tax withholding exists
	existing, err := s.repos.Payroll.GetTaxWithholdingByEmployeeID(ctx, employeeID)
	
	if err != nil {
		// Create new if doesn't exist
		tax := &models.W2TaxWithholding{
			EmployeeID:            employeeID,
			FilingStatus:          req.FilingStatus,
			FederalAllowances:     req.FederalAllowances,
			StateAllowances:       req.StateAllowances,
			AdditionalWithholding: req.AdditionalWithholding,
			ExemptFederal:         req.ExemptFederal,
			ExemptState:           req.ExemptState,
			ExemptFICA:            req.ExemptFICA,
		}
		
		if err := s.repos.Payroll.CreateTaxWithholding(ctx, tax); err != nil {
			return nil, fmt.Errorf("failed to create tax withholding: %w", err)
		}
		
		return tax, nil
	}

	// Update existing
	existing.FilingStatus = req.FilingStatus
	existing.FederalAllowances = req.FederalAllowances
	existing.StateAllowances = req.StateAllowances
	existing.AdditionalWithholding = req.AdditionalWithholding
	existing.ExemptFederal = req.ExemptFederal
	existing.ExemptState = req.ExemptState
	existing.ExemptFICA = req.ExemptFICA

	if err := s.repos.Payroll.UpdateTaxWithholding(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update tax withholding: %w", err)
	}

	return existing, nil
}

// ===============================
// Payroll Period Management
// ===============================

func (s *payrollService) CreatePayrollPeriod(ctx context.Context, req *models.CreatePayrollPeriodRequest) (*models.PayrollPeriod, error) {
	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	if req.PayDate.Before(req.EndDate) {
		return nil, fmt.Errorf("pay date must be on or after end date")
	}

	period := &models.PayrollPeriod{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		PayDate:   req.PayDate,
		Status:    "draft",
	}

	if err := s.repos.Payroll.CreatePeriod(ctx, period); err != nil {
		return nil, fmt.Errorf("failed to create payroll period: %w", err)
	}

	return period, nil
}

func (s *payrollService) GetPayrollPeriod(ctx context.Context, periodID uuid.UUID) (*models.PayrollPeriod, error) {
	return s.repos.Payroll.GetPeriodByID(ctx, periodID)
}

func (s *payrollService) ListPayrollPeriods(ctx context.Context) ([]*models.PayrollPeriod, error) {
	return s.repos.Payroll.ListPeriods(ctx, nil)
}

// ===============================
// Payroll Processing
// ===============================

func (s *payrollService) ProcessPayroll(ctx context.Context, periodID uuid.UUID, processedBy uuid.UUID) (*models.PayrollSummary, error) {
	// Get payroll period
	period, err := s.repos.Payroll.GetPeriodByID(ctx, periodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payroll period: %w", err)
	}

	if period.Status == "processed" {
		return nil, fmt.Errorf("payroll period already processed")
	}

	// Get all active employees
	employees, err := s.repos.Employee.List(ctx, map[string]interface{}{"status": "active"})
	if err != nil {
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}

	summary := &models.PayrollSummary{
		Period:          period,
		TotalEmployees:  0,
		W2Employees:     0,
		Contractors1099: 0,
	}

	// Process each employee
	for _, employee := range employees {
		if employee.Status != "active" {
			continue
		}

		// Get compensation
		comp, err := s.repos.Payroll.GetCompensationByEmployeeID(ctx, employee.ID)
		if err != nil {
			continue // Skip employees without compensation setup
		}

		var payStub *models.PayStub

		if comp.EmploymentType == "W2" {
			payStub, err = s.processW2Employee(ctx, employee, comp, period)
			summary.W2Employees++
		} else if comp.EmploymentType == "1099" {
			payStub, err = s.process1099Contractor(ctx, employee, comp, period)
			summary.Contractors1099++
		}

		if err != nil {
			return nil, fmt.Errorf("failed to process employee %s: %w", employee.ID, err)
		}

		if payStub != nil {
			summary.TotalEmployees++
			summary.TotalGrossPay += payStub.GrossPay
			summary.TotalTaxes += payStub.FederalTax + payStub.StateTax + payStub.SocialSecurity + payStub.Medicare
			summary.TotalDeductions += payStub.OtherDeductions + payStub.BenefitsDeductions
			summary.TotalNetPay += payStub.NetPay
		}
	}

	// Update period status
	now := time.Now()
	period.Status = "processed"
	period.ProcessedBy = &processedBy
	period.ProcessedAt = &now

	if err := s.repos.Payroll.UpdatePeriod(ctx, period); err != nil {
		return nil, fmt.Errorf("failed to update period status: %w", err)
	}

	summary.Status = "processed"
	return summary, nil
}

func (s *payrollService) processW2Employee(ctx context.Context, employee *models.Employee, comp *models.EmployeeCompensation, period *models.PayrollPeriod) (*models.PayStub, error) {
	// Get timesheet hours for period
	timeEntries, err := s.repos.Timesheet.GetByEmployee(ctx, employee.ID, map[string]interface{}{
		"start_date": period.StartDate,
		"end_date":   period.EndDate,
	})
	if err != nil {
		return nil, err
	}

	var regularHours, overtimeHours float64
	for _, entry := range timeEntries {
		// Simple calculation - in production, calculate per week
		totalHours := calculateTotalHours(entry)
		if totalHours > 40 && comp.OvertimeEligible {
			regularHours += 40
			overtimeHours += totalHours - 40
		} else {
			regularHours += totalHours
		}
	}

	// Calculate gross pay
	var grossPay float64
	var hourlyRate float64

	if comp.PayType == "hourly" {
		hourlyRate = *comp.HourlyRate
		grossPay = (regularHours * hourlyRate) + (overtimeHours * hourlyRate * 1.5)
	} else if comp.PayType == "salary" {
		// Calculate salary per pay period
		periodsPerYear := s.getPeriodsPerYear(comp.PayFrequency)
		grossPay = *comp.AnnualSalary / float64(periodsPerYear)
		hourlyRate = grossPay / (comp.StandardHoursPerWeek * 2) // Approximate for display
	}

	// Get tax withholding
	taxWithholding, _ := s.repos.Payroll.GetTaxWithholdingByEmployeeID(ctx, employee.ID)

	// Calculate taxes
	federalTax := s.calculateFederalTax(grossPay, taxWithholding)
	stateTax := s.calculateStateTax(grossPay, taxWithholding)
	socialSecurity := s.calculateSocialSecurity(grossPay, taxWithholding)
	medicare := s.calculateMedicare(grossPay, taxWithholding)

	// Get benefits deductions
	benefitsDeduction := s.getBenefitsDeductions(ctx, employee.ID)

	// Calculate net pay
	netPay := grossPay - federalTax - stateTax - socialSecurity - medicare - benefitsDeduction

	// Create pay stub
	payStub := &models.PayStub{
		EmployeeID:         employee.ID,
		PayrollPeriodID:    period.ID,
		GrossPay:           roundToTwoDecimals(grossPay),
		FederalTax:         roundToTwoDecimals(federalTax),
		StateTax:           roundToTwoDecimals(stateTax),
		SocialSecurity:     roundToTwoDecimals(socialSecurity),
		Medicare:           roundToTwoDecimals(medicare),
		BenefitsDeductions: roundToTwoDecimals(benefitsDeduction),
		NetPay:             roundToTwoDecimals(netPay),
		HoursWorked:        &regularHours,
		OvertimeHours:      &overtimeHours,
		HourlyRate:         &hourlyRate,
	}

	if err := s.repos.Payroll.CreatePayStub(ctx, payStub); err != nil {
		return nil, err
	}

	// Create detailed earnings
	if regularHours > 0 {
		earning := &models.PayStubEarning{
			PayStubID:   payStub.ID,
			EarningType: "regular",
			Description: "Regular Hours",
			Hours:       &regularHours,
			Rate:        &hourlyRate,
			Amount:      regularHours * hourlyRate,
		}
		s.repos.Payroll.CreatePayStubEarning(ctx, earning)
	}

	if overtimeHours > 0 {
		overtimeRate := hourlyRate * 1.5
		earning := &models.PayStubEarning{
			PayStubID:   payStub.ID,
			EarningType: "overtime",
			Description: "Overtime Hours",
			Hours:       &overtimeHours,
			Rate:        &overtimeRate,
			Amount:      overtimeHours * overtimeRate,
		}
		s.repos.Payroll.CreatePayStubEarning(ctx, earning)
	}

	// Create tax entries
	s.createTaxEntries(ctx, payStub.ID, federalTax, stateTax, socialSecurity, medicare, grossPay)

	return payStub, nil
}

func (s *payrollService) process1099Contractor(ctx context.Context, employee *models.Employee, comp *models.EmployeeCompensation, period *models.PayrollPeriod) (*models.PayStub, error) {
	// Get timesheet hours for period
	timeEntries, err := s.repos.Timesheet.GetByEmployee(ctx, employee.ID, map[string]interface{}{
		"start_date": period.StartDate,
		"end_date":   period.EndDate,
	})
	if err != nil {
		return nil, err
	}

	var totalHours float64
	for _, entry := range timeEntries {
		totalHours += calculateTotalHours(entry)
	}

	// Calculate gross pay (1099 contractors - no overtime)
	var grossPay float64
	if comp.PayType == "hourly" && comp.HourlyRate != nil {
		grossPay = totalHours * *comp.HourlyRate
	}

	// 1099 contractors - NO tax withholding (they pay their own taxes)
	netPay := grossPay

	// Create pay stub
	payStub := &models.PayStub{
		EmployeeID:      employee.ID,
		PayrollPeriodID: period.ID,
		GrossPay:        roundToTwoDecimals(grossPay),
		NetPay:          roundToTwoDecimals(netPay),
		HoursWorked:     &totalHours,
		HourlyRate:      comp.HourlyRate,
	}

	if err := s.repos.Payroll.CreatePayStub(ctx, payStub); err != nil {
		return nil, err
	}

	// Create earning entry
	earning := &models.PayStubEarning{
		PayStubID:   payStub.ID,
		EarningType: "contractor",
		Description: "Contractor Payment",
		Hours:       &totalHours,
		Rate:        comp.HourlyRate,
		Amount:      grossPay,
	}
	s.repos.Payroll.CreatePayStubEarning(ctx, earning)

	return payStub, nil
}

// ===============================
// Tax Calculations
// ===============================

func (s *payrollService) calculateFederalTax(grossPay float64, withholding *models.W2TaxWithholding) float64 {
	if withholding != nil && withholding.ExemptFederal {
		return 0
	}

	// Simplified federal tax calculation (2024 rates)
	// In production, use actual IRS tax tables based on filing status and allowances
	taxRate := 0.22 // 22% marginal rate as example
	
	if withholding != nil {
		// Reduce taxable income by allowances
		allowanceAmount := float64(withholding.FederalAllowances) * 87.50 // Weekly allowance amount
		taxableIncome := math.Max(0, grossPay-allowanceAmount)
		return taxableIncome*taxRate + withholding.AdditionalWithholding
	}

	return grossPay * taxRate
}

func (s *payrollService) calculateStateTax(grossPay float64, withholding *models.W2TaxWithholding) float64 {
	if withholding != nil && withholding.ExemptState {
		return 0
	}

	// Simplified state tax (varies by state)
	taxRate := 0.05 // 5% as example
	
	if withholding != nil {
		allowanceAmount := float64(withholding.StateAllowances) * 50.00
		taxableIncome := math.Max(0, grossPay-allowanceAmount)
		return taxableIncome * taxRate
	}

	return grossPay * taxRate
}

func (s *payrollService) calculateSocialSecurity(grossPay float64, withholding *models.W2TaxWithholding) float64 {
	if withholding != nil && withholding.ExemptFICA {
		return 0
	}

	// Social Security: 6.2% up to wage base ($168,600 in 2024)
	const rate = 0.062
	const wageBase = 168600.0
	
	// In production, check YTD earnings against wage base
	return math.Min(grossPay*rate, wageBase*rate)
}

func (s *payrollService) calculateMedicare(grossPay float64, withholding *models.W2TaxWithholding) float64 {
	if withholding != nil && withholding.ExemptFICA {
		return 0
	}

	// Medicare: 1.45% (no wage base limit)
	const rate = 0.0145
	return grossPay * rate
}

func (s *payrollService) getBenefitsDeductions(ctx context.Context, employeeID uuid.UUID) float64 {
	// Get active benefit enrollments
	enrollments, err := s.repos.Benefits.GetEmployeeEnrollments(ctx, employeeID)
	if err != nil {
		return 0
	}

	var total float64
	for _, enrollment := range enrollments {
		if enrollment.Status == models.EnrollmentStatusActive {
			total += enrollment.PayrollDeduction
		}
	}

	return total
}

func (s *payrollService) createTaxEntries(ctx context.Context, payStubID uuid.UUID, federal, state, ss, medicare, grossPay float64) {
	if federal > 0 {
		federalRate := federal / grossPay
		tax := &models.PayStubTax{
			PayStubID:   payStubID,
			TaxType:     "federal",
			Description: "Federal Income Tax",
			Amount:      federal,
			TaxableWage: grossPay,
			TaxRate:     &federalRate,
		}
		s.repos.Payroll.CreatePayStubTax(ctx, tax)
	}

	if state > 0 {
		stateRate := state / grossPay
		tax := &models.PayStubTax{
			PayStubID:   payStubID,
			TaxType:     "state",
			Description: "State Income Tax",
			Amount:      state,
			TaxableWage: grossPay,
			TaxRate:     &stateRate,
		}
		s.repos.Payroll.CreatePayStubTax(ctx, tax)
	}

	if ss > 0 {
		ssRate := 0.062
		tax := &models.PayStubTax{
			PayStubID:   payStubID,
			TaxType:     "fica_ss",
			Description: "Social Security",
			Amount:      ss,
			TaxableWage: grossPay,
			TaxRate:     &ssRate,
		}
		s.repos.Payroll.CreatePayStubTax(ctx, tax)
	}

	if medicare > 0 {
		medicareRate := 0.0145
		tax := &models.PayStubTax{
			PayStubID:   payStubID,
			TaxType:     "fica_medicare",
			Description: "Medicare",
			Amount:      medicare,
			TaxableWage: grossPay,
			TaxRate:     &medicareRate,
		}
		s.repos.Payroll.CreatePayStubTax(ctx, tax)
	}
}

// ===============================
// Pay Stub Retrieval
// ===============================

func (s *payrollService) GetEmployeePayStubs(ctx context.Context, employeeID uuid.UUID) ([]*models.PayStub, error) {
	return s.repos.Payroll.ListPayStubsByEmployee(ctx, employeeID)
}

func (s *payrollService) GetPayStubDetail(ctx context.Context, payStubID uuid.UUID) (*models.PayStubDetail, error) {
	// Get pay stub
	payStub, err := s.repos.Payroll.GetPayStubByID(ctx, payStubID)
	if err != nil {
		return nil, err
	}

	// Get employee
	employee, err := s.repos.Employee.GetByID(ctx, payStub.EmployeeID)
	if err != nil {
		return nil, err
	}

	// Get payroll period
	period, err := s.repos.Payroll.GetPeriodByID(ctx, payStub.PayrollPeriodID)
	if err != nil {
		return nil, err
	}

	// Get detailed breakdowns
	earnings, _ := s.repos.Payroll.GetPayStubEarnings(ctx, payStubID)
	deductions, _ := s.repos.Payroll.GetPayStubDeductions(ctx, payStubID)
	taxes, _ := s.repos.Payroll.GetPayStubTaxes(ctx, payStubID)

	// Calculate YTD totals
	year := period.EndDate.Year()
	ytdEarnings, _ := s.repos.Payroll.GetYTDEarnings(ctx, employee.ID, year)
	ytdTaxes, _ := s.repos.Payroll.GetYTDTaxes(ctx, employee.ID, year)

	return &models.PayStubDetail{
		PayStub:       payStub,
		Employee:      employee,
		PayrollPeriod: period,
		Earnings:      earnings,
		Deductions:    deductions,
		Taxes:         taxes,
		YTDGrossPay:   ytdEarnings,
		YTDNetPay:     ytdEarnings - ytdTaxes,
		YTDFederalTax: 0, // Would need separate query
		YTDStateTax:   0, // Would need separate query
	}, nil
}

// ===============================
// 1099 Form Generation
// ===============================

func (s *payrollService) Generate1099Forms(ctx context.Context, year int) ([]*models.Form1099, error) {
	// Get all contractors
	employees, err := s.repos.Employee.List(ctx, map[string]interface{}{"status": "active"})
	if err != nil {
		return nil, err
	}

	var forms []*models.Form1099

	for _, employee := range employees {
		// Get compensation to check if 1099
		comp, err := s.repos.Payroll.GetCompensationByEmployeeID(ctx, employee.ID)
		if err != nil || comp.EmploymentType != "1099" {
			continue
		}

		// Calculate total payments for year
		totalPayments, err := s.repos.Payroll.GetYTDEarnings(ctx, employee.ID, year)
		if err != nil || totalPayments == 0 {
			continue
		}

		// Create 1099 form
		form := &models.Form1099{
			EmployeeID:         employee.ID,
			TaxYear:            year,
			TotalPayments:      totalPayments,
			FederalTaxWithheld: 0, // 1099 contractors typically have no withholding
			StateTaxWithheld:   0,
			Status:             "draft",
		}

		if err := s.repos.Payroll.Create1099(ctx, form); err != nil {
			continue
		}

		forms = append(forms, form)
	}

	return forms, nil
}

// ===============================
// Utility Functions
// ===============================

func (s *payrollService) getPeriodsPerYear(frequency string) int {
	switch frequency {
	case "weekly":
		return 52
	case "biweekly":
		return 26
	case "semimonthly":
		return 24
	case "monthly":
		return 12
	default:
		return 26 // Default to biweekly
	}
}

func roundToTwoDecimals(val float64) float64 {
	return math.Round(val*100) / 100
}

func calculateTotalHours(entry *models.Timesheet) float64 {
	// Simple placeholder - in production, calculate from time entry details
	// This would sum up all hours from the timesheet
	return 8.0 // Placeholder
}