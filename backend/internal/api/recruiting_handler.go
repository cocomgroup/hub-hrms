package api

import (
	"encoding/json"
	"net/http"
	"log"
	"strings"
	"regexp"
	"strconv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/service"
)

// RegisterRecruitingRoutesComplete registers all recruiting-related routes including new endpoints
func RegisterRecruitingRoutes(r chi.Router, services *service.Services) {
	r.Route("/recruiting", func(r chi.Router) {
		// Apply auth middleware to all recruiting routes
		r.Use(authMiddleware(services))

		// Dashboard endpoints
		r.Get("/stats", getRecruitingStatsHandler(services))
		r.Get("/dashboard", getRecruitingDashboardHandler(services))

		// Provider endpoints
		r.Route("/providers", func(r chi.Router) {
			r.Get("/", listProvidersHandler(services))
			r.Post("/", createProviderHandler(services))
			r.Get("/{id}", getProviderHandler(services))
			r.Put("/{id}", updateProviderHandler(services))
			r.Delete("/{id}", deleteProviderHandler(services))
			r.Post("/{id}/test", testProviderConnectionHandler(services))
		})

		// Job Postings endpoints
		r.Route("/jobs", func(r chi.Router) {
			r.Get("/", listJobPostingsHandler(services))
			r.Post("/", createJobPostingHandler(services))
			r.Get("/{id}", getJobPostingHandler(services))
			r.Put("/{id}", updateJobPostingHandler(services))
			r.Delete("/{id}", deleteJobPostingHandler(services))
			r.Post("/{id}/post", postToJobBoardsHandler(services))
			r.Post("/{id}/close", closeJobPostingHandler(services))
			r.Get("/{job_id}/candidates", getCandidatesByJobHandler(services))
			r.Post("/upload", uploadJobHandler(services))
		})

		// Applicants/Candidates endpoints
		r.Route("/applicants", func(r chi.Router) {
			r.Post("/upload", uploadApplicantResumeHandler(services))
			r.Get("/", listApplicantsHandler(services))
			r.Get("/leaderboard", getApplicantLeaderboardHandler(services))
			r.Post("/{id}/analyze", analyzeCandidateHandler(services))
			r.Patch("/{id}", updateCandidateStatusHandler(services))
		})

		// Candidates endpoints (existing)
		r.Route("/candidates", func(r chi.Router) {
			r.Post("/", createCandidateHandler(services))
			r.Get("/{id}", getCandidateHandler(services))
			r.Put("/{id}", updateCandidateHandler(services))
			r.Delete("/{id}", deleteCandidateHandler(services))
		})

		// Interview endpoints
		r.Route("/interviews", func(r chi.Router) {
			r.Post("/", scheduleInterviewHandler(services))
			r.Put("/{id}", updateInterviewHandler(services))
			r.Get("/candidate/{candidate_id}", getInterviewsByCandidateHandler(services))
		})

		// Email endpoints
		r.Route("/email", func(r chi.Router) {
			r.Post("/generate", generateEmailHandler(services))
			r.Post("/send", sendEmailHandler(services))
		})
	})
}

// Dashboard Handlers

func getRecruitingStatsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		stats, err := services.Recruiting.GetDashboardStats(r.Context())
		if err != nil {
			log.Printf("Failed to get recruiting stats: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to get stats")
			return
		}

		respondJSON(w, http.StatusOK, stats)
	}
}

func getRecruitingDashboardHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		dashboard, err := services.Recruiting.GetDashboard(r.Context())
		if err != nil {
			log.Printf("Failed to get recruiting dashboard: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to get dashboard")
			return
		}

		respondJSON(w, http.StatusOK, dashboard)
	}
}

// Provider Handlers

func listProvidersHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		providers, err := services.Recruiting.GetAllProviders(r.Context())
		if err != nil {
			log.Printf("Failed to list providers: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to list providers")
			return
		}

		respondJSON(w, http.StatusOK, providers)
	}
}

func createProviderHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req models.CreateProviderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Validate required fields
		if req.Type == "" {
			respondError(w, http.StatusBadRequest, "type is required")
			return
		}
		if req.Name == "" {
			respondError(w, http.StatusBadRequest, "name is required")
			return
		}
		if req.Config == nil || len(req.Config) == 0 {
			respondError(w, http.StatusBadRequest, "config is required")
			return
		}

		provider, err := services.Recruiting.CreateProvider(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to create provider: %v", err)
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondJSON(w, http.StatusCreated, provider)
	}
}

func getProviderHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid provider ID")
			return
		}

		provider, err := services.Recruiting.GetProvider(r.Context(), id)
		if err != nil {
			log.Printf("Failed to get provider: %v", err)
			respondError(w, http.StatusNotFound, "provider not found")
			return
		}

		respondJSON(w, http.StatusOK, provider)
	}
}

func updateProviderHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid provider ID")
			return
		}

		var req models.UpdateProviderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		provider, err := services.Recruiting.UpdateProvider(r.Context(), id, &req)
		if err != nil {
			log.Printf("Failed to update provider: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to update provider")
			return
		}

		respondJSON(w, http.StatusOK, provider)
	}
}

func deleteProviderHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid provider ID")
			return
		}

		err = services.Recruiting.DeleteProvider(r.Context(), id)
		if err != nil {
			log.Printf("Failed to delete provider: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to delete provider")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"message": "provider deleted successfully"})
	}
}

func testProviderConnectionHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid provider ID")
			return
		}

		// Get provider to test
		provider, err := services.Recruiting.GetProvider(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "provider not found")
			return
		}

		// Test connection
		testReq := &models.TestProviderConnectionRequest{
			Type:   provider.Type,
			Config: provider.Config,
		}

		result, err := services.Recruiting.TestProviderConnection(r.Context(), testReq)
		if err != nil {
			log.Printf("Failed to test provider connection: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to test connection")
			return
		}

		// Update provider connection status
		if result.Success {
			updateReq := &models.UpdateProviderRequest{}
			config := provider.Config
			updateReq.Config = &config
			services.Recruiting.UpdateProvider(r.Context(), id, updateReq)
		}

		respondJSON(w, http.StatusOK, result)
	}
}

// Applicant Handlers

func listApplicantsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		status := r.URL.Query().Get("status")
		jobID := r.URL.Query().Get("job_id")

		var candidates []*models.Candidate

		if jobID != "" {
			jobUUID, err := uuid.Parse(jobID)
			if err != nil {
				respondError(w, http.StatusBadRequest, "invalid job ID")
				return
			}
			candidates, err = services.Recruiting.GetCandidatesByJob(r.Context(), jobUUID, status)
			if err != nil {
				log.Printf("Failed to get candidates: %v", err)
				respondError(w, http.StatusInternalServerError, "failed to get candidates")
				return
			}
		} else {
			// Get all candidates (you'd need to add this method)
			// For now, return empty array
			candidates = []*models.Candidate{}
		}

		respondJSON(w, http.StatusOK, candidates)
	}
}

func getApplicantLeaderboardHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		leaderboard, err := services.Recruiting.GetApplicantLeaderboard(r.Context())
		if err != nil {
			log.Printf("Failed to get leaderboard: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to get leaderboard")
			return
		}

		respondJSON(w, http.StatusOK, leaderboard)
	}
}

func analyzeCandidateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		candidateID, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid candidate ID")
			return
		}

		analysis, err := services.Recruiting.AnalyzeResume(r.Context(), candidateID)
		if err != nil {
			log.Printf("Failed to analyze candidate: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to analyze candidate")
			return
		}

		respondJSON(w, http.StatusOK, analysis)
	}
}

func updateCandidateStatusHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		candidateID, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid candidate ID")
			return
		}

		var req models.UpdateCandidateStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		updateReq := &models.UpdateCandidateRequest{
			Status: &req.Status,
			Notes:  req.Notes,
		}

		candidate, err := services.Recruiting.UpdateCandidate(r.Context(), candidateID, updateReq)
		if err != nil {
			log.Printf("Failed to update candidate: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to update candidate")
			return
		}

		respondJSON(w, http.StatusOK, candidate)
	}
}

func closeJobPostingHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		jobID, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid job ID")
			return
		}

		// Update job status to closed
		status := "closed"
		req := &models.UpdateJobPostingRequest{
			Status: &status,
		}

		job, err := services.Recruiting.UpdateJobPosting(r.Context(), jobID, req)
		if err != nil {
			log.Printf("Failed to close job posting: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to close job posting")
			return
		}

		respondJSON(w, http.StatusOK, job)
	}
}

// Interview Handlers

func scheduleInterviewHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req models.ScheduleInterviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Validate required fields
		if req.CandidateID == uuid.Nil {
			respondError(w, http.StatusBadRequest, "candidate_id is required")
			return
		}
		if req.InterviewerID == uuid.Nil {
			respondError(w, http.StatusBadRequest, "interviewer_id is required")
			return
		}
		if req.Duration <= 0 {
			respondError(w, http.StatusBadRequest, "duration must be positive")
			return
		}

		interview, err := services.Recruiting.ScheduleInterview(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to schedule interview: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to schedule interview")
			return
		}

		respondJSON(w, http.StatusCreated, interview)
	}
}

func updateInterviewHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		interviewID, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid interview ID")
			return
		}

		var req models.UpdateInterviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		interview, err := services.Recruiting.UpdateInterview(r.Context(), interviewID, &req)
		if err != nil {
			log.Printf("Failed to update interview: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to update interview")
			return
		}

		respondJSON(w, http.StatusOK, interview)
	}
}

func getInterviewsByCandidateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		candidateIDStr := chi.URLParam(r, "candidate_id")
		candidateID, err := uuid.Parse(candidateIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid candidate ID")
			return
		}

		interviews, err := services.Recruiting.GetInterviewsByCandidate(r.Context(), candidateID)
		if err != nil {
			log.Printf("Failed to get interviews: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to get interviews")
			return
		}

		respondJSON(w, http.StatusOK, interviews)
	}
}

func listJobPostingsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		status := r.URL.Query().Get("status") // active, closed, draft

		jobs, err := services.Recruiting.ListJobPostings(r.Context(), status)
		if err != nil {
			log.Printf("Failed to list job postings: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to list jobs")
			return
		}

		respondJSON(w, http.StatusOK, jobs)
	}
}

func createJobPostingHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req models.CreateJobPostingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		job, err := services.Recruiting.CreateJobPosting(r.Context(), &req, userID)
		if err != nil {
			log.Printf("Failed to create job posting: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to create job")
			return
		}

		respondJSON(w, http.StatusCreated, job)
	}
}

func getJobPostingHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid job ID")
			return
		}

		job, err := services.Recruiting.GetJobPosting(r.Context(), id)
		if err != nil {
			log.Printf("Failed to get job posting: %v", err)
			respondError(w, http.StatusNotFound, "job not found")
			return
		}

		respondJSON(w, http.StatusOK, job)
	}
}

func updateJobPostingHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid job ID")
			return
		}

		var req models.UpdateJobPostingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		job, err := services.Recruiting.UpdateJobPosting(r.Context(), id, &req)
		if err != nil {
			log.Printf("Failed to update job posting: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to update job")
			return
		}

		respondJSON(w, http.StatusOK, job)
	}
}

func deleteJobPostingHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid job ID")
			return
		}

		err = services.Recruiting.DeleteJobPosting(r.Context(), id)
		if err != nil {
			log.Printf("Failed to delete job posting: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to delete job")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"message": "job deleted successfully"})
	}
}

func uploadJobHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Parse multipart form (10MB max)
		err = r.ParseMultipartForm(10 << 20)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Failed to parse form")
			return
		}

		// Get the file
		file, header, err := r.FormFile("file")
		if err != nil {
			respondError(w, http.StatusBadRequest, "Failed to get file from form")
			return
		}
		defer file.Close()

		// Validate file size (10MB)
		if header.Size > 10*1024*1024 {
			respondError(w, http.StatusBadRequest, "File size exceeds 10MB limit")
			return
		}

		// Validate file name
		if strings.Contains(header.Filename, "..") || strings.Contains(header.Filename, "/") {
			respondError(w, http.StatusBadRequest, "Invalid filename")
			return
		}

		// Read file content
		content, err := io.ReadAll(file)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to read file")
			return
		}

		// Parse based on file extension
		var jobData models.JobUploadRequest
		filename := strings.ToLower(header.Filename)

		if strings.HasSuffix(filename, ".json") {
			err = json.Unmarshal(content, &jobData)
			if err != nil {
				respondError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse JSON: %v", err))
				return
			}
		} else if strings.HasSuffix(filename, ".md") || strings.HasSuffix(filename, ".txt") {
			jobData, err = parseJobFromText(string(content))
			if err != nil {
				respondError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse text file: %v", err))
				return
			}
		} else {
			respondError(w, http.StatusBadRequest, "Unsupported file type. Use .json, .md, or .txt")
			return
		}

		// Validate required fields
		if jobData.Title == "" {
			respondError(w, http.StatusBadRequest, "Job title is required")
			return
		}

		// Set defaults
		if jobData.EmploymentType == "" {
			jobData.EmploymentType = "full-time"
		}
		if jobData.Requirements == nil {
			jobData.Requirements = []string{}
		}
		if jobData.Responsibilities == nil {
			jobData.Responsibilities = []string{}
		}
		if jobData.Benefits == nil {
			jobData.Benefits = []string{}
		}

		// Check if we should save directly or just return parsed data
		saveDirectly := r.URL.Query().Get("save") == "true"

		if saveDirectly {
			// Create job posting directly in database
			createReq := &models.CreateJobPostingRequest{
				Title:            jobData.Title,
				Department:       jobData.Department,
				Location:         jobData.Location,
				EmploymentType:   jobData.EmploymentType,
				SalaryMin:        convertIntToFloatPtr(jobData.SalaryMin),
				SalaryMax:        convertIntToFloatPtr(jobData.SalaryMax),
				Description:      jobData.Description,
				Requirements:     jobData.Requirements,
				Responsibilities: jobData.Responsibilities,
				Benefits:         jobData.Benefits,
			}

			job, err := services.Recruiting.CreateJobPosting(r.Context(), createReq, userID)
			if err != nil {
				log.Printf("Failed to create job from upload: %v", err)
				respondError(w, http.StatusInternalServerError, "failed to create job")
				return
			}

			respondJSON(w, http.StatusCreated, job)
		} else {
			// Return parsed data for frontend review
			respondJSON(w, http.StatusOK, jobData)
		}
	}
}
// Helper function to convert int to *float64
func convertIntToFloatPtr(val int) *float64 {
	if val == 0 {
		return nil
	}
	f := float64(val)
	return &f
}

// parseJobFromText parses markdown/text format job postings
func parseJobFromText(content string) (models.JobUploadRequest, error) {
	job := models.JobUploadRequest{
		Requirements:     []string{},
		Responsibilities: []string{},
		Benefits:         []string{},
	}

	lines := strings.Split(content, "\n")
	currentSection := ""
	descriptionLines := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Extract title (first h1)
		if job.Title == "" && strings.HasPrefix(trimmed, "# ") {
			job.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
			continue
		}

		// Check for **Label:** format (Position Overview section)
		if strings.Contains(trimmed, "**") && strings.Contains(trimmed, ":") {
			// Extract bold label and value
			parts := strings.Split(trimmed, ":")
			if len(parts) >= 2 {
				label := strings.ToLower(strings.Trim(parts[0], "* "))
				value := strings.TrimSpace(strings.Join(parts[1:], ":"))
				
				// Remove trailing ** from value
				value = strings.TrimSuffix(value, "**")
				value = strings.TrimSpace(value)
				
				if strings.Contains(label, "department") {
					job.Department = value
					continue
				}
				if strings.Contains(label, "location") {
					job.Location = value
					continue
				}
				if strings.Contains(label, "title") && job.Title == "" {
					job.Title = value
					continue
				}
				if strings.Contains(label, "employment") || strings.Contains(label, "type") {
					lowerValue := strings.ToLower(value)
					if strings.Contains(lowerValue, "full-time") || strings.Contains(lowerValue, "full time") {
						job.EmploymentType = "full-time"
					} else if strings.Contains(lowerValue, "part-time") || strings.Contains(lowerValue, "part time") {
						job.EmploymentType = "part-time"
					} else if strings.Contains(lowerValue, "contract") {
						job.EmploymentType = "contract"
					} else if strings.Contains(lowerValue, "intern") {
						job.EmploymentType = "internship"
					}
					continue
				}
			}
		}

		// Detect sections by headers
		if strings.HasPrefix(trimmed, "## ") {
			sectionName := strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			sectionName = strings.ToLower(sectionName)

			if strings.Contains(sectionName, "description") || 
			   strings.Contains(sectionName, "about") || 
			   strings.Contains(sectionName, "overview") ||
			   strings.Contains(sectionName, "summary") {
				currentSection = "description"
			} else if strings.Contains(sectionName, "requirement") || 
					  strings.Contains(sectionName, "qualification") ||
					  strings.Contains(sectionName, "skills") {
				currentSection = "requirements"
			} else if strings.Contains(sectionName, "responsibilit") || 
					  strings.Contains(sectionName, "duties") ||
					  strings.Contains(sectionName, "role") {
				currentSection = "responsibilities"
			} else if strings.Contains(sectionName, "benefit") || 
					  strings.Contains(sectionName, "perk") ||
					  strings.Contains(sectionName, "compensation") {
				currentSection = "benefits"
			} else if strings.Contains(sectionName, "department") {
				currentSection = "department"
			} else if strings.Contains(sectionName, "location") {
				currentSection = "location"
			} else if strings.Contains(sectionName, "position") {
				// Position Overview section - continue processing
				continue
			} else {
				// Unknown section, keep processing in current section
			}
			continue
		}

		// Extract salary range
		salaryRegex := regexp.MustCompile(`\$?\s*([\d,]+)\s*-\s*\$?\s*([\d,]+)`)
		if matches := salaryRegex.FindStringSubmatch(trimmed); len(matches) == 3 {
			min, _ := strconv.Atoi(strings.ReplaceAll(matches[1], ",", ""))
			max, _ := strconv.Atoi(strings.ReplaceAll(matches[2], ",", ""))
			job.SalaryMin = min
			job.SalaryMax = max
			continue
		}

		// Extract employment type from inline mentions
		lowerTrimmed := strings.ToLower(trimmed)
		if strings.Contains(lowerTrimmed, "full-time") || strings.Contains(lowerTrimmed, "full time") {
			if job.EmploymentType == "" {
				job.EmploymentType = "full-time"
			}
		} else if strings.Contains(lowerTrimmed, "part-time") || strings.Contains(lowerTrimmed, "part time") {
			if job.EmploymentType == "" {
				job.EmploymentType = "part-time"
			}
		} else if strings.Contains(lowerTrimmed, "contract") {
			if job.EmploymentType == "" {
				job.EmploymentType = "contract"
			}
		} else if strings.Contains(lowerTrimmed, "intern") {
			if job.EmploymentType == "" {
				job.EmploymentType = "internship"
			}
		}

		// Simple key: value format (for text files)
		if strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "*") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				key := strings.ToLower(strings.TrimSpace(parts[0]))
				value := strings.TrimSpace(parts[1])
				
				if key == "department" {
					job.Department = value
					continue
				}
				if key == "location" {
					job.Location = value
					continue
				}
				if strings.Contains(key, "type") {
					lowerValue := strings.ToLower(value)
					if strings.Contains(lowerValue, "full-time") || strings.Contains(lowerValue, "full time") {
						job.EmploymentType = "full-time"
					} else if strings.Contains(lowerValue, "part-time") || strings.Contains(lowerValue, "part time") {
						job.EmploymentType = "part-time"
					} else if strings.Contains(lowerValue, "contract") {
						job.EmploymentType = "contract"
					} else if strings.Contains(lowerValue, "intern") {
						job.EmploymentType = "internship"
					}
					continue
				}
			}
		}

		// Add content to sections
		if currentSection != "" {
			// Clean bullet points and numbering
			content := trimmed
			content = strings.TrimPrefix(content, "- ")
			content = strings.TrimPrefix(content, "* ")
			content = strings.TrimPrefix(content, "• ")
			content = regexp.MustCompile(`^\d+\.\s+`).ReplaceAllString(content, "")
			
			// Skip if it's a header or bold label line
			if strings.HasPrefix(content, "#") || strings.HasPrefix(content, "**") {
				continue
			}
			
			if content == "" {
				continue
			}

			switch currentSection {
			case "description":
				// Skip Position Overview section content (already processed)
				if !strings.Contains(strings.ToLower(content), "title:") && 
				   !strings.Contains(strings.ToLower(content), "department:") && 
				   !strings.Contains(strings.ToLower(content), "location:") {
					descriptionLines = append(descriptionLines, content)
				}
			case "department":
				if job.Department == "" {
					job.Department = content
				}
				currentSection = "" // Single value
			case "location":
				if job.Location == "" {
					job.Location = content
				}
				currentSection = "" // Single value
			case "requirements":
				job.Requirements = append(job.Requirements, content)
			case "responsibilities":
				job.Responsibilities = append(job.Responsibilities, content)
			case "benefits":
				job.Benefits = append(job.Benefits, content)
			}
		}
	}

	// Join description lines
	if len(descriptionLines) > 0 {
		job.Description = strings.Join(descriptionLines, "\n")
	}

	// Set defaults if missing
	if job.EmploymentType == "" {
		job.EmploymentType = "full-time"
	}

	// Set default currency
	job.SalaryCurrency = "USD"

	return job, nil
}


func postToJobBoardsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		jobID, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid job ID")
			return
		}

		var req models.PostToJobBoardsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req.JobPostingID = jobID

		err = services.Recruiting.PostToJobBoards(r.Context(), &req, userID)
		if err != nil {
			log.Printf("Failed to post to job boards: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to post job")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"message": "job posted successfully"})
	}
}

func getCandidatesByJobHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		jobIDStr := chi.URLParam(r, "job_id")
		jobID, err := uuid.Parse(jobIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid job ID")
			return
		}

		status := r.URL.Query().Get("status") // applied, screening, interviewing, etc.

		candidates, err := services.Recruiting.GetCandidatesByJob(r.Context(), jobID, status)
		if err != nil {
			log.Printf("Failed to get candidates by job: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to get candidates")
			return
		}

		respondJSON(w, http.StatusOK, candidates)
	}
}

// Candidates Handlers

func createCandidateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req models.CreateCandidateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		candidate, err := services.Recruiting.CreateCandidate(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to create candidate: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to create candidate")
			return
		}

		respondJSON(w, http.StatusCreated, candidate)
	}
}

func getCandidateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid candidate ID")
			return
		}

		candidate, err := services.Recruiting.GetCandidate(r.Context(), id)
		if err != nil {
			log.Printf("Failed to get candidate: %v", err)
			respondError(w, http.StatusNotFound, "candidate not found")
			return
		}

		respondJSON(w, http.StatusOK, candidate)
	}
}

func updateCandidateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid candidate ID")
			return
		}

		var req models.UpdateCandidateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		candidate, err := services.Recruiting.UpdateCandidate(r.Context(), id, &req)
		if err != nil {
			log.Printf("Failed to update candidate: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to update candidate")
			return
		}

		respondJSON(w, http.StatusOK, candidate)
	}
}

func deleteCandidateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid candidate ID")
			return
		}

		err = services.Recruiting.DeleteCandidate(r.Context(), id)
		if err != nil {
			log.Printf("Failed to delete candidate: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to delete candidate")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"message": "candidate deleted successfully"})
	}
}

// Email Handlers

func generateEmailHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req models.EmailGenerationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		response, err := services.Recruiting.GenerateEmail(r.Context(), &req)
		if err != nil {
			log.Printf("Failed to generate email: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to generate email")
			return
		}

		respondJSON(w, http.StatusOK, response)
	}
}

func sendEmailHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req models.SendEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		err = services.Recruiting.SendEmail(r.Context(), &req, userID)
		if err != nil {
			log.Printf("Failed to send email: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to send email")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"message": "email sent successfully"})
	}
}

// uploadApplicantResumeHandler handles resume file uploads
func uploadApplicantResumeHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form (10 MB max)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			respondError(w, http.StatusBadRequest, "File too large. Maximum size is 10MB")
			return
		}

		// Get form fields
		name := r.FormValue("name")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		position := r.FormValue("position")
		source := r.FormValue("source")

		// Validate required fields
		if name == "" || email == "" || position == "" {
			respondError(w, http.StatusBadRequest, "Name, email, and position are required")
			return
		}

		// Get uploaded file
		file, header, err := r.FormFile("resume")
		if err != nil {
			respondError(w, http.StatusBadRequest, "Resume file is required")
			return
		}
		defer file.Close()

		// Validate file extension
		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowedExts := map[string]bool{
			".pdf":  true,
			".doc":  true,
			".docx": true,
		}
		if !allowedExts[ext] {
			respondError(w, http.StatusBadRequest, "Invalid file type. Only PDF and Word documents are allowed")
			return
		}

		// Validate file size (5MB)
		if header.Size > 5*1024*1024 {
			respondError(w, http.StatusBadRequest, "File size must be less than 5MB")
			return
		}

		// Generate unique filename
		applicantID := uuid.New()
		filename := fmt.Sprintf("%s_%s%s", applicantID.String(), sanitizeFilename(name), ext)
		
		// Determine upload path based on environment
		var uploadPath string
		if os.Getenv("ENVIRONMENT") == "production" || os.Getenv("AWS_REGION") != "" {
			// AWS environment - use /tmp for Lambda or writable directory
			uploadPath = "/tmp/resumes"
		} else {
			// Local/Windows development
			uploadPath = "./uploads/resumes"
		}
		
		// Create directory if it doesn't exist
		if err := os.MkdirAll(uploadPath, 0755); err != nil {
			log.Printf("Failed to create upload directory: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to create upload directory")
			return
		}
		
		filepath := filepath.Join(uploadPath, filename)

		// Save file to disk
		dst, err := os.Create(filepath)
		if err != nil {
			log.Printf("Failed to create file: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to save file")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			os.Remove(filepath)
			log.Printf("Failed to copy file: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to save file")
			return
		}

		// Create applicant record
		applicant := &models.Applicant{
			ID:          applicantID,
			Name:        name,
			Email:       email,
			Phone:       phone,
			Position:    position,
			Source:      source,
			ResumeURL:   fmt.Sprintf("/uploads/resumes/%s", filename),
			AppliedDate: time.Now(),
			Status:      "new",
			AIScore:     0.0,
			Notes:       "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Save to database via service
		if err := services.Recruiting.CreateApplicant(r.Context(), applicant); err != nil {
			os.Remove(filepath)
			log.Printf("Failed to create applicant: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to create applicant record")
			return
		}

		// TODO: Trigger AI analysis in background
		// go services.AI.AnalyzeResume(applicant.ID)

		respondJSON(w, http.StatusOK, applicant)
	}
}

// sanitizeFilename removes special characters from filename
func sanitizeFilename(name string) string {
	// Remove special characters and replace spaces
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	
	sanitized := replacer.Replace(name)
	// Limit length
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}
	
	return strings.ToLower(sanitized)
}

