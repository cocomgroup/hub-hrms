package api

import (
	"encoding/json"
	"net/http"
	"time"
	"log"
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// RegisterTimesheetRoutes registers all timesheet-related routes
func RegisterTimesheetRoutes(r chi.Router, services *service.Services) {
	r.Route("/timesheet", func(r chi.Router) {
		r.Use(authMiddleware(services))
		r.Get("/", listTimesheetsHandler(services))

		// Clock in/out
		r.Post("/clock-in", clockInHandler(services))
		r.Post("/clock-out", clockOutHandler(services))
		r.Get("/active", getActiveClockInHandler(services))
		
		// Time entries
		r.Get("/entries", getTimeEntriesHandler(services))
		r.Post("/entries", createTimeEntryHandler(services))
		r.Get("/entries/{id}", getTimeEntryHandler(services))
		r.Put("/entries/{id}", updateTimeEntryHandler(services))
		r.Put("/entries/{id}/submit", submitEntryHandler(services))
		r.Put("/entries/{id}/submit", recallEntryHandler(services))
		r.Put("/entries/{id}/approve", approveEntryHandler(services))
		r.Put("/entries/{id}/reject", rejectEntryHandler(services))
		
		// Submit and approval

		r.Get("/pending", getPendingApprovalsHandler(services))

		r.Get("/projects", getProjectsHandler(services))
		r.Get("/periods/current", getCurrentPeriodHandler(services))
		r.Post("/periods/{id}/submit", submitPeriodHandler(services))

				
		// Reports
		r.Get("/reports/summary", getEmployeeSummaryHandler(services))
	})
}

// clockInHandler handles clock-in requests
func clockInHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get employee ID from JWT context
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req models.ClockInRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Empty body is okay for clock in
			req = models.ClockInRequest{}
		}

		timesheet, err := services.Timesheet.ClockIn(r.Context(), employeeID, req.Notes)
		if err != nil {
			if err == service.ErrAlreadyClockedIn {
				respondError(w, http.StatusBadRequest, "already clocked in")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to clock in")
			return
		}

		respondJSON(w, http.StatusCreated, timesheet)
	}
}

// clockOutHandler handles clock-out requests
func clockOutHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get employee ID from JWT context
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req models.ClockOutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		timesheet, err := services.Timesheet.ClockOut(r.Context(), employeeID, req.BreakMinutes, req.Notes)
		if err != nil {
			if err == service.ErrNotClockedIn {
				respondError(w, http.StatusBadRequest, "not clocked in")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to clock out")
			return
		}

		respondJSON(w, http.StatusOK, timesheet)
	}
}

// getActiveClockInHandler gets the active clock-in for the employee
func getActiveClockInHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		clockIn, err := services.Timesheet.GetActiveClockIn(r.Context(), employeeID)
		if err != nil {
			// ✅ FIX: Return null instead of 500 when no active clock-in
			if err.Error() == "no rows in result set" || 
			   err.Error() == "no active clock-in" ||
			   err == sql.ErrNoRows {
				respondJSON(w, http.StatusOK, nil)
				return
			}
			log.Printf("ERROR: Failed to get active clock-in: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to get active clock-in")
			return
		}

		respondJSON(w, http.StatusOK, clockIn)
	}
}

// getTimeEntriesHandler gets time entries for the employee
func getTimeEntriesHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

        // Check if admin is viewing another employee
        targetEmployeeID := r.URL.Query().Get("employee_id")
        if targetEmployeeID != "" {
            // Verify user is admin
            userRole := r.Context().Value("user_role").(string)
            if userRole == "admin" || userRole == "hr-manager" {
                employeeID = uuid.MustParse(targetEmployeeID)
            }
        }

		// Parse query parameters
		startDateStr := r.URL.Query().Get("start_date")
		endDateStr := r.URL.Query().Get("end_date")

		var startDate, endDate time.Time
		if startDateStr != "" {
			startDate, _ = time.Parse("2006-01-02", startDateStr)
		} else {
			// Default to current week
			startDate = getWeekStart()
		}

		if endDateStr != "" {
			endDate, _ = time.Parse("2006-01-02", endDateStr)
		} else {
			endDate = getWeekEnd()
		}

		timesheets, err := services.Timesheet.GetTimeEntries(r.Context(), employeeID, startDate, endDate)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get time entries")
			return
		}

		respondJSON(w, http.StatusOK, timesheets)
	}
}

// createTimeEntryHandler creates a new time entry
func createTimeEntryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			log.Printf("ERROR: unauthorized: %s", err)
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		log.Printf("createTimeEntryHandler: Employee id: %s", employeeID)

		var req models.TimesheetCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("ERROR: invalid request body: %s", err)
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Set employee ID from context
		req.EmployeeID = employeeID

		timesheet, err := services.Timesheet.CreateTimeEntry(r.Context(), &req)
		if err != nil {
			if err == service.ErrInvalidTimeRange {
				log.Printf("ERROR: invalid time range: %v", err)
				respondError(w, http.StatusBadRequest, "invalid time range")
				return
			}
			log.Printf("ERROR: internal error: %s", err)
			respondError(w, http.StatusInternalServerError, "failed to create time entry")
			return
		}

		respondJSON(w, http.StatusCreated, timesheet)
	}
}

// getTimeEntryHandler gets a specific time entry
func getTimeEntryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid time entry ID")
			return
		}

		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Get all entries and find the specific one
		// Note: This is inefficient - in production, add GetByID to the service
		timesheets, err := services.Timesheet.GetTimeEntries(r.Context(), employeeID, time.Time{}, time.Now())
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get time entry")
			return
		}

		// Find the specific entry
		for _, ts := range timesheets {
			if ts.ID == id {
				respondJSON(w, http.StatusOK, ts)
				return
			}
		}

		respondError(w, http.StatusNotFound, "time entry not found")
	}
}

// updateTimeEntryHandler updates a time entry
func updateTimeEntryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid time entry ID")
			return
		}

		var req models.TimesheetUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		timesheet, err := services.Timesheet.UpdateTimeEntry(r.Context(), id, &req)
		if err != nil {
			if err == service.ErrTimesheetNotDraft {
				respondError(w, http.StatusBadRequest, "timesheet must be in draft status")
				return
			}
			if err == service.ErrInvalidTimeRange {
				respondError(w, http.StatusBadRequest, "invalid time range")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to update time entry")
			return
		}

		respondJSON(w, http.StatusOK, timesheet)
	}
}

// deleteTimeEntryHandler deletes a time entry
func deleteTimeEntryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid time entry ID")
			return
		}

		if err := services.Timesheet.DeleteTimeEntry(r.Context(), id); err != nil {
			if err == service.ErrTimesheetNotDraft {
				respondError(w, http.StatusBadRequest, "timesheet must be in draft status")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to delete time entry")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// submitEntryHandler submits an individual time entry for approval
func submitEntryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entryIDStr := chi.URLParam(r, "id")
		entryID, err := uuid.Parse(entryIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid entry ID")
			return
		}

		userID := getUserIDFromJWTContext(r)
		if userID == uuid.Nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		err = services.Timesheet.SubmitEntry(r.Context(), entryID, userID)
		if err != nil {
			log.Printf("ERROR: failed to submit entry: %v", err)
			if err == service.ErrUnauthorized {
				respondError(w, http.StatusForbidden, "unauthorized")
				return
			}
			if err == service.ErrInvalidStatus {
				respondError(w, http.StatusBadRequest, err.Error())
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to submit entry")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "Time entry submitted for approval",
		})
	}
}

// recallEntryHandler recalls a submitted entry back to draft
func recallEntryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entryIDStr := chi.URLParam(r, "id")
		entryID, err := uuid.Parse(entryIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid entry ID")
			return
		}

		userID := getUserIDFromJWTContext(r)
		if userID == uuid.Nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		err = services.Timesheet.RecallEntry(r.Context(), entryID, userID)
		if err != nil {
			log.Printf("ERROR: failed to recall entry: %v", err)
			if err == service.ErrUnauthorized {
				respondError(w, http.StatusForbidden, "unauthorized")
				return
			}
			if err == service.ErrInvalidStatus {
				respondError(w, http.StatusBadRequest, err.Error())
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to recall entry")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "Time entry recalled to draft",
		})
	}
}

// approveEntryHandler approves a time entry (manager only)
func approveEntryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entryIDStr := chi.URLParam(r, "id")
		entryID, err := uuid.Parse(entryIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid entry ID")
			return
		}

		userID := getUserIDFromJWTContext(r)
		if userID == uuid.Nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// TODO: Verify user is manager of the employee
		// For now, any authenticated user can approve

		err = services.Timesheet.ApproveEntry(r.Context(), entryID, userID)
		if err != nil {
			log.Printf("ERROR: failed to approve entry: %v", err)
			if err == service.ErrUnauthorized {
				respondError(w, http.StatusForbidden, "unauthorized")
				return
			}
			if err == service.ErrInvalidStatus {
				respondError(w, http.StatusBadRequest, err.Error())
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to approve entry")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "Time entry approved",
		})
	}
}

// rejectEntryHandler rejects a time entry with a reason (manager only)
func rejectEntryHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entryIDStr := chi.URLParam(r, "id")
		entryID, err := uuid.Parse(entryIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid entry ID")
			return
		}

		userID := getUserIDFromJWTContext(r)
		if userID == uuid.Nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Reason == "" {
			respondError(w, http.StatusBadRequest, "rejection reason is required")
			return
		}

		// TODO: Verify user is manager of the employee

		err = services.Timesheet.RejectEntry(r.Context(), entryID, userID, req.Reason)
		if err != nil {
			log.Printf("ERROR: failed to reject entry: %v", err)
			if err == service.ErrUnauthorized {
				respondError(w, http.StatusForbidden, "unauthorized")
				return
			}
			if err == service.ErrInvalidStatus {
				respondError(w, http.StatusBadRequest, err.Error())
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to reject entry")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "Time entry rejected",
		})
	}
}

// getPendingApprovalsHandler gets all pending timesheet approvals
func getPendingApprovalsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Note: In production, verify user has manager role

		timesheets, err := services.Timesheet.GetPendingApprovals(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get pending approvals")
			return
		}

		respondJSON(w, http.StatusOK, timesheets)
	}
}

// getEmployeeSummaryHandler gets hours summary for an employee
func getEmployeeSummaryHandler(services *service.Services) http.HandlerFunc {
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
			startDate, _ = time.Parse("2006-01-02", startDateStr)
		} else {
			startDate = getWeekStart()
		}

		if endDateStr != "" {
			endDate, _ = time.Parse("2006-01-02", endDateStr)
		} else {
			endDate = getWeekEnd()
		}

		summary, err := services.Timesheet.GetEmployeeSummary(r.Context(), employeeID, startDate, endDate)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get summary")
			return
		}

		respondJSON(w, http.StatusOK, summary)
	}
}

// List all timesheets with optional status filter
func listTimesheetsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		
		// TODO: Implement full list when service is ready
		// For now, return empty array
		_ = status
		respondJSON(w, http.StatusOK, []interface{}{})
	}
}

// getCurrentPeriodHandler gets the current timesheet period
func getCurrentPeriodHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// For now, return a mock current period
		// TODO: Implement actual period logic
		period := map[string]interface{}{
			"id": uuid.New().String(),
			"employee_id": employeeID.String(),
			"start_date": time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
			"end_date": time.Now().Format("2006-01-02"),
			"status": "open",
		}

		respondJSON(w, http.StatusOK, period)
	}
}

// submitPeriodHandler submits a timesheet period
func submitPeriodHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		periodID := chi.URLParam(r, "id")
		
		id, err := uuid.Parse(periodID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid period ID")
			return
		}

		employeeID, err := getEmployeeIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// TODO: Implement actual period submission logic
		log.Printf("DEBUG: Submitting period %s for employee %s", id, employeeID)

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "period submitted successfully",
		})
	}
}

// getProjectsHandler gets all active projects
// (This function should already exist around line 371, but here's the correct version)
func getProjectsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get projects from database
		// For now return empty array
		projects := []interface{}{}
		
		// TODO: Get actual projects from services
		// projects, err := services.Timesheet.GetProjects(r.Context())
		// if err != nil {
		//     respondError(w, http.StatusInternalServerError, "failed to get projects")
		//     return
		// }

		respondJSON(w, http.StatusOK, projects)
	}
}

// Helper functions


func getWeekStart() time.Time {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
}

func getWeekEnd() time.Time {
	return getWeekStart().AddDate(0, 0, 6)
}