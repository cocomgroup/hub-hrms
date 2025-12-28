package service

import (
	"context"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
	"time"

	"github.com/google/uuid"
)

// OnboardingService handles onboarding operations
type OnboardingService interface {
	CreateTask(ctx context.Context, task *models.OnboardingTask) error
	GetTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.OnboardingTask, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error)
	UpdateTask(ctx context.Context, task *models.OnboardingTask) error
	CreateOnboardingPlan(ctx context.Context, employeeID uuid.UUID, department string) error
}

type onboardingService struct {
	repos *repository.Repositories
}

func NewOnboardingService(repos *repository.Repositories) OnboardingService {
	return &onboardingService{repos: repos}
}

func (s *onboardingService) CreateTask(ctx context.Context, task *models.OnboardingTask) error {
	return s.repos.Onboarding.CreateTask(ctx, task)
}

func (s *onboardingService) GetTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.OnboardingTask, error) {
	return s.repos.Onboarding.GetTasksByEmployee(ctx, employeeID)
}

func (s *onboardingService) GetTaskByID(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error) {
	return s.repos.Onboarding.GetTaskByID(ctx, id)
}

func (s *onboardingService) UpdateTask(ctx context.Context, task *models.OnboardingTask) error {
	if task.Status == "completed" && task.CompletedAt == nil {
		now := time.Now()
		task.CompletedAt = &now
	}
	return s.repos.Onboarding.UpdateTask(ctx, task)
}

func (s *onboardingService) CreateOnboardingPlan(ctx context.Context, employeeID uuid.UUID, department string) error {
	// Default onboarding tasks
	tasks := []models.OnboardingTask{
		{
			EmployeeID:        employeeID,
			TaskName:          "Complete I-9 Form",
			Description:       strPtr("Complete employment eligibility verification"),
			Category:          strPtr("HR Documents"),
			Status:            "pending",
			DueDate:           timePtr(time.Now().AddDate(0, 0, 3)),
			DocumentsRequired: true,
		},
		{
			EmployeeID:  employeeID,
			TaskName:    "Setup Direct Deposit",
			Description: strPtr("Provide bank account information for payroll"),
			Category:    strPtr("Payroll"),
			Status:      "pending",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 7)),
		},
		{
			EmployeeID:  employeeID,
			TaskName:    "Complete Benefits Enrollment",
			Description: strPtr("Select health insurance and other benefits"),
			Category:    strPtr("Benefits"),
			Status:      "pending",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 30)),
		},
		{
			EmployeeID:  employeeID,
			TaskName:    "IT Account Setup",
			Description: strPtr("Receive email, system access credentials"),
			Category:    strPtr("IT"),
			Status:      "pending",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 1)),
		},
		{
			EmployeeID:  employeeID,
			TaskName:    "Review Employee Handbook",
			Description: strPtr("Read and acknowledge company policies"),
			Category:    strPtr("HR Documents"),
			Status:      "pending",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 7)),
		},
	}

	for _, task := range tasks {
		if err := s.repos.Onboarding.CreateTask(ctx, &task); err != nil {
			return err
		}
	}

	return nil
}

