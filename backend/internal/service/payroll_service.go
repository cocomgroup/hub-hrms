package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"math"

	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
)

type PayrollService interface {
	CreatePayrollPeriod(ctx context.Context, req *models.PayrollPeriodRequest) (*models.PayrollPeriod, error)
	GetPayrollPeriod(ctx context.Context, id uuid.UUID) (*models.PayrollPeriod, error)
	ListPayrollPeriods(ctx context.Context) ([]*models.PayrollPeriod, error)
	UpdatePayrollPeriod(ctx context.Context, id uuid.UUID, req *models.PayrollPeriodRequest) (*models.PayrollPeriod, error)
	ProcessPayroll(ctx context.Context, periodID uuid.UUID) (*models.PayrollSummary, error)
	GetPayStub(ctx context.Context, id uuid.UUID) (*models.PayStub, error)
	GetPayStubsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.PayStub, error)
	GetPayStubsByPeriod(ctx context.Context, periodID uuid.UUID) ([]*models.PayStub, error)
}

type payrollService struct {
	repos *repository.Repositories
}

func NewPayrollService(repos *repository.Repositories) PayrollService {
	return &payrollService{
		repos: repos,
	}
}

// CreatePayrollPeriod creates a new payroll period
func (s *payrollService) CreatePayrollPeriod(ctx context.Context, req *models.PayrollPeriodRequest) (*models.PayrollPeriod, error) {
	period := &models.PayrollPeriod{
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		PayDate:    req.PayDate,
		PeriodType: req.PeriodType,
		Status:     "pending",
	}

	if err := s.repos.Payroll.CreatePayrollPeriod(ctx, period); err != nil {
		return nil, fmt.Errorf("failed to create payroll period: %w", err)
	}

	return period, nil
}

// GetPayrollPeriod retrieves a payroll period by ID
func (s *payrollService) GetPayrollPeriod(ctx context.Context, id uuid.UUID) (*models.PayrollPeriod, error) {
	period, err := s.repos.Payroll.GetPayrollPeriod(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payroll period: %w", err)
	}
	return period, nil
}

// ListPayrollPeriods retrieves all payroll periods
func (s *payrollService) ListPayrollPeriods(ctx context.Context) ([]*models.PayrollPeriod, error) {
	periods, err := s.repos.Payroll.ListPayrollPeriods(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list payroll periods: %w", err)
	}
	return periods, nil
}

// UpdatePayrollPeriod updates a payroll period
func (s *payrollService) UpdatePayrollPeriod(ctx context.Context, id uuid.UUID, req *models.PayrollPeriodRequest) (*models.PayrollPeriod, error) {
	period, err := s.repos.Payroll.GetPayrollPeriod(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("payroll period not found: %w", err)
	}

	// Update fields
	period.StartDate = req.StartDate
	period.EndDate = req.EndDate
	period.PayDate = req.PayDate
	period.PeriodType = req.PeriodType

	if err := s.repos.Payroll.UpdatePayrollPeriod(ctx, period); err != nil {
		return nil, fmt.Errorf("failed to update payroll period: %w", err)
	}

	return period, nil
}

// ProcessPayroll processes payroll for a given period
func (s *payrollService) ProcessPayroll(ctx context.Context, periodID uuid.UUID) (*models.PayrollSummary, error) {
	// Get payroll period
	period, err := s.repos.Payroll.GetPayrollPeriod(ctx, periodID)
	if err != nil {
		return nil, fmt.Errorf("payroll period not found: %w", err)
	}

	// Check if already processed
	if period.Status == "processed" {
		return nil, fmt.Errorf("payroll period already processed")
	}

	// Get all active employees
	employees, err := s.repos.Employee.List(ctx, map[string]interface{}{
		"status": "active",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}

	summary := &models.PayrollSummary{
		PeriodID:        periodID,
		StartDate:       period.StartDate,
		EndDate:         period.EndDate,
		EmployeeCount:   len(employees),
		ProcessedAt:     time.Now(),
		TotalGrossPay:   0,
		TotalNetPay:     0,
		TotalTaxes:      0,
	}

	// Process each employee
	for _, employee := range employees {
		// Get employee compensation - FIXED: using Payroll repository
		comp, err := s.repos.Payroll.GetByEmployeeID(ctx, employee.ID)
		if err != nil {
			// Skip employees without compensation setup
			continue
		}

		var payStub *models.PayStub
		
		// Process based on employment type
		if employee.EmploymentType == "W2" || employee.EmploymentType == "full-time" || employee.EmploymentType == "part-time" {
			payStub, err = s.processW2Employee(ctx, employee, comp, period)
		} else if employee.EmploymentType == "1099" || employee.EmploymentType == "contractor" {
			payStub, err = s.process1099Contractor(ctx, employee, comp, period)
		} else {
			continue // Skip unknown employment types
		}

		if err != nil {
			// Log error but continue processing other employees
			continue
		}

		// Update summary
		summary.TotalGrossPay += payStub.GrossPay
		summary.TotalNetPay += payStub.NetPay
		if payStub.FederalTax != nil {
			summary.TotalTaxes += *payStub.FederalTax
		}
		if payStub.StateTax != nil {
			summary.TotalTaxes += *payStub.StateTax
		}
		if payStub.SocialSecurity != nil {
			summary.TotalTaxes += *payStub.SocialSecurity
		}
		if payStub.Medicare != nil {
			summary.TotalTaxes += *payStub.Medicare
		}
	}

	// Update period status
	period.Status = "processed"
	if err := s.repos.Payroll.UpdatePayrollPeriod(ctx, period); err != nil {
		return nil, fmt.Errorf("failed to update period status: %w", err)
	}

	summary.Status = "processed"
	return summary, nil
}

func (s *payrollService) processW2Employee(ctx context.Context, employee *models.Employee, comp *models.EmployeeCompensation, period *models.PayrollPeriod) (*models.PayStub, error) {
	// Get time entries for the payroll period
	timeEntries, err := s.repos.Timesheet.GetTimeEntriesByEmployee(ctx, employee.ID, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get time entries: %w", err)
	}

	// Calculate hours from time entries
	var regularHours, overtimeHours float64
	
	// Group entries by week to calculate overtime properly
	weeklyHours := make(map[string]float64) // key: week start date
	
	for _, entry := range timeEntries {
		if entry.Type != "regular" {
			// Skip PTO and holiday hours for now (handle separately if needed)
			continue
		}
		
		// Get the Monday of the week for this entry
		weekStart := getWeekStart(entry.Date)
		weekKey := weekStart.Format("2006-01-02")
		weeklyHours[weekKey] += entry.Hours
	}
	
	// Calculate regular and overtime hours per week
	for _, weekHours := range weeklyHours {
		if weekHours > 40 && comp.OvertimeEligible {
			regularHours += 40
			overtimeHours += weekHours - 40
		} else {
			regularHours += weekHours
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
		annualSalary := *comp.Salary
		
		var periodsPerYear int
		switch period.PeriodType {
		case "weekly":
			periodsPerYear = 52
		case "bi-weekly":
			periodsPerYear = 26
		case "semi-monthly":
			periodsPerYear = 24
		case "monthly":
			periodsPerYear = 12
		default:
			periodsPerYear = 26 // Default to bi-weekly
		}
		
		grossPay = annualSalary / float64(periodsPerYear)
		
		// Calculate effective hourly rate for display
		if regularHours > 0 {
			hourlyRate = grossPay / regularHours
		}
	}

	// Calculate taxes (simplified)
	federalTax := calculateFederalTax(grossPay, comp.FilingStatus)
	stateTax := calculateStateTax(grossPay, comp.State)
	socialSecurity := grossPay * 0.062 // 6.2%
	medicare := grossPay * 0.0145      // 1.45%

	// Calculate net pay
	totalDeductions := federalTax + stateTax + socialSecurity + medicare
	netPay := grossPay - totalDeductions

	// Create pay stub
	totalHours := regularHours + overtimeHours
	payStub := &models.PayStub{
		EmployeeID:      employee.ID,
		PayrollPeriodID: period.ID,
		GrossPay:        roundToTwoDecimals(grossPay),
		NetPay:          roundToTwoDecimals(netPay),
		FederalTax:      floatPtr(roundToTwoDecimals(federalTax)),
		StateTax:        floatPtr(roundToTwoDecimals(stateTax)),
		SocialSecurity:  floatPtr(roundToTwoDecimals(socialSecurity)),
		Medicare:        floatPtr(roundToTwoDecimals(medicare)),
		HoursWorked:     &totalHours,
		RegularHours:    &regularHours,
		OvertimeHours:   &overtimeHours,
		HourlyRate:      &hourlyRate,
	}

	if err := s.repos.Payroll.CreatePayStub(ctx, payStub); err != nil {
		return nil, fmt.Errorf("failed to create pay stub: %w", err)
	}

	// Create tax entries
	s.createTaxEntries(ctx, payStub.ID, federalTax, stateTax, socialSecurity, medicare, grossPay)

	return payStub, nil
}

func (s *payrollService) process1099Contractor(ctx context.Context, employee *models.Employee, comp *models.EmployeeCompensation, period *models.PayrollPeriod) (*models.PayStub, error) {
	// Get time entries for the payroll period
	timeEntries, err := s.repos.Timesheet.GetTimeEntriesByEmployee(ctx, employee.ID, period.StartDate, period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get time entries: %w", err)
	}

	// Calculate total hours from time entries
	var totalHours float64
	for _, entry := range timeEntries {
		if entry.Type == "regular" {
			totalHours += entry.Hours
		}
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
		// No tax withholding for 1099
		FederalTax:     nil,
		StateTax:       nil,
		SocialSecurity: nil,
		Medicare:       nil,
	}

	if err := s.repos.Payroll.CreatePayStub(ctx, payStub); err != nil {
		return nil, fmt.Errorf("failed to create pay stub: %w", err)
	}

	return payStub, nil
}

// GetPayStub retrieves a pay stub by ID
func (s *payrollService) GetPayStub(ctx context.Context, id uuid.UUID) (*models.PayStub, error) {
	payStub, err := s.repos.Payroll.GetPayStub(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get pay stub: %w", err)
	}
	return payStub, nil
}

// GetPayStubsByEmployee retrieves all pay stubs for an employee
func (s *payrollService) GetPayStubsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.PayStub, error) {
	payStubs, err := s.repos.Payroll.GetPayStubsByEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pay stubs: %w", err)
	}
	return payStubs, nil
}

// GetPayStubsByPeriod retrieves all pay stubs for a payroll period
func (s *payrollService) GetPayStubsByPeriod(ctx context.Context, periodID uuid.UUID) ([]*models.PayStub, error) {
	payStubs, err := s.repos.Payroll.GetPayStubsByPeriod(ctx, periodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pay stubs: %w", err)
	}
	return payStubs, nil
}

// Helper functions

func (s *payrollService) createTaxEntries(ctx context.Context, payStubID uuid.UUID, federalTax, stateTax, socialSecurity, medicare, grossPay float64) {
	// Create tax entry records for audit trail
	taxes := []struct {
		taxType string
		amount  float64
	}{
		{"federal", federalTax},
		{"state", stateTax},
		{"social_security", socialSecurity},
		{"medicare", medicare},
	}

	for _, tax := range taxes {
		if tax.amount > 0 {
			taxEntry := &models.TaxEntry{
				PayStubID:   payStubID,
				TaxType:     tax.taxType,
				Amount:      roundToTwoDecimals(tax.amount),
				TaxableWage: grossPay,
				Percentage: sql.NullFloat64{
					Float64: (tax.amount / grossPay) * 100,
					Valid:   true,
				},
			}
			s.repos.Payroll.CreateTaxEntry(ctx, taxEntry)
		}
	}
}

func calculateFederalTax(grossPay float64, filingStatus string) float64 {
	// Simplified federal tax calculation
	// In production, use IRS tax tables
	annualizedIncome := grossPay * 26 // Assume bi-weekly
	
	var tax float64
	if filingStatus == "single" {
		if annualizedIncome <= 10275 {
			tax = annualizedIncome * 0.10
		} else if annualizedIncome <= 41775 {
			tax = 1027.50 + (annualizedIncome-10275)*0.12
		} else if annualizedIncome <= 89075 {
			tax = 4807.50 + (annualizedIncome-41775)*0.22
		} else {
			tax = 15213.50 + (annualizedIncome-89075)*0.24
		}
	} else { // married
		if annualizedIncome <= 20550 {
			tax = annualizedIncome * 0.10
		} else if annualizedIncome <= 83550 {
			tax = 2055 + (annualizedIncome-20550)*0.12
		} else {
			tax = 9615 + (annualizedIncome-83550)*0.22
		}
	}
	
	return tax / 26 // Convert back to per-period
}

func calculateStateTax(grossPay float64, state string) float64 {
	// Simplified state tax calculation
	// In production, use state-specific tax tables
	stateTaxRates := map[string]float64{
		"CA": 0.09, // California
		"NY": 0.06, // New York
		"TX": 0.00, // Texas (no state income tax)
		"FL": 0.00, // Florida (no state income tax)
		"PA": 0.03, // Pennsylvania
	}
	
	rate, exists := stateTaxRates[state]
	if !exists {
		rate = 0.05 // Default 5%
	}
	
	return grossPay * rate
}

func getWeekStart(t time.Time) time.Time {
	// Get Monday of the week
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
}

func roundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

func floatPtr(f float64) *float64 {
	return &f
}
