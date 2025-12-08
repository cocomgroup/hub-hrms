package api

import (
	"encoding/json"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
	"net/http"
	"strings"

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

// RegisterTimesheetRoutes registers timesheet routes
func RegisterTimesheetRoutes(r chi.Router, services *service.Services) {
	r.Route("/timesheets", func(r chi.Router) {
		r.Use(authMiddleware(services))
		r.Post("/clock-in", clockInHandler(services))
		r.Post("/clock-out", clockOutHandler(services))
		r.Get("/employee/{employeeId}", getEmployeeTimesheetsHandler(services))
		r.Put("/{id}/approve", approveTimesheetHandler(services))
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

			// Add user info to context if needed
			_ = claims

			next.ServeHTTP(w, r)
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
		var employee models.Employee
		if err := json.NewDecoder(r.Body).Decode(&employee); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := services.Employee.Create(r.Context(), &employee); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create employee")
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

		var employee models.Employee
		if err := json.NewDecoder(r.Body).Decode(&employee); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		employee.ID = id
		if err := services.Employee.Update(r.Context(), &employee); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update employee")
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

		var task models.OnboardingTask
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		task.ID = taskID
		if err := services.Onboarding.UpdateTask(r.Context(), &task); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update task")
			return
		}

		respondJSON(w, http.StatusOK, task)
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

		var task models.OnboardingTask
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		task.EmployeeID = employeeID
		if err := services.Onboarding.CreateTask(r.Context(), &task); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create task")
			return
		}

		respondJSON(w, http.StatusCreated, task)
	}
}

// Timesheet handlers
func clockInHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ClockInRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		timesheet, err := services.Timesheet.ClockIn(r.Context(), &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to clock in")
			return
		}

		respondJSON(w, http.StatusCreated, timesheet)
	}
}

func clockOutHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ClockOutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		timesheet, err := services.Timesheet.ClockOut(r.Context(), &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to clock out")
			return
		}

		respondJSON(w, http.StatusOK, timesheet)
	}
}

func getEmployeeTimesheetsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeId")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		timesheets, err := services.Timesheet.GetByEmployee(r.Context(), employeeID, nil)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get timesheets")
			return
		}

		respondJSON(w, http.StatusOK, timesheets)
	}
}

func approveTimesheetHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timesheetIDStr := chi.URLParam(r, "id")
		timesheetID, err := uuid.Parse(timesheetIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid timesheet ID")
			return
		}

		// In a real app, get approver ID from JWT token
		approverID := uuid.New() // Placeholder

		if err := services.Timesheet.ApproveTimesheet(r.Context(), timesheetID, approverID); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to approve timesheet")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"status": "approved"})
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

// Benefits handlers
func listBenefitPlansHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plans, err := services.Benefits.ListPlans(r.Context())
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

		enrollment, err := services.Benefits.Enroll(r.Context(), employeeID, &req)
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

		enrollments, err := services.Benefits.GetEnrollmentsByEmployee(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get enrollments")
			return
		}

		respondJSON(w, http.StatusOK, enrollments)
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
