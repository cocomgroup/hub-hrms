package api

import (
	"encoding/json"
	"net/http"
	"log"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// RegisterOnboardingRoutes registers all onboarding-related routes
func RegisterOnboardingRoutes(r chi.Router, services *service.Services) {
	r.Route("/onboarding", func(r chi.Router) {
		r.Use(authMiddleware(services))
		
		// Dashboard - overview of all onboarding workflows
		r.Get("/dashboard", getOnboardingDashboardHandler(services))
		
		// Workflow management
		r.Post("/workflows", createOnboardingWorkflowHandler(services))
		r.Get("/workflows", listOnboardingWorkflowsHandler(services))
		r.Get("/workflows/{id}", getOnboardingWorkflowHandler(services))
		r.Put("/workflows/{id}", updateOnboardingWorkflowHandler(services))
		r.Delete("/workflows/{id}", deleteOnboardingWorkflowHandler(services))
		r.Get("/workflows/employee/{employeeId}", getOnboardingWorkflowByEmployeeHandler(services))
		
		// Task management
		r.Post("/tasks", createOnboardingTaskHandler(services))
		r.Get("/tasks/{id}", getOnboardingTaskHandler(services))
		r.Put("/tasks/{id}", updateOnboardingTaskHandler(services))
		r.Post("/tasks/{id}/complete", completeOnboardingTaskHandler(services))
		r.Delete("/tasks/{id}", deleteOnboardingTaskHandler(services))
		r.Get("/workflows/{workflowId}/tasks", listTasksByOnboardingWorkflowHandler(services))
		
		// AI interaction endpoints
		r.Post("/ai/chat", onboardingAIChatHandler(services))
		r.Get("/workflows/{workflowId}/interactions", listOnboardingInteractionsHandler(services))
		
		// Template management
		r.Get("/templates", listOnboardingTemplatesHandler(services))
		r.Get("/templates/{id}", getOnboardingTemplateHandler(services))
		
		// Milestone tracking
		r.Post("/milestones", createOnboardingMilestoneHandler(services))
		r.Get("/workflows/{workflowId}/milestones", listOnboardingMilestonesHandler(services))
		r.Put("/milestones/{id}/complete", completeOnboardingMilestoneHandler(services))
		
		// Statistics and analytics
		r.Get("/workflows/{workflowId}/statistics", getOnboardingStatisticsHandler(services))
	})
}

// ============================================================================
// DASHBOARD HANDLERS
// ============================================================================

// getOnboardingDashboardHandler returns dashboard overview for managers
func getOnboardingDashboardHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters := make(map[string]interface{})
		
		// Get current user's employee ID for manager filtering
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err == nil {
			filters["manager_id"] = employeeID
			log.Printf("DEBUG getOnboardingDashboardHandler: Filtering by manager_id=%s", employeeID)
		}
		
		// Allow filtering by status
		if status := r.URL.Query().Get("status"); status != "" {
			filters["status"] = status
		}
		
		dashboard, err := services.Onboarding.GetDashboard(r.Context(), filters)
		if err != nil {
			log.Printf("ERROR getOnboardingDashboardHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to get dashboard")
			return
		}
		
		respondJSON(w, http.StatusOK, dashboard)
	}
}

// ============================================================================
// WORKFLOW HANDLERS
// ============================================================================

// createOnboardingWorkflowHandler creates a new onboarding workflow
func createOnboardingWorkflowHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateWorkflowRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		// Get creator from context
		createdBy, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			log.Printf("WARNING createOnboardingWorkflowHandler: Could not get employee_id, using user_id")
			userID, userErr := getUserIDFromContext(r.Context())
			if userErr != nil {
				respondError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			createdBy = userID
		}
		
		log.Printf("DEBUG createOnboardingWorkflowHandler: Creating workflow for employee=%s, created_by=%s", req.EmployeeID, createdBy)
		
		workflow, err := services.Onboarding.CreateWorkflow(r.Context(), &req, createdBy)
		if err != nil {
			log.Printf("ERROR createOnboardingWorkflowHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to create workflow: "+err.Error())
			return
		}
		
		respondJSON(w, http.StatusCreated, workflow)
	}
}

// getOnboardingWorkflowHandler retrieves a specific workflow with all details
func getOnboardingWorkflowHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow ID")
			return
		}
		
		workflow, err := services.Onboarding.GetWorkflow(r.Context(), id)
		if err != nil {
			log.Printf("ERROR getOnboardingWorkflowHandler: %v", err)
			respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		
		respondJSON(w, http.StatusOK, workflow)
	}
}

// getOnboardingWorkflowByEmployeeHandler retrieves workflow for a specific employee
func getOnboardingWorkflowByEmployeeHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := uuid.Parse(chi.URLParam(r, "employeeId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}
		
		workflow, err := services.Onboarding.GetWorkflowByEmployee(r.Context(), employeeID)
		if err != nil {
			log.Printf("ERROR getOnboardingWorkflowByEmployeeHandler: %v", err)
			respondError(w, http.StatusNotFound, "workflow not found for employee")
			return
		}
		
		respondJSON(w, http.StatusOK, workflow)
	}
}

// listOnboardingWorkflowsHandler lists workflows with optional filters
func listOnboardingWorkflowsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters := make(map[string]interface{})
		
		// Parse query parameters
		if status := r.URL.Query().Get("status"); status != "" {
			filters["status"] = status
		}
		
		if managerIDStr := r.URL.Query().Get("manager_id"); managerIDStr != "" {
			if managerID, err := uuid.Parse(managerIDStr); err == nil {
				filters["manager_id"] = managerID
			}
		}
		
		if employeeIDStr := r.URL.Query().Get("employee_id"); employeeIDStr != "" {
			if employeeID, err := uuid.Parse(employeeIDStr); err == nil {
				filters["employee_id"] = employeeID
			}
		}
		
		workflows, err := services.Onboarding.ListWorkflows(r.Context(), filters)
		if err != nil {
			log.Printf("ERROR listOnboardingWorkflowsHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to list workflows")
			return
		}
		
		respondJSON(w, http.StatusOK, workflows)
	}
}

// updateOnboardingWorkflowHandler updates a workflow
func updateOnboardingWorkflowHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow ID")
			return
		}
		
		var req models.UpdateWorkflowRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		workflow, err := services.Onboarding.UpdateWorkflow(r.Context(), id, &req)
		if err != nil {
			log.Printf("ERROR updateOnboardingWorkflowHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to update workflow")
			return
		}
		
		respondJSON(w, http.StatusOK, workflow)
	}
}

// deleteOnboardingWorkflowHandler deletes a workflow (soft delete recommended)
func deleteOnboardingWorkflowHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow ID")
			return
		}
		
		// Note: Implement soft delete in service layer if needed
		err = services.Onboarding.DeleteWorkflow(r.Context(), id)
		if err != nil {
			log.Printf("ERROR deleteOnboardingWorkflowHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to delete workflow")
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "workflow deleted successfully"})
	}
}

// ============================================================================
// TASK HANDLERS
// ============================================================================

// createOnboardingTaskHandler creates a new task
func createOnboardingTaskHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		task, err := services.Onboarding.CreateTask(r.Context(), &req)
		if err != nil {
			log.Printf("ERROR createOnboardingTaskHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to create task")
			return
		}
		
		respondJSON(w, http.StatusCreated, task)
	}
}

// getOnboardingTaskHandler retrieves a specific task
func getOnboardingTaskHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid task ID")
			return
		}
		
		task, err := services.Onboarding.GetTask(r.Context(), id)
		if err != nil {
			log.Printf("ERROR getOnboardingTaskHandler: %v", err)
			respondError(w, http.StatusNotFound, "task not found")
			return
		}
		
		respondJSON(w, http.StatusOK, task)
	}
}

// updateOnboardingTaskHandler updates a task
func updateOnboardingTaskHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid task ID")
			return
		}
		
		var req models.UpdateTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		task, err := services.Onboarding.UpdateTask(r.Context(), id, &req)
		if err != nil {
			log.Printf("ERROR updateOnboardingTaskHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to update task")
			return
		}
		
		respondJSON(w, http.StatusOK, task)
	}
}

// completeOnboardingTaskHandler marks a task as complete
func completeOnboardingTaskHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid task ID")
			return
		}
		
		completedBy, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			log.Printf("WARNING completeOnboardingTaskHandler: Could not get employee_id")
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		var req models.CompleteTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Optional body, use defaults
			req = models.CompleteTaskRequest{}
		}
		
		err = services.Onboarding.CompleteTask(r.Context(), taskID, completedBy, &req)
		if err != nil {
			log.Printf("ERROR completeOnboardingTaskHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to complete task")
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "task completed successfully"})
	}
}

// deleteOnboardingTaskHandler deletes a task
func deleteOnboardingTaskHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid task ID")
			return
		}
		
		err = services.Onboarding.DeleteTask(r.Context(), id)
		if err != nil {
			log.Printf("ERROR deleteOnboardingTaskHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to delete task")
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "task deleted successfully"})
	}
}

// listTasksByOnboardingWorkflowHandler lists all tasks for a workflow
func listTasksByOnboardingWorkflowHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow ID")
			return
		}
		
		tasks, err := services.Onboarding.ListTasksByWorkflow(r.Context(), workflowID)
		if err != nil {
			log.Printf("ERROR listTasksByOnboardingWorkflowHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to list tasks")
			return
		}
		
		respondJSON(w, http.StatusOK, tasks)
	}
}

// ============================================================================
// AI INTERACTION HANDLERS
// ============================================================================

// onboardingAIChatHandler handles AI chat interactions
func onboardingAIChatHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.AIInteractionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		// Get employee ID for logging interaction
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			log.Printf("WARNING onboardingAIChatHandler: Could not get employee_id")
		} else {
			log.Printf("DEBUG onboardingAIChatHandler: AI chat request from employee=%s for workflow=%s", employeeID, req.WorkflowID)
		}
		
		response, err := services.Onboarding.HandleAIInteraction(r.Context(), &req)
		if err != nil {
			log.Printf("ERROR onboardingAIChatHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "AI interaction failed")
			return
		}
		
		respondJSON(w, http.StatusOK, response)
	}
}

// listOnboardingInteractionsHandler lists AI interactions for a workflow
func listOnboardingInteractionsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow ID")
			return
		}
		
		// Parse limit from query params (default 10)
		limit := 10
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}
		
		interactions, err := services.Onboarding.ListInteractionsByWorkflow(r.Context(), workflowID, limit)
		if err != nil {
			log.Printf("ERROR listOnboardingInteractionsHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to list interactions")
			return
		}
		
		respondJSON(w, http.StatusOK, interactions)
	}
}

// ============================================================================
// TEMPLATE HANDLERS
// ============================================================================

// listOnboardingTemplatesHandler lists available templates
func listOnboardingTemplatesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		department := r.URL.Query().Get("department")
		roleType := r.URL.Query().Get("role_type")
		
		templates, err := services.Onboarding.ListTemplates(r.Context(), department, roleType)
		if err != nil {
			log.Printf("ERROR listOnboardingTemplatesHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to list templates")
			return
		}
		
		respondJSON(w, http.StatusOK, templates)
	}
}

// getOnboardingTemplateHandler retrieves a specific template with items
func getOnboardingTemplateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid template ID")
			return
		}
		
		template, err := services.Onboarding.GetTemplate(r.Context(), id)
		if err != nil {
			log.Printf("ERROR getOnboardingTemplateHandler: %v", err)
			respondError(w, http.StatusNotFound, "template not found")
			return
		}
		
		respondJSON(w, http.StatusOK, template)
	}
}

// ============================================================================
// MILESTONE HANDLERS
// ============================================================================

// createOnboardingMilestoneHandler creates a new milestone
func createOnboardingMilestoneHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateMilestoneRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		milestone, err := services.Onboarding.CreateMilestone(r.Context(), &req)
		if err != nil {
			log.Printf("ERROR createOnboardingMilestoneHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to create milestone")
			return
		}
		
		respondJSON(w, http.StatusCreated, milestone)
	}
}

// listOnboardingMilestonesHandler lists milestones for a workflow
func listOnboardingMilestonesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow ID")
			return
		}
		
		milestones, err := services.Onboarding.ListMilestonesByWorkflow(r.Context(), workflowID)
		if err != nil {
			log.Printf("ERROR listOnboardingMilestonesHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to list milestones")
			return
		}
		
		respondJSON(w, http.StatusOK, milestones)
	}
}

// completeOnboardingMilestoneHandler marks a milestone as complete
func completeOnboardingMilestoneHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid milestone ID")
			return
		}
		
		err = services.Onboarding.CompleteMilestone(r.Context(), id)
		if err != nil {
			log.Printf("ERROR completeOnboardingMilestoneHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to complete milestone")
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "milestone completed successfully"})
	}
}

// ============================================================================
// STATISTICS HANDLERS
// ============================================================================

// getOnboardingStatisticsHandler retrieves statistics for a workflow
func getOnboardingStatisticsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID, err := uuid.Parse(chi.URLParam(r, "workflowId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid workflow ID")
			return
		}
		
		statistics, err := services.Onboarding.GetWorkflowStatistics(r.Context(), workflowID)
		if err != nil {
			log.Printf("ERROR getOnboardingStatisticsHandler: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to get statistics")
			return
		}
		
		respondJSON(w, http.StatusOK, statistics)
	}
}
