package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// RegisterCompensationRoutes registers all compensation-related routes
func RegisterCompensationRoutes(r chi.Router, services *service.Services) {
	r.Route("/compensation", func(r chi.Router) {
		r.Use(authMiddleware(services))

		// Compensation Plans
		r.Get("/plans", getAllPlansHandler(services))
		r.Post("/plans", createPlanHandler(services))
		r.Get("/plans/{id}", getPlanHandler(services))
		r.Put("/plans/{id}", updatePlanHandler(services))
		r.Delete("/plans/{id}", deletePlanHandler(services))
		r.Get("/plans/employee/{employee_id}", getPlansByEmployeeHandler(services))
		r.Get("/plans/employee/{employee_id}/active", getActivePlanHandler(services))
		r.Get("/plans/employee/{employee_id}/total", getTotalCompensationHandler(services))

		// Bonuses
		r.Get("/bonuses", getAllBonusesHandler(services))
		r.Post("/bonuses", createBonusHandler(services))
		r.Get("/bonuses/{id}", getBonusHandler(services))
		r.Put("/bonuses/{id}", updateBonusHandler(services))
		r.Delete("/bonuses/{id}", deleteBonusHandler(services))
		r.Get("/bonuses/employee/{employee_id}", getBonusesByEmployeeHandler(services))
		r.Get("/bonuses/status/{status}", getBonusesByStatusHandler(services))
		r.Post("/bonuses/{id}/approve", approveBonusHandler(services))
		r.Post("/bonuses/{id}/mark-paid", markBonusPaidHandler(services))
		r.Get("/bonuses/pending", getPendingBonusesHandler(services))
	})
}

// === COMPENSATION PLANS HANDLERS ===

// createPlanHandler creates a new compensation plan
func createPlanHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateCompensationPlanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		plan, err := services.Compensation.CreatePlan(r.Context(), &req)
		if err != nil {
			log.Printf("Error creating compensation plan: %v", err)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, plan)
	}
}

// getPlanHandler retrieves a compensation plan by ID
func getPlanHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid plan ID")
			return
		}

		plan, err := services.Compensation.GetPlan(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, plan)
	}
}

// getAllPlansHandler retrieves all compensation plans
func getAllPlansHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plans, err := services.Compensation.GetAllPlans(r.Context())
		if err != nil {
			log.Printf("Error getting compensation plans: %v", err)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, plans)
	}
}

// getPlansByEmployeeHandler retrieves all compensation plans for an employee
func getPlansByEmployeeHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employee_id")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		plans, err := services.Compensation.GetPlansByEmployee(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, plans)
	}
}

// getActivePlanHandler retrieves the active compensation plan for an employee
func getActivePlanHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employee_id")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		plan, err := services.Compensation.GetActivePlan(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, plan)
	}
}

// updatePlanHandler updates a compensation plan
func updatePlanHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid plan ID")
			return
		}

		var req models.UpdateCompensationPlanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		plan, err := services.Compensation.UpdatePlan(r.Context(), id, &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, plan)
	}
}

// deletePlanHandler deletes a compensation plan
func deletePlanHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid plan ID")
			return
		}

		if err := services.Compensation.DeletePlan(r.Context(), id); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// getTotalCompensationHandler calculates total annual compensation for an employee
func getTotalCompensationHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employee_id")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		total, err := services.Compensation.CalculateTotalCompensation(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]float64{
			"total_annual_compensation": total,
		})
	}
}

// === BONUS HANDLERS ===

// createBonusHandler creates a new bonus
func createBonusHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateBonusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		bonus, err := services.Compensation.CreateBonus(r.Context(), &req)
		if err != nil {
			log.Printf("Error creating bonus: %v", err)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, bonus)
	}
}

// getBonusHandler retrieves a bonus by ID
func getBonusHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid bonus ID")
			return
		}

		bonus, err := services.Compensation.GetBonus(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, bonus)
	}
}

// getAllBonusesHandler retrieves all bonuses
func getAllBonusesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bonuses, err := services.Compensation.GetAllBonuses(r.Context())
		if err != nil {
			log.Printf("Error getting bonuses: %v", err)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, bonuses)
	}
}

// getBonusesByEmployeeHandler retrieves all bonuses for an employee
func getBonusesByEmployeeHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employee_id")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		bonuses, err := services.Compensation.GetBonusesByEmployee(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, bonuses)
	}
}

// getBonusesByStatusHandler retrieves bonuses by status
func getBonusesByStatusHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := chi.URLParam(r, "status")

		bonuses, err := services.Compensation.GetBonusesByStatus(r.Context(), status)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, bonuses)
	}
}

// updateBonusHandler updates a bonus
func updateBonusHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid bonus ID")
			return
		}

		var req models.UpdateBonusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		bonus, err := services.Compensation.UpdateBonus(r.Context(), id, &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, bonus)
	}
}

// approveBonusHandler approves a bonus
func approveBonusHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid bonus ID")
			return
		}

		// Get approver ID from JWT context
		approverID := getUserIDFromJWTContext(r)
		if approverID == uuid.Nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		bonus, err := services.Compensation.ApproveBonus(r.Context(), id, approverID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, bonus)
	}
}

// markBonusPaidHandler marks a bonus as paid
func markBonusPaidHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid bonus ID")
			return
		}

		bonus, err := services.Compensation.MarkBonusPaid(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, bonus)
	}
}

// deleteBonusHandler deletes a bonus
func deleteBonusHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid bonus ID")
			return
		}

		if err := services.Compensation.DeleteBonus(r.Context(), id); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// getPendingBonusesHandler retrieves all pending bonuses
func getPendingBonusesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bonuses, err := services.Compensation.GetPendingBonuses(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, bonuses)
	}
}