package api

import (
	"encoding/json"
	"net/http"
	"log"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// RegisterProjectRoutes registers all project management routes
func RegisterProjectRoutes(r chi.Router, services *service.Services) {
	r.Route("/projects", func(r chi.Router) {
		r.Use(authMiddleware(services))
		
		// Project CRUD
		r.Get("/", listProjectsHandler(services))
		r.Post("/", createProjectHandler(services))
		r.Get("/{id}", getProjectHandler(services))
		r.Put("/{id}", updateProjectHandler(services))
		r.Delete("/{id}", deleteProjectHandler(services))
		
		// Project members
		r.Get("/{id}/members", getProjectMembersHandler(services))
		r.Post("/{id}/members", assignProjectMemberHandler(services))
		r.Delete("/{id}/members/{employeeId}", removeProjectMemberHandler(services))
		
		// Employee projects
		r.Get("/employee/{employeeId}", getEmployeeProjectsHandler(services))
		r.Get("/", listProjectsHandler(services))
		r.Post("/", createProjectHandler(services))
		r.Get("/{id}", getProjectHandler(services))
		r.Put("/{id}", updateProjectHandler(services))
		r.Delete("/{id}", deleteProjectHandler(services))
		
		// NEW: Project member management endpoints
		r.Get("/employee/{employeeID}", getEmployeeProjectsHandler(services))
		r.Post("/{projectID}/members", assignProjectMemberHandler(services))
		r.Delete("/{projectID}/members/{employeeID}", revokeProjectMemberHandler(services))
	})
	
	// Manager assignment endpoint
	r.Route("/managers", func(r chi.Router) {
		r.Use(authMiddleware(services))
		r.Post("/assign", assignEmployeeToManagerHandler(services))
	})
}

// createProjectHandler creates a new project
func createProjectHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (created_by)
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			log.Printf("ERROR: Failed to get user ID from context: %v", err)
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		log.Printf("DEBUG: User ID from context: %s", userID)

		var req models.CreateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("ERROR: Failed to decode request body: %v", err)
			respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
			return
		}

		log.Printf("DEBUG: Decoded project request successfully")
		log.Printf("DEBUG: Project Name: %s", req.Name)
		log.Printf("DEBUG: Project Status: %s", req.Status)
		log.Printf("DEBUG: Project Priority: %s", req.Priority)
		if req.ManagerID != nil {
			log.Printf("DEBUG: Manager ID: %s", req.ManagerID)
		}
		if req.StartDate != nil {
			log.Printf("DEBUG: Start Date: %s", req.StartDate)
		}
		if req.EndDate != nil {
			log.Printf("DEBUG: End Date: %s", req.EndDate)
		}
		if req.Budget != nil {
			log.Printf("DEBUG: Budget: %f", *req.Budget)
		}

		// Validate
		if req.Name == "" {
			log.Printf("ERROR: Project name is empty")
			respondError(w, http.StatusBadRequest, "project name is required")
			return
		}

		log.Printf("DEBUG: Calling CreateProject service...")
		project, err := services.Project.CreateProject(r.Context(), &req, userID)
		if err != nil {
			log.Printf("ERROR: CreateProject failed: %v", err)
			log.Printf("ERROR: Error type: %T", err)
			
			if err == service.ErrInvalidManager {
				log.Printf("ERROR: Invalid manager provided")
				respondError(w, http.StatusBadRequest, "invalid manager")
				return
			}
			
			// Log the full error for debugging
			log.Printf("ERROR: Full error details: %+v", err)
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create project: %v", err))
			return
		}

		log.Printf("SUCCESS: Created project with ID: %s", project.ID)
		respondJSON(w, http.StatusCreated, project)
	}
}


// listProjectsHandler lists all projects with optional filters
func listProjectsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		managerIDStr := r.URL.Query().Get("manager_id")

		var managerID *uuid.UUID
		if managerIDStr != "" {
			id, err := uuid.Parse(managerIDStr)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid manager ID")
				return
			}
			managerID = &id
		}

		projects, err := services.Project.ListProjects(r.Context(), status, managerID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list projects")
			return
		}

		respondJSON(w, http.StatusOK, projects)
	}
}

// getProjectHandler gets a project by ID with details
func getProjectHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid project ID")
			return
		}

		project, err := services.Project.GetProject(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "project not found")
			return
		}

		respondJSON(w, http.StatusOK, project)
	}
}

// updateProjectHandler updates a project
func updateProjectHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid project ID")
			return
		}

		var req models.UpdateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		project, err := services.Project.UpdateProject(r.Context(), id, &req)
		if err != nil {
			if err == service.ErrProjectNotFound {
				respondError(w, http.StatusNotFound, "project not found")
				return
			}
			if err == service.ErrInvalidManager {
				respondError(w, http.StatusBadRequest, "invalid manager")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to update project")
			return
		}

		respondJSON(w, http.StatusOK, project)
	}
}

// deleteProjectHandler deletes a project
func deleteProjectHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid project ID")
			return
		}

		if err := services.Project.DeleteProject(r.Context(), id); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to delete project")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// assignProjectMemberHandler assigns an employee to a project
// assignProjectMemberHandler assigns an employee to a project
func assignProjectMemberHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(chi.URLParam(r, "projectID"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid project ID")
			return
		}
		
		var req struct {
			EmployeeID string `json:"employee_id"`
			Role       string `json:"role"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		employeeID, err := uuid.Parse(req.EmployeeID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}
		
		// Default role to 'member' if not specified
		role := req.Role
		if role == "" {
			role = "member"
		}
		
		// Validate role
		validRoles := map[string]bool{
			"member":  true,
			"lead":    true,
			"manager": true,
		}
		if !validRoles[role] {
			respondError(w, http.StatusBadRequest, "invalid role: must be member, lead, or manager")
			return
		}
		
		// Assign the project member
		if err := services.Project.AssignMember(r.Context(), projectID, employeeID, role); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{
			"message": "project member assigned successfully",
		})
	}
}

// removeProjectMemberHandler removes an employee from a project
func removeProjectMemberHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectIDStr := chi.URLParam(r, "id")
		employeeIDStr := chi.URLParam(r, "employeeId")

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid project ID")
			return
		}

		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		if err := services.Project.RemoveMember(r.Context(), projectID, employeeID); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to remove member")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func revokeProjectMemberHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(chi.URLParam(r, "projectID"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid project ID")
			return
		}
		
		employeeID, err := uuid.Parse(chi.URLParam(r, "employeeID"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}
		
		// Revoke the project member
		if err := services.Project.RemoveMember(r.Context(), projectID, employeeID); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{
			"message": "project member revoked successfully",
		})
	}
}

// getProjectMembersHandler gets all members of a project
func getProjectMembersHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID, err := uuid.Parse(chi.URLParam(r, "projectID"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid project ID")
			return
		}
		
		// Get project members
		members, err := services.Project.GetProjectMembers(r.Context(), projectID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get project members")
			return
		}
		
		if members == nil {
			members = []*models.ProjectMember{}
		}
		
		respondJSON(w, http.StatusOK, members)
	}
}

// getEmployeeProjectsHandler gets all projects for an employee
func getEmployeeProjectsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := uuid.Parse(chi.URLParam(r, "employeeID"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}
		
		// Get projects for this employee
		projects, err := services.Project.GetEmployeeProjects(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get employee projects")
			return
		}
		
		if projects == nil {
			projects = []*models.Project{}
		}
		
		respondJSON(w, http.StatusOK, projects)
	}
}

// assignEmployeeToManagerHandler assigns an employee to a manager
func assignEmployeeToManagerHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.AssignManagerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := services.Project.AssignEmployeeToManager(r.Context(), &req); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "Employee assigned to manager successfully",
		})
	}
}