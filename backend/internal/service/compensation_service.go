package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
)


// CompensationService interface handles compensation business logic
type CompensationService interface {
	// Compensation Plans
	CreatePlan(ctx context.Context, req *models.CreateCompensationPlanRequest) (*models.CompensationPlan, error)
	GetPlan(ctx context.Context, id uuid.UUID) (*models.CompensationPlan, error)
	GetPlansByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.CompensationPlan, error)
	GetActivePlan(ctx context.Context, employeeID uuid.UUID) (*models.CompensationPlan, error)
	GetAllPlans(ctx context.Context) ([]*models.CompensationPlan, error)
	UpdatePlan(ctx context.Context, id uuid.UUID, req *models.UpdateCompensationPlanRequest) (*models.CompensationPlan, error)
	DeletePlan(ctx context.Context, id uuid.UUID) error
	CalculateTotalCompensation(ctx context.Context, employeeID uuid.UUID) (float64, error)
	
	// Bonuses
	CreateBonus(ctx context.Context, req *models.CreateBonusRequest) (*models.Bonus, error)
	GetBonus(ctx context.Context, id uuid.UUID) (*models.Bonus, error)
	GetBonusesByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.Bonus, error)
	GetAllBonuses(ctx context.Context) ([]*models.Bonus, error)
	GetBonusesByStatus(ctx context.Context, status string) ([]*models.Bonus, error)
	UpdateBonus(ctx context.Context, id uuid.UUID, req *models.UpdateBonusRequest) (*models.Bonus, error)
	ApproveBonus(ctx context.Context, id uuid.UUID, approverID uuid.UUID) (*models.Bonus, error)
	MarkBonusPaid(ctx context.Context, id uuid.UUID) (*models.Bonus, error)
	DeleteBonus(ctx context.Context, id uuid.UUID) error
	GetPendingBonuses(ctx context.Context) ([]*models.Bonus, error)
}

// compensationService implements CompensationService interface
type compensationService struct {
	repo repository.CompensationRepository
}

// NewCompensationService creates a new CompensationService
func NewCompensationService(repos *repository.Repositories) CompensationService {
	return &compensationService{repo: repos.Compensation}
}

// === COMPENSATION PLANS ===

// CreatePlan creates a new compensation plan
func (s *compensationService) CreatePlan(ctx context.Context, req *models.CreateCompensationPlanRequest) (*models.CompensationPlan, error) {
	// Validate end date if provided
	if req.EndDate != nil && req.EndDate.Before(req.EffectiveDate) {
		return nil, fmt.Errorf("end date must be after effective date")
	}

	// Check for overlapping active plans
	activePlan, err := s.repo.GetActivePlan(ctx, req.EmployeeID)
	if err == nil && activePlan != nil {
		// If there's an active plan with no end date, we need to set one
		if activePlan.EndDate == nil {
			// Automatically set end date to one day before new plan's effective date
			endDate := req.EffectiveDate.AddDate(0, 0, -1)
			activePlan.EndDate = &endDate
			if err := s.repo.UpdatePlan(ctx, activePlan); err != nil {
				return nil, fmt.Errorf("failed to update existing plan: %w", err)
			}
		}
	}

	plan := &models.CompensationPlan{
		EmployeeID:       req.EmployeeID,
		CompensationType: req.CompensationType,
		BaseAmount:       req.BaseAmount,
		Currency:         req.Currency,
		PayFrequency:     req.PayFrequency,
		EffectiveDate:    req.EffectiveDate,
		EndDate:          req.EndDate,
		Status:           req.Status,
	}

	if plan.Currency == "" {
		plan.Currency = "USD"
	}
	if plan.Status == "" {
		plan.Status = "active"
	}

	if err := s.repo.CreatePlan(ctx, plan); err != nil {
		return nil, err
	}

	return s.repo.GetPlan(ctx, plan.ID)
}

// GetPlan retrieves a compensation plan by ID
func (s *compensationService) GetPlan(ctx context.Context, id uuid.UUID) (*models.CompensationPlan, error) {
	return s.repo.GetPlan(ctx, id)
}

// GetPlansByEmployee retrieves all compensation plans for an employee
func (s *compensationService) GetPlansByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.CompensationPlan, error) {
	return s.repo.GetPlansByEmployee(ctx, employeeID)
}

// GetActivePlan retrieves the active compensation plan for an employee
func (s *compensationService) GetActivePlan(ctx context.Context, employeeID uuid.UUID) (*models.CompensationPlan, error) {
	return s.repo.GetActivePlan(ctx, employeeID)
}

// GetAllPlans retrieves all compensation plans
func (s *compensationService) GetAllPlans(ctx context.Context) ([]*models.CompensationPlan, error) {
	return s.repo.GetAllPlans(ctx)
}

// UpdatePlan updates a compensation plan
func (s *compensationService) UpdatePlan(ctx context.Context, id uuid.UUID, req *models.UpdateCompensationPlanRequest) (*models.CompensationPlan, error) {
	// Get existing plan
	plan, err := s.repo.GetPlan(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate end date if provided
	if req.EndDate != nil && req.EndDate.Before(req.EffectiveDate) {
		return nil, fmt.Errorf("end date must be after effective date")
	}

	// Update fields
	if req.CompensationType != "" {
		plan.CompensationType = req.CompensationType
	}
	if req.BaseAmount > 0 {
		plan.BaseAmount = req.BaseAmount
	}
	if req.Currency != "" {
		plan.Currency = req.Currency
	}
	if req.PayFrequency != "" {
		plan.PayFrequency = req.PayFrequency
	}
	if !req.EffectiveDate.IsZero() {
		plan.EffectiveDate = req.EffectiveDate
	}
	if req.EndDate != nil {
		plan.EndDate = req.EndDate
	}
	if req.Status != "" {
		plan.Status = req.Status
	}

	if err := s.repo.UpdatePlan(ctx, plan); err != nil {
		return nil, err
	}

	return s.repo.GetPlan(ctx, id)
}

// DeletePlan deletes a compensation plan
func (s *compensationService) DeletePlan(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePlan(ctx, id)
}

// === BONUSES ===

// CreateBonus creates a new bonus
func (s *compensationService) CreateBonus(ctx context.Context, req *models.CreateBonusRequest) (*models.Bonus, error) {
	bonus := &models.Bonus{
		EmployeeID:  req.EmployeeID,
		BonusType:   req.BonusType,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		PaymentDate: req.PaymentDate,
		Status:      req.Status,
	}

	if bonus.Currency == "" {
		bonus.Currency = "USD"
	}
	if bonus.Status == "" {
		bonus.Status = "pending"
	}

	if err := s.repo.CreateBonus(ctx, bonus); err != nil {
		return nil, err
	}

	return s.repo.GetBonus(ctx, bonus.ID)
}

// GetBonus retrieves a bonus by ID
func (s *compensationService) GetBonus(ctx context.Context, id uuid.UUID) (*models.Bonus, error) {
	return s.repo.GetBonus(ctx, id)
}

// GetBonusesByEmployee retrieves all bonuses for an employee
func (s *compensationService) GetBonusesByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.Bonus, error) {
	return s.repo.GetBonusesByEmployee(ctx, employeeID)
}

// GetAllBonuses retrieves all bonuses
func (s *compensationService) GetAllBonuses(ctx context.Context) ([]*models.Bonus, error) {
	return s.repo.GetAllBonuses(ctx)
}

// GetBonusesByStatus retrieves bonuses by status
func (s *compensationService) GetBonusesByStatus(ctx context.Context, status string) ([]*models.Bonus, error) {
	return s.repo.GetBonusesByStatus(ctx, status)
}

// UpdateBonus updates a bonus
func (s *compensationService) UpdateBonus(ctx context.Context, id uuid.UUID, req *models.UpdateBonusRequest) (*models.Bonus, error) {
	// Get existing bonus
	bonus, err := s.repo.GetBonus(ctx, id)
	if err != nil {
		return nil, err
	}

	// Don't allow updates to paid bonuses
	if bonus.Status == "paid" {
		return nil, fmt.Errorf("cannot update a paid bonus")
	}

	// Update fields
	if req.BonusType != "" {
		bonus.BonusType = req.BonusType
	}
	if req.Amount > 0 {
		bonus.Amount = req.Amount
	}
	if req.Currency != "" {
		bonus.Currency = req.Currency
	}
	if req.Description != "" {
		bonus.Description = req.Description
	}
	if !req.PaymentDate.IsZero() {
		bonus.PaymentDate = req.PaymentDate
	}
	if req.Status != "" {
		bonus.Status = req.Status
	}

	if err := s.repo.UpdateBonus(ctx, bonus); err != nil {
		return nil, err
	}

	return s.repo.GetBonus(ctx, id)
}

// ApproveBonus approves a bonus
func (s *compensationService) ApproveBonus(ctx context.Context, id uuid.UUID, approverID uuid.UUID) (*models.Bonus, error) {
	if err := s.repo.ApproveBonus(ctx, id, approverID); err != nil {
		return nil, err
	}
	return s.repo.GetBonus(ctx, id)
}

// MarkBonusPaid marks a bonus as paid
func (s *compensationService) MarkBonusPaid(ctx context.Context, id uuid.UUID) (*models.Bonus, error) {
	if err := s.repo.MarkBonusPaid(ctx, id); err != nil {
		return nil, err
	}
	return s.repo.GetBonus(ctx, id)
}

// DeleteBonus deletes a bonus
func (s *compensationService) DeleteBonus(ctx context.Context, id uuid.UUID) error {
	bonus, err := s.repo.GetBonus(ctx, id)
	if err != nil {
		return err
	}

	// Don't allow deletion of paid bonuses
	if bonus.Status == "paid" {
		return fmt.Errorf("cannot delete a paid bonus")
	}

	return s.repo.DeleteBonus(ctx, id)
}

// === ANALYTICS & REPORTING ===

// CalculateTotalCompensation calculates total annual compensation for an employee
func (s *compensationService) CalculateTotalCompensation(ctx context.Context, employeeID uuid.UUID) (float64, error) {
	// Get active compensation plan
	plan, err := s.repo.GetActivePlan(ctx, employeeID)
	if err != nil {
		return 0, err
	}

	// Convert to annual amount based on pay frequency
	var annualBase float64
	switch plan.PayFrequency {
	case "hourly":
		// Assume 2080 hours per year (40 hours/week * 52 weeks)
		annualBase = plan.BaseAmount * 2080
	case "weekly":
		annualBase = plan.BaseAmount * 52
	case "biweekly":
		annualBase = plan.BaseAmount * 26
	case "monthly":
		annualBase = plan.BaseAmount * 12
	case "annually":
		annualBase = plan.BaseAmount
	default:
		annualBase = plan.BaseAmount
	}

	return annualBase, nil
}

// GetPendingBonuses retrieves all pending bonuses for approval
func (s *compensationService) GetPendingBonuses(ctx context.Context) ([]*models.Bonus, error) {
	return s.repo.GetBonusesByStatus(ctx, "pending")
}