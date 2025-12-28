package api

import (
	"encoding/json"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// RegisterOnboardingRoutes registers onboarding routes
func RegisterOnboardingRoutes(r chi.Router, services *service.Services) {
	r.Route("/onboarding", func(r chi.Router) {
		r.Use(authMiddleware(services))
		r.Get("/", listOnboardingHandler(services))
		r.Get("/{employeeId}/tasks", getOnboardingTasksHandler(services))
		r.Put("/{employeeId}/tasks/{taskId}", updateOnboardingTaskHandler(services))
		r.Post("/{employeeId}/tasks", createOnboardingTaskHandler(services))
	})
}

// List all onboarding tasks (for HR Dashboard)
func listOnboardingHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		
		// TODO: Implement services.Onboarding.ListAll(ctx, status)
		// For now, return empty array
		_ = status
		respondJSON(w, http.StatusOK, []interface{}{})
	}
}

// Onboarding handlers
func getOnboardingTasksHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeId")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		tasks, err := services.Onboarding.GetTasksByEmployee(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get tasks")
			return
		}

		respondJSON(w, http.StatusOK, tasks)
	}
}

func updateOnboardingTaskHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskIDStr := chi.URLParam(r, "taskId")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid task ID")
			return
		}

		// Parse update request - only status and completed_at typically updated
		var reqBody struct {
			Status      string  `json:"status"`
			CompletedAt *string `json:"completed_at,omitempty"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}
		
		// Parse completed_at if provided
		var completedAt *time.Time
		if reqBody.CompletedAt != nil && *reqBody.CompletedAt != "" {
			parsed, err := time.Parse(time.RFC3339, *reqBody.CompletedAt)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid completed_at format: "+err.Error())
				return
			}
			completedAt = &parsed
		}
		
		// Get existing task first
		existingTask, err := services.Onboarding.GetTaskByID(r.Context(), taskID)
		if err != nil {
			respondError(w, http.StatusNotFound, "task not found")
			return
		}
		
		// Update fields
		existingTask.Status = reqBody.Status
		if completedAt != nil {
			existingTask.CompletedAt = completedAt
		}

		if err := services.Onboarding.UpdateTask(r.Context(), existingTask); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update task: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, existingTask)
	}
}

func createOnboardingTaskHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeId")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		// Parse into temporary struct for date handling
		var reqBody struct {
			TaskName          string  `json:"task_name"`
			Description       *string `json:"description,omitempty"`
			Category          *string `json:"category,omitempty"`
			Status            string  `json:"status"`
			DueDate           *string `json:"due_date,omitempty"`
			DocumentsRequired bool    `json:"documents_required"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}
		
		// Parse due_date if provided
		var dueDate *time.Time
		if reqBody.DueDate != nil && *reqBody.DueDate != "" {
			parsed, err := time.Parse("2006-01-02", *reqBody.DueDate)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid due_date format, use YYYY-MM-DD: "+err.Error())
				return
			}
			dueDate = &parsed
		}
		
		task := models.OnboardingTask{
			EmployeeID:        employeeID,
			TaskName:          reqBody.TaskName,
			Description:       reqBody.Description,
			Category:          reqBody.Category,
			Status:            reqBody.Status,
			DueDate:           dueDate,
			DocumentsRequired: reqBody.DocumentsRequired,
		}

		if err := services.Onboarding.CreateTask(r.Context(), &task); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create task: "+err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, task)
	}
}
