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

// RegisterOrganizationRoutes registers organization endpoints
func RegisterOrganizationRoutes(r chi.Router, services *service.Services) {
	r.Route("/organizations", func(r chi.Router) {
		r.Use(authMiddleware(services))
		
		// Organization CRUD
		r.Get("/", listOrganizationsHandler(services))
		r.Post("/", createOrganizationHandler(services))
		r.Get("/hierarchy", getOrganizationHierarchyHandler(services))
		r.Get("/{id}", getOrganizationHandler(services))
		r.Put("/{id}", updateOrganizationHandler(services))
		r.Delete("/{id}", deleteOrganizationHandler(services))
		
		// Employee assignments
		r.Post("/{id}/employees", assignEmployeeHandler(services))
		r.Post("/{id}/employees/bulk", bulkAssignEmployeesHandler(services))
		r.Delete("/{id}/employees/{employeeId}", unassignEmployeeHandler(services))
		r.Get("/{id}/employees", getOrganizationEmployeesHandler(services))
		
		// Stats
		r.Get("/{id}/stats", getOrganizationStatsHandler(services))
	})
	
	// Employee's organizations
	r.Get("/employees/{employeeId}/organizations", func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeId")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}
		
		orgs, err := services.Organization.GetEmployeeOrganizations(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get employee organizations")
			return
		}
		
		respondJSON(w, http.StatusOK, orgs)
	})
}

func createOrganizationHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateOrganizationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid user ID")
			return
		}
		
		org, err := services.Organization.CreateOrganization(r.Context(), &req, userID)
		if err != nil {
			if err == service.ErrOrganizationExists {
				respondError(w, http.StatusConflict, "organization code already exists")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to create organization")
			return
		}
		
		respondJSON(w, http.StatusCreated, org)
	}
}

func getOrganizationHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid organization ID")
			return
		}
		
		org, err := services.Organization.GetOrganization(r.Context(), id)
		if err != nil {
			if err == service.ErrOrganizationNotFound {
				respondError(w, http.StatusNotFound, "organization not found")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to get organization")
			return
		}
		
		respondJSON(w, http.StatusOK, org)
	}
}

func listOrganizationsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters := make(map[string]interface{})
		
		// Parse query parameters
		if parentIDStr := r.URL.Query().Get("parent_id"); parentIDStr != "" {
			parentID, err := uuid.Parse(parentIDStr)
			if err == nil {
				filters["parent_id"] = parentID
			}
		}
		
		if isActiveStr := r.URL.Query().Get("is_active"); isActiveStr != "" {
			if isActiveStr == "true" {
				filters["is_active"] = true
			} else if isActiveStr == "false" {
				filters["is_active"] = false
			}
		}
		
		if orgType := r.URL.Query().Get("type"); orgType != "" {
			filters["type"] = orgType
		}
		
		orgs, err := services.Organization.ListOrganizations(r.Context(), filters)
		if err != nil {
			log.Printf("ERROR listing organizations: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to list organizations")
			return
		}
		
		respondJSON(w, http.StatusOK, orgs)
	}
}

func updateOrganizationHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid organization ID")
			return
		}
		
		var req models.UpdateOrganizationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		org, err := services.Organization.UpdateOrganization(r.Context(), id, &req)
		if err != nil {
			if err == service.ErrOrganizationNotFound {
				respondError(w, http.StatusNotFound, "organization not found")
				return
			}
			if err == service.ErrCircularReference {
				respondError(w, http.StatusBadRequest, "circular reference detected")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to update organization")
			return
		}
		
		respondJSON(w, http.StatusOK, org)
	}
}

func deleteOrganizationHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid organization ID")
			return
		}
		
		err = services.Organization.DeleteOrganization(r.Context(), id)
		if err != nil {
			if err == service.ErrCannotDeleteOrg {
				respondError(w, http.StatusBadRequest, "cannot delete organization with employees or children")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to delete organization")
			return
		}
		
		w.WriteHeader(http.StatusNoContent)
	}
}

func getOrganizationHierarchyHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rootID *uuid.UUID
		if rootIDStr := r.URL.Query().Get("root_id"); rootIDStr != "" {
			parsed, err := uuid.Parse(rootIDStr)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid root ID")
				return
			}
			rootID = &parsed
		}
		
		hierarchy, err := services.Organization.GetHierarchy(r.Context(), rootID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get hierarchy")
			return
		}
		
		respondJSON(w, http.StatusOK, hierarchy)
	}
}

func assignEmployeeHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		orgID, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid organization ID")
			return
		}
		
		var req models.AssignEmployeeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid user ID")
			return
		}

		
		err = services.Organization.AssignEmployee(r.Context(), orgID, &req, userID)
		if err != nil {
			if err == service.ErrOrganizationNotFound {
				respondError(w, http.StatusNotFound, "organization not found")
				return
			}
			if err == service.ErrEmployeeNotFound {
				respondError(w, http.StatusNotFound, "employee not found")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to assign employee")
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "employee assigned successfully"})
	}
}

func bulkAssignEmployeesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		orgID, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid organization ID")
			return
		}
		
		var req models.BulkAssignEmployeesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid user ID")
			return
		}
		
		err = services.Organization.BulkAssignEmployees(r.Context(), orgID, &req, userID)
		if err != nil {
			if err == service.ErrOrganizationNotFound {
				respondError(w, http.StatusNotFound, "organization not found")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to bulk assign employees")
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{
			"message": "employees assigned successfully",
			"count":   string(rune(len(req.EmployeeIDs))),
		})
	}
}

func unassignEmployeeHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgIDStr := chi.URLParam(r, "id")
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid organization ID")
			return
		}
		
		empIDStr := chi.URLParam(r, "employeeId")
		empID, err := uuid.Parse(empIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}
		
		err = services.Organization.UnassignEmployee(r.Context(), orgID, empID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to unassign employee")
			return
		}
		
		w.WriteHeader(http.StatusNoContent)
	}
}

func getOrganizationEmployeesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		orgID, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid organization ID")
			return
		}
		
		employees, err := services.Organization.GetOrganizationEmployees(r.Context(), orgID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get employees")
			return
		}
		
		respondJSON(w, http.StatusOK, employees)
	}
}

func getOrganizationStatsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		orgID, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid organization ID")
			return
		}
		
		stats, err := services.Organization.GetOrganizationStats(r.Context(), orgID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get stats")
			return
		}
		
		respondJSON(w, http.StatusOK, stats)
	}
}
