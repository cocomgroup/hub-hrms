package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// RegisterPayrollRoutes registers all payroll-related routes
func RegisterPayrollRoutes(r chi.Router, services *service.Services) {
	r.Route("/payroll", func(r chi.Router) {
		// Compensation endpoints
		r.Post("/compensation", createCompensationHandler(services))
		r.Get("/compensation/employee/{employeeID}", getEmployeeCompensationHandler(services))

		// Tax Withholding endpoints (W2 only)
		r.Put("/tax-withholding/{employeeID}", updateTaxWithholdingHandler(services))
		r.Get("/tax-withholding/{employeeID}", getTaxWithholdingHandler(services))

		// Payroll Periods endpoints
		r.Post("/periods", createPayrollPeriodHandler(services))
		r.Get("/periods", listPayrollPeriodsHandler(services))
		r.Get("/periods/{periodID}", getPayrollPeriodHandler(services))
		r.Post("/periods/{periodID}/process", processPayrollHandler(services))

		// Pay Stubs endpoints
		r.Get("/paystubs/employee/{employeeID}", getEmployeePayStubsHandler(services))
		r.Get("/paystubs/{payStubID}", getPayStubDetailHandler(services))
		r.Get("/paystubs/{payStubID}/pdf", downloadPayStubPDFHandler(services))

		// 1099 Forms endpoints
		r.Post("/1099/generate/{year}", generate1099FormsHandler(services))
		r.Get("/1099/{year}", list1099ByYearHandler(services))
		r.Get("/1099/employee/{employeeID}/year/{year}", get1099ForEmployeeHandler(services))
	})
}

// ===============================
// Compensation Handlers
// ===============================

// createCompensationHandler creates compensation for an employee
func createCompensationHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateCompensationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		comp, err := services.Payroll.CreateCompensation(r.Context(), &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create compensation")
			return
		}

		respondJSON(w, http.StatusCreated, comp)
	}
}

// getEmployeeCompensationHandler gets compensation for a specific employee
func getEmployeeCompensationHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeID")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		comp, err := services.Payroll.GetEmployeeCompensation(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusNotFound, "compensation not found")
			return
		}

		respondJSON(w, http.StatusOK, comp)
	}
}

// ===============================
// Tax Withholding Handlers
// ===============================

// updateTaxWithholdingHandler updates tax withholding for an employee
func updateTaxWithholdingHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeID")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		var req models.UpdateTaxWithholdingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		tax, err := services.Payroll.UpdateTaxWithholding(r.Context(), employeeID, &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update tax withholding")
			return
		}

		respondJSON(w, http.StatusOK, tax)
	}
}

// getTaxWithholdingHandler gets tax withholding for an employee
func getTaxWithholdingHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeID")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		// TODO: Implement GetTaxWithholding in service
		// For now, return placeholder
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"employee_id": employeeID,
			"message":     "tax withholding retrieved",
		})
	}
}

// ===============================
// Payroll Period Handlers
// ===============================

// createPayrollPeriodHandler creates a new payroll period
func createPayrollPeriodHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreatePayrollPeriodRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		period, err := services.Payroll.CreatePayrollPeriod(r.Context(), &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create payroll period")
			return
		}

		respondJSON(w, http.StatusCreated, period)
	}
}

// listPayrollPeriodsHandler lists all payroll periods
func listPayrollPeriodsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		periods, err := services.Payroll.ListPayrollPeriods(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list payroll periods")
			return
		}

		respondJSON(w, http.StatusOK, periods)
	}
}

// getPayrollPeriodHandler gets a specific payroll period
func getPayrollPeriodHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		periodIDStr := chi.URLParam(r, "periodID")
		periodID, err := uuid.Parse(periodIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid period ID")
			return
		}

		period, err := services.Payroll.GetPayrollPeriod(r.Context(), periodID)
		if err != nil {
			respondError(w, http.StatusNotFound, "payroll period not found")
			return
		}

		respondJSON(w, http.StatusOK, period)
	}
}

// processPayrollHandler processes payroll for a period
func processPayrollHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		periodIDStr := chi.URLParam(r, "periodID")
		periodID, err := uuid.Parse(periodIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid period ID")
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		summary, err := services.Payroll.ProcessPayroll(r.Context(), periodID, userID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to process payroll")
			return
		}

		respondJSON(w, http.StatusOK, summary)
	}
}

// ===============================
// Pay Stub Handlers
// ===============================

// getEmployeePayStubsHandler gets all pay stubs for an employee
func getEmployeePayStubsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeID")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		stubs, err := services.Payroll.GetEmployeePayStubs(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get pay stubs")
			return
		}

		respondJSON(w, http.StatusOK, stubs)
	}
}

// getPayStubDetailHandler gets details for a specific pay stub
func getPayStubDetailHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payStubIDStr := chi.URLParam(r, "payStubID")
		payStubID, err := uuid.Parse(payStubIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid pay stub ID")
			return
		}

		detail, err := services.Payroll.GetPayStubDetail(r.Context(), payStubID)
		if err != nil {
			respondError(w, http.StatusNotFound, "pay stub not found")
			return
		}

		respondJSON(w, http.StatusOK, detail)
	}
}

// downloadPayStubPDFHandler generates and downloads a pay stub PDF
func downloadPayStubPDFHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payStubIDStr := chi.URLParam(r, "payStubID")
		payStubID, err := uuid.Parse(payStubIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid pay stub ID")
			return
		}

		detail, err := services.Payroll.GetPayStubDetail(r.Context(), payStubID)
		if err != nil {
			respondError(w, http.StatusNotFound, "pay stub not found")
			return
		}

		// Generate PDF
		pdfBytes := generatePayStubPDF(detail)

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=paystub.pdf")
		w.Write(pdfBytes)
	}
}

// ===============================
// 1099 Form Handlers
// ===============================

// generate1099FormsHandler generates 1099 forms for a tax year
func generate1099FormsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		yearStr := chi.URLParam(r, "year")
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid year")
			return
		}

		forms, err := services.Payroll.Generate1099Forms(r.Context(), year)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to generate 1099 forms")
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"year":    year,
			"count":   len(forms),
			"forms":   forms,
			"message": "1099 forms generated successfully",
		})
	}
}

// list1099ByYearHandler lists all 1099 forms for a tax year
func list1099ByYearHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		yearStr := chi.URLParam(r, "year")
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid year")
			return
		}

		// TODO: Implement List1099ByYear in service
		// For now, return placeholder
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"year":  year,
			"forms": []interface{}{},
		})
	}
}

// get1099ForEmployeeHandler gets 1099 form for a specific employee and year
func get1099ForEmployeeHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeID")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid employee ID")
			return
		}

		yearStr := chi.URLParam(r, "year")
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid year")
			return
		}

		// TODO: Implement Get1099ForEmployee in service
		// For now, return placeholder
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"employee_id": employeeID,
			"year":        year,
			"message":     "1099 form retrieved",
		})
	}
}

// ===============================
// Utility Functions
// ===============================

// generatePayStubPDF generates a PDF for a pay stub
func generatePayStubPDF(detail *models.PayStubDetail) []byte {
	// Simple PDF generation - in production, use a proper PDF library like gofpdf
	// For now, return placeholder PDF with basic structure
	return []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Count 1\n/Kids [3 0 R]\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/Resources <<\n/Font <<\n/F1 4 0 R\n>>\n>>\n/MediaBox [0 0 612 792]\n/Contents 5 0 R\n>>\nendobj\n4 0 obj\n<<\n/Type /Font\n/Subtype /Type1\n/BaseFont /Helvetica\n>>\nendobj\n5 0 obj\n<< /Length 44 >>\nstream\nBT\n/F1 24 Tf\n100 700 Td\n(Pay Stub) Tj\nET\nendstream\nendobj\nxref\n0 6\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000262 00000 n \n0000000341 00000 n \ntrailer\n<<\n/Size 6\n/Root 1 0 R\n>>\nstartxref\n437\n%%EOF")
}