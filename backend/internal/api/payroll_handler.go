package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// RegisterPayrollRoutes registers all payroll-related routes
func RegisterPayrollRoutes(r chi.Router, services *service.Services) {
	r.Route("/payroll", func(r chi.Router) {
		r.Use(authMiddleware(services))
		
		// Payroll Periods endpoints
		r.Post("/periods", createPayrollPeriodHandler(services))
		r.Get("/periods", listPayrollPeriodsHandler(services))
		r.Get("/periods/{periodID}", getPayrollPeriodHandler(services))
		r.Put("/periods/{periodID}", updatePayrollPeriodHandler(services))
		r.Post("/periods/{periodID}/process", processPayrollHandler(services))

		// Pay Stubs endpoints
		r.Get("/paystubs/{payStubID}", getPayStubHandler(services))
		r.Get("/paystubs/employee/{employeeID}", getEmployeePayStubsHandler(services))
		r.Get("/paystubs/period/{periodID}", getPayStubsByPeriodHandler(services))
		r.Get("/paystubs/{payStubID}/pdf", downloadPayStubPDFHandler(services))
	})
}

// ===============================
// Payroll Period Handlers
// ===============================

// createPayrollPeriodHandler creates a new payroll period
func createPayrollPeriodHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.PayrollPeriodRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		period, err := services.Payroll.CreatePayrollPeriod(r.Context(), &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create payroll period: "+err.Error())
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

		if periods == nil {
			periods = []*models.PayrollPeriod{}
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

// updatePayrollPeriodHandler updates a payroll period
func updatePayrollPeriodHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		periodIDStr := chi.URLParam(r, "periodID")
		periodID, err := uuid.Parse(periodIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid period ID")
			return
		}

		var req models.PayrollPeriodRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		period, err := services.Payroll.UpdatePayrollPeriod(r.Context(), periodID, &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to update payroll period: "+err.Error())
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

		summary, err := services.Payroll.ProcessPayroll(r.Context(), periodID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to process payroll: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, summary)
	}
}

// ===============================
// Pay Stub Handlers
// ===============================

// getPayStubHandler gets a specific pay stub
func getPayStubHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payStubIDStr := chi.URLParam(r, "payStubID")
		payStubID, err := uuid.Parse(payStubIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid pay stub ID")
			return
		}

		stub, err := services.Payroll.GetPayStub(r.Context(), payStubID)
		if err != nil {
			respondError(w, http.StatusNotFound, "pay stub not found")
			return
		}

		respondJSON(w, http.StatusOK, stub)
	}
}

// getEmployeePayStubsHandler gets all pay stubs for an employee
func getEmployeePayStubsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeID")
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

		if stubs == nil {
			stubs = []*models.PayStub{}
		}

		respondJSON(w, http.StatusOK, stubs)
	}
}

// getPayStubsByPeriodHandler gets all pay stubs for a payroll period
func getPayStubsByPeriodHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		periodIDStr := chi.URLParam(r, "periodID")
		periodID, err := uuid.Parse(periodIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid period ID")
			return
		}

		stubs, err := services.Payroll.GetPayStubsByPeriod(r.Context(), periodID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get pay stubs")
			return
		}

		if stubs == nil {
			stubs = []*models.PayStub{}
		}

		respondJSON(w, http.StatusOK, stubs)
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

		stub, err := services.Payroll.GetPayStub(r.Context(), payStubID)
		if err != nil {
			respondError(w, http.StatusNotFound, "pay stub not found")
			return
		}

		// Generate PDF
		pdfBytes := generatePayStubPDF(stub)

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=paystub.pdf")
		w.Write(pdfBytes)
	}
}

// ===============================
// Utility Functions
// ===============================

// generatePayStubPDF generates a PDF for a pay stub
func generatePayStubPDF(stub *models.PayStub) []byte {
	// Simple PDF generation - in production, use a proper PDF library like gofpdf
	// For now, return placeholder PDF with basic structure
	return []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj\n2 0 obj\n<<\n/Type /Pages\n/Count 1\n/Kids [3 0 R]\n>>\nendobj\n3 0 obj\n<<\n/Type /Page\n/Parent 2 0 R\n/Resources <<\n/Font <<\n/F1 4 0 R\n>>\n>>\n/MediaBox [0 0 612 792]\n/Contents 5 0 R\n>>\nendobj\n4 0 obj\n<<\n/Type /Font\n/Subtype /Type1\n/BaseFont /Helvetica\n>>\nendobj\n5 0 obj\n<< /Length 44 >>\nstream\nBT\n/F1 24 Tf\n100 700 Td\n(Pay Stub) Tj\nET\nendstream\nendobj\nxref\n0 6\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000262 00000 n \n0000000341 00000 n \ntrailer\n<<\n/Size 6\n/Root 1 0 R\n>>\nstartxref\n437\n%%EOF")
}
