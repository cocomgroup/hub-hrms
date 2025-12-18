package api

import (
	"context"
	"encoding/json"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
	"io"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(r chi.Router, services *service.Services) {
	r.Post("/auth/login", loginHandler(services))
}

// RegisterEmployeeRoutes registers employee routes
func RegisterEmployeeRoutes(r chi.Router, services *service.Services) {
	r.Route("/employees", func(r chi.Router) {
		r.Use(authMiddleware(services))
		r.Get("/", listEmployeesHandler(services))
		r.Post("/", createEmployeeHandler(services))
		r.Get("/{id}", getEmployeeHandler(services))
		r.Put("/{id}", updateEmployeeHandler(services))
	})
}

// RegisterOnboardingRoutes registers onboarding routes
func RegisterOnboardingRoutes(r chi.Router, services *service.Services) {
	r.Route("/onboarding", func(r chi.Router) {
		r.Use(authMiddleware(services))
		r.Get("/{employeeId}", getOnboardingTasksHandler(services))
		r.Put("/{employeeId}/tasks/{taskId}", updateOnboardingTaskHandler(services))
		r.Post("/{employeeId}/tasks", createOnboardingTaskHandler(services))
	})
}


// Middleware
func authMiddleware(services *service.Services) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondError(w, http.StatusUnauthorized, "invalid authorization header")
				return
			}

			token, err := services.Auth.ValidateToken(parts[1])
			if err != nil || !token.Valid {
				respondError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				respondError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			// Debug: log claims
			log.Printf("DEBUG JWT Middleware: Claims = %+v", claims)

			// Add user_id to context
			ctx := r.Context()
			if userID, ok := claims["user_id"].(string); ok {
				log.Printf("DEBUG JWT Middleware: Setting user_id in context: %s", userID)
				ctx = context.WithValue(ctx, "user_id", userID)
			} else {
				log.Printf("WARNING JWT Middleware: user_id not found in claims or wrong type")
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Auth handlers
func loginHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		response, err := services.Auth.Login(r.Context(), &req)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		respondJSON(w, http.StatusOK, response)
	}
}

// Employee handlers
func listEmployeesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters := make(map[string]interface{})
		status := r.URL.Query().Get("status")
		if status != "" {
			filters["status"] = status
		}

		employees, err := services.Employee.List(r.Context(), filters)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list employees")
			return
		}

		respondJSON(w, http.StatusOK, employees)
	}
}

func createEmployeeHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read and log the raw body for debugging
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			respondError(w, http.StatusBadRequest, "error reading request body")
			return
		}
		log.Printf("Raw request body: %s", string(bodyBytes))
		
		// Parse into a temporary struct that accepts string dates
		var reqBody struct {
			FirstName              string  `json:"first_name"`
			LastName               string  `json:"last_name"`
			Email                  string  `json:"email"`
			Phone                  string  `json:"phone"`
			DateOfBirth            *string `json:"date_of_birth,omitempty"`
			HireDate               string  `json:"hire_date"`
			Department             string  `json:"department"`
			Position               string  `json:"position"`
			EmploymentType         string  `json:"employment_type"`
			Status                 string  `json:"status"`
			StreetAddress          string  `json:"street_address"`
			City                   string  `json:"city"`
			State                  string  `json:"state"`
			ZipCode                string  `json:"zip_code"`
			Country                string  `json:"country"`
			EmergencyContactName   string  `json:"emergency_contact_name"`
			EmergencyContactPhone  string  `json:"emergency_contact_phone"`
		}
		
		if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
			log.Printf("Error decoding employee: %v", err)
			respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}
		
		// Parse hire_date from YYYY-MM-DD format
		hireDate, err := time.Parse("2006-01-02", reqBody.HireDate)
		if err != nil {
			log.Printf("Error parsing hire_date: %v", err)
			respondError(w, http.StatusBadRequest, "invalid hire_date format, use YYYY-MM-DD: "+err.Error())
			return
		}
		
		// Parse date_of_birth if provided
		var dateOfBirth *time.Time
		if reqBody.DateOfBirth != nil && *reqBody.DateOfBirth != "" {
			dob, err := time.Parse("2006-01-02", *reqBody.DateOfBirth)
			if err != nil {
				log.Printf("Error parsing date_of_birth: %v", err)
				respondError(w, http.StatusBadRequest, "invalid date_of_birth format, use YYYY-MM-DD: "+err.Error())
				return
			}
			dateOfBirth = &dob
		}
		
		// Create employee struct
		employee := models.Employee{
			FirstName:              reqBody.FirstName,
			LastName:               reqBody.LastName,
			Email:                  reqBody.Email,
			Phone:                  reqBody.Phone,
			DateOfBirth:            dateOfBirth,
			HireDate:               hireDate,
			Department:             reqBody.Department,
			Position:               reqBody.Position,
			EmploymentType:         reqBody.EmploymentType,
			Status:                 reqBody.Status,
			StreetAddress:          reqBody.StreetAddress,
			City:                   reqBody.City,
			State:                  reqBody.State,
			ZipCode:                reqBody.ZipCode,
			Country:                reqBody.Country,
			EmergencyContactName:   reqBody.EmergencyContactName,
			EmergencyContactPhone:  reqBody.EmergencyContactPhone,
		}

		// Log the received employee data for debugging
		log.Printf("Creating employee: %+v", employee)

		if err := services.Employee.Create(r.Context(), &employee); err != nil {
			log.Printf("Error creating employee: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to create employee: "+err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, employee)
	}
}

func getEmployeeHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		employee, err := services.Employee.GetByID(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "employee not found")
			return
		}

		respondJSON(w, http.StatusOK, employee)
	}
}

func updateEmployeeHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		// Parse into a temporary struct that accepts string dates
		var reqBody struct {
			FirstName              string  `json:"first_name"`
			LastName               string  `json:"last_name"`
			Email                  string  `json:"email"`
			Phone                  string  `json:"phone"`
			DateOfBirth            *string `json:"date_of_birth,omitempty"`
			HireDate               string  `json:"hire_date"`
			Department             string  `json:"department"`
			Position               string  `json:"position"`
			EmploymentType         string  `json:"employment_type"`
			Status                 string  `json:"status"`
			StreetAddress          string  `json:"street_address"`
			City                   string  `json:"city"`
			State                  string  `json:"state"`
			ZipCode                string  `json:"zip_code"`
			Country                string  `json:"country"`
			EmergencyContactName   string  `json:"emergency_contact_name"`
			EmergencyContactPhone  string  `json:"emergency_contact_phone"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}
		
		// Parse hire_date from YYYY-MM-DD format
		hireDate, err := time.Parse("2006-01-02", reqBody.HireDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid hire_date format, use YYYY-MM-DD: "+err.Error())
			return
		}
		
		// Parse date_of_birth if provided
		var dateOfBirth *time.Time
		if reqBody.DateOfBirth != nil && *reqBody.DateOfBirth != "" {
			dob, err := time.Parse("2006-01-02", *reqBody.DateOfBirth)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid date_of_birth format, use YYYY-MM-DD: "+err.Error())
				return
			}
			dateOfBirth = &dob
		}
		
		// Create employee struct
		employee := models.Employee{
			ID:                     id,
			FirstName:              reqBody.FirstName,
			LastName:               reqBody.LastName,
			Email:                  reqBody.Email,
			Phone:                  reqBody.Phone,
			DateOfBirth:            dateOfBirth,
			HireDate:               hireDate,
			Department:             reqBody.Department,
			Position:               reqBody.Position,
			EmploymentType:         reqBody.EmploymentType,
			Status:                 reqBody.Status,
			StreetAddress:          reqBody.StreetAddress,
			City:                   reqBody.City,
			State:                  reqBody.State,
			ZipCode:                reqBody.ZipCode,
			Country:                reqBody.Country,
			EmergencyContactName:   reqBody.EmergencyContactName,
			EmergencyContactPhone:  reqBody.EmergencyContactPhone,
		}

		if err := services.Employee.Update(r.Context(), &employee); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update employee: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, employee)
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


// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// Helper functions for context
func getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userIDStr, ok := ctx.Value("user_id").(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user_id not found in context")
	}
	return uuid.Parse(userIDStr)
}

// getUserIDFromContext extracts user ID from JWT context
func getUserIDFromJWTContext(r *http.Request) uuid.UUID {
	// Debug: log what's in context
	log.Printf("DEBUG: Attempting to get user_id from context")
	
	// This assumes the JWT middleware sets the user_id in context
	if userID := r.Context().Value("user_id"); userID != nil {
		log.Printf("DEBUG: Found user_id in context: %v (type: %T)", userID, userID)
		if id, ok := userID.(uuid.UUID); ok {
			log.Printf("DEBUG: Successfully parsed as uuid.UUID: %s", id)
			return id
		}
		if idStr, ok := userID.(string); ok {
			log.Printf("DEBUG: user_id is string: %s", idStr)
			if id, err := uuid.Parse(idStr); err == nil {
				log.Printf("DEBUG: Successfully parsed string to UUID: %s", id)
				return id
			} else {
				log.Printf("DEBUG: Failed to parse string to UUID: %v", err)
			}
		}
	} else {
		log.Printf("DEBUG: user_id not found in context")
	}
	
	// Fallback: try to get from claims in context
	if claims := r.Context().Value("claims"); claims != nil {
		log.Printf("DEBUG: Found claims in context")
		if claimsMap, ok := claims.(map[string]interface{}); ok {
			if userIDStr, ok := claimsMap["user_id"].(string); ok {
				if id, err := uuid.Parse(userIDStr); err == nil {
					log.Printf("DEBUG: Got user_id from claims: %s", id)
					return id
				}
			}
		}
	}
	
	// Return nil UUID as fallback (should be handled by auth middleware)
	log.Printf("WARNING: Returning nil UUID - user_id not found in context")
	return uuid.Nil
}

func getEmployeeIDFromContext(ctx context.Context) (uuid.UUID, error) {
	// Extract employee ID from JWT claims in context
	// This is a placeholder - implement based on your auth middleware
	claims, ok := ctx.Value("claims").(map[string]interface{})
	if !ok {
		return uuid.Nil, service.ErrUnauthorized
	}

	employeeIDStr, ok := claims["employee_id"].(string)
	if !ok {
		return uuid.Nil, service.ErrUnauthorized
	}

	return uuid.Parse(employeeIDStr)
}
