package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// Benefits handlers
func listBenefitPlansHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plans, err := services.Benefits.GetAllBenefitPlans(r.Context(), true) 
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list plans")
			return
		}

		respondJSON(w, http.StatusOK, plans)
	}
}

func createEnrollmentHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.EnrollmentCreate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// In a real app, get employee ID from JWT token
		employeeID := uuid.New() // Placeholder

		var input models.CreateEnrollmentInput
		input.PlanID = req.PlanID
		
		enrollment, err := services.Benefits.CreateEnrollment(r.Context(), employeeID, &input)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create enrollment")
			return
		}

		respondJSON(w, http.StatusCreated, enrollment)
	}
}

func getEnrollmentsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeId")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		enrollments, err := services.Benefits.GetEmployeeEnrollments(r.Context(), employeeID) 
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get enrollments")
			return
		}

		respondJSON(w, http.StatusOK, enrollments)
	}
}