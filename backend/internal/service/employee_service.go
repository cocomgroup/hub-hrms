package service

import (
	"context"

	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
)

// EmployeeService handles employee operations
type EmployeeService interface {
	Create(ctx context.Context, employee *models.Employee) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*models.Employee, error)
	Update(ctx context.Context, employee *models.Employee) error
}

type employeeService struct {
	repos *repository.Repositories
}

func NewEmployeeService(repos *repository.Repositories) EmployeeService {
	return &employeeService{repos: repos}
}

func (s *employeeService) Create(ctx context.Context, employee *models.Employee) error {
	return s.repos.Employee.Create(ctx, employee)
}

func (s *employeeService) GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	return s.repos.Employee.GetByID(ctx, id)
}

func (s *employeeService) List(ctx context.Context, filters map[string]interface{}) ([]*models.Employee, error) {
	return s.repos.Employee.List(ctx, filters)
}

func (s *employeeService) Update(ctx context.Context, employee *models.Employee) error {
	return s.repos.Employee.Update(ctx, employee)
}
