package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"hub-hrms/backend/internal/models"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, search string, role string) ([]*models.User, error) {
	args := m.Called(ctx, search, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockEmployeeRepository is a mock implementation of EmployeeRepository
type MockEmployeeRepository struct {
	mock.Mock
}

func (m *MockEmployeeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) GetByEmail(ctx context.Context, email string) (*models.Employee, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) Create(ctx context.Context, employee *models.Employee) error {
	args := m.Called(ctx, employee)
	return args.Error(0)
}

func (m *MockEmployeeRepository) Update(ctx context.Context, employee *models.Employee) error {
	args := m.Called(ctx, employee)
	return args.Error(0)
}

func (m *MockEmployeeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEmployeeRepository) List(ctx context.Context, search, department, status string) ([]*models.Employee, error) {
	args := m.Called(ctx, search, department, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) GetAllEmployees(ctx context.Context) ([]*models.Employee, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) GetEmployeesByManager(ctx context.Context, managerID uuid.UUID) ([]*models.Employee, error) {
	args := m.Called(ctx, managerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Employee), args.Error(1)
}

// MockApplicantRepository is a mock implementation of ApplicantRepository
type MockApplicantRepository struct {
	mock.Mock
}

func (m *MockApplicantRepository) Create(ctx context.Context, applicant *models.Applicant) error {
	args := m.Called(ctx, applicant)
	return args.Error(0)
}

func (m *MockApplicantRepository) GetAll(ctx context.Context) ([]*models.Applicant, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Applicant), args.Error(1)
}

func (m *MockApplicantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Applicant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Applicant), args.Error(1)
}

func (m *MockApplicantRepository) Update(ctx context.Context, applicant *models.Applicant) error {
	args := m.Called(ctx, applicant)
	return args.Error(0)
}

func (m *MockApplicantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Add more mock repositories as needed for other entities
