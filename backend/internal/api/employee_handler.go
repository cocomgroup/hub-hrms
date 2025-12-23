package api

import (
	"encoding/json"
	"net/http"
	"time"
	"fmt"
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

		// Get existing employee first
		existingEmployee, err := services.Employee.GetByID(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "employee not found")
			return
		}

		// Parse request body - all fields optional for partial updates
		var reqBody struct {
			FirstName              *string `json:"first_name,omitempty"`
			LastName               *string `json:"last_name,omitempty"`
			Email                  *string `json:"email,omitempty"`
			Phone                  *string `json:"phone,omitempty"`
			DateOfBirth            *string `json:"date_of_birth,omitempty"`
			HireDate               *string `json:"hire_date,omitempty"`
			Department             *string `json:"department,omitempty"`
			Position               *string `json:"position,omitempty"`
			ManagerID              *string `json:"manager_id,omitempty"`
			EmploymentType         *string `json:"employment_type,omitempty"`
			Status                 *string `json:"status,omitempty"`
			StreetAddress          *string `json:"street_address,omitempty"`
			City                   *string `json:"city,omitempty"`
			State                  *string `json:"state,omitempty"`
			ZipCode                *string `json:"zip_code,omitempty"`
			Country                *string `json:"country,omitempty"`
			EmergencyContactName   *string `json:"emergency_contact_name,omitempty"`
			EmergencyContactPhone  *string `json:"emergency_contact_phone,omitempty"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}

		// Update only provided fields
		if reqBody.FirstName != nil {
			existingEmployee.FirstName = *reqBody.FirstName
		}
		if reqBody.LastName != nil {
			existingEmployee.LastName = *reqBody.LastName
		}
		if reqBody.Email != nil {
			existingEmployee.Email = *reqBody.Email
		}
		if reqBody.Phone != nil {
			existingEmployee.Phone = *reqBody.Phone
		}
		if reqBody.Department != nil {
			existingEmployee.Department = *reqBody.Department
		}
		if reqBody.Position != nil {
			existingEmployee.Position = *reqBody.Position
		}
		if reqBody.EmploymentType != nil {
			existingEmployee.EmploymentType = *reqBody.EmploymentType
		}
		if reqBody.Status != nil {
			existingEmployee.Status = *reqBody.Status
		}
		if reqBody.StreetAddress != nil {
			existingEmployee.StreetAddress = *reqBody.StreetAddress
		}
		if reqBody.City != nil {
			existingEmployee.City = *reqBody.City
		}
		if reqBody.State != nil {
			existingEmployee.State = *reqBody.State
		}
		if reqBody.ZipCode != nil {
			existingEmployee.ZipCode = *reqBody.ZipCode
		}
		if reqBody.Country != nil {
			existingEmployee.Country = *reqBody.Country
		}
		if reqBody.EmergencyContactName != nil {
			existingEmployee.EmergencyContactName = *reqBody.EmergencyContactName
		}
		if reqBody.EmergencyContactPhone != nil {
			existingEmployee.EmergencyContactPhone = *reqBody.EmergencyContactPhone
		}

		// Parse hire_date if provided (flexible format support)
		if reqBody.HireDate != nil && *reqBody.HireDate != "" {
			hireDate, err := parseFlexibleDate(*reqBody.HireDate)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid hire_date format: "+err.Error())
				return
			}
			existingEmployee.HireDate = hireDate
		}
		
		// Parse date_of_birth if provided
		if reqBody.DateOfBirth != nil && *reqBody.DateOfBirth != "" {
			dob, err := parseFlexibleDate(*reqBody.DateOfBirth)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid date_of_birth format: "+err.Error())
				return
			}
			existingEmployee.DateOfBirth = &dob
		}

		// Parse manager_id if provided
		if reqBody.ManagerID != nil && *reqBody.ManagerID != "" {
			managerID, err := uuid.Parse(*reqBody.ManagerID)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid manager_id format")
				return
			}
			existingEmployee.ManagerID = &managerID
		}

		// Update employee
		if err := services.Employee.Update(r.Context(), existingEmployee); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update employee")
			return
		}

		respondJSON(w, http.StatusOK, existingEmployee)
	}
}

// Helper function to parse dates in multiple formats
func parseFlexibleDate(dateStr string) (time.Time, error) {
	// Try common formats
	formats := []string{
		"2006-01-02",                 // YYYY-MM-DD
		"2006-01-02T15:04:05Z",       // RFC3339
		"2006-01-02T15:04:05-07:00",  // RFC3339 with timezone
		"2006-01-02T15:04:05.000Z",   // With milliseconds
		time.RFC3339,                 // Standard RFC3339
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
