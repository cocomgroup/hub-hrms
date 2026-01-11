package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
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
	GetTask(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error)
	ListTasksByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*models.OnboardingTask, error)
	UpdateTask(ctx context.Context, id uuid.UUID, req *models.UpdateTaskRequest) (*models.OnboardingTask, error)
	CompleteTask(ctx context.Context, taskID, completedBy uuid.UUID, req *models.CompleteTaskRequest) error
	DeleteTask(ctx context.Context, id uuid.UUID) error

	// AI Interactions
	HandleAIInteraction(ctx context.Context, req *models.AIInteractionRequest) (*models.AIInteractionResponse, error)
	ListInteractionsByWorkflow(ctx context.Context, workflowID uuid.UUID, limit int) ([]*models.OnboardingInteraction, error)

	// Milestones
	CreateMilestone(ctx context.Context, req *models.CreateMilestoneRequest) (*models.OnboardingMilestone, error)
	ListMilestonesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*models.OnboardingMilestone, error)
	CompleteMilestone(ctx context.Context, id uuid.UUID) error

	// Templates
	GetTemplate(ctx context.Context, id uuid.UUID) (*models.OnboardingChecklistTemplate, error)
	ListTemplates(ctx context.Context, department, roleType string) ([]*models.OnboardingChecklistTemplate, error)

	// Statistics
	GetWorkflowStatistics(ctx context.Context, workflowID uuid.UUID) (*models.OnboardingStatistics, error)
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

// Workflows

func (s *onboardingService) CreateWorkflow(ctx context.Context, req *models.CreateWorkflowRequest, createdBy uuid.UUID) (*models.OnboardingWorkflow, error) {
	now := time.Now()
	workflow := &models.OnboardingWorkflow{
		ID:                     uuid.New(),
		EmployeeID:             req.EmployeeID,
		Status:                 "in_progress",
		StartDate:              req.StartDate,
		ExpectedCompletionDate: req.ExpectedCompletionDate,
		OverallProgress:        0,
		AssignedBuddyID:        req.AssignedBuddyID,
		AssignedManagerID:      req.AssignedManagerID,
		Notes:                  req.Notes,
		CreatedBy:              createdBy,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	// Link to workflow template if provided
	if req.WorkflowTemplateID != nil {
		workflow.WorkflowTemplateID = req.WorkflowTemplateID
	}

	err := s.repos.Onboarding.CreateOnboarding(ctx, workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return workflow, nil
}

func (s *onboardingService) GetWorkflow(ctx context.Context, id uuid.UUID) (*models.OnboardingWorkflow, error) {
	return s.repos.Onboarding.GetOnboarding(ctx, id)
}

func (s *onboardingService) GetWorkflowByEmployee(ctx context.Context, employeeID uuid.UUID) (*models.OnboardingWorkflow, error) {
	return s.repos.Onboarding.GetOnboardingByEmployee(ctx, employeeID)
}

func (s *onboardingService) ListWorkflows(ctx context.Context, filters map[string]interface{}) ([]*models.OnboardingWorkflow, error) {
	return s.repos.Onboarding.ListOnboardings(ctx, filters)
}

func (s *onboardingService) UpdateWorkflow(ctx context.Context, id uuid.UUID, req *models.UpdateWorkflowRequest) (*models.OnboardingWorkflow, error) {
	workflow, err := s.repos.Onboarding.GetOnboarding(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// Update fields
	if req.Status != "" {
		workflow.Status = req.Status
		if req.Status == "completed" && workflow.ActualCompletionDate == nil {
			now := time.Now()
			workflow.ActualCompletionDate = &now
		}
	}
	if req.ExpectedCompletionDate != nil {
		workflow.ExpectedCompletionDate = req.ExpectedCompletionDate
	}
	if req.AssignedBuddyID != nil {
		workflow.AssignedBuddyID = req.AssignedBuddyID
	}
	if req.AssignedManagerID != nil {
		workflow.AssignedManagerID = req.AssignedManagerID
	}
	if req.Notes != "" {
		workflow.Notes = req.Notes
	}

	workflow.UpdatedAt = time.Now()

	err = s.repos.Onboarding.UpdateOnboarding(ctx, workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	return workflow, nil
}

func (s *onboardingService) DeleteWorkflow(ctx context.Context, id uuid.UUID) error {
	return s.repos.Onboarding.DeleteOnboarding(ctx, id)
}

// Tasks

func (s *onboardingService) CreateTask(ctx context.Context, req *models.CreateTaskRequest) (*models.OnboardingTask, error) {
	now := time.Now()
	
	// Set default priority if not provided
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	
	task := &models.OnboardingTask{
		ID:              uuid.New(),
		WorkflowID:      req.WorkflowID,
		Title:           req.Title,
		Description:     req.Description,
		Category:        req.Category,
		Priority:        priority,
		Status:          "pending",
		AssignedTo:      req.AssignedTo,
		DueDate:         req.DueDate,
		IsMandatory:     req.IsMandatory,
		EstimatedHours:  req.EstimatedHours,
		OrderIndex:      req.OrderIndex,
		AIGenerated:     false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err := s.repos.Onboarding.CreateTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

func (s *onboardingService) GetTask(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error) {
	return s.repos.Onboarding.GetTask(ctx, id)
}

func (s *onboardingService) GetTaskByID(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error) {
	return s.repos.Onboarding.GetTask(ctx, id)
}

func (s *onboardingService) ListTasksByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*models.OnboardingTask, error) {
	return s.repos.Onboarding.GetTasksByWorkflow(ctx, workflowID)
}

func (s *onboardingService) GetTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.OnboardingTask, error) {
	return s.repos.Onboarding.GetTasksByEmployee(ctx, employeeID)
}

func (s *onboardingService) UpdateTask(ctx context.Context, id uuid.UUID, req *models.UpdateTaskRequest) (*models.OnboardingTask, error) {
	task, err := s.repos.Onboarding.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	// Update fields
	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status != "" {
		task.Status = req.Status
		if req.Status == "completed" && task.CompletedAt == nil {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
	if req.Priority != "" {
		task.Priority = req.Priority
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}
	if req.AssignedTo != nil {
		task.AssignedTo = req.AssignedTo
	}
	if req.ActualHours != nil {
		task.ActualHours = req.ActualHours
	}
	if req.AISuggestions != "" {
		task.AISuggestions = req.AISuggestions
	}

	task.UpdatedAt = time.Now()

	err = s.repos.Onboarding.UpdateTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

func (s *onboardingService) CompleteTask(ctx context.Context, taskID, completedBy uuid.UUID, req *models.CompleteTaskRequest) error {
	task, err := s.repos.Onboarding.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	now := time.Now()
	task.Status = "completed"
	task.CompletedAt = &now
	task.CompletedBy = &completedBy
	task.UpdatedAt = now

	if req.ActualHours != nil {
		task.ActualHours = req.ActualHours
	}
	
	// Note: OnboardingTask doesn't have a Notes field in the model
	// The notes would typically be stored separately or in the Description field

	err = s.repos.Onboarding.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	return nil
}

func (s *onboardingService) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return s.repos.Onboarding.DeleteTask(ctx, id)
}

// AI Interactions

func (s *onboardingService) HandleAIInteraction(ctx context.Context, req *models.AIInteractionRequest) (*models.AIInteractionResponse, error) {
	// Get workflow to get employee ID
	workflow, err := s.repos.Onboarding.GetOnboarding(ctx, req.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// Create interaction record
	now := time.Now()
	interaction := &models.OnboardingInteraction{
		ID:              uuid.New(),
		WorkflowID:      req.WorkflowID,
		EmployeeID:      workflow.EmployeeID,
		InteractionType: "chat",
		Message:         req.Message,
		AIResponse:      "", // Will be filled after AI call
		RequiresAction:  false,
		ActionTaken:     false,
		CreatedAt:       now,
	}

	// TODO: Call actual AI service here with req.Context for additional context
	// For now, return a simple response
	aiResponse := fmt.Sprintf("I understand you asked: '%s'. How can I help you with your onboarding?", req.Message)
	sentiment := "positive"
	
	interaction.AIResponse = aiResponse
	interaction.Sentiment = sentiment

	// Save interaction
	err = s.repos.Onboarding.CreateInteraction(ctx, interaction)
	if err != nil {
		return nil, fmt.Errorf("failed to save interaction: %w", err)
	}

	response := &models.AIInteractionResponse{
		InteractionID: interaction.ID,
		Response:      aiResponse,
		Suggestions:   []string{"View my tasks", "What documents do I need?", "When is my first day?"},
		Sentiment:     sentiment,
	}

	return response, nil
}

func (s *onboardingService) ListInteractionsByWorkflow(ctx context.Context, workflowID uuid.UUID, limit int) ([]*models.OnboardingInteraction, error) {
	return s.repos.Onboarding.ListInteractionsByWorkflow(ctx, workflowID, limit)
}

// Milestones

func (s *onboardingService) CreateMilestone(ctx context.Context, req *models.CreateMilestoneRequest) (*models.OnboardingMilestone, error) {
	now := time.Now()
	milestone := &models.OnboardingMilestone{
		ID:              uuid.New(),
		WorkflowID:      req.WorkflowID,
		Name:            req.Name,
		Description:     req.Description,
		TargetDate:      req.TargetDate,
		Status:          "pending",
		CelebrationSent: false,
		CreatedAt:       now,
	}

	err := s.repos.Onboarding.CreateMilestone(ctx, milestone)
	if err != nil {
		return nil, fmt.Errorf("failed to create milestone: %w", err)
	}

	return milestone, nil
}

func (s *onboardingService) ListMilestonesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*models.OnboardingMilestone, error) {
	return s.repos.Onboarding.ListMilestonesByWorkflow(ctx, workflowID)
}

func (s *onboardingService) CompleteMilestone(ctx context.Context, id uuid.UUID) error {
	milestone, err := s.repos.Onboarding.GetMilestone(ctx, id)
	if err != nil {
		return fmt.Errorf("milestone not found: %w", err)
	}

	now := time.Now()
	milestone.CompletedDate = &now
	milestone.Status = "completed"

	err = s.repos.Onboarding.UpdateMilestone(ctx, milestone)
	if err != nil {
		return fmt.Errorf("failed to complete milestone: %w", err)
	}

	return nil
}

// Templates

func (s *onboardingService) GetTemplate(ctx context.Context, id uuid.UUID) (*models.OnboardingChecklistTemplate, error) {
	return s.repos.Onboarding.GetTemplate(ctx, id)
}

func (s *onboardingService) ListTemplates(ctx context.Context, department, roleType string) ([]*models.OnboardingChecklistTemplate, error) {
	return s.repos.Onboarding.ListTemplates(ctx, department, roleType)
}

// Statistics

func (s *onboardingService) GetWorkflowStatistics(ctx context.Context, workflowID uuid.UUID) (*models.OnboardingStatistics, error) {
	return s.repos.Onboarding.GetStatistics(ctx, workflowID)
}

func (s *onboardingService) GetDashboard(ctx context.Context, filters map[string]interface{}) (*models.OnboardingDashboardResponse, error) {
	return s.repos.Onboarding.GetDashboardData(ctx, filters)
}

// CreateOnboardingPlan creates default onboarding tasks for a new employee
func (s *onboardingService) CreateOnboardingPlan(ctx context.Context, employeeID uuid.UUID, department string) error {
	// First, create a workflow for the employee
	now := time.Now()
	workflow := &models.OnboardingWorkflow{
		ID:              uuid.New(),
		EmployeeID:      employeeID,
		Status:          "in_progress",
		StartDate:       now,
		OverallProgress: 0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err := s.repos.Onboarding.CreateOnboarding(ctx, workflow)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	// Default onboarding tasks
	tasks := []models.OnboardingTask{
		{
			ID:             uuid.New(),
			WorkflowID:     workflow.ID,
			Title:          "Complete I-9 Form",
			Description:    "Complete employment eligibility verification",
			Category:       "documentation",
			Priority:       "high",
			Status:         "pending",
			IsMandatory:    true,
			DueDate:        timePtr(now.AddDate(0, 0, 3)),
			OrderIndex:     1,
			AIGenerated:    false,
			EstimatedHours: float64Ptr(0.5),
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             uuid.New(),
			WorkflowID:     workflow.ID,
			Title:          "Setup Direct Deposit",
			Description:    "Provide bank account information for payroll",
			Category:       "administrative",
			Priority:       "high",
			Status:         "pending",
			IsMandatory:    true,
			DueDate:        timePtr(now.AddDate(0, 0, 7)),
			OrderIndex:     2,
			AIGenerated:    false,
			EstimatedHours: float64Ptr(0.25),
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             uuid.New(),
			WorkflowID:     workflow.ID,
			Title:          "Complete Benefits Enrollment",
			Description:    "Select health insurance and other benefits",
			Category:       "administrative",
			Priority:       "medium",
			Status:         "pending",
			IsMandatory:    true,
			DueDate:        timePtr(now.AddDate(0, 0, 30)),
			OrderIndex:     3,
			AIGenerated:    false,
			EstimatedHours: float64Ptr(1.0),
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             uuid.New(),
			WorkflowID:     workflow.ID,
			Title:          "IT Account Setup",
			Description:    "Receive email, system access credentials",
			Category:       "access",
			Priority:       "high",
			Status:         "pending",
			IsMandatory:    true,
			DueDate:        timePtr(now.AddDate(0, 0, 1)),
			OrderIndex:     4,
			AIGenerated:    false,
			EstimatedHours: float64Ptr(0.5),
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             uuid.New(),
			WorkflowID:     workflow.ID,
			Title:          "Review Employee Handbook",
			Description:    "Read and acknowledge company policies",
			Category:       "documentation",
			Priority:       "medium",
			Status:         "pending",
			IsMandatory:    true,
			DueDate:        timePtr(now.AddDate(0, 0, 7)),
			OrderIndex:     5,
			AIGenerated:    false,
			EstimatedHours: float64Ptr(2.0),
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	for _, task := range tasks {
		if err := s.repos.Onboarding.CreateTask(ctx, &task); err != nil {
			return fmt.Errorf("failed to create task: %w", err)
		}
	}

	return nil
}
