package api

import (
	"encoding/json"
	"net/http"
	"log"
	"database/sql"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// RegisterPTORoutes registers all PTO-related routes
func RegisterPTORoutes(r chi.Router, services *service.Services) {
	r.Route("/pto", func(r chi.Router) {
		r.Use(authMiddleware(services))
		
		// Balance endpoints
		r.Get("/balance", getPTOBalanceHandler(services))
		
		// Request endpoints
		r.Get("/requests", getPTORequestsHandler(services))
		r.Post("/requests", createPTORequestHandler(services))
		r.Get("/requests/{id}", getPTORequestHandler(services))
		r.Post("/requests/{id}/review", reviewPTORequestHandler(services))
		
		// Manager endpoints
		r.Get("/pending", getPendingPTORequestsHandler(services))
	})
}

// getPTOBalanceHandler gets the PTO balance for the current employee
func getPTOBalanceHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		balance, err := services.PTO.GetBalance(r.Context(), employeeID)
		if err != nil {
			// ✅ FIX: Return default balance instead of 500 error
			if err.Error() == "no rows in result set" || err == sql.ErrNoRows {
				// Return default empty balance
				defaultBalance := &models.PTOBalance{
					ID:           uuid.New(),
					EmployeeID:   employeeID,
					VacationDays: 0,
					SickDays:     0,
					PersonalDays: 0,
					Year:         time.Now().Year(),
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				respondJSON(w, http.StatusOK, defaultBalance)
				return
			}
			
			log.Printf("Error getting pto balance: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to get PTO balance")
			return
		}

		respondJSON(w, http.StatusOK, balance)
	}
}

// getPTORequestsHandler gets all PTO requests for the current employee
func getPTORequestsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		requests, err := services.PTO.GetRequestsByEmployee(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get PTO requests")
			return
		}

		respondJSON(w, http.StatusOK, requests)
	}
}

// createPTORequestHandler creates a new PTO request
func createPTORequestHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req models.PTORequestCreate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Validate request
		if req.PTOType == "" {
			respondError(w, http.StatusBadRequest, "pto_type is required")
			return
		}

		if req.StartDate.IsZero() || req.EndDate.IsZero() {
			respondError(w, http.StatusBadRequest, "start_date and end_date are required")
			return
		}

		if req.EndDate.Before(req.StartDate) {
			respondError(w, http.StatusBadRequest, "end_date must be after start_date")
			return
		}

		if req.DaysRequested <= 0 {
			respondError(w, http.StatusBadRequest, "days_requested must be greater than 0")
			return
		}

		// Check balance
		balance, err := services.PTO.GetBalance(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to check balance")
			return
		}

		// Verify sufficient balance
		var available float64
		switch req.PTOType {
		case "vacation":
			available = balance.VacationDays
		case "sick":
			available = balance.SickDays
		case "personal":
			available = balance.PersonalDays
		default:
			respondError(w, http.StatusBadRequest, "invalid pto_type (must be vacation, sick, or personal)")
			return
		}

		if req.DaysRequested > available {
			respondError(w, http.StatusBadRequest, "insufficient PTO balance")
			return
		}

		// Create request
		ptoRequest, err := services.PTO.CreateRequest(r.Context(), employeeID, &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create PTO request")
			return
		}

		respondJSON(w, http.StatusCreated, ptoRequest)
	}
}

// getPTORequestHandler gets a specific PTO request
func getPTORequestHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid request ID")
			return
		}

		// Note: In production, verify the employee owns this request or is a manager
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Get the request
		requests, err := services.PTO.GetRequestsByEmployee(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get PTO request")
			return
		}

		// Find the specific request
		for _, req := range requests {
			if req.ID == id {
				respondJSON(w, http.StatusOK, req)
				return
			}
		}

		respondError(w, http.StatusNotFound, "PTO request not found")
	}
}

// reviewPTORequestHandler reviews (approves/denies) a PTO request
func reviewPTORequestHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		requestID, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid request ID")
			return
		}

		// Get reviewer ID (manager/HR)
		reviewerID := getUserIDFromJWTContext(r)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// TODO: Verify reviewer has permission (is manager or HR)

		var review models.PTORequestReview
		if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Validate status
		if review.Status != "approved" && review.Status != "denied" {
			respondError(w, http.StatusBadRequest, "status must be 'approved' or 'denied'")
			return
		}

		// Review the request
		if err := services.PTO.ReviewRequest(r.Context(), requestID, reviewerID, &review); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to review PTO request")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "PTO request reviewed successfully",
			"status":  review.Status,
		})
	}
}

// getPendingPTORequestsHandler gets all pending PTO requests (for managers)
func getPendingPTORequestsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Verify user has manager/HR role

		// For now, this would need to be implemented in the service layer
		// to get all pending requests across all employees
		respondError(w, http.StatusNotImplemented, "not implemented yet")
	}
}
