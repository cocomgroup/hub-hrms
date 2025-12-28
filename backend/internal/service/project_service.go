package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
)

var (
	ErrProjectNotFound   = errors.New("project not found")
	ErrDuplicateMember   = errors.New("employee already assigned to project")
	ErrInvalidManager    = errors.New("invalid manager")
)

type ProjectService interface {
	CreateProject(ctx context.Context, req *models.CreateProjectRequest, createdBy uuid.UUID) (*models.Project, error)
	GetProject(ctx context.Context, id uuid.UUID) (*models.ProjectWithDetails, error)
	ListProjects(ctx context.Context, status string, managerID *uuid.UUID) ([]*models.ProjectWithDetails, error)
	UpdateProject(ctx context.Context, id uuid.UUID, req *models.UpdateProjectRequest) (*models.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
	
	// Project members
	AssignMember(ctx context.Context, projectID uuid.UUID, req *models.AssignProjectMemberRequest) error
	RemoveMember(ctx context.Context, projectID, employeeID uuid.UUID) error
	GetProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMemberInfo, error)
	GetEmployeeProjects(ctx context.Context, employeeID uuid.UUID) ([]*models.Project, error)
	
	// Manager assignment
	AssignEmployeeToManager(ctx context.Context, req *models.AssignManagerRequest) error
}

type projectService struct {
	repo         repository.ProjectRepository
	employeeRepo repository.EmployeeRepository
}

func NewProjectService(repos *repository.Repositories) ProjectService {
	return &projectService{
		repo:         repos.Project,
		employeeRepo: repos.Employee,
	}
}

func (s *projectService) CreateProject(ctx context.Context, req *models.CreateProjectRequest, createdBy uuid.UUID) (*models.Project, error) {
	// Validate manager if provided
	if req.ManagerID != nil {
		_, err := s.employeeRepo.GetByID(ctx, *req.ManagerID)
		if err != nil {
			return nil, ErrInvalidManager
		}
	}

	// Set defaults
	if req.Status == "" {
		req.Status = "active"
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}

	project := &models.Project{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		ManagerID:   req.ManagerID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Budget:      req.Budget,
		CreatedBy:   createdBy,
	}

	err := s.repo.Create(ctx, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) GetProject(ctx context.Context, id uuid.UUID) (*models.ProjectWithDetails, error) {
	return s.repo.GetWithDetails(ctx, id)
}

func (s *projectService) ListProjects(ctx context.Context, status string, managerID *uuid.UUID) ([]*models.ProjectWithDetails, error) {
	return s.repo.ListWithDetails(ctx, status, managerID)
}

func (s *projectService) UpdateProject(ctx context.Context, id uuid.UUID, req *models.UpdateProjectRequest) (*models.Project, error) {
	// Get existing project
	project, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	// Validate manager if provided
	if req.ManagerID != nil {
		_, err := s.employeeRepo.GetByID(ctx, *req.ManagerID)
		if err != nil {
			return nil, ErrInvalidManager
		}
	}

	// Update fields
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.Status != nil {
		project.Status = *req.Status
	}
	if req.Priority != nil {
		project.Priority = *req.Priority
	}
	if req.ManagerID != nil {
		project.ManagerID = req.ManagerID
	}
	if req.StartDate != nil {
		project.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		project.EndDate = req.EndDate
	}
	if req.Budget != nil {
		project.Budget = req.Budget
	}

	err = s.repo.Update(ctx, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *projectService) AssignMember(ctx context.Context, projectID uuid.UUID, req *models.AssignProjectMemberRequest) error {
	// Verify project exists
	_, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return ErrProjectNotFound
	}

	// Verify employee exists
	_, err = s.employeeRepo.GetByID(ctx, req.EmployeeID)
	if err != nil {
		return errors.New("employee not found")
	}

	// Set default role
	if req.Role == "" {
		req.Role = "member"
	}

	member := &models.ProjectMember{
		ProjectID:  projectID,
		EmployeeID: req.EmployeeID,
		Role:       req.Role,
	}

	return s.repo.AddMember(ctx, member)
}

func (s *projectService) RemoveMember(ctx context.Context, projectID, employeeID uuid.UUID) error {
	return s.repo.RemoveMember(ctx, projectID, employeeID)
}

func (s *projectService) GetProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMemberInfo, error) {
	return s.repo.GetMembers(ctx, projectID)
}

func (s *projectService) GetEmployeeProjects(ctx context.Context, employeeID uuid.UUID) ([]*models.Project, error) {
	return s.repo.GetEmployeeProjects(ctx, employeeID)
}

func (s *projectService) AssignEmployeeToManager(ctx context.Context, req *models.AssignManagerRequest) error {
	// Verify both employee and manager exist
	employee, err := s.employeeRepo.GetByID(ctx, req.EmployeeID)
	if err != nil {
		return errors.New("employee not found")
	}

	_, err = s.employeeRepo.GetByID(ctx, req.ManagerID)
	if err != nil {
		return errors.New("manager not found")
	}

	// Update employee's manager
	employee.ManagerID = &req.ManagerID
	return s.employeeRepo.Update(ctx, employee)
}