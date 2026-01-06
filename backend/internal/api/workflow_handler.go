package api

import (
	"encoding/json"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
	"net/http"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// RegisterWorkflowRoutes registers all workflow-related routes
func RegisterWorkflowRoutes(r chi.Router, services *service.Services) {
	r.Route("/workflows", func(r chi.Router) {
		r.Use(authMiddleware(services))
		r.Post("/", createWorkflowHandler(services))
		r.Get("/", listWorkflowsHandler(services))
		r.Get("/{workflowId}", getWorkflowHandler(services))
		r.Delete("/{workflowId}", cancelWorkflowHandler(services))
		
		// Step operations
		r.Put("/{workflowId}/steps/{stepId}/start", startStepHandler(services))
		r.Put("/{workflowId}/steps/{stepId}/complete", completeStepHandler(services))
		r.Put("/{workflowId}/steps/{stepId}/skip", skipStepHandler(services))

		// Onboarding-specific endpoints
		r.Get("/{workflowId}/tasks", getOnboardingTasksHandler(services))
		r.Get("/{workflowId}/milestones", getOnboardingMilestonesHandler(services))

		// Integration triggers
		r.Post("/{workflowId}/integrations/docusign", triggerDocuSignHandler(services))
		r.Post("/{workflowId}/integrations/background-check", triggerBackgroundCheckHandler(services))
		r.Post("/{workflowId}/integrations/doc-search", triggerDocSearchHandler(services))
		
		// Exception management
		r.Get("/{workflowId}/exceptions", getExceptionsHandler(services))
		r.Post("/{workflowId}/exceptions", createExceptionHandler(services))
		r.Put("/{workflowId}/exceptions/{exceptionId}/resolve", resolveExceptionHandler(services))
		
		// Progress monitoring
		r.Get("/{workflowId}/progress", getProgressHandler(services))
		r.Put("/{workflowId}/stage/advance", advanceStageHandler(services))

		// Template management endpoints
		r.Get("/templates", listWorkflowTemplatesHandler(services))
		r.Post("/templates", createWorkflowTemplateHandler(services))
		r.Get("/templates/{id}", getWorkflowTemplateHandler(services))
		r.Put("/templates/{id}", updateWorkflowTemplateHandler(services))
		r.Delete("/templates/{id}", deleteWorkflowTemplateHandler(services))

	})
}

// createWorkflowHandler creates a new workflow from template
func createWorkflowHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			EmployeeID   string `json:"employee_id"`
			TemplateName string `json:"template_name"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		employeeID, err := uuid.Parse(req.EmployeeID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee_id")
			return
		}
		
		// Get user from context (from JWT middleware)
		userID := getUserIDFromJWTContext(r)
		log.Printf("DEBUG createWorkflowHandler: userID from context = %s", userID)
		log.Printf("DEBUG createWorkflowHandler: userID is nil? %v", userID == uuid.Nil)
		
		workflow, err := services.Workflow.InitiateWorkflow(r.Context(), employeeID, req.TemplateName, userID)
		if err != nil {
			log.Printf("ERROR createWorkflowHandler: InitiateWorkflow failed with userID=%s, error=%v", userID, err)
			respondError(w, http.StatusInternalServerError, "failed to create workflow: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusCreated, workflow)
	}
}

// listWorkflowsHandler lists all workflows with optional filters
func listWorkflowsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters := make(map[string]interface{})
		
		// Parse query parameters
		if status := r.URL.Query().Get("status"); status != "" {
			filters["status"] = status
		}
		
		if employeeID := r.URL.Query().Get("employee_id"); employeeID != "" {
			if id, err := uuid.Parse(employeeID); err == nil {
				filters["employee_id"] = id
			}
		}
		
		workflows, err := services.Workflow.ListWorkflows(r.Context(), filters)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list workflows: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, workflows)
	}
}

// getWorkflowHandler retrieves workflow with all details
func getWorkflowHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow_id")
			return
		}
		
		workflow, err := services.Workflow.GetWorkflow(r.Context(), workflowID)
		if err != nil {
			respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		
		respondJSON(w, http.StatusOK, workflow)
	}
}

// cancelWorkflowHandler cancels an active workflow
func cancelWorkflowHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow_id")
			return
		}
		
		err = services.Workflow.CancelWorkflow(r.Context(), workflowID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to cancel workflow: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "workflow cancelled"})
	}
}

// startStepHandler starts a workflow step
func startStepHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stepID, err := uuid.Parse(chi.URLParam(r, "stepId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid step_id")
			return
		}
		
		err = services.Workflow.StartStep(r.Context(), stepID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to start step: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "step started"})
	}
}

// completeStepHandler marks a step as completed
func completeStepHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stepID, err := uuid.Parse(chi.URLParam(r, "stepId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid step_id")
			return
		}
		
		userID := getUserIDFromJWTContext(r)
		
		err = services.Workflow.CompleteStep(r.Context(), stepID, userID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to complete step: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "step completed"})
	}
}

// skipStepHandler skips a step with reason
func skipStepHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stepID, err := uuid.Parse(chi.URLParam(r, "stepId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid step_id")
			return
		}
		
		var req struct {
			Reason string `json:"reason"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		userID := getUserIDFromJWTContext(r)
		
		err = services.Workflow.SkipStep(r.Context(), stepID, userID, req.Reason)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to skip step: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "step skipped"})
	}
}

// triggerDocuSignHandler triggers DocuSign integration for a step
func triggerDocuSignHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			StepID       string `json:"step_id"`
			DocumentType string `json:"document_type"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		stepID, err := uuid.Parse(req.StepID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid step_id")
			return
		}
		
		err = services.Workflow.TriggerDocuSign(r.Context(), stepID, req.DocumentType)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to trigger DocuSign: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "DocuSign triggered successfully"})
	}
}

// triggerBackgroundCheckHandler triggers background check integration
func triggerBackgroundCheckHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			StepID     string   `json:"step_id"`
			CheckTypes []string `json:"check_types"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		stepID, err := uuid.Parse(req.StepID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid step_id")
			return
		}
		
		err = services.Workflow.TriggerBackgroundCheck(r.Context(), stepID, req.CheckTypes)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to trigger background check: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "Background check initiated successfully"})
	}
}

// triggerDocSearchHandler triggers document search integration
func triggerDocSearchHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			StepID string `json:"step_id"`
			Query  string `json:"query"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		stepID, err := uuid.Parse(req.StepID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid step_id")
			return
		}
		
		err = services.Workflow.TriggerDocSearch(r.Context(), stepID, req.Query)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to trigger document search: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "Document search completed successfully"})
	}
}

// getExceptionsHandler lists all exceptions for a workflow
func getExceptionsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := uuid.Parse(chi.URLParam(r, "workflowId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow_id")
			return
		}
		
		// Onboarding workflows use tasks/milestones, not exceptions
		// Return empty array for compatibility
		respondJSON(w, http.StatusOK, []interface{}{})
	}
}

// createExceptionHandler creates a new exception for a workflow
func createExceptionHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow_id")
			return
		}
		
		var req struct {
			StepID        *string `json:"step_id,omitempty"`
			ExceptionType string  `json:"exception_type"`
			Severity      string  `json:"severity"`
			Title         string  `json:"title"`
			Description   string  `json:"description,omitempty"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		exception := &models.WorkflowException{
			ExceptionType: req.ExceptionType,
			Severity:      req.Severity,
			Title:         req.Title,
			Description:   req.Description,
		}
		
		if req.StepID != nil {
			stepID, err := uuid.Parse(*req.StepID)
			if err == nil {
				exception.StepID = &stepID
			}
		}
		
		err = services.Workflow.RaiseException(r.Context(), workflowID, exception)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create exception: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusCreated, exception)
	}
}

// resolveExceptionHandler resolves an exception
func resolveExceptionHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exceptionID, err := uuid.Parse(chi.URLParam(r, "exceptionId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid exception_id")
			return
		}
		
		var req struct {
			Notes string `json:"notes"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		userID := getUserIDFromJWTContext(r)
		
		err = services.Workflow.ResolveException(r.Context(), exceptionID, userID, req.Notes)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to resolve exception: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "exception resolved"})
	}
}

// getProgressHandler retrieves workflow progress metrics
func getProgressHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow_id")
			return
		}
		
		progress, err := services.Workflow.CheckWorkflowProgress(r.Context(), workflowID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get progress: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, progress)
	}
}

// advanceStageHandler advances workflow to next stage
func advanceStageHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow_id")
			return
		}
		
		err = services.Workflow.AdvanceStage(r.Context(), workflowID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to advance stage: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "stage advanced"})
	}
}

// createWorkflowTemplateHandler creates a new workflow template/definition
func createWorkflowTemplateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name         string                    `json:"name"`
			Description  string                    `json:"description"`
			WorkflowType string                    `json:"workflow_type"`
			Status       string                    `json:"status"`
			Steps        []models.WorkflowStepDef  `json:"steps"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("ERROR: invalid request body: %v", err)
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Validate required fields
		if req.Name == "" {
			respondError(w, http.StatusBadRequest, "workflow name is required")
			return
		}

		if len(req.Steps) == 0 {
			respondError(w, http.StatusBadRequest, "at least one step is required")
			return
		}

		// Get user from context
		userID := getUserIDFromJWTContext(r)
		if userID == uuid.Nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Create workflow template
		template := &models.WorkflowTemplate{
			Name:         req.Name,
			Description:  req.Description,
			WorkflowType: req.WorkflowType,
			Status:       req.Status,
			CreatedBy:    userID,
		}

		// Create template with steps
		err := services.Workflow.CreateWorkflowTemplate(r.Context(), template, req.Steps)
		if err != nil {
			log.Printf("ERROR: failed to create workflow template: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to create workflow template")
			return
		}

		respondJSON(w, http.StatusCreated, template)
	}
}

// updateWorkflowTemplateHandler updates an existing workflow template
func updateWorkflowTemplateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templateID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid template id")
			return
		}

		var req struct {
			Name         string                    `json:"name"`
			Description  string                    `json:"description"`
			WorkflowType string                    `json:"workflow_type"`
			Status       string                    `json:"status"`
			Steps        []models.WorkflowStepDef  `json:"steps"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Get user from context
		userID := getUserIDFromJWTContext(r)
		if userID == uuid.Nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Update template
		template := &models.WorkflowTemplate{
			ID:           templateID,
			Name:         req.Name,
			Description:  req.Description,
			WorkflowType: req.WorkflowType,
			Status:       req.Status,
		}

		err = services.Workflow.UpdateWorkflowTemplate(r.Context(), template, req.Steps)
		if err != nil {
			log.Printf("ERROR: failed to update workflow template: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to update workflow template")
			return
		}

		respondJSON(w, http.StatusOK, template)
	}
}

// deleteWorkflowTemplateHandler deletes a workflow template
func deleteWorkflowTemplateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templateID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid template id")
			return
		}

		err = services.Workflow.DeleteWorkflowTemplate(r.Context(), templateID)
		if err != nil {
			log.Printf("ERROR: failed to delete workflow template: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to delete workflow template")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// listWorkflowTemplatesHandler lists all workflow templates
func listWorkflowTemplatesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates, err := services.Workflow.ListWorkflowTemplates(r.Context())
		if err != nil {
			log.Printf("ERROR: failed to list workflow templates: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to list workflow templates")
			return
		}

		respondJSON(w, http.StatusOK, templates)
	}
}

// getWorkflowTemplateHandler gets a single workflow template with steps
func getWorkflowTemplateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templateID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid template id")
			return
		}

		template, err := services.Workflow.GetWorkflowTemplate(r.Context(), templateID)
		if err != nil {
			log.Printf("ERROR: failed to get workflow template: %v", err)
			respondError(w, http.StatusNotFound, "workflow template not found")
			return
		}

		respondJSON(w, http.StatusOK, template)
	}
}

func getOnboardingTasksHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow_id")
			return
		}
		
		// Get workflow with details including tasks
		workflow, err := services.Workflow.GetWorkflow(r.Context(), workflowID)
		if err != nil {
			respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		
		respondJSON(w, http.StatusOK, workflow.Tasks)
	}
}

func getOnboardingMilestonesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow_id")
			return
		}
		
		// Get workflow with details including milestones
		workflow, err := services.Workflow.GetWorkflow(r.Context(), workflowID)
		if err != nil {
			respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		
		respondJSON(w, http.StatusOK, workflow.Milestones)
	}
}