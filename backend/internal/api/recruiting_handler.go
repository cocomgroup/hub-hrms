package api

import (
	"encoding/json"
	"net/http"
	"log"

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
		})

		// Applicants/Candidates endpoints
		r.Route("/applicants", func(r chi.Router) {
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