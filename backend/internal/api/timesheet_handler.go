package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// ============================================================================
// SIMPLIFIED TIMESHEET HANDLER
// ============================================================================
// NOTE: This handler ONLY manages time entries and timesheets.
// Project management (create/update/delete) is handled by project_handler.go
// Employees can only SELECT from projects they're assigned to.
// ============================================================================

// RegisterTimesheetRoutes registers simplified timesheet routes
func RegisterTimesheetRoutes(r chi.Router, services *service.Services) {
	r.Route("/timesheet", func(r chi.Router) {
		// Apply auth middleware to all timesheet routes
		r.Use(authMiddleware(services))
		
		// Time Entry endpoints (daily)
		r.Get("/entries", getTimeEntriesHandler(services))
		r.Post("/entries", createTimeEntryHandler(services))
		r.Get("/entries/{id}", getTimeEntryHandler(services))
		r.Put("/entries/{id}", updateTimeEntryHandler(services))
		r.Delete("/entries/{id}", deleteTimeEntryHandler(services))
		r.Post("/entries/bulk", bulkCreateTimeEntriesHandler(services))
		
		// Timesheet endpoints (weekly)
		r.Get("/summary", getWeeklySummaryHandler(services))
		r.Post("/submit", submitTimesheetHandler(services))
		r.Get("/timesheets", getEmployeeTimesheetsHandler(services))
		r.Get("/timesheets/{id}", getTimesheetHandler(services))
		
		// Manager endpoints
		r.Get("/pending", getPendingTimesheetsHandler(services))
		r.Post("/timesheets/{id}/approve", approveTimesheetHandler(services))
		
		// Project endpoints (READ-ONLY - employee's assigned projects)
		r.Get("/projects", getAvailableProjectsHandler(services))
		
		// NOTE: Project creation/management is in /api/projects (project_handler.go)
		// Only HR managers and project managers can create/update/delete projects
	})
}

// ============================================================================
// TIME ENTRY HANDLERS (Daily)
// ============================================================================

func getTimeEntriesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		// Parse query parameters
		startDateStr := r.URL.Query().Get("start_date")
		endDateStr := r.URL.Query().Get("end_date")
		
		var startDate, endDate time.Time
		if startDateStr != "" {
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid start_date format")
				return
			}
		} else {
			// Default to current week
			startDate = getWeekStart(time.Now())
		}
		
		if endDateStr != "" {
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid end_date format")
				return
			}
		} else {
			endDate = getWeekEnd(time.Now())
		}
		
		entries, err := services.Timesheet.GetTimeEntries(r.Context(), employeeID, startDate, endDate)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get time entries")
			return
		}
		
		if entries == nil {
			entries = []*models.TimeEntry{}
		}
		
		respondJSON(w, http.StatusOK, entries)
	}
}

func createTimeEntryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		var req models.TimeEntryCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		entry, err := services.Timesheet.CreateTimeEntry(r.Context(), employeeID, &req)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		
		respondJSON(w, http.StatusCreated, entry)
	}
}

func getTimeEntryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entryID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid entry ID")
			return
		}
		
		entry, err := services.Timesheet.GetTimeEntry(r.Context(), entryID)
		if err != nil {
			respondError(w, http.StatusNotFound, "time entry not found")
			return
		}
		
		respondJSON(w, http.StatusOK, entry)
	}
}

func updateTimeEntryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		entryID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid entry ID")
			return
		}
		
		var req models.TimeEntryUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		entry, err := services.Timesheet.UpdateTimeEntry(r.Context(), entryID, employeeID, &req)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, entry)
	}
}

func deleteTimeEntryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		entryID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid entry ID")
			return
		}
		
		if err := services.Timesheet.DeleteTimeEntry(r.Context(), entryID, employeeID); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, map[string]string{"message": "time entry deleted"})
	}
}

func bulkCreateTimeEntriesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		var req models.TimeEntryBulkCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		entries, err := services.Timesheet.BulkCreateTimeEntries(r.Context(), employeeID, &req)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		
		respondJSON(w, http.StatusCreated, entries)
	}
}

// ============================================================================
// TIMESHEET HANDLERS (Weekly)
// ============================================================================

func getWeeklySummaryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		// Parse query parameters
		startDateStr := r.URL.Query().Get("start_date")
		endDateStr := r.URL.Query().Get("end_date")
		
		var startDate, endDate time.Time
		if startDateStr != "" {
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid start_date format")
				return
			}
		} else {
			startDate = getWeekStart(time.Now())
		}
		
		if endDateStr != "" {
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid end_date format")
				return
			}
		} else {
			endDate = getWeekEnd(time.Now())
		}
		
		summary, err := services.Timesheet.GetWeeklySummary(r.Context(), employeeID, startDate, endDate)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get summary")
			return
		}
		
		respondJSON(w, http.StatusOK, summary)
	}
}

func submitTimesheetHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		var req models.TimesheetSubmitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		timesheet, err := services.Timesheet.SubmitTimesheet(r.Context(), employeeID, &req)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, timesheet)
	}
}

func getEmployeeTimesheetsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		timesheets, err := services.Timesheet.GetTimesheetsByEmployee(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get timesheets")
			return
		}
		
		if timesheets == nil {
			timesheets = []*models.Timesheet{}
		}
		
		respondJSON(w, http.StatusOK, timesheets)
	}
}

func getTimesheetHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timesheetID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid timesheet ID")
			return
		}
		
		timesheet, err := services.Timesheet.GetTimesheet(r.Context(), timesheetID)
		if err != nil {
			respondError(w, http.StatusNotFound, "timesheet not found")
			return
		}
		
		respondJSON(w, http.StatusOK, timesheet)
	}
}

// ============================================================================
// MANAGER HANDLERS
// ============================================================================

func getPendingTimesheetsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get manager ID from context (assuming it's the employee ID)
		managerID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		timesheets, err := services.Timesheet.GetPendingTimesheets(r.Context(), managerID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get pending timesheets")
			return
		}
		
		if timesheets == nil {
			timesheets = []*models.Timesheet{}
		}
		
		respondJSON(w, http.StatusOK, timesheets)
	}
}

func approveTimesheetHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		managerID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		timesheetID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid timesheet ID")
			return
		}
		
		var req models.TimesheetApprovalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		timesheet, err := services.Timesheet.ApproveTimesheet(r.Context(), timesheetID, managerID, &req)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		
		respondJSON(w, http.StatusOK, timesheet)
	}
}

// ============================================================================
// PROJECT HANDLERS (READ-ONLY)
// ============================================================================

// getAvailableProjectsHandler returns projects the employee is assigned to
// NOTE: This is READ-ONLY. Project creation/management is in /api/projects
func getAvailableProjectsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		
		projects, err := services.Timesheet.GetAvailableProjects(r.Context(), employeeID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get available projects")
			return
		}
		
		if projects == nil {
			projects = []*models.Project{}
		}
		
		respondJSON(w, http.StatusOK, projects)
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func getWeekStart(t time.Time) time.Time {
	// Get Monday of current week
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
}

func getWeekEnd(t time.Time) time.Time {
	// Get Sunday of current week
	return getWeekStart(t).AddDate(0, 0, 6)
}
