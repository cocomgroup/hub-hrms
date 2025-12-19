package api

import (
	"encoding/json"
	"net/http"
	"time"
	"log"
	"io"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

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
