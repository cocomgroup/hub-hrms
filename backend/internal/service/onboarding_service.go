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
	// Workflows
	CreateWorkflow(ctx context.Context, req *models.CreateWorkflowRequest, createdBy uuid.UUID) (*models.OnboardingWorkflow, error)
	GetWorkflow(ctx context.Context, id uuid.UUID) (*models.OnboardingWorkflow, error)
	GetWorkflowByEmployee(ctx context.Context, employeeID uuid.UUID) (*models.OnboardingWorkflow, error)
	ListWorkflows(ctx context.Context, filters map[string]interface{}) ([]*models.OnboardingWorkflow, error)
	UpdateWorkflow(ctx context.Context, id uuid.UUID, req *models.UpdateWorkflowRequest) (*models.OnboardingWorkflow, error)
	DeleteWorkflow(ctx context.Context, id uuid.UUID) error 

	// Tasks
	CreateTask(ctx context.Context, req *models.CreateTaskRequest) (*models.OnboardingTask, error)
	GetTask(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error)  // Add this
	ListTasksByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*models.OnboardingTask, error)  // Add this
	UpdateTask(ctx context.Context, id uuid.UUID, req *models.UpdateTaskRequest) (*models.OnboardingTask, error)
	CompleteTask(ctx context.Context, taskID, completedBy uuid.UUID, req *models.CompleteTaskRequest) error
	DeleteTask(ctx context.Context, id uuid.UUID) error

	// AI Interactions
	HandleAIInteraction(ctx context.Context, req *models.AIInteractionRequest) (*models.AIInteractionResponse, error)
	ListInteractionsByWorkflow(ctx context.Context, workflowID uuid.UUID, limit int) ([]*models.OnboardingInteraction, error)  // Add this
	
	// Milestones
	CreateMilestone(ctx context.Context, req *models.CreateMilestoneRequest) (*models.OnboardingMilestone, error)  // Add this
	ListMilestonesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*models.OnboardingMilestone, error)  // Add this
	CompleteMilestone(ctx context.Context, id uuid.UUID) error  

	// Templates
	GetTemplate(ctx context.Context, id uuid.UUID) (*models.OnboardingChecklistTemplate, error)  // Add this
	ListTemplates(ctx context.Context, department, roleType string) ([]*models.OnboardingChecklistTemplate, error)
	
	// Statistics
	GetWorkflowStatistics(ctx context.Context, workflowID uuid.UUID) (*models.OnboardingStatistics, error)  // Add this
	GetDashboard(ctx context.Context, filters map[string]interface{}) (*models.OnboardingDashboardResponse, error)

	// Additional helper methods
	GetTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.OnboardingTask, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error)
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
func (s *onboardingService) DeleteWorkflow(ctx context.Context, id uuid.UUID) error {
	return s.repos.Onboarding.DeleteWorkflow(ctx, id)
}

func (s *onboardingService) GetTask(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error) {
	return s.repo.GetTask(ctx, id)
}

func (s *onboardingService) ListTasksByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*models.OnboardingTask, error) {
	return s.repo.ListTasksByWorkflow(ctx, workflowID)
}

func (s *onboardingService) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteTask(ctx, id)
}

func (s *onboardingService) ListInteractionsByWorkflow(ctx context.Context, workflowID uuid.UUID, limit int) ([]*models.OnboardingInteraction, error) {
	return s.repo.ListInteractionsByWorkflow(ctx, workflowID, limit)
}

func (s *onboardingService) CreateMilestone(ctx context.Context, req *models.CreateMilestoneRequest) (*models.OnboardingMilestone, error) {
	milestone := &models.OnboardingMilestone{
		ID:          uuid.New(),
		WorkflowID:  req.WorkflowID,
		Name:        req.Name,
		Description: req.Description,
		TargetDate:  req.TargetDate,
		Status:      "pending",
	}
	
	err := s.repo.CreateMilestone(ctx, milestone)
	if err != nil {
		return nil, err
	}
	
	return milestone, nil
}

func (s *onboardingService) ListMilestonesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*models.OnboardingMilestone, error) {
	return s.repo.ListMilestonesByWorkflow(ctx, workflowID)
}

func (s *onboardingService) CompleteMilestone(ctx context.Context, id uuid.UUID) error {
	milestones, err := s.repo.GetMilestone(ctx, id)
	if err != nil {
		return err
	}
	
	now := time.Now()
	milestones.CompletedDate = &now
	milestones.Status = "completed"
	
	return s.repo.UpdateMilestone(ctx, milestones)
}

func (s *onboardingService) GetTemplate(ctx context.Context, id uuid.UUID) (*models.OnboardingChecklistTemplate, error) {
	return s.repo.GetTemplate(ctx, id)
}

func (s *onboardingService) GetWorkflowStatistics(ctx context.Context, workflowID uuid.UUID) (*models.OnboardingStatistics, error) {
	return s.repo.GetWorkflowStatistics(ctx, workflowID)
}
