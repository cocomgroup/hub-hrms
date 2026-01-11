package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// RegisterBackgroundCheckRoutes registers all background check routes
func RegisterBackgroundCheckRoutes(r chi.Router, services *service.Services) {
	r.Route("/background-checks", func(r chi.Router) {
		r.Use(authMiddleware(services))
		
		r.Post("/", initiateCheckHandler(services))
		r.Get("/{id}", getCheckHandler(services))
		r.Post("/{id}/cancel", cancelCheckHandler(services))
		
		// Packages
		r.Get("/packages", listPackagesHandler(services))
		r.Post("/packages", createPackageHandler(services))

		// Background check
		r.Get("/employee/{employeeId}/background-checks", getEmployeeChecksHandler(services))
	})

	// Webhook endpoint (no auth middleware for webhooks)
	r.Route("/webhooks/background-checks", func(r chi.Router) {
		r.Post("/{provider}", handleWebhookHandler(services))
	})
}

// InitiateCheckRequest represents the request to initiate a background check
type InitiateCheckRequest struct {
	EmployeeID string                 `json:"employee_id"`
	PackageID  string                 `json:"package_id"`
	Candidate  models.CandidateInfo   `json:"candidate"`
	Consent    ConsentInfo            `json:"consent"`
}

// ConsentInfo represents FCRA consent information
type ConsentInfo struct {
	FCRADisclosureProvided bool      `json:"fcra_disclosure_provided"`
	FCRAConsentObtained    bool      `json:"fcra_consent_obtained"`
	FCRAConsentDate        time.Time `json:"fcra_consent_date"`
	SignatureData          string    `json:"signature_data,omitempty"`
}

// initiateCheckHandler handles POST /api/v1/background-checks
func initiateCheckHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("=== initiateCheckHandler START ===")

		// Read and log the raw body for debugging
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			respondError(w, http.StatusBadRequest, "error reading request body")
			return
		}
		log.Printf("Raw request body: %s", string(bodyBytes))

		var req InitiateCheckRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			log.Printf("Error decoding background check request: %v", err)
			respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}

		// Validate required fields
		if req.EmployeeID == "" || req.PackageID == "" {
			log.Printf("Missing required fields - EmployeeID: %s, PackageID: %s", req.EmployeeID, req.PackageID)
			respondError(w, http.StatusBadRequest, "employee_id and package_id are required")
			return
		}

		if !req.Consent.FCRAConsentObtained {
			log.Printf("FCRA consent not obtained")
			respondError(w, http.StatusBadRequest, "FCRA consent must be obtained before initiating check")
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, err := getEmployeeIDStrFromContext(r.Context())
		if err != nil {
			log.Printf("ERROR: Failed to get user ID from context: %v", err)
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Convert UUID to string for service call
		userIDStr := userID.String()

		log.Printf("Initiating background check for employee: %s, package: %s, by user: %s", 
			req.EmployeeID, req.PackageID, userIDStr)

		// Initiate the background check
		check, err := services.BackgroundCheck.InitiateBackgroundCheck(
			r.Context(),
			req.EmployeeID,
			req.PackageID,
			req.Candidate,
			userIDStr,
		)

		if err != nil {
			log.Printf("ERROR: Failed to initiate background check: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to initiate background check: "+err.Error())
			return
		}

		log.Printf("SUCCESS: Background check initiated with ID: %s", check.ID)
		log.Printf("=== initiateCheckHandler END ===")

		respondJSON(w, http.StatusCreated, check)
	}
}

// getCheckHandler handles GET /api/v1/background-checks/{id}
func getCheckHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("=== getCheckHandler START ===")
		
		checkID := chi.URLParam(r, "id")
		log.Printf("Check ID: %s", checkID)

		check, err := services.BackgroundCheck.GetCheckStatus(r.Context(), checkID)
		if err != nil {
			log.Printf("ERROR: Background check not found: %v", err)
			respondError(w, http.StatusNotFound, "background check not found")
			return
		}

		log.Printf("SUCCESS: Retrieved background check: %s", checkID)
		log.Printf("=== getCheckHandler END ===")

		respondJSON(w, http.StatusOK, check)
	}
}

// cancelCheckHandler handles POST /api/v1/background-checks/{id}/cancel
func cancelCheckHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("=== cancelCheckHandler START ===")
		
		checkID := chi.URLParam(r, "id")
		log.Printf("Cancelling check ID: %s", checkID)

		if err := services.BackgroundCheck.CancelCheck(r.Context(), checkID); err != nil {
			log.Printf("ERROR: Failed to cancel check: %v", err)
			respondError(w, http.StatusBadRequest, "failed to cancel check: "+err.Error())
			return
		}

		log.Printf("SUCCESS: Background check cancelled: %s", checkID)
		log.Printf("=== cancelCheckHandler END ===")

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "Background check cancelled successfully",
		})
	}
}

// getEmployeeChecksHandler handles GET /api/v1/employees/{employeeId}/background-checks
func getEmployeeChecksHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("=== getEmployeeChecksHandler START ===")
		
		employeeID := chi.URLParam(r, "employeeId")
		log.Printf("Employee ID: %s", employeeID)

		checks, err := services.BackgroundCheck.GetEmployeeChecks(r.Context(), employeeID)
		if err != nil {
			log.Printf("ERROR: Failed to get employee checks: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to get background checks: "+err.Error())
			return
		}

		if checks == nil {
			checks = []*models.BackgroundCheck{}
		}

		log.Printf("SUCCESS: Retrieved %d background checks for employee: %s", len(checks), employeeID)
		log.Printf("=== getEmployeeChecksHandler END ===")

		respondJSON(w, http.StatusOK, checks)
	}
}

// CreatePackageRequest represents a request to create a package
type CreatePackageRequest struct {
	Name           string                       `json:"name"`
	Description    string                       `json:"description"`
	CheckTypes     []models.BackgroundCheckType `json:"check_types"`
	ProviderID     string                       `json:"provider_id"`
	TurnaroundDays int                          `json:"turnaround_days"`
	Cost           float64                      `json:"cost"`
}

// createPackageHandler handles POST /api/v1/background-checks/packages
func createPackageHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("=== createPackageHandler START ===")

		// Read and log the raw body for debugging
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			respondError(w, http.StatusBadRequest, "error reading request body")
			return
		}
		log.Printf("Raw request body: %s", string(bodyBytes))

		var req CreatePackageRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			log.Printf("Error decoding package: %v", err)
			respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}

		// Validate required fields
		if req.Name == "" || req.ProviderID == "" {
			log.Printf("Missing required fields - Name: %s, ProviderID: %s", req.Name, req.ProviderID)
			respondError(w, http.StatusBadRequest, "name and provider_id are required")
			return
		}

		pkg := &models.BackgroundCheckPackage{
			ID:             uuid.New().String(),
			Name:           req.Name,
			Description:    req.Description,
			CheckTypes:     req.CheckTypes,
			ProviderID:     req.ProviderID,
			TurnaroundDays: req.TurnaroundDays,
			Cost:           req.Cost,
			Active:         true,
		}

		log.Printf("Creating package: %+v", pkg)

		if err := services.BackgroundCheck.CreatePackage(r.Context(), pkg); err != nil {
			log.Printf("ERROR: Failed to create package: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to create package: "+err.Error())
			return
		}

		log.Printf("SUCCESS: Package created with ID: %s", pkg.ID)
		log.Printf("=== createPackageHandler END ===")

		respondJSON(w, http.StatusCreated, pkg)
	}
}

// listPackagesHandler handles GET /api/v1/background-checks/packages
func listPackagesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("=== listPackagesHandler START ===")

		packages, err := services.BackgroundCheck.ListPackages(r.Context())
		if err != nil {
			log.Printf("ERROR: Failed to list packages: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to list packages: "+err.Error())
			return
		}

		if packages == nil {
			packages = []*models.BackgroundCheckPackage{}
		}

		log.Printf("SUCCESS: Retrieved %d packages", len(packages))
		log.Printf("=== listPackagesHandler END ===")

		respondJSON(w, http.StatusOK, packages)
	}
}

// handleWebhookHandler handles POST /webhooks/background-checks/{provider}
func handleWebhookHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("=== handleWebhookHandler START ===")
		
		provider := chi.URLParam(r, "provider")
		log.Printf("Provider: %s", provider)

		// Get signature from header (varies by provider)
		signature := r.Header.Get("X-Checkr-Signature")
		if signature == "" {
			signature = r.Header.Get("X-Webhook-Signature")
		}
		log.Printf("Signature header: %s", signature)

		// Read body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("ERROR: Failed to read request body: %v", err)
			respondError(w, http.StatusBadRequest, "failed to read request body")
			return
		}
		log.Printf("Webhook payload size: %d bytes", len(bodyBytes))

		// Process webhook
		if err := services.BackgroundCheck.HandleWebhook(r.Context(), provider, bodyBytes, signature); err != nil {
			log.Printf("ERROR: Failed to process webhook: %v", err)
			respondError(w, http.StatusBadRequest, "failed to process webhook: "+err.Error())
			return
		}

		log.Printf("SUCCESS: Webhook processed successfully")
		log.Printf("=== handleWebhookHandler END ===")

		// Return success
		w.WriteHeader(http.StatusOK)
	}
}

// Helper functions

// getEmployeeIDFromContext retrieves the employee ID from context
// This should be set by your auth middleware
func getEmployeeIDStrFromContext(ctx context.Context) (uuid.UUID, error) {
	employeeID := ctx.Value("employee_id")
	if employeeID == nil {
		// Fallback to user_id if employee_id not found
		userID := ctx.Value("user_id")
		if userID == nil {
			return uuid.UUID{}, fmt.Errorf("user_id not found in context")
		}
		
		// Try to parse as string first
		if idStr, ok := userID.(string); ok {
			return uuid.Parse(idStr)
		}
		
		// Try as UUID directly
		if id, ok := userID.(uuid.UUID); ok {
			return id, nil
		}
		
		return uuid.UUID{}, fmt.Errorf("user_id is not a valid type")
	}
	
	// Try to parse employee_id as string first
	if idStr, ok := employeeID.(string); ok {
		return uuid.Parse(idStr)
	}
	
	// Try as UUID directly
	if id, ok := employeeID.(uuid.UUID); ok {
		return id, nil
	}
	
	return uuid.UUID{}, fmt.Errorf("employee_id is not a valid type")
}