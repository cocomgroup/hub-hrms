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


// RegisterRecruitingRoutes registers all recruiting-related routes
func RegisterRecruitingRoutes(r chi.Router, services *service.Services) {
	r.Route("/recruiting", func(r chi.Router) {
		// Apply auth middleware to all recruiting routes
		r.Use(authMiddleware(services))

		// Job Postings endpoints
		r.Route("/jobs", func(r chi.Router) {
			r.Get("/", listJobPostingsHandler(services))
			r.Post("/", createJobPostingHandler(services))
			r.Get("/{id}", getJobPostingHandler(services))
			r.Put("/{id}", updateJobPostingHandler(services))
			r.Delete("/{id}", deleteJobPostingHandler(services))
			r.Post("/{id}/post", postToJobBoardsHandler(services))
			r.Get("/{job_id}/candidates", getCandidatesByJobHandler(services))
		})

		// Candidates endpoints
		r.Route("/candidates", func(r chi.Router) {
			r.Post("/", createCandidateHandler(services))
			r.Get("/{id}", getCandidateHandler(services))
			r.Put("/{id}", updateCandidateHandler(services))
			r.Delete("/{id}", deleteCandidateHandler(services))
			r.Post("/{id}/analyze", analyzeResumeHandler(services))
		})

		// Email endpoints
		r.Route("/email", func(r chi.Router) {
			r.Post("/generate", generateEmailHandler(services))
			r.Post("/send", sendEmailHandler(services))
		})
	})
}

// Job Postings Handlers

// createJobPostingHandler creates a new job posting
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

		// Validate required fields
		if req.Title == "" {
			respondError(w, http.StatusBadRequest, "title is required")
			return
		}
		if req.Department == "" {
			respondError(w, http.StatusBadRequest, "department is required")
			return
		}
		if req.Location == "" {
			respondError(w, http.StatusBadRequest, "location is required")
			return
		}
		if req.EmploymentType == "" {
			respondError(w, http.StatusBadRequest, "employment_type is required")
			return
		}
		if req.Description == "" {
			respondError(w, http.StatusBadRequest, "description is required")
			return
		}

		job, err := services.Recruiting.CreateJobPosting(r.Context(), &req, userID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create job posting")
			return
		}

		respondJSON(w, http.StatusCreated, job)
	}
}

// getJobPostingHandler gets a specific job posting
func getJobPostingHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid job ID")
			return
		}

		job, err := services.Recruiting.GetJobPosting(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "job posting not found")
			return
		}

		respondJSON(w, http.StatusOK, job)
	}
}

// listJobPostingsHandler lists all job postings with optional status filter
func listJobPostingsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")

		jobs, err := services.Recruiting.ListJobPostings(r.Context(), status)
		if err != nil {
			log.Printf("ERROR: ListJobPostings failed: %v", err)  
			respondError(w, http.StatusInternalServerError, err.Error())  
			return
		}

		respondJSON(w, http.StatusOK, jobs)
	}
}

// updateJobPostingHandler updates a job posting
func updateJobPostingHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			respondError(w, http.StatusInternalServerError, "failed to update job posting")
			return
		}

		respondJSON(w, http.StatusOK, job)
	}
}

// deleteJobPostingHandler deletes a job posting
func deleteJobPostingHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid job ID")
			return
		}

		err = services.Recruiting.DeleteJobPosting(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to delete job posting")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// postToJobBoardsHandler posts a job to multiple job boards
func postToJobBoardsHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		jobID, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid job ID")
			return
		}

		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req struct {
			Boards []string `json:"boards"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if len(req.Boards) == 0 {
			respondError(w, http.StatusBadRequest, "at least one job board must be selected")
			return
		}

		postReq := &models.PostToJobBoardsRequest{
			JobPostingID: jobID,
			Boards:       req.Boards,
		}

		err = services.Recruiting.PostToJobBoards(r.Context(), postReq, userID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to post to job boards")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "Job posted to selected boards successfully",
		})
	}
}

// Candidates Handlers

// createCandidateHandler creates a new candidate
func createCandidateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateCandidateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Validate required fields
		if req.JobPostingID == uuid.Nil {
			respondError(w, http.StatusBadRequest, "job_posting_id is required")
			return
		}
		if req.FirstName == "" {
			respondError(w, http.StatusBadRequest, "first_name is required")
			return
		}
		if req.LastName == "" {
			respondError(w, http.StatusBadRequest, "last_name is required")
			return
		}
		if req.Email == "" {
			respondError(w, http.StatusBadRequest, "email is required")
			return
		}

		candidate, err := services.Recruiting.CreateCandidate(r.Context(), &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create candidate")
			return
		}

		respondJSON(w, http.StatusCreated, candidate)
	}
}

// getCandidateHandler gets a specific candidate
func getCandidateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid candidate ID")
			return
		}

		candidate, err := services.Recruiting.GetCandidate(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "candidate not found")
			return
		}

		respondJSON(w, http.StatusOK, candidate)
	}
}

// getCandidatesByJobHandler gets all candidates for a specific job
func getCandidatesByJobHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobIDStr := chi.URLParam(r, "job_id")
		jobID, err := uuid.Parse(jobIDStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid job ID")
			return
		}

		status := r.URL.Query().Get("status")

		candidates, err := services.Recruiting.GetCandidatesByJob(r.Context(), jobID, status)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get candidates")
			return
		}

		respondJSON(w, http.StatusOK, candidates)
	}
}

// updateCandidateHandler updates a candidate
func updateCandidateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			respondError(w, http.StatusInternalServerError, "failed to update candidate")
			return
		}

		respondJSON(w, http.StatusOK, candidate)
	}
}

// deleteCandidateHandler deletes a candidate
func deleteCandidateHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid candidate ID")
			return
		}

		err = services.Recruiting.DeleteCandidate(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to delete candidate")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// AI Services Handlers

// analyzeResumeHandler analyzes a candidate's resume using AI
func analyzeResumeHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		candidateID, err := uuid.Parse(idStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid candidate ID")
			return
		}

		analysis, err := services.Recruiting.AnalyzeResume(r.Context(), candidateID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to analyze resume")
			return
		}

		respondJSON(w, http.StatusOK, analysis)
	}
}

// generateEmailHandler generates an email using AI
func generateEmailHandler(services *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.EmailGenerationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Validate required fields
		if req.CandidateID == uuid.Nil {
			respondError(w, http.StatusBadRequest, "candidate_id is required")
			return
		}
		if req.JobID == uuid.Nil {
			respondError(w, http.StatusBadRequest, "job_id is required")
			return
		}
		if req.Context == "" {
			respondError(w, http.StatusBadRequest, "context is required")
			return
		}

		response, err := services.Recruiting.GenerateEmail(r.Context(), &req)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to generate email")
			return
		}

		respondJSON(w, http.StatusOK, response)
	}
}

// sendEmailHandler sends an email to a candidate
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

		// Validate required fields
		if req.CandidateID == uuid.Nil {
			respondError(w, http.StatusBadRequest, "candidate_id is required")
			return
		}
		if req.Subject == "" {
			respondError(w, http.StatusBadRequest, "subject is required")
			return
		}
		if req.Body == "" {
			respondError(w, http.StatusBadRequest, "body is required")
			return
		}

		err = services.Recruiting.SendEmail(r.Context(), &req, userID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to send email")
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"message": "Email sent successfully",
		})
	}
}