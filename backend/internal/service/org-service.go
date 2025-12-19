package service

import (
	"context"
	"errors"
	"fmt"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrOrganizationExists   = errors.New("organization with this code already exists")
	ErrCircularReference    = errors.New("circular reference detected in organization hierarchy")
	ErrCannotDeleteOrg      = errors.New("cannot delete organization with active employees or children")
)

// OrganizationService handles organization business logic
type OrganizationService interface {
	CreateOrganization(ctx context.Context, req *models.CreateOrganizationRequest, createdBy uuid.UUID) (*models.Organization, error)
	GetOrganization(ctx context.Context, id uuid.UUID) (*models.OrganizationWithDetails, error)
	ListOrganizations(ctx context.Context, filters map[string]interface{}) ([]*models.OrganizationWithDetails, error)
	UpdateOrganization(ctx context.Context, id uuid.UUID, req *models.UpdateOrganizationRequest) (*models.Organization, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
	GetHierarchy(ctx context.Context, rootID *uuid.UUID) (*models.OrganizationHierarchy, error)
	
	// Employee management
	AssignEmployee(ctx context.Context, orgID uuid.UUID, req *models.AssignEmployeeRequest, assignedBy uuid.UUID) error
	UnassignEmployee(ctx context.Context, orgID, employeeID uuid.UUID) error
	GetOrganizationEmployees(ctx context.Context, orgID uuid.UUID) ([]*models.Employee, error)
	GetEmployeeOrganizations(ctx context.Context, employeeID uuid.UUID) ([]*models.Organization, error)
	BulkAssignEmployees(ctx context.Context, orgID uuid.UUID, req *models.BulkAssignEmployeesRequest, assignedBy uuid.UUID) error
	
	// Stats and reporting
	GetOrganizationStats(ctx context.Context, orgID uuid.UUID) (*models.OrganizationStats, error)
}

type organizationService struct {
	repos *repository.Repositories
}

func NewOrganizationService(repos *repository.Repositories) OrganizationService {
	return &organizationService{repos: repos}
}

func (s *organizationService) CreateOrganization(ctx context.Context, req *models.CreateOrganizationRequest, createdBy uuid.UUID) (*models.Organization, error) {
	// Check if code already exists
	existing, err := s.repos.Organization.GetByCode(ctx, req.Code)
	if err == nil && existing != nil {
		return nil, ErrOrganizationExists
	}
	
	// Calculate level based on parent
	level := 0
	if req.ParentID != nil {
		parent, err := s.repos.Organization.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent organization not found: %w", err)
		}
		level = parent.Level + 1
	}
	
	// Create organization
	org := &models.Organization{
		ID:            uuid.New(),
		Name:          req.Name,
		Code:          req.Code,
		Description:   req.Description,
		ParentID:      req.ParentID,
		ManagerID:     req.ManagerID,
		Type:          req.Type,
		Level:         level,
		CostCenter:    req.CostCenter,
		Location:      req.Location,
		IsActive:      true,
		EmployeeCount: 0,
		CreatedBy:     createdBy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	err = s.repos.Organization.Create(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}
	
	return org, nil
}

func (s *organizationService) GetOrganization(ctx context.Context, id uuid.UUID) (*models.OrganizationWithDetails, error) {
	org, err := s.repos.Organization.GetByID(ctx, id)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}
	return org, nil
}

func (s *organizationService) ListOrganizations(ctx context.Context, filters map[string]interface{}) ([]*models.OrganizationWithDetails, error) {
	return s.repos.Organization.List(ctx, filters)
}

func (s *organizationService) UpdateOrganization(ctx context.Context, id uuid.UUID, req *models.UpdateOrganizationRequest) (*models.Organization, error) {
	// Get existing organization
	existing, err := s.repos.Organization.GetByID(ctx, id)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}
	
	// Check for circular reference if parent is being changed
	if req.ParentID != nil && *req.ParentID != uuid.Nil {
		if err := s.checkCircularReference(ctx, id, *req.ParentID); err != nil {
			return nil, err
		}
	}
	
	// Update fields
	org := &models.Organization{
		ID:            id,
		Name:          existing.Name,
		Code:          existing.Code,
		Description:   existing.Description,
		ParentID:      existing.ParentID,
		ManagerID:     existing.ManagerID,
		Type:          existing.Type,
		Level:         existing.Level,
		CostCenter:    existing.CostCenter,
		Location:      existing.Location,
		IsActive:      existing.IsActive,
		EmployeeCount: existing.EmployeeCount,
		CreatedBy:     existing.CreatedBy,
		CreatedAt:     existing.CreatedAt,
		UpdatedAt:     time.Now(),
	}
	
	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.Description != nil {
		org.Description = *req.Description
	}
	if req.ParentID != nil {
		org.ParentID = req.ParentID
		// Recalculate level
		if *req.ParentID != uuid.Nil {
			parent, err := s.repos.Organization.GetByID(ctx, *req.ParentID)
			if err != nil {
				return nil, fmt.Errorf("parent organization not found: %w", err)
			}
			org.Level = parent.Level + 1
		} else {
			org.Level = 0
		}
	}
	if req.ManagerID != nil {
		org.ManagerID = req.ManagerID
	}
	if req.Type != nil {
		org.Type = *req.Type
	}
	if req.CostCenter != nil {
		org.CostCenter = *req.CostCenter
	}
	if req.Location != nil {
		org.Location = *req.Location
	}
	if req.IsActive != nil {
		org.IsActive = *req.IsActive
	}
	
	err = s.repos.Organization.Update(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}
	
	return org, nil
}

func (s *organizationService) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	// Check if organization has children
	children, err := s.repos.Organization.GetChildren(ctx, id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return ErrCannotDeleteOrg
	}
	
	// Check if organization has active employees
	employees, err := s.repos.Organization.GetEmployees(ctx, id)
	if err != nil {
		return err
	}
	if len(employees) > 0 {
		return ErrCannotDeleteOrg
	}
	
	return s.repos.Organization.Delete(ctx, id)
}

func (s *organizationService) GetHierarchy(ctx context.Context, rootID *uuid.UUID) (*models.OrganizationHierarchy, error) {
	return s.repos.Organization.GetHierarchy(ctx, rootID)
}

func (s *organizationService) AssignEmployee(ctx context.Context, orgID uuid.UUID, req *models.AssignEmployeeRequest, assignedBy uuid.UUID) error {
	// Verify organization exists
	_, err := s.repos.Organization.GetByID(ctx, orgID)
	if err != nil {
		return ErrOrganizationNotFound
	}
	
	// Verify employee exists
	_, err = s.repos.Employee.GetByID(ctx, req.EmployeeID)
	if err != nil {
		return ErrEmployeeNotFound
	}
	
	assignment := &models.OrganizationEmployee{
		ID:             uuid.New(),
		OrganizationID: orgID,
		EmployeeID:     req.EmployeeID,
		Role:           req.Role,
		IsPrimary:      req.IsPrimary,
		StartDate:      req.StartDate,
		AssignedBy:     assignedBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	return s.repos.Organization.AssignEmployee(ctx, assignment)
}

func (s *organizationService) UnassignEmployee(ctx context.Context, orgID, employeeID uuid.UUID) error {
	return s.repos.Organization.UnassignEmployee(ctx, orgID, employeeID)
}

func (s *organizationService) GetOrganizationEmployees(ctx context.Context, orgID uuid.UUID) ([]*models.Employee, error) {
	return s.repos.Organization.GetEmployees(ctx, orgID)
}

func (s *organizationService) GetEmployeeOrganizations(ctx context.Context, employeeID uuid.UUID) ([]*models.Organization, error) {
	return s.repos.Organization.GetEmployeeOrganizations(ctx, employeeID)
}

func (s *organizationService) BulkAssignEmployees(ctx context.Context, orgID uuid.UUID, req *models.BulkAssignEmployeesRequest, assignedBy uuid.UUID) error {
	// Verify organization exists
	_, err := s.repos.Organization.GetByID(ctx, orgID)
	if err != nil {
		return ErrOrganizationNotFound
	}
	
	// Verify all employees exist
	for _, empID := range req.EmployeeIDs {
		_, err := s.repos.Employee.GetByID(ctx, empID)
		if err != nil {
			return fmt.Errorf("employee %s not found: %w", empID, ErrEmployeeNotFound)
		}
	}
	
	return s.repos.Organization.BulkAssignEmployees(ctx, orgID, req.EmployeeIDs, req)
}

func (s *organizationService) GetOrganizationStats(ctx context.Context, orgID uuid.UUID) (*models.OrganizationStats, error) {
	return s.repos.Organization.GetStats(ctx, orgID)
}

// checkCircularReference ensures no circular references in org hierarchy
func (s *organizationService) checkCircularReference(ctx context.Context, orgID, newParentID uuid.UUID) error {
	if orgID == newParentID {
		return ErrCircularReference
	}
	
	// Walk up the parent chain to check for circular reference
	currentID := newParentID
	visited := make(map[uuid.UUID]bool)
	
	for currentID != uuid.Nil {
		if visited[currentID] {
			return ErrCircularReference
		}
		if currentID == orgID {
			return ErrCircularReference
		}
		
		visited[currentID] = true
		
		parent, err := s.repos.Organization.GetByID(ctx, currentID)
		if err != nil {
			return err
		}
		
		if parent.ParentID == nil {
			break
		}
		currentID = *parent.ParentID
	}
	
	return nil
}
