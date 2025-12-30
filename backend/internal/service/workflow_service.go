package service

import (
	"context"
	"fmt"
	"log"
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
	GetWorkflow(ctx context.Context, workflowID uuid.UUID) (*models.WorkflowWithDetails, error)
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

// InitiateWorkflow creates a new workflow from a template
func (s *workflowService) InitiateWorkflow(ctx context.Context, employeeID uuid.UUID, templateName string, createdBy uuid.UUID) (*models.OnboardingWorkflow, error) {
	log.Printf("DEBUG InitiateWorkflow: employeeID=%s, templateName=%s, createdBy=%s", employeeID, templateName, createdBy)
	log.Printf("DEBUG InitiateWorkflow: createdBy is nil? %v", createdBy == uuid.Nil)
	
	// Get employee details
	employee, err := s.repos.Employee.GetByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}
	
	// Create workflow
	workflow := &models.OnboardingWorkflow{
		EmployeeID:   employeeID,
		TemplateName: templateName,
		Status:       "active",
		CurrentStage: "pre-boarding",
		CreatedBy:    &createdBy,
	}
	
	log.Printf("DEBUG InitiateWorkflow: workflow.CreatedBy = %v", workflow.CreatedBy)
	if workflow.CreatedBy != nil {
		log.Printf("DEBUG InitiateWorkflow: *workflow.CreatedBy = %s", *workflow.CreatedBy)
	}
	
	// Calculate expected completion (30 days from now)
	expectedCompletion := time.Now().Add(30 * 24 * time.Hour)
	workflow.ExpectedCompletion = &expectedCompletion
	
	log.Printf("DEBUG InitiateWorkflow: About to call CreateWorkflow")
	err = s.repos.Workflow.CreateWorkflow(ctx, workflow)
	if err != nil {
		log.Printf("ERROR InitiateWorkflow: CreateWorkflow failed: %v", err)
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}
	
	// Generate steps from template
	steps, err := s.generateStepsFromTemplate(templateName, workflow.ID, employee)
	if err != nil {
		return nil, fmt.Errorf("failed to generate steps: %w", err)
	}
	
	// Create all steps
	for _, step := range steps {
		if err := s.repos.Workflow.CreateStep(ctx, step); err != nil {
			return nil, fmt.Errorf("failed to create step: %w", err)
		}
	}
	
	return workflow, nil
}

// generateStepsFromTemplate creates steps based on template
func (s *workflowService) generateStepsFromTemplate(templateName string, workflowID uuid.UUID, employee *models.Employee) ([]*models.WorkflowStep, error) {
	var steps []*models.WorkflowStep
	
	switch templateName {
	case "software-engineer":
		steps = s.createSoftwareEngineerTemplate(workflowID, employee)
	case "sales-representative":
		steps = s.createSalesRepTemplate(workflowID, employee)
	case "manager":
		steps = s.createManagerTemplate(workflowID, employee)
	case "generic":
		steps = s.createGenericTemplate(workflowID, employee)
	default:
		steps = s.createGenericTemplate(workflowID, employee)
	}
	
	return steps, nil
}

// createSoftwareEngineerTemplate generates steps for software engineer onboarding
func (s *workflowService) createSoftwareEngineerTemplate(workflowID uuid.UUID, employee *models.Employee) []*models.WorkflowStep {
	steps := []*models.WorkflowStep{
		// Pre-boarding stage
		{
			WorkflowID:      workflowID,
			StepOrder:       1,
			StepName:        "Send Offer Letter",
			StepType:        "integration",
			Stage:           "pre-boarding",
			Status:          "pending",
			Description:     "Send offer letter via DocuSign for signature",
			IntegrationType: "docusign",
			IntegrationConfig: map[string]interface{}{
				"document_type": "offer-letter",
				"template_id":   "offer-letter-tech",
			},
			DueDate: timePtr(time.Now().Add(-7 * 24 * time.Hour)), // 7 days before start
		},
		{
			WorkflowID:      workflowID,
			StepOrder:       2,
			StepName:        "Send I-9 Form",
			StepType:        "integration",
			Stage:           "pre-boarding",
			Status:          "pending",
			Description:     "Send I-9 Employment Eligibility form via DocuSign",
			IntegrationType: "docusign",
			IntegrationConfig: map[string]interface{}{
				"document_type": "i9",
			},
			DueDate: timePtr(time.Now().Add(-5 * 24 * time.Hour)),
		},
		{
			WorkflowID:      workflowID,
			StepOrder:       3,
			StepName:        "Send W-4 Form",
			StepType:        "integration",
			Stage:           "pre-boarding",
			Status:          "pending",
			Description:     "Send W-4 tax withholding form via DocuSign",
			IntegrationType: "docusign",
			IntegrationConfig: map[string]interface{}{
				"document_type": "w4",
			},
			DueDate: timePtr(time.Now().Add(-5 * 24 * time.Hour)),
		},
		{
			WorkflowID:      workflowID,
			StepOrder:       4,
			StepName:        "Initiate Background Check",
			StepType:        "integration",
			Stage:           "pre-boarding",
			Status:          "pending",
			Description:     "Start criminal and employment background check",
			IntegrationType: "background-check",
			IntegrationConfig: map[string]interface{}{
				"check_types": []string{"criminal", "employment"},
			},
			DueDate: timePtr(time.Now().Add(-7 * 24 * time.Hour)),
		},
		{
			WorkflowID:  workflowID,
			StepOrder:   5,
			StepName:    "Order Equipment",
			StepType:    "manual",
			Stage:       "pre-boarding",
			Status:      "pending",
			Description: "Order laptop, monitor, keyboard, mouse",
			DueDate:     timePtr(time.Now().Add(-5 * 24 * time.Hour)),
		},
		{
			WorkflowID:      workflowID,
			StepOrder:       6,
			StepName:        "Fetch Onboarding Documents",
			StepType:        "integration",
			Stage:           "pre-boarding",
			Status:          "pending",
			Description:     "Retrieve employee handbook and policies",
			IntegrationType: "doc-search",
			IntegrationConfig: map[string]interface{}{
				"query":         "handbook",
				"document_type": "handbook",
			},
			DueDate: timePtr(time.Now().Add(-3 * 24 * time.Hour)),
		},
		{
			WorkflowID:  workflowID,
			StepOrder:   7,
			StepName:    "Create Email Account",
			StepType:    "manual",
			Stage:       "pre-boarding",
			Status:      "pending",
			Description: "Setup company email and calendar access",
			DueDate:     timePtr(time.Now().Add(-2 * 24 * time.Hour)),
		},
		{
			WorkflowID:  workflowID,
			StepOrder:   8,
			StepName:    "Setup Development Environment Access",
			StepType:    "manual",
			Stage:       "pre-boarding",
			Status:      "pending",
			Description: "Create GitHub, Jira, and Confluence accounts",
			Dependencies: []uuid.UUID{}, // Will be set after steps are created
			DueDate:     timePtr(time.Now().Add(-1 * 24 * time.Hour)),
		},
		
		// Day 1 stage
		{
			WorkflowID:  workflowID,
			StepOrder:   9,
			StepName:    "Send Welcome Email",
			StepType:    "integration",
			Stage:       "day-1",
			Status:      "blocked",
			Description: "Send welcome email with first day instructions",
			DueDate:     timePtr(time.Now()),
		},
		{
			WorkflowID:  workflowID,
			StepOrder:   10,
			StepName:    "Office Tour",
			StepType:    "manual",
			Stage:       "day-1",
			Status:      "blocked",
			Description: "Conduct office tour and introductions",
			DueDate:     timePtr(time.Now()),
		},
		{
			WorkflowID:  workflowID,
			StepOrder:   11,
			StepName:    "IT Setup - Laptop Configuration",
			StepType:    "manual",
			Stage:       "day-1",
			Status:      "blocked",
			Description: "Setup laptop with required software and tools",
			DueDate:     timePtr(time.Now()),
		},
		{
			WorkflowID:  workflowID,
			StepOrder:   12,
			StepName:    "Building Access Card",
			StepType:    "manual",
			Stage:       "day-1",
			Status:      "blocked",
			Description: "Issue building access card and parking pass",
			DueDate:     timePtr(time.Now()),
		},
		
		// Week 1 stage
		{
			WorkflowID:      workflowID,
			StepOrder:       13,
			StepName:        "Benefits Enrollment",
			StepType:        "integration",
			Stage:           "week-1",
			Status:          "blocked",
			Description:     "Complete benefits enrollment forms",
			IntegrationType: "docusign",
			IntegrationConfig: map[string]interface{}{
				"document_type": "benefits",
			},
			DueDate: timePtr(time.Now().Add(3 * 24 * time.Hour)),
		},
		{
			WorkflowID:  workflowID,
			StepOrder:   14,
			StepName:    "Codebase Onboarding",
			StepType:    "manual",
			Stage:       "week-1",
			Status:      "blocked",
			Description: "Review codebase architecture and setup local environment",
			DueDate:     timePtr(time.Now().Add(4 * 24 * time.Hour)),
		},
		{
			WorkflowID:  workflowID,
			StepOrder:   15,
			StepName:    "First Code Review",
			StepType:    "manual",
			Stage:       "week-1",
			Status:      "blocked",
			Description: "Submit first pull request and participate in code review",
			DueDate:     timePtr(time.Now().Add(5 * 24 * time.Hour)),
		},
		
		// Month 1 stage
		{
			WorkflowID:  workflowID,
			StepOrder:   16,
			StepName:    "30-Day Check-in",
			StepType:    "manual",
			Stage:       "month-1",
			Status:      "blocked",
			Description: "Conduct 30-day check-in meeting with manager",
			DueDate:     timePtr(time.Now().Add(30 * 24 * time.Hour)),
		},
		{
			WorkflowID:  workflowID,
			StepOrder:   17,
			StepName:    "Goal Setting Session",
			StepType:    "manual",
			Stage:       "month-1",
			Status:      "blocked",
			Description: "Set quarterly goals and expectations",
			DueDate:     timePtr(time.Now().Add(30 * 24 * time.Hour)),
		},
	}
	
	return steps
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
func (s *workflowService) GetWorkflow(ctx context.Context, workflowID uuid.UUID) (*models.WorkflowWithDetails, error) {
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
	
	progress := &WorkflowProgress{
		WorkflowID:   workflowID,
		TotalSteps:   len(steps),
		CurrentStage: workflow.CurrentStage,
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
	progress.DaysElapsed = int(time.Since(workflow.StartedAt).Hours() / 24)
	
	// Calculate expected days
	if workflow.ExpectedCompletion != nil {
		progress.ExpectedDays = int(workflow.ExpectedCompletion.Sub(workflow.StartedAt).Hours() / 24)
		
		// Check if on track
		if progress.ExpectedDays > 0 {
			expectedProgress := (progress.DaysElapsed * 100) / progress.ExpectedDays
			progress.IsOnTrack = progress.ProgressPercentage >= expectedProgress
		}
	}
	
	// Count open exceptions
	for _, ex := range exceptions {
		if ex.ResolutionStatus == "open" || ex.ResolutionStatus == "in-progress" {
			progress.OpenExceptions++
		}
	}
	
	return progress, nil
}

// AdvanceStage advances workflow to next stage
func (s *workflowService) AdvanceStage(ctx context.Context, workflowID uuid.UUID) error {
	workflow, err := s.repos.Workflow.GetWorkflow(ctx, workflowID)
	if err != nil {
		return err
	}
	
	// Determine next stage
	var nextStage string
	switch workflow.CurrentStage {
	case "pre-boarding":
		nextStage = "day-1"
	case "day-1":
		nextStage = "week-1"
	case "week-1":
		nextStage = "month-1"
	case "month-1":
		nextStage = "completed"
		// Mark workflow as completed
		s.repos.Workflow.UpdateWorkflowStatus(ctx, workflowID, "completed")
	default:
		return fmt.Errorf("unknown stage: %s", workflow.CurrentStage)
	}
	
	return s.repos.Workflow.UpdateWorkflowStage(ctx, workflowID, nextStage)
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
	
	// Check if all steps in current stage are completed or skipped
	allDone := true
	for _, step := range steps {
		if step.Stage == workflow.CurrentStage {
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
