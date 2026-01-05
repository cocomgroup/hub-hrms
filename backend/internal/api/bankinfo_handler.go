package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// getRoleFromContext extracts role from JWT claims
func getRoleFromContext(r *http.Request) string {
	if claims, ok := r.Context().Value("claims").(jwt.MapClaims); ok {
		if role, ok := claims["role"].(string); ok {
			return role
		}
	}
	return ""
}

// RegisterBankInfoRoutes registers all bank info routes
func RegisterBankInfoRoutes(r chi.Router, services *service.Services) {
	r.Route("/bank-info", func(r chi.Router) {
		r.Use(authMiddleware(services))
		
		// Bank information CRUD
		r.Post("/", createBankInfoHandler(services))
		r.Get("/employee/{employeeId}", getBankInfoByEmployeeHandler(services))
		r.Get("/{id}", getBankInfoHandler(services))
		r.Put("/{id}", updateBankInfoHandler(services))
		r.Delete("/{id}", deleteBankInfoHandler(services))
		
		// Primary account management
		r.Put("/{id}/set-primary", setPrimaryBankInfoHandler(services))
		
		// Verification
		r.Post("/{id}/verify", verifyBankInfoHandler(services))
		
		// Admin/HR list
		r.Get("/", listBankInfoHandler(services))
	})
}

// createBankInfoHandler handles creating new bank information
func createBankInfoHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		var req models.BankInfoCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		// Validate required fields
		if req.AccountHolderName == "" {
			respondError(w, http.StatusBadRequest, "account holder name is required")
			return
		}
		if req.BankName == "" {
			respondError(w, http.StatusBadRequest, "bank name is required")
			return
		}
		if req.AccountNumber == "" {
			respondError(w, http.StatusBadRequest, "account number is required")
			return
		}
		if req.RoutingNumber == "" {
			respondError(w, http.StatusBadRequest, "routing number is required")
			return
		}
		if req.AccountType == "" {
			respondError(w, http.StatusBadRequest, "account type is required")
			return
		}
		
		// If no employee ID provided, use current user's employee ID
		if req.EmployeeID == uuid.Nil {
			employeeID, err := getEmployeeIDFromContext(r.Context())
			if err != nil {
				respondError(w, http.StatusBadRequest, "employee_id is required")
				return
			}
			req.EmployeeID = employeeID
		}
		
		// Create bank info
		bankInfo, err := services.BankInfo.CreateBankInfo(r.Context(), &req, userID)
		if err != nil {
			if err == service.ErrInvalidRoutingNumber {
				respondError(w, http.StatusBadRequest, "routing number must be exactly 9 digits")
				return
			}
			if err == service.ErrInvalidAccountNumber {
				respondError(w, http.StatusBadRequest, "account number must be between 8 and 17 digits")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to create bank information")
			return
		}
		
		// Return response without sensitive data
		respondJSON(w, http.StatusCreated, bankInfo.ToBankInfoResponse())
	}
}

// getBankInfoHandler retrieves bank information by ID
func getBankInfoHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid bank info ID")
			return
		}
		
		bankInfo, err := services.BankInfo.GetBankInfo(r.Context(), id)
		if err != nil {
			if err == service.ErrBankInfoNotFound {
				respondError(w, http.StatusNotFound, "bank information not found")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to get bank information")
			return
		}
		
		// Check authorization - user can only view their own bank info or must be admin/hr
		employeeID, _ := getEmployeeIDFromContext(r.Context())
		role := getRoleFromContext(r)
		
		if role != "admin" && role != "hr-manager" {
			if bankInfo.EmployeeID != employeeID {
				respondError(w, http.StatusForbidden, "unauthorized to view this bank information")
				return
			}
		}
		
		// Admin/HR can optionally decrypt sensitive data (add query param ?decrypt=true)
		if r.URL.Query().Get("decrypt") == "true" && (role == "admin" || role == "hr-manager") {
			if err := services.BankInfo.DecryptSensitiveData(r.Context(), bankInfo); err != nil {
				respondError(w, http.StatusInternalServerError, "failed to decrypt sensitive data")
				return
			}
			// Return full data including decrypted fields
			respondJSON(w, http.StatusOK, bankInfo)
			return
		}
		
		// Return masked data
		respondJSON(w, http.StatusOK, bankInfo.ToBankInfoResponse())
	}
}

// getBankInfoByEmployeeHandler retrieves all bank information for an employee
func getBankInfoByEmployeeHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeId")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}
		
		// Check authorization
		role := getRoleFromContext(r)
		if role != "admin" && role != "hr-manager" {
			currentEmployeeID, _ := getEmployeeIDFromContext(r.Context())
			if employeeID != currentEmployeeID {
				respondError(w, http.StatusForbidden, "unauthorized to view this employee's bank information")
				return
			}
		}
		
		bankInfos, err := services.BankInfo.GetBankInfoByEmployeeID(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get bank information")
			return
		}
		
		// Convert to response format (masked)
		responses := make([]*models.BankInfoResponse, len(bankInfos))
		for i, bankInfo := range bankInfos {
			responses[i] = bankInfo.ToBankInfoResponse()
		}
		
		respondJSON(w, http.StatusOK, responses)
	}
}

// updateBankInfoHandler handles updating bank information
func updateBankInfoHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid bank info ID")
			return
		}
		
		// Check authorization
		bankInfo, err := services.BankInfo.GetBankInfo(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "bank information not found")
			return
		}
		
		role := getRoleFromContext(r)
		if role != "admin" && role != "hr-manager" {
			employeeID, _ := getEmployeeIDFromContext(r.Context())
			if bankInfo.EmployeeID != employeeID {
				respondError(w, http.StatusForbidden, "unauthorized to update this bank information")
				return
			}
		}
		
		var req models.BankInfoUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		// Update bank info
		updatedBankInfo, err := services.BankInfo.UpdateBankInfo(r.Context(), id, &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update bank information")
			return
		}
		
		respondJSON(w, http.StatusOK, updatedBankInfo.ToBankInfoResponse())
	}
}

// deleteBankInfoHandler handles soft deleting bank information
func deleteBankInfoHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid bank info ID")
			return
		}
		
		// Check authorization
		bankInfo, err := services.BankInfo.GetBankInfo(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "bank information not found")
			return
		}
		
		role := getRoleFromContext(r)
		if role != "admin" && role != "hr-manager" {
			employeeID, _ := getEmployeeIDFromContext(r.Context())
			if bankInfo.EmployeeID != employeeID {
				respondError(w, http.StatusForbidden, "unauthorized to delete this bank information")
				return
			}
		}
		
		// Delete bank info
		if err := services.BankInfo.DeleteBankInfo(r.Context(), id); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to delete bank information")
			return
		}
		
		w.WriteHeader(http.StatusNoContent)
	}
}

// setPrimaryBankInfoHandler sets a bank account as primary
func setPrimaryBankInfoHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid bank info ID")
			return
		}
		
		// Check authorization
		bankInfo, err := services.BankInfo.GetBankInfo(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "bank information not found")
			return
		}
		
		role := getRoleFromContext(r)
		if role != "admin" && role != "hr-manager" {
			employeeID, _ := getEmployeeIDFromContext(r.Context())
			if bankInfo.EmployeeID != employeeID {
				respondError(w, http.StatusForbidden, "unauthorized to modify this bank information")
				return
			}
		}
		
		// Set as primary
		if err := services.BankInfo.SetPrimaryBankInfo(r.Context(), id, bankInfo.EmployeeID); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to set primary bank account")
			return
		}
		
		// Return updated bank info
		updatedBankInfo, _ := services.BankInfo.GetBankInfo(r.Context(), id)
		respondJSON(w, http.StatusOK, updatedBankInfo.ToBankInfoResponse())
	}
}

// verifyBankInfoHandler marks bank information as verified
func verifyBankInfoHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only admins/HR can verify
		role := getRoleFromContext(r)
		if role != "admin" && role != "hr-manager" {
			respondError(w, http.StatusForbidden, "unauthorized to verify bank information")
			return
		}
		
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid bank info ID")
			return
		}
		
		userID, _ := getUserIDFromContext(r.Context())
		
		// Verify bank info
		if err := services.BankInfo.VerifyBankInfo(r.Context(), id, userID); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to verify bank information")
			return
		}
		
		// Return updated bank info
		bankInfo, _ := services.BankInfo.GetBankInfo(r.Context(), id)
		respondJSON(w, http.StatusOK, bankInfo.ToBankInfoResponse())
	}
}

// listBankInfoHandler lists all bank information (admin/HR only)
func listBankInfoHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only admins/HR can list all
		role := getRoleFromContext(r)
		if role != "admin" && role != "hr-manager" {
			respondError(w, http.StatusForbidden, "unauthorized to list all bank information")
			return
		}
		
		// Parse filters
		filters := make(map[string]interface{})
		if status := r.URL.Query().Get("status"); status != "" {
			filters["status"] = status
		}
		if verified := r.URL.Query().Get("verified"); verified == "true" {
			filters["verified"] = true
		} else if verified == "false" {
			filters["verified"] = false
		}
		
		bankInfos, err := services.BankInfo.ListBankInfo(r.Context(), filters)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list bank information")
			return
		}
		
		// Convert to response format
		responses := make([]*models.BankInfoResponse, len(bankInfos))
		for i, bankInfo := range bankInfos {
			responses[i] = bankInfo.ToBankInfoResponse()
		}
		
		respondJSON(w, http.StatusOK, responses)
	}
}