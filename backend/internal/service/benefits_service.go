package service

import (
	"context"
	"fmt"
	"time"

	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"

	"github.com/google/uuid"
)

// BenefitsService interface defines service operations
type BenefitsService interface {
	GetAllBenefitPlans(ctx context.Context, activeOnly bool) ([]models.BenefitPlan, error)
	GetBenefitPlanByID(ctx context.Context, id uuid.UUID) (*models.BenefitPlan, error)
	CreateBenefitPlan(ctx context.Context, input *models.CreateBenefitPlanInput) (*models.BenefitPlan, error)
	UpdateBenefitPlan(ctx context.Context, id uuid.UUID, input *models.UpdateBenefitPlanInput) (*models.BenefitPlan, error)
	GetEmployeeEnrollments(ctx context.Context, employeeID uuid.UUID) ([]models.BenefitEnrollment, error)
	CreateEnrollment(ctx context.Context, employeeID uuid.UUID, input *models.CreateEnrollmentInput) (*models.BenefitEnrollment, error)
	CancelEnrollment(ctx context.Context, id uuid.UUID, employeeID uuid.UUID) error
	GetAllEnrollments(ctx context.Context, status *models.EnrollmentStatus) ([]models.BenefitEnrollment, error)
	GetBenefitsSummary(ctx context.Context, employeeID uuid.UUID) (*models.BenefitsSummary, error)
}

// BenefitsService handles benefits operations
type benefitsService struct {
	repo repository.BenefitsRepository
}

func NewBenefitsService(repos *repository.Repositories) BenefitsService {
	return &benefitsService{repo: repos.Benefits}
}

// GetAllBenefitPlans retrieves all benefit plans
func (s *benefitsService) GetAllBenefitPlans(ctx context.Context, activeOnly bool) ([]models.BenefitPlan, error) {
	return s.repo.GetAllBenefitPlans(ctx, activeOnly)
}

// GetBenefitPlanByID retrieves a specific benefit plan
func (s *benefitsService) GetBenefitPlanByID(ctx context.Context, id uuid.UUID) (*models.BenefitPlan, error) {
	return s.repo.GetBenefitPlanByID(ctx, id)
}

// CreateBenefitPlan creates a new benefit plan
func (s *benefitsService) CreateBenefitPlan(ctx context.Context, input *models.CreateBenefitPlanInput) (*models.BenefitPlan, error) {
	// Parse dates
	enrollmentStart, err := time.Parse("2006-01-02", input.EnrollmentStartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid enrollment start date")
	}

	enrollmentEnd, err := time.Parse("2006-01-02", input.EnrollmentEndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid enrollment end date")
	}

	effectiveDate, err := time.Parse("2006-01-02", input.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date")
	}

	// Validate dates
	if enrollmentEnd.Before(enrollmentStart) {
		return nil, fmt.Errorf("enrollment end date must be after start date")
	}

	plan := &models.BenefitPlan{
		Name:                 input.Name,
		Category:             input.Category,
		PlanType:             input.PlanType,
		Provider:             input.Provider,
		Description:          input.Description,
		EmployeeCost:         input.EmployeeCost,
		EmployerCost:         input.EmployerCost,
		DeductibleSingle:     input.DeductibleSingle,
		DeductibleFamily:     input.DeductibleFamily,
		OutOfPocketMaxSingle: input.OutOfPocketMaxSingle,
		OutOfPocketMaxFamily: input.OutOfPocketMaxFamily,
		CopayPrimaryCare:     input.CopayPrimaryCare,
		CopaySpecialist:      input.CopaySpecialist,
		CopayEmergency:       input.CopayEmergency,
		CoinsuranceRate:      input.CoinsuranceRate,
		Active:               true,
		EnrollmentStartDate:  enrollmentStart,
		EnrollmentEndDate:    enrollmentEnd,
		EffectiveDate:        effectiveDate,
	}

	err = s.repo.CreateBenefitPlan(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to create benefit plan: %w", err)
	}

	return plan, nil
}

// UpdateBenefitPlan updates a benefit plan
func (s *benefitsService) UpdateBenefitPlan(ctx context.Context, id uuid.UUID, input *models.UpdateBenefitPlanInput) (*models.BenefitPlan, error) {
	// Get existing plan
	plan, err := s.repo.GetBenefitPlanByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if input.Name != "" {
		plan.Name = input.Name
	}
	if input.Category != "" {
		plan.Category = input.Category
	}
	if input.PlanType != "" {
		plan.PlanType = input.PlanType
	}
	if input.Provider != "" {
		plan.Provider = input.Provider
	}
	if input.Description != "" {
		plan.Description = input.Description
	}
	if input.EmployeeCost != nil {
		plan.EmployeeCost = *input.EmployeeCost
	}
	if input.EmployerCost != nil {
		plan.EmployerCost = *input.EmployerCost
	}
	if input.DeductibleSingle != nil {
		plan.DeductibleSingle = *input.DeductibleSingle
	}
	if input.DeductibleFamily != nil {
		plan.DeductibleFamily = *input.DeductibleFamily
	}
	if input.OutOfPocketMaxSingle != nil {
		plan.OutOfPocketMaxSingle = *input.OutOfPocketMaxSingle
	}
	if input.OutOfPocketMaxFamily != nil {
		plan.OutOfPocketMaxFamily = *input.OutOfPocketMaxFamily
	}
	if input.CopayPrimaryCare != nil {
		plan.CopayPrimaryCare = *input.CopayPrimaryCare
	}
	if input.CopaySpecialist != nil {
		plan.CopaySpecialist = *input.CopaySpecialist
	}
	if input.CopayEmergency != nil {
		plan.CopayEmergency = *input.CopayEmergency
	}
	if input.CoinsuranceRate != nil {
		plan.CoinsuranceRate = *input.CoinsuranceRate
	}
	if input.Active != nil {
		plan.Active = *input.Active
	}

	err = s.repo.UpdateBenefitPlan(ctx, id, plan)
	if err != nil {
		return nil, err
	}

	return s.repo.GetBenefitPlanByID(ctx, id)
}

// GetEmployeeEnrollments retrieves enrollments for an employee
func (s *benefitsService) GetEmployeeEnrollments(ctx context.Context, employeeID uuid.UUID) ([]models.BenefitEnrollment, error) {
	return s.repo.GetEmployeeEnrollments(ctx, employeeID)
}

// CreateEnrollment creates a new benefit enrollment
func (s *benefitsService) CreateEnrollment(ctx context.Context, employeeID uuid.UUID, input *models.CreateEnrollmentInput) (*models.BenefitEnrollment, error) {
	// Get the plan
	plan, err := s.repo.GetBenefitPlanByID(ctx, input.PlanID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	if !plan.Active {
		return nil, fmt.Errorf("plan is not active")
	}

	// Parse effective date
	effectiveDate, err := time.Parse("2006-01-02", input.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date")
	}

	// Validate effective date is in the future
	if effectiveDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, fmt.Errorf("effective date must be in the future")
	}

	// Check if enrollment period is open
	now := time.Now()
	if now.Before(plan.EnrollmentStartDate) || now.After(plan.EnrollmentEndDate) {
		return nil, fmt.Errorf("enrollment period is not open for this plan")
	}

	// Calculate costs based on coverage level
	employeeCost := plan.EmployeeCost
	employerCost := plan.EmployerCost

	// Adjust costs based on coverage level
	switch input.CoverageLevel {
	case models.CoverageLevelEmployeeSpouse:
		employeeCost *= 1.5
		employerCost *= 1.5
	case models.CoverageLevelEmployeeChild:
		employeeCost *= 1.3
		employerCost *= 1.3
	case models.CoverageLevelFamily:
		employeeCost *= 2.0
		employerCost *= 2.0
	}

	totalCost := employeeCost + employerCost
	payrollDeduction := employeeCost / 12 // Monthly deduction

	// Parse dependents
	var dependents []models.Dependent
	for _, depInput := range input.Dependents {
		dob, err := time.Parse("2006-01-02", depInput.DateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("invalid dependent date of birth")
		}

		dependent := models.Dependent{
			FirstName:    depInput.FirstName,
			LastName:     depInput.LastName,
			Relationship: depInput.Relationship,
			DateOfBirth:  dob,
			SSN:          depInput.SSN,
		}
		dependents = append(dependents, dependent)
	}

	enrollment := &models.BenefitEnrollment{
		EmployeeID:       employeeID,
		PlanID:           input.PlanID,
		CoverageLevel:    input.CoverageLevel,
		EffectiveDate:    effectiveDate,
		EmployeeCost:     employeeCost,
		EmployerCost:     employerCost,
		TotalCost:        totalCost,
		PayrollDeduction: payrollDeduction,
		Dependents:       dependents,
	}

	err = s.repo.CreateEnrollment(ctx, enrollment)
	if err != nil {
		return nil, fmt.Errorf("failed to create enrollment: %w", err)
	}

	return s.repo.GetEnrollmentByID(ctx, enrollment.ID)
}

// CancelEnrollment cancels a benefit enrollment
func (s *benefitsService) CancelEnrollment(ctx context.Context, id uuid.UUID, employeeID uuid.UUID) error {
	// Get the enrollment
	enrollment, err := s.repo.GetEnrollmentByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify ownership
	if enrollment.EmployeeID != employeeID {
		return fmt.Errorf("unauthorized: cannot cancel another employee's enrollment")
	}

	// Can only cancel active enrollments
	if enrollment.Status != models.EnrollmentStatusActive {
		return fmt.Errorf("can only cancel active enrollments")
	}

	// Set termination date to end of current month
	now := time.Now()
	terminationDate := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())

	return s.repo.CancelEnrollment(ctx, id, terminationDate)
}

// GetAllEnrollments retrieves all enrollments (admin)
func (s *benefitsService) GetAllEnrollments(ctx context.Context, status *models.EnrollmentStatus) ([]models.BenefitEnrollment, error) {
	return s.repo.GetAllEnrollments(ctx, status)
}

// GetBenefitsSummary gets a summary of employee benefits
func (s *benefitsService) GetBenefitsSummary(ctx context.Context, employeeID uuid.UUID) (*models.BenefitsSummary, error) {
	// Get employee enrollments
	enrollments, err := s.repo.GetEmployeeEnrollments(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	// Get available plans
	availablePlans, err := s.repo.GetAllBenefitPlans(ctx, true)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	var totalEmployeeCost float64
	var totalEmployerCost float64
	var monthlyDeduction float64
	activeCount := 0

	for _, enrollment := range enrollments {
		if enrollment.Status == models.EnrollmentStatusActive {
			activeCount++
			totalEmployeeCost += enrollment.EmployeeCost
			totalEmployerCost += enrollment.EmployerCost
			monthlyDeduction += enrollment.PayrollDeduction
		}
	}

	summary := &models.BenefitsSummary{
		EmployeeID:        employeeID,
		ActiveEnrollments: activeCount,
		TotalEmployeeCost: totalEmployeeCost,
		TotalEmployerCost: totalEmployerCost,
		MonthlyDeduction:  monthlyDeduction,
		Enrollments:       enrollments,
		AvailablePlans:    availablePlans,
	}

	return summary, nil
}
