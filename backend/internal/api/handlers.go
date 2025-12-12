package api

import (
	"context"
	"encoding/json"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
	"io"
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


// RegisterPTORoutes registers PTO routes
func RegisterPTORoutes(r chi.Router, services *service.Services) {
	r.Route("/pto", func(r chi.Router) {
		r.Use(authMiddleware(services))
		r.Get("/balance/{employeeId}", getPTOBalanceHandler(services))
		r.Post("/requests", createPTORequestHandler(services))
		r.Get("/requests/{employeeId}", getPTORequestsHandler(services))
		r.Put("/requests/{id}/review", reviewPTORequestHandler(services))
	})
}

// RegisterBenefitsRoutes registers benefits routes
func RegisterBenefitsRoutes(r chi.Router, services *service.Services) {
	r.Route("/benefits", func(r chi.Router) {
		r.Use(authMiddleware(services))
		r.Get("/plans", listBenefitPlansHandler(services))
		r.Post("/enrollments", createEnrollmentHandler(services))
		r.Get("/enrollments/{employeeId}", getEnrollmentsHandler(services))
	})
}

// RegisterPayrollRoutes registers payroll routes
func RegisterPayrollRoutes(r chi.Router, services *service.Services) {
	r.Route("/payroll", func(r chi.Router) {
		r.Use(authMiddleware(services))
		r.Get("/periods", listPayrollPeriodsHandler(services))
		r.Get("/stubs/{employeeId}", getPayStubsHandler(services))
		r.Post("/process/{periodId}", processPayrollHandler(services))
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


// PTO handlers
func getPTOBalanceHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeId")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		balance, err := services.PTO.GetBalance(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get PTO balance")
			return
		}

		respondJSON(w, http.StatusOK, balance)
	}
}

func createPTORequestHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.PTORequestCreate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// In a real app, get employee ID from JWT token
		employeeID := uuid.New() // Placeholder

		request, err := services.PTO.CreateRequest(r.Context(), employeeID, &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create PTO request")
			return
		}

		respondJSON(w, http.StatusCreated, request)
	}
}

func getPTORequestsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeId")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
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

func reviewPTORequestHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestIDStr := chi.URLParam(r, "id")
		requestID, err := uuid.Parse(requestIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid request ID")
			return
		}

		var review models.PTORequestReview
		if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// In a real app, get reviewer ID from JWT token
		reviewerID := uuid.New() // Placeholder

		if err := services.PTO.ReviewRequest(r.Context(), requestID, reviewerID, &review); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to review request")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"status": review.Status})
	}
}



// Payroll handlers
func listPayrollPeriodsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		periods, err := services.Payroll.ListPeriods(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list periods")
			return
		}

		respondJSON(w, http.StatusOK, periods)
	}
}

func getPayStubsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeId")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		stubs, err := services.Payroll.GetPayStubsByEmployee(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get pay stubs")
			return
		}

		respondJSON(w, http.StatusOK, stubs)
	}
}

func processPayrollHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		periodIDStr := chi.URLParam(r, "periodId")
		periodID, err := uuid.Parse(periodIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid period ID")
			return
		}

		// In a real app, get processor ID from JWT token
		processorID := uuid.New() // Placeholder

		if err := services.Payroll.ProcessPayroll(r.Context(), periodID, processorID); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to process payroll")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"status": "processed"})
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