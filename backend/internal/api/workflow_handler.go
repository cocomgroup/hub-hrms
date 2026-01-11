package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// RegisterWorkflowRoutes registers workflow management routes
func RegisterWorkflowRoutes(r chi.Router, services *service.Services) {
	r.Route("/workflows", func(r chi.Router) {
		// Add auth middleware to all workflow routes
		r.Use(authMiddleware(services))
		
		// Stats endpoint
		r.Get("/stats", getWorkflowStatsHandler(services))
		
		// Workflow assignment endpoint - POST /workflows
		r.Post("/", assignWorkflowHandler(services))
		
		// Template endpoints
		r.Get("/templates", listWorkflowTemplatesHandler(services))
		r.Post("/templates", createWorkflowTemplateHandler(services))
		r.Get("/templates/{id}", getWorkflowTemplateHandler(services))
		r.Put("/templates/{id}", updateWorkflowTemplateHandler(services))
		r.Delete("/templates/{id}", deleteWorkflowTemplateHandler(services))
		r.Put("/templates/{id}/toggle", toggleWorkflowTemplateHandler(services))
		
		// Recent assignments
		r.Get("/assignments/recent", getRecentAssignmentsHandler(services))
	})
}

// getWorkflowStatsHandler returns workflow statistics
func getWorkflowStatsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Get templates
		templates, err := services.Workflow.ListWorkflowTemplates(ctx)
		if err != nil {
			templates = []*models.WorkflowTemplate{}
		}
		
		activeTemplates := 0
		for _, t := range templates {
			// WorkflowTemplate uses Status field: "active", "inactive", "draft"
			if t.Status == "active" {
				activeTemplates++
			}
		}
		
		// Get workflows (OnboardingWorkflow = NewHireOnboarding)
		workflows, err := services.Workflow.ListWorkflows(ctx, map[string]interface{}{})
		if err != nil {
			workflows = []*models.OnboardingWorkflow{}
		}
		
		activeWorkflows := 0
		completedThisMonth := 0
		totalDays := 0
		
		now := time.Now()
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		
		for _, wf := range workflows {
			switch wf.Status {
			case "in_progress":
				activeWorkflows++
			case "completed":
				// NewHireOnboarding uses ActualCompletionDate, not CompletedDate
				if wf.ActualCompletionDate != nil && wf.ActualCompletionDate.After(monthStart) {
					completedThisMonth++
					days := int(wf.ActualCompletionDate.Sub(wf.StartDate).Hours() / 24)
					totalDays += days
				}
			}
		}
		
		avgDays := 0
		if completedThisMonth > 0 {
			avgDays = totalDays / completedThisMonth
		}
		
		stats := map[string]interface{}{
			"templates_count":      activeTemplates,
			"active_workflows":     activeWorkflows,
			"completed_this_month": completedThisMonth,
			"avg_completion_time":  avgDays,
			"pending_assignments":  0,
		}
		
		respondJSON(w, http.StatusOK, stats)
	}
}

// listWorkflowTemplatesHandler lists all workflow templates
func listWorkflowTemplatesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		activeOnly := r.URL.Query().Get("active") == "true"
		
		templates, err := services.Workflow.ListWorkflowTemplates(ctx)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list templates")
			return
		}
		
		if activeOnly {
			var activeTemplates []*models.WorkflowTemplate
			for _, t := range templates {
				if t.Status == "active" {
					activeTemplates = append(activeTemplates, t)
				}
			}
			templates = activeTemplates
		}
		
		respondJSON(w, http.StatusOK, templates)
	}
}

// createWorkflowTemplateHandler creates a new workflow template
func createWorkflowTemplateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		var req struct {
			Name          string `json:"name"`
			Description   string `json:"description"`
			Type          string `json:"type"`
			EstimatedDays int    `json:"estimated_days"`
			Steps         []struct {
				Name          string `json:"name"`
				Description   string `json:"description"`
				Order         int    `json:"order"`
				EstimatedDays int    `json:"estimated_days"`
				Required      bool   `json:"required"`
				AssigneeRole  string `json:"assignee_role"`
				StepType      string `json:"step_type"`
				AutoTrigger   bool   `json:"auto_trigger"`
			} `json:"steps"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		// Get user ID from context for CreatedBy
		userID, _ := getUserIDFromContext(ctx)
		
		template := &models.WorkflowTemplate{
			ID:           uuid.New(),
			Name:         req.Name,
			Description:  req.Description,
			WorkflowType: req.Type, // Maps "type" to "workflow_type"
			Status:       "active",  // New templates are active by default
			CreatedBy:    userID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		
		// Convert request steps to workflow step definitions
		var steps []models.WorkflowStepDef
		for _, s := range req.Steps {
			dueDays := req.EstimatedDays
			if s.EstimatedDays > 0 {
				dueDays = s.EstimatedDays
			}
			
			step := models.WorkflowStepDef{
				ID:           uuid.New(),
				WorkflowID:   template.ID,
				StepName:     s.Name,
				Description:  s.Description,
				StepOrder:    s.Order,
				StepType:     s.StepType,
				Required:     s.Required,
				AssignedRole: s.AssigneeRole,
				AutoTrigger:  s.AutoTrigger,
				DueDays:      &dueDays,
				CreatedAt:    time.Now(),
			}
			steps = append(steps, step)
		}
		
		if err := services.Workflow.CreateWorkflowTemplate(ctx, template, steps); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create template: "+err.Error())
			return
		}
		
		// Load steps into response
		template.Steps = steps
		
		respondJSON(w, http.StatusCreated, template)
	}
}

// getWorkflowTemplateHandler retrieves a specific template
func getWorkflowTemplateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := chi.URLParam(r, "id")
		
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid template ID")
			return
		}
		
		template, err := services.Workflow.GetWorkflowTemplate(ctx, id)
		if err != nil {
			respondError(w, http.StatusNotFound, "template not found")
			return
		}
		
		respondJSON(w, http.StatusOK, template)
	}
}

// updateWorkflowTemplateHandler updates a template
func updateWorkflowTemplateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := chi.URLParam(r, "id")
		
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid template ID")
			return
		}
		
		var req struct {
			Name          string `json:"name"`
			Description   string `json:"description"`
			Type          string `json:"type"`
			EstimatedDays int    `json:"estimated_days"`
			Steps         []struct {
				Name          string `json:"name"`
				Description   string `json:"description"`
				Order         int    `json:"order"`
				EstimatedDays int    `json:"estimated_days"`
				Required      bool   `json:"required"`
				AssigneeRole  string `json:"assignee_role"`
				StepType      string `json:"step_type"`
				AutoTrigger   bool   `json:"auto_trigger"`
			} `json:"steps"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		// Get existing template
		template, err := services.Workflow.GetWorkflowTemplate(ctx, id)
		if err != nil {
			respondError(w, http.StatusNotFound, "template not found")
			return
		}
		
		// Update fields
		template.Name = req.Name
		template.Description = req.Description
		template.WorkflowType = req.Type
		template.UpdatedAt = time.Now()
		
		// Convert request steps
		var steps []models.WorkflowStepDef
		for _, s := range req.Steps {
			dueDays := req.EstimatedDays
			if s.EstimatedDays > 0 {
				dueDays = s.EstimatedDays
			}
			
			step := models.WorkflowStepDef{
				ID:           uuid.New(),
				WorkflowID:   template.ID,
				StepName:     s.Name,
				Description:  s.Description,
				StepOrder:    s.Order,
				StepType:     s.StepType,
				Required:     s.Required,
				AssignedRole: s.AssigneeRole,
				AutoTrigger:  s.AutoTrigger,
				DueDays:      &dueDays,
				CreatedAt:    time.Now(),
			}
			steps = append(steps, step)
		}
		
		if err := services.Workflow.UpdateWorkflowTemplate(ctx, template, steps); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update template")
			return
		}
		
		// Load steps into response
		template.Steps = steps
		
		respondJSON(w, http.StatusOK, template)
	}
}

// deleteWorkflowTemplateHandler deletes a template
func deleteWorkflowTemplateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := chi.URLParam(r, "id")
		
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid template ID")
			return
		}
		
		if err := services.Workflow.DeleteWorkflowTemplate(ctx, id); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to delete template")
			return
		}
		
		w.WriteHeader(http.StatusNoContent)
	}
}

// toggleWorkflowTemplateHandler toggles template active status
func toggleWorkflowTemplateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := chi.URLParam(r, "id")
		
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid template ID")
			return
		}
		
		template, err := services.Workflow.GetWorkflowTemplate(ctx, id)
		if err != nil {
			respondError(w, http.StatusNotFound, "template not found")
			return
		}
		
		// Toggle status: active <-> inactive
		if template.Status == "active" {
			template.Status = "inactive"
		} else {
			template.Status = "active"
		}
		template.UpdatedAt = time.Now()
		
		// Update without changing steps (pass nil for steps to keep existing)
		if err := services.Workflow.UpdateWorkflowTemplate(ctx, template, nil); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to toggle template")
			return
		}
		
		respondJSON(w, http.StatusOK, template)
	}
}

// getRecentAssignmentsHandler gets recent workflow assignments
func getRecentAssignmentsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		workflows, err := services.Workflow.ListWorkflows(ctx, map[string]interface{}{})
		if err != nil {
			// Log the error for debugging
			log.Printf("ERROR: Failed to list workflows: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to get workflows")
			return
		}
		
		// If no workflows, return empty array
		if len(workflows) == 0 {
			respondJSON(w, http.StatusOK, []map[string]interface{}{})
			return
		}
		
		// Convert to assignment format and limit to 10 most recent
		var assignments []map[string]interface{}
		limit := 10
		count := 0
		
		// Reverse iterate to get most recent first
		for i := len(workflows) - 1; i >= 0 && count < limit; i-- {
			wf := workflows[i]
			
			// Skip if workflow is nil
			if wf == nil {
				continue
			}
			
			// NewHireOnboarding has OverallProgress field (0-100)
			progress := float64(wf.OverallProgress)
			
			dueDate := ""
			if wf.ExpectedCompletionDate != nil {
				dueDate = wf.ExpectedCompletionDate.Format(time.RFC3339)
			}
			
			// Get employee name - handle case where it might not be set
			employeeName := wf.EmployeeName
			if employeeName == "" {
				// Try to fetch employee name
				if employee, err := services.Employee.GetByID(ctx, wf.EmployeeID); err == nil {
					employeeName = employee.FirstName + " " + employee.LastName
				} else {
					employeeName = "Unknown Employee"
				}
			}
			
			assignment := map[string]interface{}{
				"id":            wf.ID.String(),
				"employee_id":   wf.EmployeeID.String(),
				"employee_name": employeeName,
				"template_name": "Onboarding", // Would need to fetch from template if linked
				"status":        wf.Status,
				"progress":      progress,
				"start_date":    wf.StartDate.Format(time.RFC3339),
				"due_date":      dueDate,
			}
			
			assignments = append(assignments, assignment)
			count++
		}
		
		respondJSON(w, http.StatusOK, assignments)
	}
}

// assignWorkflowHandler assigns a workflow template to an employee
func assignWorkflowHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		var req struct {
			EmployeeID      string `json:"employee_id"`
			TemplateID      string `json:"template_id"`
			StartDate       string `json:"start_date"`
			Notes           string `json:"notes"`
			AssignToManager bool   `json:"assign_to_manager"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		// Validate required fields
		if req.EmployeeID == "" || req.TemplateID == "" {
			respondError(w, http.StatusBadRequest, "employee_id and template_id are required")
			return
		}
		
		// Parse employee UUID
		employeeID, err := uuid.Parse(req.EmployeeID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee_id")
			return
		}
		
		// Parse template UUID
		templateID, err := uuid.Parse(req.TemplateID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid template_id")
			return
		}
		
		// Get user ID from context
		userID, err := getUserIDFromContext(ctx)
		if err != nil {
			// If no user in context, log and continue with nil creator
			log.Printf("WARN: No user ID in context: %v", err)
			userID = uuid.Nil
		}
		
		// Verify user exists in database before using as foreign key
		// If user doesn't exist, we'll set created_by to nil to avoid FK constraint violation
		//var createdByPtr *uuid.UUID
		/*
		createdByPtr = *uuid.UUID
		if userID != uuid.Nil {
			// Check if user exists (optional - if you have a User service)
			// For now, we'll just use the userID and let the service handle it
			createdByPtr = &userID
		}
		*/
		// Get template to get the template name
		template, err := services.Workflow.GetWorkflowTemplate(ctx, templateID)
		if err != nil {
			respondError(w, http.StatusNotFound, "template not found")
			return
		}
		
		// Use the service method to initiate the workflow
		// Note: InitiateWorkflow creates the workflow with all necessary data
		// If notes are needed, they would be added to the InitiateWorkflow signature
		workflow, err := services.Workflow.InitiateWorkflow(ctx, employeeID, template.Name, userID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to assign workflow: "+err.Error())
			return
		}
		
		// TODO: If assign_to_manager is true, send notification to manager
		// This would require integration with notification service
		if req.AssignToManager {
			// Get employee to find manager
			employee, err := services.Employee.GetByID(ctx, employeeID)
			if err == nil && employee.ManagerID != nil {
				// TODO: Send notification to manager
				// services.Notification.NotifyManager(ctx, *employee.ManagerID, workflow)
			}
		}
		
		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"id":          workflow.ID.String(),
			"employee_id": workflow.EmployeeID.String(),
			"status":      workflow.Status,
			"start_date":  workflow.StartDate.Format(time.RFC3339),
			"due_date":    workflow.ExpectedCompletionDate.Format(time.RFC3339),
			"message":     "Workflow assigned successfully",
		})
	}
}