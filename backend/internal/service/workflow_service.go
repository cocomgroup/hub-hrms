package service

import (
	"context"
	"fmt"
	"log"
	"encoding/json"
	"hub-hrms/backend/internal/integrations"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
	"time"

	"github.com/google/uuid"
)

// WorkflowService handles workflow business logic
type WorkflowService interface {
	// Workflow management
	InitiateWorkflow(ctx context.Context, employeeID uuid.UUID, templateName string, createdBy uuid.UUID) (*models.OnboardingWorkflow, error)
	GetWorkflow(ctx context.Context, workflowID uuid.UUID) (*models.OnboardingWithDetails, error)
	ListWorkflows(ctx context.Context, filters map[string]interface{}) ([]*models.OnboardingWorkflow, error)
	CancelWorkflow(ctx context.Context, workflowID uuid.UUID) error
	
	// Step execution
	StartStep(ctx context.Context, stepID uuid.UUID) error
	CompleteStep(ctx context.Context, stepID, completedBy uuid.UUID) error
	SkipStep(ctx context.Context, stepID, userID uuid.UUID, reason string) error
	
	// Integration triggers
	TriggerDocuSign(ctx context.Context, stepID uuid.UUID, documentType string) error
	TriggerBackgroundCheck(ctx context.Context, stepID uuid.UUID, checkTypes []string) error
	TriggerDocSearch(ctx context.Context, stepID uuid.UUID, query string) error
	
	// Exception handling
	RaiseException(ctx context.Context, workflowID uuid.UUID, exception *models.WorkflowException) error
	ResolveException(ctx context.Context, exceptionID, resolvedBy uuid.UUID, notes string) error
	
	// Progress monitoring
	CheckWorkflowProgress(ctx context.Context, workflowID uuid.UUID) (*WorkflowProgress, error)
	AdvanceStage(ctx context.Context, workflowID uuid.UUID) error

	// Template management
	CreateWorkflowTemplate(ctx context.Context, template *models.WorkflowTemplate, steps []models.WorkflowStepDef) error
	GetWorkflowTemplate(ctx context.Context, templateID uuid.UUID) (*models.WorkflowTemplate, error)
	ListWorkflowTemplates(ctx context.Context) ([]*models.WorkflowTemplate, error)
	UpdateWorkflowTemplate(ctx context.Context, template *models.WorkflowTemplate, steps []models.WorkflowStepDef) error
	DeleteWorkflowTemplate(ctx context.Context, templateID uuid.UUID) error

	GetStats(ctx context.Context) (map[string]interface{}, error)
	ListTemplates(ctx context.Context, activeOnly bool) ([]*models.WorkflowTemplate, error)
	CreateTemplate(ctx context.Context, req interface{}) (*models.WorkflowTemplate, error)
	GetTemplate(ctx context.Context, id string) (*models.WorkflowTemplate, error)
	UpdateTemplate(ctx context.Context, id string, req interface{}) (*models.WorkflowTemplate, error) 
	DeleteTemplate(ctx context.Context, id string) error
	ToggleTemplate(ctx context.Context, id string) (*models.WorkflowTemplate, error)
	GetRecentAssignments(ctx context.Context, limit int) ([]map[string]interface{}, error)
}

type workflowService struct {
	repos             *repository.Repositories
	docuSign          integrations.DocuSignService
	backgroundCheck   integrations.BackgroundCheckService
	docSearch         integrations.DocSearchService
}

// WorkflowProgress represents workflow progress metrics
type WorkflowProgress struct {
	WorkflowID         uuid.UUID `json:"workflow_id"`
	TotalSteps         int       `json:"total_steps"`
	CompletedSteps     int       `json:"completed_steps"`
	InProgressSteps    int       `json:"in_progress_steps"`
	PendingSteps       int       `json:"pending_steps"`
	BlockedSteps       int       `json:"blocked_steps"`
	FailedSteps        int       `json:"failed_steps"`
	ProgressPercentage int       `json:"progress_percentage"`
	CurrentStage       string    `json:"current_stage"`
	DaysElapsed        int       `json:"days_elapsed"`
	ExpectedDays       int       `json:"expected_days"`
	IsOnTrack          bool      `json:"is_on_track"`
	OpenExceptions     int       `json:"open_exceptions"`
}

func NewWorkflowService(repos *repository.Repositories) WorkflowService {
	return &workflowService{
		repos:           repos,
		docuSign:        integrations.NewMockDocuSignService(),
		backgroundCheck: integrations.NewMockBackgroundCheckService(),
		docSearch:       integrations.NewMockDocSearchService(),
	}
}

// getCurrentStage calculates the current onboarding stage based on progress
func getCurrentStage(workflow *models.OnboardingWorkflow) string {
	if workflow.Status == "not_started" {
		return "pre-boarding"
	}
	if workflow.Status == "completed" {
		return "completed"
	}
	
	// Calculate stage based on progress percentage
	progress := workflow.OverallProgress
	
	if progress < 25 {
		return "pre-boarding"
	} else if progress < 50 {
		return "day-1"
	} else if progress < 75 {
		return "week-1"
	} else if progress < 100 {
		return "month-1"
	}
	
	return "completed"
}

// InitiateWorkflow creates a new workflow from a template
func (s *workflowService) InitiateWorkflow(ctx context.Context, employeeID uuid.UUID, templateName string, createdByUserID uuid.UUID) (*models.OnboardingWorkflow, error) {
	log.Printf("DEBUG InitiateWorkflow: employeeID=%s, templateName=%s, createdByUserID=%s", employeeID, templateName, createdByUserID)
	log.Printf("DEBUG InitiateWorkflow: createdBy is nil? %v", createdByUserID == uuid.Nil)
	
	// Get employee details
	employee, err := s.repos.Employee.GetByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}
	
	now := time.Now()
	
	hrMgr, err := s.repos.Employee.GetByUserID(ctx, createdByUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get createdBy employee: %w", err)
	}
	createdBy := hrMgr.ID
	
	// Create workflow using NewHireOnboarding (OnboardingWorkflow) structure
	workflow := &models.OnboardingWorkflow{
		ID:                     uuid.New(),
		EmployeeID:             employeeID,
		EmployeeName:           employee.FirstName + " " + employee.LastName,
		EmployeeEmail:          employee.Email,
		Status:                 "in_progress",  // not_started, in_progress, completed, overdue
		StartDate:              now,
		OverallProgress:        0,
		CreatedBy:              createdBy,  
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	
	// Calculate expected completion (30 days from now)
	expectedCompletion := now.Add(30 * 24 * time.Hour)
	workflow.ExpectedCompletionDate = &expectedCompletion
	
	log.Printf("DEBUG InitiateWorkflow: workflow.CreatedBy = %v", workflow.CreatedBy)
	if workflow.CreatedBy != uuid.Nil {
		log.Printf("DEBUG InitiateWorkflow: *workflow.CreatedBy = %s", workflow.CreatedBy)
	} else {
		log.Printf("DEBUG InitiateWorkflow: CreatedBy is nil (no user in context or user doesn't exist)")
	}
	
	log.Printf("DEBUG InitiateWorkflow: About to call CreateOnboarding")
	err = s.repos.Onboarding.CreateOnboarding(ctx, workflow)
	if err != nil {
		log.Printf("ERROR InitiateWorkflow: CreateOnboarding failed: %v", err)
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}
	
	// Generate tasks from template
	tasks, err := s.generateTasksFromTemplate(templateName, workflow.ID, employee)
	if err != nil {
		log.Printf("WARN InitiateWorkflow: Failed to generate tasks: %v", err)
		// Don't fail the entire workflow creation if task generation fails
	} else {
		// Create the generated tasks
		for _, task := range tasks {
			if err := s.repos.Onboarding.CreateTask(ctx, task); err != nil {
				log.Printf("WARN InitiateWorkflow: Failed to create task %s: %v", task.Title, err)
			}
		}
	}
	
	log.Printf("DEBUG InitiateWorkflow: Workflow created successfully with ID %s", workflow.ID)
	return workflow, nil
}

// generateStepsFromTemplate creates steps based on template
func (s *workflowService) generateTasksFromTemplate(templateName string, workflowID uuid.UUID, employee *models.Employee) ([]*models.OnboardingTask, error) {
	// This is a simplified version - in production you'd look up the template
	// and generate tasks based on the template definition
	
	now := time.Now()
	tasks := []*models.OnboardingTask{}
	
	// Default tasks based on template name
	switch templateName {
	case "standard-onboarding":
		tasks = append(tasks,
			&models.OnboardingTask{
				ID:             uuid.New(),
				WorkflowID:     workflowID,
				Title:          "Complete I-9 Form",
				Description:    "Complete employment eligibility verification",
				Category:       "documentation",
				Priority:       "high",
				Status:         "pending",
				IsMandatory:    true,
				OrderIndex:     1,
				DueDate:        timePtr(now.AddDate(0, 0, 3)),
				EstimatedHours: float64Ptr(0.5),
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			&models.OnboardingTask{
				ID:             uuid.New(),
				WorkflowID:     workflowID,
				Title:          "Setup Direct Deposit",
				Description:    "Provide bank account information",
				Category:       "administrative",
				Priority:       "high",
				Status:         "pending",
				IsMandatory:    true,
				OrderIndex:     2,
				DueDate:        timePtr(now.AddDate(0, 0, 7)),
				EstimatedHours: float64Ptr(0.25),
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			&models.OnboardingTask{
				ID:             uuid.New(),
				WorkflowID:     workflowID,
				Title:          "IT Account Setup",
				Description:    "Receive email and system access",
				Category:       "access",
				Priority:       "high",
				Status:         "pending",
				IsMandatory:    true,
				OrderIndex:     3,
				DueDate:        timePtr(now.AddDate(0, 0, 1)),
				EstimatedHours: float64Ptr(0.5),
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		)
	case "engineering-onboarding":
		tasks = append(tasks,
			&models.OnboardingTask{
				ID:             uuid.New(),
				WorkflowID:     workflowID,
				Title:          "Development Environment Setup",
				Description:    "Install required development tools and access",
				Category:       "access",
				Priority:       "high",
				Status:         "pending",
				IsMandatory:    true,
				OrderIndex:     1,
				DueDate:        timePtr(now.AddDate(0, 0, 2)),
				EstimatedHours: float64Ptr(2.0),
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			&models.OnboardingTask{
				ID:             uuid.New(),
				WorkflowID:     workflowID,
				Title:          "Code Repository Access",
				Description:    "Get access to GitHub/GitLab repositories",
				Category:       "access",
				Priority:       "high",
				Status:         "pending",
				IsMandatory:    true,
				OrderIndex:     2,
				DueDate:        timePtr(now.AddDate(0, 0, 1)),
				EstimatedHours: float64Ptr(0.5),
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		)
	}
	
	return tasks, nil
}

// createGenericTemplate generates steps for generic employee onboarding
func (s *workflowService) createGenericTemplate(workflowID uuid.UUID, employee *models.Employee) []*models.WorkflowStep {
	steps := []*models.WorkflowStep{
		{
			WorkflowID:      workflowID,
			StepOrder:       1,
			StepName:        "Send Offer Letter",
			StepType:        "integration",
			Stage:           "pre-boarding",
			Status:          "pending",
			IntegrationType: "docusign",
			IntegrationConfig: map[string]interface{}{
				"document_type": "offer-letter",
			},
			DueDate: timePtr(time.Now().Add(-7 * 24 * time.Hour)),
		},
		{
			WorkflowID:      workflowID,
			StepOrder:       2,
			StepName:        "Send I-9 Form",
			StepType:        "integration",
			Stage:           "pre-boarding",
			Status:          "pending",
			IntegrationType: "docusign",
			IntegrationConfig: map[string]interface{}{
				"document_type": "i9",
			},
			DueDate: timePtr(time.Now().Add(-5 * 24 * time.Hour)),
		},
		{
			WorkflowID:  workflowID,
			StepOrder:   3,
			StepName:    "Welcome Email",
			StepType:    "manual",
			Stage:       "day-1",
			Status:      "blocked",
			Description: "Send welcome email",
			DueDate:     timePtr(time.Now()),
		},
		{
			WorkflowID:  workflowID,
			StepOrder:   4,
			StepName:    "Office Tour",
			StepType:    "manual",
			Stage:       "day-1",
			Status:      "blocked",
			Description: "Conduct office tour",
			DueDate:     timePtr(time.Now()),
		},
	}
	
	return steps
}

// createSalesRepTemplate generates steps for sales representative onboarding
func (s *workflowService) createSalesRepTemplate(workflowID uuid.UUID, employee *models.Employee) []*models.WorkflowStep {
	// Similar to generic but with CRM setup, sales training, etc.
	return s.createGenericTemplate(workflowID, employee)
}

// createManagerTemplate generates steps for manager onboarding
func (s *workflowService) createManagerTemplate(workflowID uuid.UUID, employee *models.Employee) []*models.WorkflowStep {
	// Similar to generic but with leadership training, team intros, etc.
	return s.createGenericTemplate(workflowID, employee)
}

// GetWorkflow retrieves workflow with all details
func (s *workflowService) GetWorkflow(ctx context.Context, workflowID uuid.UUID) (*models.OnboardingWithDetails, error) {
	return s.repos.Workflow.GetWorkflowWithDetails(ctx, workflowID)
}

// ListWorkflows retrieves workflows with filters
func (s *workflowService) ListWorkflows(ctx context.Context, filters map[string]interface{}) ([]*models.OnboardingWorkflow, error) {
	return s.repos.Workflow.ListWorkflows(ctx, filters)
}

// CancelWorkflow cancels an active workflow
func (s *workflowService) CancelWorkflow(ctx context.Context, workflowID uuid.UUID) error {
	return s.repos.Workflow.UpdateWorkflowStatus(ctx, workflowID, "cancelled")
}

// StartStep marks a step as in-progress
func (s *workflowService) StartStep(ctx context.Context, stepID uuid.UUID) error {
	step, err := s.repos.Workflow.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	
	// Check if dependencies are met
	if !s.checkDependencies(ctx, step) {
		return fmt.Errorf("step dependencies not met")
	}
	
	now := time.Now()
	step.Status = "in-progress"
	step.StartedAt = &now
	
	return s.repos.Workflow.UpdateStep(ctx, step)
}

// CompleteStep marks a step as completed
func (s *workflowService) CompleteStep(ctx context.Context, stepID, completedBy uuid.UUID) error {
	err := s.repos.Workflow.CompleteStep(ctx, stepID, completedBy)
	if err != nil {
		return err
	}
	
	// Check if we should advance stage
	step, _ := s.repos.Workflow.GetStepByID(ctx, stepID)
	if step != nil {
		s.checkAndAdvanceStage(ctx, step.WorkflowID)
	}
	
	return nil
}

// SkipStep skips a step with reason
func (s *workflowService) SkipStep(ctx context.Context, stepID, userID uuid.UUID, reason string) error {
	step, err := s.repos.Workflow.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	
	step.Status = "skipped"
	step.CompletedBy = &userID
	now := time.Now()
	step.CompletedAt = &now
	
	// Add reason to metadata
	if step.Metadata == nil {
		step.Metadata = make(map[string]interface{})
	}
	step.Metadata["skip_reason"] = reason
	step.Metadata["skipped_by"] = userID.String()
	
	return s.repos.Workflow.UpdateStep(ctx, step)
}

// TriggerDocuSign triggers DocuSign integration
func (s *workflowService) TriggerDocuSign(ctx context.Context, stepID uuid.UUID, documentType string) error {
	step, err := s.repos.Workflow.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	
	// Get workflow to get employee info
	workflow, err := s.repos.Workflow.GetWorkflow(ctx, step.WorkflowID)
	if err != nil {
		return err
	}
	
	employee, err := s.repos.Employee.GetByID(ctx, workflow.EmployeeID)
	if err != nil {
		return err
	}
	
	// Create integration record
	integration := &models.WorkflowIntegration{
		WorkflowID:      step.WorkflowID,
		StepID:          step.ID,
		IntegrationType: "docusign",
		Status:          "pending",
		RequestPayload: map[string]interface{}{
			"document_type": documentType,
			"signer_email":  employee.Email,
			"signer_name":   employee.FirstName + " " + employee.LastName,
		},
		MaxRetries: 3,
	}
	
	err = s.repos.Workflow.CreateIntegration(ctx, integration)
	if err != nil {
		return err
	}
	
	// Call DocuSign API
	req := &integrations.DocuSignEnvelopeRequest{
		DocumentType: documentType,
		SignerEmail:  employee.Email,
		SignerName:   employee.FirstName + " " + employee.LastName,
		EmployeeID:   employee.ID,
	}
	
	response, err := s.docuSign.SendEnvelope(ctx, req)
	if err != nil {
		integration.Status = "failed"
		integration.ErrorMessage = err.Error()
		s.repos.Workflow.UpdateIntegration(ctx, integration)
		
		// Raise exception
		s.raiseIntegrationException(ctx, step.WorkflowID, step.ID, "docusign", err.Error())
		
		return err
	}
	
	// Update integration with response
	integration.Status = "completed"
	integration.ExternalID = response.EnvelopeID
	integration.ResponsePayload = map[string]interface{}{
		"envelope_id":  response.EnvelopeID,
		"status":       response.Status,
		"sent_at":      response.SentAt,
		"signer_email": response.SignerEmail,
	}
	
	return s.repos.Workflow.UpdateIntegration(ctx, integration)
}

// TriggerBackgroundCheck triggers background check integration
func (s *workflowService) TriggerBackgroundCheck(ctx context.Context, stepID uuid.UUID, checkTypes []string) error {
	step, err := s.repos.Workflow.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	
	workflow, err := s.repos.Workflow.GetWorkflow(ctx, step.WorkflowID)
	if err != nil {
		return err
	}
	
	employee, err := s.repos.Employee.GetByID(ctx, workflow.EmployeeID)
	if err != nil {
		return err
	}
	
	// Create integration record
	integration := &models.WorkflowIntegration{
		WorkflowID:      step.WorkflowID,
		StepID:          step.ID,
		IntegrationType: "background-check",
		Status:          "pending",
		RequestPayload: map[string]interface{}{
			"first_name":  employee.FirstName,
			"last_name":   employee.LastName,
			"email":       employee.Email,
			"check_types": checkTypes,
		},
		MaxRetries: 3,
	}
	
	err = s.repos.Workflow.CreateIntegration(ctx, integration)
	if err != nil {
		return err
	}
	
	// Call Background Check API
	req := &integrations.BackgroundCheckRequest{
		FirstName:  employee.FirstName,
		LastName:   employee.LastName,
		Email:      employee.Email,
		CheckTypes: checkTypes,
		EmployeeID: employee.ID,
	}
	
	response, err := s.backgroundCheck.InitiateCheck(ctx, req)
	if err != nil {
		integration.Status = "failed"
		integration.ErrorMessage = err.Error()
		s.repos.Workflow.UpdateIntegration(ctx, integration)
		s.raiseIntegrationException(ctx, step.WorkflowID, step.ID, "background-check", err.Error())
		return err
	}
	
	// Update integration
	integration.Status = "in-progress"
	integration.ExternalID = response.CheckID
	integration.ResponsePayload = map[string]interface{}{
		"check_id":     response.CheckID,
		"status":       response.Status,
		"candidate":    response.Candidate,
		"check_types":  response.CheckTypes,
		"initiated_at": response.InitiatedAt,
	}
	
	return s.repos.Workflow.UpdateIntegration(ctx, integration)
}

// TriggerDocSearch triggers document search integration
func (s *workflowService) TriggerDocSearch(ctx context.Context, stepID uuid.UUID, query string) error {
	step, err := s.repos.Workflow.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	
	// Create integration record
	integration := &models.WorkflowIntegration{
		WorkflowID:      step.WorkflowID,
		StepID:          step.ID,
		IntegrationType: "doc-search",
		Status:          "pending",
		RequestPayload: map[string]interface{}{
			"query": query,
		},
		MaxRetries: 3,
	}
	
	err = s.repos.Workflow.CreateIntegration(ctx, integration)
	if err != nil {
		return err
	}
	
	// Call Doc Search API
	req := &integrations.DocSearchRequest{
		Query: query,
		Limit: 10,
	}
	
	response, err := s.docSearch.SearchDocuments(ctx, req)
	if err != nil {
		integration.Status = "failed"
		integration.ErrorMessage = err.Error()
		s.repos.Workflow.UpdateIntegration(ctx, integration)
		s.raiseIntegrationException(ctx, step.WorkflowID, step.ID, "doc-search", err.Error())
		return err
	}
	
	// Update integration
	integration.Status = "completed"
	integration.ResponsePayload = map[string]interface{}{
		"total_count": response.TotalCount,
		"documents":   response.Documents,
	}
	
	// Create document records for found documents
	for _, doc := range response.Documents {
		workflowDoc := &models.WorkflowDocument{
			WorkflowID:   step.WorkflowID,
			StepID:       &step.ID,
			DocumentName: doc.Name,
			DocumentType: doc.DocumentType,
			S3Key:        doc.S3Key,
			FileType:     doc.FileType,
			FileSize:     doc.FileSize,
			Status:       "available",
			Metadata:     doc.Metadata,
		}
		s.repos.Workflow.CreateDocument(ctx, workflowDoc)
	}
	
	return s.repos.Workflow.UpdateIntegration(ctx, integration)
}

// RaiseException creates an exception record
func (s *workflowService) RaiseException(ctx context.Context, workflowID uuid.UUID, exception *models.WorkflowException) error {
	exception.WorkflowID = workflowID
	exception.ResolutionStatus = "open"
	return s.repos.Workflow.CreateException(ctx, exception)
}

// raiseIntegrationException helper for integration failures
func (s *workflowService) raiseIntegrationException(ctx context.Context, workflowID, stepID uuid.UUID, integrationType, errorMsg string) {
	exception := &models.WorkflowException{
		WorkflowID:    workflowID,
		StepID:        &stepID,
		ExceptionType: "integration_failure",
		Severity:      "high",
		Title:         fmt.Sprintf("%s integration failed", integrationType),
		Description:   errorMsg,
	}
	s.repos.Workflow.CreateException(ctx, exception)
}

// ResolveException marks an exception as resolved
func (s *workflowService) ResolveException(ctx context.Context, exceptionID, resolvedBy uuid.UUID, notes string) error {
	return s.repos.Workflow.ResolveException(ctx, exceptionID, resolvedBy, notes)
}

// CheckWorkflowProgress calculates workflow progress metrics
func (s *workflowService) CheckWorkflowProgress(ctx context.Context, workflowID uuid.UUID) (*WorkflowProgress, error) {
	workflow, err := s.repos.Workflow.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	
	steps, err := s.repos.Workflow.GetSteps(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	
	exceptions, err := s.repos.Workflow.GetExceptions(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	
	// ✓ FIXED: Calculate current stage from progress
	currentStage := getCurrentStage(workflow)
	
	progress := &WorkflowProgress{
		WorkflowID:   workflowID,
		TotalSteps:   len(steps),
		CurrentStage: currentStage,
	}
	
	// Count step statuses
	for _, step := range steps {
		switch step.Status {
		case "completed":
			progress.CompletedSteps++
		case "in-progress":
			progress.InProgressSteps++
		case "pending":
			progress.PendingSteps++
		case "blocked":
			progress.BlockedSteps++
		case "failed":
			progress.FailedSteps++
		}
	}
	
	// Calculate progress percentage
	if progress.TotalSteps > 0 {
		progress.ProgressPercentage = (progress.CompletedSteps * 100) / progress.TotalSteps
	}
	
	// Calculate days elapsed
	progress.DaysElapsed = int(time.Since(workflow.StartDate).Hours() / 24)
	if workflow.ExpectedCompletionDate != nil {
		progress.ExpectedDays = int(workflow.ExpectedCompletionDate.Sub(workflow.StartDate).Hours() / 24)
		progress.IsOnTrack = progress.DaysElapsed <= progress.ExpectedDays
	}
	
	// Count open exceptions
	progress.OpenExceptions = len(exceptions)
	
	return progress, nil
}

// AdvanceStage advances workflow to next stage
func (s *workflowService) AdvanceStage(ctx context.Context, workflowID uuid.UUID) error {
	workflow, err := s.repos.Workflow.GetWorkflow(ctx, workflowID)
	if err != nil {
		return err
	}
	
	// ✓ FIXED: Calculate current stage from progress
	currentStage := getCurrentStage(workflow)
	
	// Determine next stage and progress
	var nextStage string
	var newProgress int
	
	switch currentStage {
	case "pre-boarding":
		nextStage = "day-1"
		newProgress = 25
	case "day-1":
		nextStage = "week-1"
		newProgress = 50
	case "week-1":
		nextStage = "month-1"
		newProgress = 75
	case "month-1":
		nextStage = "completed"
		newProgress = 100
		workflow.Status = "completed"
		now := time.Now()
		workflow.ActualCompletionDate = &now
	default:
		return fmt.Errorf("unknown stage: %s nextStage %s", currentStage, nextStage)
	}
	
	// Update progress
	workflow.OverallProgress = newProgress
	
	return s.repos.Onboarding.UpdateOnboarding(ctx, workflow)
}

// checkAndAdvanceStage checks if all steps in current stage are done
func (s *workflowService) checkAndAdvanceStage(ctx context.Context, workflowID uuid.UUID) error {
	workflow, err := s.repos.Workflow.GetWorkflow(ctx, workflowID)
	if err != nil {
		return err
	}
	
	steps, err := s.repos.Workflow.GetSteps(ctx, workflowID)
	if err != nil {
		return err
	}
	
	// ✓ FIXED: Calculate current stage from progress
	currentStage := getCurrentStage(workflow)
	
	// Check if all steps in current stage are completed or skipped
	allDone := true
	for _, step := range steps {
		if step.Stage == currentStage {
			if step.Status != "completed" && step.Status != "skipped" {
				allDone = false
				break
			}
		}
	}
	
	// If all done, advance stage
	if allDone {
		return s.AdvanceStage(ctx, workflowID)
	}
	
	return nil
}

// checkDependencies checks if step dependencies are met
func (s *workflowService) checkDependencies(ctx context.Context, step *models.WorkflowStep) bool {
	if len(step.Dependencies) == 0 {
		return true
	}
	
	for _, depID := range step.Dependencies {
		depStep, err := s.repos.Workflow.GetStepByID(ctx, depID)
		if err != nil || (depStep.Status != "completed" && depStep.Status != "skipped") {
			return false
		}
	}
	
	return true
}

// CreateWorkflowTemplate creates a new workflow template with steps
func (s *workflowService) CreateWorkflowTemplate(ctx context.Context, template *models.WorkflowTemplate, steps []models.WorkflowStepDef) error {
	// Validate template
	if template.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(steps) == 0 {
		return fmt.Errorf("at least one step is required")
	}

	// Set default status if not provided
	if template.Status == "" {
		template.Status = "active"
	}

	// Create template
	err := s.repos.Workflow.CreateTemplate(ctx, template)
	if err != nil {
		return fmt.Errorf("failed to create workflow template: %w", err)
	}

	// Create step definitions
	for i := range steps {
		steps[i].WorkflowID = template.ID

		// Validate step order
		if steps[i].StepOrder == 0 {
			steps[i].StepOrder = i + 1
		}

		// Set default role if not provided
		if steps[i].AssignedRole == "" {
			steps[i].AssignedRole = "hr"
		}

		err := s.repos.Workflow.CreateStepDef(ctx, &steps[i])
		if err != nil {
			// Rollback: delete template if step creation fails
			_ = s.repos.Workflow.DeleteTemplate(ctx, template.ID)
			return fmt.Errorf("failed to create step definition: %w", err)
		}
	}

	// Load steps into template
	template.Steps = steps

	return nil
}

// GetWorkflowTemplate retrieves a workflow template by ID with all steps
func (s *workflowService) GetWorkflowTemplate(ctx context.Context, templateID uuid.UUID) (*models.WorkflowTemplate, error) {
	template, err := s.repos.Workflow.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow template: %w", err)
	}

	return template, nil
}

// ListWorkflowTemplates retrieves all workflow templates with their steps
func (s *workflowService) ListWorkflowTemplates(ctx context.Context) ([]*models.WorkflowTemplate, error) {
	templates, err := s.repos.Workflow.ListTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow templates: %w", err)
	}

	return templates, nil
}

// UpdateWorkflowTemplate updates a workflow template and its steps
func (s *workflowService) UpdateWorkflowTemplate(ctx context.Context, template *models.WorkflowTemplate, steps []models.WorkflowStepDef) error {
	// Validate
	if template.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(steps) == 0 {
		return fmt.Errorf("at least one step is required")
	}

	// Update template
	err := s.repos.Workflow.UpdateTemplate(ctx, template)
	if err != nil {
		return fmt.Errorf("failed to update workflow template: %w", err)
	}

	// Delete existing step definitions
	err = s.repos.Workflow.DeleteStepDefsByWorkflowID(ctx, template.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing steps: %w", err)
	}

	// Create new step definitions
	for i := range steps {
		steps[i].WorkflowID = template.ID
		steps[i].ID = uuid.Nil // Reset ID to create new

		// Validate step order
		if steps[i].StepOrder == 0 {
			steps[i].StepOrder = i + 1
		}

		// Set default role if not provided
		if steps[i].AssignedRole == "" {
			steps[i].AssignedRole = "hr"
		}

		err := s.repos.Workflow.CreateStepDef(ctx, &steps[i])
		if err != nil {
			return fmt.Errorf("failed to create step definition: %w", err)
		}
	}

	// Load steps into template
	template.Steps = steps

	return nil
}

// DeleteWorkflowTemplate deletes a workflow template and all its steps
func (s *workflowService) DeleteWorkflowTemplate(ctx context.Context, templateID uuid.UUID) error {
	// Check if template exists
	_, err := s.repos.Workflow.GetTemplateByID(ctx, templateID)
	if err != nil {
		return fmt.Errorf("workflow template not found")
	}

	// TODO: Check if template is in use by any active workflow instances
	// This would prevent deletion of templates that have active workflows

	// Delete template (cascade deletes steps)
	err = s.repos.Workflow.DeleteTemplate(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to delete workflow template: %w", err)
	}

	return nil
}

// ============================================================================
// HELPER METHODS (Optional)
// ============================================================================

// DuplicateWorkflowTemplate creates a copy of an existing template
func (s *workflowService) DuplicateWorkflowTemplate(ctx context.Context, sourceTemplateID uuid.UUID, newName string, createdBy uuid.UUID) (*models.WorkflowTemplate, error) {
	// Get source template
	source, err := s.GetWorkflowTemplate(ctx, sourceTemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source template: %w", err)
	}

	// Create new template
	newTemplate := &models.WorkflowTemplate{
		Name:         newName,
		Description:  source.Description + " (Copy)",
		WorkflowType: source.WorkflowType,
		Status:       "draft",
		CreatedBy:    createdBy,
	}

	// Copy steps
	newSteps := make([]models.WorkflowStepDef, len(source.Steps))
	for i, step := range source.Steps {
		newSteps[i] = models.WorkflowStepDef{
			StepOrder:    step.StepOrder,
			StepType:     step.StepType,
			StepName:     step.StepName,
			Description:  step.Description,
			Required:     step.Required,
			AutoTrigger:  step.AutoTrigger,
			AssignedRole: step.AssignedRole,
			DueDays:      step.DueDays,
		}
	}

	// Create new template with steps
	err = s.CreateWorkflowTemplate(ctx, newTemplate, newSteps)
	if err != nil {
		return nil, fmt.Errorf("failed to duplicate template: %w", err)
	}

	return newTemplate, nil
}

// GetTemplatesByType retrieves templates filtered by workflow type
func (s *workflowService) GetTemplatesByType(ctx context.Context, workflowType string) ([]*models.WorkflowTemplate, error) {
	allTemplates, err := s.ListWorkflowTemplates(ctx)
	if err != nil {
		return nil, err
	}

	// Filter by type
	var filtered []*models.WorkflowTemplate
	for _, template := range allTemplates {
		if template.WorkflowType == workflowType {
			filtered = append(filtered, template)
		}
	}

	return filtered, nil
}

// GetActiveTemplates retrieves only active templates
func (s *workflowService) GetActiveTemplates(ctx context.Context) ([]*models.WorkflowTemplate, error) {
	allTemplates, err := s.ListWorkflowTemplates(ctx)
	if err != nil {
		return nil, err
	}

	// Filter active only
	var active []*models.WorkflowTemplate
	for _, template := range allTemplates {
		if template.Status == "active" {
			active = append(active, template)
		}
	}

	return active, nil
}
// GetStats returns workflow statistics for the dashboard
func (s *workflowService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// Count active workflow templates
	templates, err := s.ListWorkflowTemplates(ctx)
	if err != nil {
		return nil, err
	}
	
	activeTemplates := 0
	for _, t := range templates {
		if t.Status == "active" {
			activeTemplates++
		}
	}
	
	// Count active workflows
	workflows, err := s.ListWorkflows(ctx, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	
	activeWorkflows := 0
	completedThisMonth := 0
	totalDays := 0
	pendingAssignments := 0
	
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	for _, wf := range workflows {
		switch wf.Status {
		case "active", "in_progress":
			activeWorkflows++
		case "completed":
			if wf.ActualCompletionDate != nil && wf.ActualCompletionDate.After(monthStart) {
				completedThisMonth++
				days := int(wf.ActualCompletionDate.Sub(wf.StartDate).Hours() / 24)
				totalDays += days
			}
		case "pending":
			pendingAssignments++
		}
	}
	
	avgDays := 0
	if completedThisMonth > 0 {
		avgDays = totalDays / completedThisMonth
	}
	
	return map[string]interface{}{
		"templates_count":      activeTemplates,
		"active_workflows":     activeWorkflows,
		"completed_this_month": completedThisMonth,
		"avg_completion_time":  avgDays,
		"pending_assignments":  pendingAssignments,
	}, nil
}

// ListTemplates returns workflow templates with optional active filter
func (s *workflowService) ListTemplates(ctx context.Context, activeOnly bool) ([]*models.WorkflowTemplate, error) {
	templates, err := s.ListWorkflowTemplates(ctx)
	if err != nil {
		return nil, err
	}
	
	if !activeOnly {
		return templates, nil
	}
	
	// Filter to only active templates
	var activeTemplates []*models.WorkflowTemplate
	for _, t := range templates {
		if t.Status == "active" {
			activeTemplates = append(activeTemplates, t)
		}
	}
	
	return activeTemplates, nil
}

// CreateTemplate creates a new workflow template
// CreateTemplate creates a new workflow template
func (s *workflowService) CreateTemplate(ctx context.Context, req interface{}) (*models.WorkflowTemplate, error) {
	// Parse the request
	reqData, ok := req.(map[string]interface{})
	if !ok {
		// Try to convert to map
		jsonData, _ := json.Marshal(req)
		json.Unmarshal(jsonData, &reqData)
	}
	
	// Get user ID from context (if available)
	userID := uuid.Nil
	// In a real implementation, extract from context:
	// userID := getUserIDFromContext(ctx)
	
	template := &models.WorkflowTemplate{
		ID:           uuid.New(),
		Name:         getStringOrEmpty(reqData, "name"),
		Description:  getStringOrEmpty(reqData, "description"),
		WorkflowType: getStringOrEmpty(reqData, "type"),
		Status:       "active",
		CreatedBy:    userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	// Parse steps
	var steps []models.WorkflowStepDef
	if stepsData, ok := reqData["steps"].([]interface{}); ok {
		for _, stepData := range stepsData {
			if stepMap, ok := stepData.(map[string]interface{}); ok {
				step := models.WorkflowStepDef{
					ID:           uuid.New(),
					WorkflowID:   template.ID,  // FIXED: Use template.ID (uuid.UUID)
					StepOrder:    getIntOrZero(stepMap, "order"),
					StepType:     getStringOrEmpty(stepMap, "step_type"),  // ADDED: Required field
					StepName:     getStringOrEmpty(stepMap, "name"),
					Description:  getStringOrEmpty(stepMap, "description"),
					Required:     getBoolOrFalse(stepMap, "required"),
					AutoTrigger:  getBoolOrFalse(stepMap, "auto_trigger"),
					AssignedRole: getStringOrEmpty(stepMap, "assignee_role"),
					DueDays:      getIntPointer(stepMap, "estimated_days"),  // FIXED: Pointer
					CreatedAt:    time.Now(),
				}
				steps = append(steps, step)
			}
		}
	}
	
	// Create template with steps
	if err := s.CreateWorkflowTemplate(ctx, template, steps); err != nil {
		return nil, err
	}
	
	return template, nil
}


// GetTemplate retrieves a workflow template by ID
func (s *workflowService) GetTemplate(ctx context.Context, id string) (*models.WorkflowTemplate, error) {
	templateID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	
	return s.GetWorkflowTemplate(ctx, templateID)
}

// UpdateTemplate updates a workflow template
func (s *workflowService) UpdateTemplate(ctx context.Context, id string, req interface{}) (*models.WorkflowTemplate, error) {
	templateID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	
	// Get existing template
	template, err := s.GetWorkflowTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}
	
	// Parse the request
	reqData, ok := req.(map[string]interface{})
	if !ok {
		jsonData, _ := json.Marshal(req)
		json.Unmarshal(jsonData, &reqData)
	}
	
	// Update template fields
	if name, ok := reqData["name"].(string); ok {
		template.Name = name
	}
	if desc, ok := reqData["description"].(string); ok {
		template.Description = desc
	}
	if typ, ok := reqData["type"].(string); ok {
		template.WorkflowType = typ
	}

	template.UpdatedAt = time.Now()
	
	// Parse steps
	var steps []models.WorkflowStepDef
	if stepsData, ok := reqData["steps"].([]interface{}); ok {
		for _, stepData := range stepsData {
			if stepMap, ok := stepData.(map[string]interface{}); ok {
				step := models.WorkflowStepDef{
					ID:           uuid.New(),
					WorkflowID:   template.ID,  
					StepOrder:    getIntOrZero(stepMap, "order"),
					StepType:     getStringOrEmpty(stepMap, "step_type"),  
					StepName:     getStringOrEmpty(stepMap, "name"),
					Description:  getStringOrEmpty(stepMap, "description"),
					Required:     getBoolOrFalse(stepMap, "required"),
					AutoTrigger:  getBoolOrFalse(stepMap, "auto_trigger"),
					AssignedRole: getStringOrEmpty(stepMap, "assignee_role"),
					DueDays:      getIntPointer(stepMap, "estimated_days"),  
					CreatedAt:    time.Now(),
				}

				steps = append(steps, step)
			}
		}
	}
	
	// Update template with steps
	if err := s.UpdateWorkflowTemplate(ctx, template, steps); err != nil {
		return nil, err
	}
	
	return template, nil
}

// DeleteTemplate deletes a workflow template
func (s *workflowService) DeleteTemplate(ctx context.Context, id string) error {
	templateID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	
	return s.DeleteWorkflowTemplate(ctx, templateID)
}

// ToggleTemplate toggles the active status of a template
func (s *workflowService) ToggleTemplate(ctx context.Context, id string) (*models.WorkflowTemplate, error) {
	templateID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	
	template, err := s.GetWorkflowTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}
	
	// Toggle active status
	//template.Status = !template.Status
	if template.Status == "draft" { 
		template.Status = "active"
	} else if template.Status == "active" { 
		template.Status = "inactive" 
	} else if template.Status == "inactive" {
		template.Status = "active"
	}
	template.UpdatedAt = time.Now()
	
	// Update without changing steps
	if err := s.UpdateWorkflowTemplate(ctx, template, nil); err != nil {
		return nil, err
	}
	
	return template, nil
}

// GetRecentAssignments returns recent workflow assignments
func (s *workflowService) GetRecentAssignments(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	workflows, err := s.ListWorkflows(ctx, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	
	// Convert to assignment format
	var assignments []map[string]interface{}
	count := 0
	
	// Sort by start date (most recent first)
	for i := len(workflows) - 1; i >= 0 && count < limit; i-- {
		wf := workflows[i]
		
		// Calculate progress
		var progress = float64(wf.OverallProgress)
		
		assignment := map[string]interface{}{
			"id":            wf.ID.String(),
			"employee_id":   wf.EmployeeID.String(),
			"employee_name": wf.EmployeeName, // Helper function
			"template_name": s.getTemplateNameFromID(ctx, wf.WorkflowTemplateID ), // Helper function
			"status":        wf.Status,
			"progress":      progress,
			"start_date":    wf.StartDate.Format(time.RFC3339),
			"due_date":      wf.ExpectedCompletionDate.Format(time.RFC3339),
		}
		
		assignments = append(assignments, assignment)
		count++
	}
	
	return assignments, nil
}

// Helper functions

func getStringOrEmpty(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getIntOrZero(m map[string]interface{}, key string) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	if val, ok := m[key].(int); ok {
		return val
	}
	return 0
}

// Add this helper function near the other helper functions (around line 1290)
func getIntPointer(m map[string]interface{}, key string) *int {
	if val, ok := m[key].(float64); ok {
		intVal := int(val)
		return &intVal
	}
	if val, ok := m[key].(int); ok {
		return &val
	}
	return nil
}


func getBoolOrFalse(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

func (s *workflowService) getTemplateNameFromID(ctx context.Context, templateID *uuid.UUID) string {
    // Handle nil pointer
    if templateID == nil {
        return "Direct Assignment"
    }
    
    // Dereference pointer when calling GetTemplateByID
    template, err := s.repos.Workflow.GetTemplateByID(ctx, *templateID)
    if err != nil {
        return "Template Not Found"
    }
    return template.Name
}