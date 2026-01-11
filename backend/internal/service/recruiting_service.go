package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
)

type RecruitingService interface {
	// Job Postings
	CreateJobPosting(ctx context.Context, req *models.CreateJobPostingRequest, userID uuid.UUID) (*models.JobPosting, error)
	GetJobPosting(ctx context.Context, id uuid.UUID) (*models.JobPosting, error)
	ListJobPostings(ctx context.Context, status string) ([]*models.JobPosting, error)
	UpdateJobPosting(ctx context.Context, id uuid.UUID, req *models.UpdateJobPostingRequest) (*models.JobPosting, error)
	DeleteJobPosting(ctx context.Context, id uuid.UUID) error
	PostToJobBoards(ctx context.Context, req *models.PostToJobBoardsRequest, userID uuid.UUID) error
	PostToJobBoardsByProviderIDs(ctx context.Context, jobID uuid.UUID, providerIDs []uuid.UUID, userID uuid.UUID) error


	// Applicants
	CreateApplicant(ctx context.Context, applicant *models.Applicant) error

	// Candidates
	CreateCandidate(ctx context.Context, req *models.CreateCandidateRequest) (*models.Candidate, error)
	GetCandidate(ctx context.Context, id uuid.UUID) (*models.Candidate, error)
	GetCandidatesByJob(ctx context.Context, jobID uuid.UUID, status string) ([]*models.Candidate, error)
	UpdateCandidate(ctx context.Context, id uuid.UUID, req *models.UpdateCandidateRequest) (*models.Candidate, error)
	DeleteCandidate(ctx context.Context, id uuid.UUID) error

	// AI Services
	AnalyzeResume(ctx context.Context, candidateID uuid.UUID) (*models.ResumeAnalysisResponse, error)
	GenerateEmail(ctx context.Context, req *models.EmailGenerationRequest) (*models.EmailGenerationResponse, error)

	// Email
	SendEmail(ctx context.Context, req *models.SendEmailRequest, sentBy uuid.UUID) error

		// Providers
	CreateProvider(ctx context.Context, req *models.CreateProviderRequest) (*models.RecruitingProvider, error)
	GetProvider(ctx context.Context, id uuid.UUID) (*models.RecruitingProvider, error)
	GetAllProviders(ctx context.Context) ([]*models.RecruitingProvider, error)
	UpdateProvider(ctx context.Context, id uuid.UUID, req *models.UpdateProviderRequest) (*models.RecruitingProvider, error)
	DeleteProvider(ctx context.Context, id uuid.UUID) error
	TestProviderConnection(ctx context.Context, req *models.TestProviderConnectionRequest) (*models.TestProviderConnectionResponse, error)

	// Dashboard
	GetDashboardStats(ctx context.Context) (*models.RecruitingDashboardStats, error)
	GetDashboard(ctx context.Context) (*models.RecruitingDashboard, error)
	GetApplicantLeaderboard(ctx context.Context) ([]*models.ApplicantLeaderboard, error)

	// Interviews
	ScheduleInterview(ctx context.Context, req *models.ScheduleInterviewRequest) (*models.Interview, error)
	UpdateInterview(ctx context.Context, id uuid.UUID, req *models.UpdateInterviewRequest) (*models.Interview, error)
	GetInterviewsByCandidate(ctx context.Context, candidateID uuid.UUID) ([]*models.Interview, error)
}

type recruitingService struct {
	repos *repository.Repositories
}

func NewRecruitingService(repos *repository.Repositories) RecruitingService {
	return &recruitingService{repos: repos}
}

// Job Postings
func (s *recruitingService) CreateJobPosting(ctx context.Context, req *models.CreateJobPostingRequest, userID uuid.UUID) (*models.JobPosting, error) {
	now := time.Now()
	job := &models.JobPosting{
		ID:                uuid.New(),
		Title:             req.Title,
		Department:        req.Department,
		Location:          req.Location,
		EmploymentType:    req.EmploymentType,
		SalaryMin:         req.SalaryMin,
		SalaryMax:         req.SalaryMax,
		SalaryCurrency:    "USD", // Default to USD
		Description:       req.Description,
		Requirements:      models.FromSlice(req.Requirements),
		Responsibilities:  models.FromSlice(req.Responsibilities),
		Benefits:          models.FromSlice(req.Benefits),
		Status:            "draft",
		Providers:         models.StringArray{}, // Initialize empty providers array
		PostedDate:        nil,
		ApplicationsCount: 0,
		CreatedBy:         userID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err := s.repos.Recruiting.CreateJobPosting(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to create job posting: %w", err)
	}

	return job, nil
}

func (s *recruitingService) GetJobPosting(ctx context.Context, id uuid.UUID) (*models.JobPosting, error) {
	return s.repos.Recruiting.GetJobPosting(ctx, id)
}

func (s *recruitingService) ListJobPostings(ctx context.Context, status string) ([]*models.JobPosting, error) {
	return s.repos.Recruiting.ListJobPostings(ctx, status)
}

func (s *recruitingService) UpdateJobPosting(ctx context.Context, id uuid.UUID, req *models.UpdateJobPostingRequest) (*models.JobPosting, error) {
	// Build updates map
	updates := make(map[string]interface{})
	
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Department != nil {
		updates["department"] = *req.Department
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.EmploymentType != nil {
		updates["employment_type"] = *req.EmploymentType
	}
	if req.SalaryMin != nil {
		updates["salary_min"] = *req.SalaryMin
	}
	if req.SalaryMax != nil {
		updates["salary_max"] = *req.SalaryMax
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Requirements != nil {
		updates["requirements"] = req.Requirements
	}
	if req.Responsibilities != nil {
		updates["responsibilities"] = req.Responsibilities
	}
	if req.Benefits != nil {
		updates["benefits"] = req.Benefits
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	err := s.repos.Recruiting.UpdateJobPosting(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update job posting: %w", err)
	}

	return s.repos.Recruiting.GetJobPosting(ctx, id)
}

func (s *recruitingService) DeleteJobPosting(ctx context.Context, id uuid.UUID) error {
	return s.repos.Recruiting.DeleteJobPosting(ctx, id)
}

func (s *recruitingService) PostToJobBoards(ctx context.Context, req *models.PostToJobBoardsRequest, userID uuid.UUID) error {
	// Get the job posting
	job, err := s.repos.Recruiting.GetJobPosting(ctx, req.JobPostingID)
	if err != nil {
		return fmt.Errorf("failed to get job posting: %w", err)
	}

	// Create job board postings for each selected board
	now := time.Now()
	expiresAt := now.AddDate(0, 1, 0) // 1 month from now

	for _, boardName := range req.Boards {
		posting := &models.JobBoardPosting{
			ID:           uuid.New(),
			JobPostingID: job.ID,
			BoardName:    boardName,
			ExternalID:   nil, // Would be set by actual API integration
			PostedAt:     now,
			ExpiresAt:    &expiresAt,
			Status:       "active",
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		err := s.repos.Recruiting.CreateJobBoardPosting(ctx, posting)
		if err != nil {
			return fmt.Errorf("failed to post to %s: %w", boardName, err)
		}
	}

	// Update job status to active if it was draft
	if job.Status == "draft" {
		updates := map[string]interface{}{
			"status":      "active",
			"posted_date": &now,
		}
		err = s.repos.Recruiting.UpdateJobPosting(ctx, job.ID, updates)
		if err != nil {
			return fmt.Errorf("failed to update job status: %w", err)
		}
	}

	return nil
}

func (s *recruitingService) PostToJobBoardsByProviderIDs(ctx context.Context, jobID uuid.UUID, providerIDs []uuid.UUID, userID uuid.UUID) error {
	// Get job posting
	job, err := s.repos.Recruiting.GetJobPosting(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job posting: %w", err)
	}

	// Update job status to active
	if job.Status == "draft" {
		now := time.Now()
		updates := map[string]interface{}{
			"status":      "active",
			"posted_date": now,
		}
		err = s.repos.Recruiting.UpdateJobPosting(ctx, jobID, updates)
		if err != nil {
			return fmt.Errorf("failed to update job status: %w", err)
		}
	}

	// Post to each selected provider
	now := time.Now()
	for _, providerID := range providerIDs {
		provider, err := s.repos.Recruiting.GetProvider(ctx, providerID)
		if err != nil {
			continue // Skip failed providers
		}

		if !provider.IsConnected {
			continue // Skip disconnected providers
		}

		// In production, make actual API calls to each provider
		// For now, just create a job board posting record
		posting := &models.JobBoardPosting{
			ID:           uuid.New(),
			JobPostingID: jobID,
			BoardName:    provider.Type,
			PostedAt:     now,
			Status:       "active",
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		err = s.repos.Recruiting.CreateJobBoardPosting(ctx, posting)
		if err != nil {
			continue // Skip failed postings
		}

		// Update provider stats
		s.repos.Recruiting.UpdateProviderStats(ctx, providerID, 1, 0)
	}

	return nil
}

// Candidates
func (s *recruitingService) CreateCandidate(ctx context.Context, req *models.CreateCandidateRequest) (*models.Candidate, error) {
	now := time.Now()
	candidate := &models.Candidate{
		ID:           uuid.New(),
		JobPostingID: req.JobPostingID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Phone:        req.Phone,
		ResumeURL:    req.ResumeURL,
		CoverLetter:  req.CoverLetter,
		LinkedInURL:  req.LinkedInURL,
		PortfolioURL: req.PortfolioURL,
		Status:       "new",
		Skills:       req.Skills,
		AppliedDate:  now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Initialize empty arrays if nil
	if candidate.Skills == nil {
		candidate.Skills = []string{}
	}
	if candidate.Strengths == nil {
		candidate.Strengths = []string{}
	}
	if candidate.Weaknesses == nil {
		candidate.Weaknesses = []string{}
	}

	err := s.repos.Recruiting.CreateCandidate(ctx, candidate)
	if err != nil {
		return nil, fmt.Errorf("failed to create candidate: %w", err)
	}

	// Increment application count for the job
	err = s.repos.Recruiting.IncrementApplicationCount(ctx, req.JobPostingID)
	if err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to increment application count: %v\n", err)
	}

	return candidate, nil
}

func (s *recruitingService) GetCandidate(ctx context.Context, id uuid.UUID) (*models.Candidate, error) {
	return s.repos.Recruiting.GetCandidateByID(ctx, id)
}

func (s *recruitingService) GetCandidatesByJob(ctx context.Context, jobID uuid.UUID, status string) ([]*models.Candidate, error) {
	return s.repos.Recruiting.GetCandidatesByJob(ctx, jobID, status)
}

func (s *recruitingService) UpdateCandidate(ctx context.Context, id uuid.UUID, req *models.UpdateCandidateRequest) (*models.Candidate, error) {
	updates := make(map[string]interface{})

	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Score != nil {
		updates["score"] = *req.Score
	}
	if req.AISummary != nil {
		updates["ai_summary"] = *req.AISummary
	}
	if req.Strengths != nil {
		updates["strengths"] = req.Strengths
	}
	if req.Weaknesses != nil {
		updates["weaknesses"] = req.Weaknesses
	}
	if req.ExperienceYears != nil {
		updates["experience_years"] = *req.ExperienceYears
	}
	if req.Skills != nil {
		updates["skills"] = req.Skills
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	err := s.repos.Recruiting.UpdateCandidate(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update candidate: %w", err)
	}

	return s.repos.Recruiting.GetCandidateByID(ctx, id)
}

func (s *recruitingService) DeleteCandidate(ctx context.Context, id uuid.UUID) error {
	return s.repos.Recruiting.DeleteCandidate(ctx, id)
}

// AI Services
func (s *recruitingService) AnalyzeResume(ctx context.Context, candidateID uuid.UUID) (*models.ResumeAnalysisResponse, error) {
	// Get candidate
	candidate, err := s.repos.Recruiting.GetCandidateByID(ctx, candidateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate: %w", err)
	}

	// Get job posting
	job, err := s.repos.Recruiting.GetJobPosting(ctx, candidate.JobPostingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job posting: %w", err)
	}

	// In production, this would call an AI service (OpenAI, Anthropic, etc.)
	// For now, we'll provide a mock analysis
	analysis := s.mockResumeAnalysis(candidate, job)

	// Update candidate with AI analysis
	updates := map[string]interface{}{
		"score":            analysis.Score,
		"ai_summary":       analysis.Summary,
		"strengths":        analysis.Strengths,
		"weaknesses":       analysis.Weaknesses,
		"experience_years": analysis.ExperienceYears,
		"skills":           analysis.Skills,
	}

	err = s.repos.Recruiting.UpdateCandidate(ctx, candidateID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update candidate with analysis: %w", err)
	}

	return analysis, nil
}

func (s *recruitingService) mockResumeAnalysis(candidate *models.Candidate, job *models.JobPosting) *models.ResumeAnalysisResponse {
	// Simple scoring based on skills match
	score := 65
	matchedSkills := 0
	for _, candidateSkill := range candidate.Skills {
		for _, req := range job.Requirements {
			if strings.Contains(strings.ToLower(req), strings.ToLower(candidateSkill)) {
				matchedSkills++
				break
			}
		}
	}

	if matchedSkills > 0 {
		score += matchedSkills * 5
		if score > 95 {
			score = 95
		}
	}

	summary := fmt.Sprintf(
		"%s %s is a %s candidate for the %s position. "+
			"They have demonstrated experience in %d relevant areas and bring %d years of professional experience. "+
			"Their background aligns well with the requirements, particularly in %s.",
		candidate.FirstName,
		candidate.LastName,
		getScoreDescription(score),
		job.Title,
		matchedSkills,
		5, // Mock experience years
		strings.Join(candidate.Skills[:min(2, len(candidate.Skills))], " and "),
	)

	strengths := []string{
		fmt.Sprintf("Strong background in %s", strings.Join(candidate.Skills[:min(3, len(candidate.Skills))], ", ")),
		"Relevant industry experience",
		"Clear communication skills demonstrated in cover letter",
	}

	weaknesses := []string{
		"Could benefit from more specific examples of past achievements",
		"Some required skills not explicitly mentioned in resume",
	}

	return &models.ResumeAnalysisResponse{
		Score:             score,
		Summary:           summary,
		Strengths:         strengths,
		Weaknesses:        weaknesses,
		ExperienceYears:   5,
		Skills:            candidate.Skills,
		KeyQualifications: candidate.Skills[:min(5, len(candidate.Skills))],
		RedFlags:          []string{},
	}
}

func (s *recruitingService) GenerateEmail(ctx context.Context, req *models.EmailGenerationRequest) (*models.EmailGenerationResponse, error) {
	// Get candidate
	candidate, err := s.repos.Recruiting.GetCandidateByID(ctx, req.CandidateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate: %w", err)
	}

	// Get job posting
	job, err := s.repos.Recruiting.GetJobPosting(ctx, req.JobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job posting: %w", err)
	}

	// In production, this would call an AI service to generate contextual emails
	// For now, we'll use templates
	response := s.generateEmailFromTemplate(candidate, job, req.Context, req.Tone)

	return response, nil
}

func (s *recruitingService) generateEmailFromTemplate(candidate *models.Candidate, job *models.JobPosting, context, tone string) *models.EmailGenerationResponse {
	candidateName := fmt.Sprintf("%s %s", candidate.FirstName, candidate.LastName)

	var subject, body string

	switch context {
	case "screening":
		subject = fmt.Sprintf("Next Steps - %s Position at Your Company", job.Title)
		body = fmt.Sprintf(`Dear %s,

Thank you for your application for the %s position. We've reviewed your resume and are impressed with your background in %s.

We'd like to schedule a brief phone screening to discuss your qualifications and learn more about your experience.

Please let us know your availability for a 30-minute conversation in the next week.

Best regards,
Hiring Team
Your Company`, candidateName, job.Title, strings.Join(candidate.Skills[:min(2, len(candidate.Skills))], " and "))

	case "interview":
		subject = fmt.Sprintf("Interview Invitation - %s Position", job.Title)
		body = fmt.Sprintf(`Dear %s,

We're pleased to invite you to interview for the %s position at Your Company.

Interview Details:
- Date & Time: [Please select from available times]
- Duration: 60 minutes
- Format: Video Conference
- Interviewers: Hiring Manager and Team Lead

Please confirm your attendance by responding to this email.

We look forward to speaking with you!

Best regards,
Hiring Team`, candidateName, job.Title)

	case "offer":
		subject = fmt.Sprintf("🎉 Job Offer - %s at Your Company", job.Title)
		body = fmt.Sprintf(`Dear %s,

Congratulations! We're thrilled to extend an offer for the %s position at Your Company.

We were impressed by your experience and believe you'll be a great addition to our team.

Please review the attached formal offer letter for complete details including compensation, benefits, and start date.

We'd appreciate your response by [date]. Please don't hesitate to reach out with any questions.

Welcome to the team!

Best regards,
Hiring Team`, candidateName, job.Title)

	case "rejection":
		subject = fmt.Sprintf("Update on Your Application - %s", job.Title)
		body = fmt.Sprintf(`Dear %s,

Thank you for taking the time to apply for the %s position at Your Company and for your interest in joining our team.

After careful consideration, we've decided to move forward with other candidates whose qualifications more closely match our current needs.

We were impressed by your background in %s, and we encourage you to apply for future positions that align with your skills.

We wish you the best in your job search.

Best regards,
Hiring Team`, candidateName, job.Title, strings.Join(candidate.Skills[:min(2, len(candidate.Skills))], " and "))

	default:
		subject = fmt.Sprintf("Regarding Your Application - %s", job.Title)
		body = fmt.Sprintf(`Dear %s,

Thank you for your interest in the %s position at Your Company.

We wanted to reach out regarding your application.

Best regards,
Hiring Team`, candidateName, job.Title)
	}

	return &models.EmailGenerationResponse{
		Subject: subject,
		Body:    body,
	}
}

func (s *recruitingService) SendEmail(ctx context.Context, req *models.SendEmailRequest, sentBy uuid.UUID) error {
	// Get candidate to validate
	_, err := s.repos.Recruiting.GetCandidateByID(ctx, req.CandidateID)
	if err != nil {
		return fmt.Errorf("failed to get candidate: %w", err)
	}

	// In production, this would integrate with an email service (SendGrid, SES, etc.)
	// For now, we'll just log the email to the database
	
	emailType := "custom"
	if req.EmailType != nil {
		emailType = *req.EmailType
	}

	email := &models.CandidateEmail{
		ID:          uuid.New(),
		CandidateID: req.CandidateID,
		SentBy:      sentBy,
		Subject:     req.Subject,
		Body:        req.Body,
		EmailType:   emailType,
		SentAt:      time.Now(),
		CreatedAt:   time.Now(),
	}

	err = s.repos.Recruiting.CreateCandidateEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to save email record: %w", err)
	}

	// TODO: Actually send the email via email service
	fmt.Printf("Email sent to candidate %s: %s\n", req.CandidateID, req.Subject)

	return nil
}

// Helper functions
func getScoreDescription(score int) string {
	if score >= 85 {
		return "excellent"
	} else if score >= 70 {
		return "strong"
	} else if score >= 60 {
		return "good"
	} else {
		return "potential"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Provider default configurations
var providerDefaults = map[string]struct {
	Icon  string
	Color string
}{
	"linkedin":      {Icon: "💼", Color: "#0077b5"},
	"indeed":        {Icon: "🔍", Color: "#2164f3"},
	"ziprecruiter":  {Icon: "⚡", Color: "#1ca774"},
	"glassdoor":     {Icon: "🏢", Color: "#0caa41"},
	"monster":       {Icon: "👹", Color: "#6f42c1"},
	"custom":        {Icon: "🔧", Color: "#6b7280"},
}

// Provider methods

func (s *recruitingService) CreateProvider(ctx context.Context, req *models.CreateProviderRequest) (*models.RecruitingProvider, error) {
	// Validate provider type
	defaults, ok := providerDefaults[req.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported provider type: %s", req.Type)
	}

	now := time.Now()
	provider := &models.RecruitingProvider{
		ID:              uuid.New(),
		Type:            req.Type,
		Name:            req.Name,
		Icon:            defaults.Icon,
		Color:           defaults.Color,
		IsConnected:     false, // Will be set to true after successful connection test
		Config:          req.Config,
		JobsPosted:      0,
		ApplicantsTotal: 0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Test connection before creating
	testReq := &models.TestProviderConnectionRequest{
		Type:   req.Type,
		Config: req.Config,
	}
	testResult, err := s.TestProviderConnection(ctx, testReq)
	if err == nil && testResult.Success {
		provider.IsConnected = true
	}

	err = s.repos.Recruiting.CreateProvider(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, nil
}

func (s *recruitingService) GetProvider(ctx context.Context, id uuid.UUID) (*models.RecruitingProvider, error) {
	return s.repos.Recruiting.GetProvider(ctx, id)
}

func (s *recruitingService) GetAllProviders(ctx context.Context) ([]*models.RecruitingProvider, error) {
	return s.repos.Recruiting.GetAllProviders(ctx)
}

func (s *recruitingService) UpdateProvider(ctx context.Context, id uuid.UUID, req *models.UpdateProviderRequest) (*models.RecruitingProvider, error) {
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Config != nil {
		updates["config"] = *req.Config
		
		// Test new connection if config changed
		provider, err := s.GetProvider(ctx, id)
		if err != nil {
			return nil, err
		}

		testReq := &models.TestProviderConnectionRequest{
			Type:   provider.Type,
			Config: *req.Config,
		}
		testResult, _ := s.TestProviderConnection(ctx, testReq)
		updates["is_connected"] = testResult != nil && testResult.Success
	}

	if len(updates) == 0 {
		return s.GetProvider(ctx, id)
	}

	err := s.repos.Recruiting.UpdateProvider(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}

	return s.GetProvider(ctx, id)
}

func (s *recruitingService) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	return s.repos.Recruiting.DeleteProvider(ctx, id)
}

func (s *recruitingService) TestProviderConnection(ctx context.Context, req *models.TestProviderConnectionRequest) (*models.TestProviderConnectionResponse, error) {
	// This is a simplified version. In production, you would:
	// 1. Make actual API calls to each provider
	// 2. Verify credentials
	// 3. Check API quotas/limits
	// 4. Return specific error messages

	response := &models.TestProviderConnectionResponse{
		Success: false,
		Message: "",
	}

	switch req.Type {
	case "linkedin":
		// Validate required fields
		if req.Config["client_id"] == nil || req.Config["client_secret"] == nil {
			response.Message = "Missing required fields: client_id, client_secret"
			return response, nil
		}
		// In production: Make OAuth validation call to LinkedIn API
		response.Success = true
		response.Message = "Successfully connected to LinkedIn"

	case "indeed":
		if req.Config["publisher_id"] == nil || req.Config["api_token"] == nil {
			response.Message = "Missing required fields: publisher_id, api_token"
			return response, nil
		}
		// In production: Validate Indeed API credentials
		response.Success = true
		response.Message = "Successfully connected to Indeed"

	case "ziprecruiter":
		if req.Config["api_key"] == nil {
			response.Message = "Missing required field: api_key"
			return response, nil
		}
		// In production: Test ZipRecruiter API
		response.Success = true
		response.Message = "Successfully connected to ZipRecruiter"

	case "glassdoor":
		if req.Config["partner_id"] == nil || req.Config["partner_key"] == nil {
			response.Message = "Missing required fields: partner_id, partner_key"
			return response, nil
		}
		// In production: Validate Glassdoor credentials
		response.Success = true
		response.Message = "Successfully connected to Glassdoor"

	case "monster":
		if req.Config["api_key"] == nil {
			response.Message = "Missing required field: api_key"
			return response, nil
		}
		// In production: Test Monster API
		response.Success = true
		response.Message = "Successfully connected to Monster"

	case "custom":
		if req.Config["api_url"] == nil || req.Config["api_key"] == nil {
			response.Message = "Missing required fields: api_url, api_key"
			return response, nil
		}
		// In production: Make test call to custom API
		response.Success = true
		response.Message = "Successfully connected to custom provider"

	default:
		response.Message = fmt.Sprintf("Unsupported provider type: %s", req.Type)
		return response, nil
	}

	return response, nil
}

// Dashboard methods

func (s *recruitingService) GetDashboardStats(ctx context.Context) (*models.RecruitingDashboardStats, error) {
	return s.repos.Recruiting.GetDashboardStats(ctx)
}

func (s *recruitingService) GetDashboard(ctx context.Context) (*models.RecruitingDashboard, error) {
	return s.repos.Recruiting.GetDashboard(ctx)
}

func (s *recruitingService) GetApplicantLeaderboard(ctx context.Context) ([]*models.ApplicantLeaderboard, error) {
	return s.repos.Recruiting.GetApplicantLeaderboard(ctx, 50) // Top 50
}

// Interview methods

func (s *recruitingService) ScheduleInterview(ctx context.Context, req *models.ScheduleInterviewRequest) (*models.Interview, error) {
	now := time.Now()
	interview := &models.Interview{
		ID:            uuid.New(),
		CandidateID:   req.CandidateID,
		InterviewerID: req.InterviewerID,
		ScheduledAt:   req.ScheduledAt,
		Duration:      req.Duration,
		InterviewType: req.InterviewType,
		Location:      req.Location,
		MeetingURL:    req.MeetingURL,
		Status:        "scheduled",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err := s.repos.Recruiting.CreateInterview(ctx, interview)
	if err != nil {
		return nil, fmt.Errorf("failed to schedule interview: %w", err)
	}

	return interview, nil
}

func (s *recruitingService) UpdateInterview(ctx context.Context, id uuid.UUID, req *models.UpdateInterviewRequest) (*models.Interview, error) {
	updates := make(map[string]interface{})

	if req.ScheduledAt != nil {
		updates["scheduled_at"] = *req.ScheduledAt
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.InterviewType != nil {
		updates["interview_type"] = *req.InterviewType
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.MeetingURL != nil {
		updates["meeting_url"] = *req.MeetingURL
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Feedback != nil {
		updates["feedback"] = *req.Feedback
	}
	if req.Rating != nil {
		updates["rating"] = *req.Rating
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}

	err := s.repos.Recruiting.UpdateInterview(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update interview: %w", err)
	}

	// Return updated interview (you'd need to add GetInterview method to repo)
	// For now, return nil
	return nil, nil
}

func (s *recruitingService) GetInterviewsByCandidate(ctx context.Context, candidateID uuid.UUID) ([]*models.Interview, error) {
	return s.repos.Recruiting.GetInterviewsByCandidate(ctx, candidateID)
}

// Enhanced PostToJobBoards with provider integration
func (s *recruitingService) PostToJobBoardsEnhanced(ctx context.Context, req *models.PostToJobBoardsRequest, userID uuid.UUID) error {
	// Get job posting
	job, err := s.repos.Recruiting.GetJobPosting(ctx, req.JobPostingID)
	if err != nil {
		return fmt.Errorf("failed to get job posting: %w", err)
	}

	// Update job status to active
	if job.Status == "draft" {
		now := time.Now()
		updates := map[string]interface{}{
			"status":      "active",
			"posted_date": now,
		}
		err = s.repos.Recruiting.UpdateJobPosting(ctx, req.JobPostingID, updates)
		if err != nil {
			return fmt.Errorf("failed to update job status: %w", err)
		}
	}

	// Post to each selected board
	now := time.Now()
	for _, boardName := range req.Boards {  // ✓ FIXED: Use req.Boards instead of req.Providers
		// Get all providers and find matching one by type
		providers, err := s.repos.Recruiting.GetAllProviders(ctx)
		if err != nil {
			continue // Skip if can't get providers
		}

		// Find provider matching this board name
		var matchedProvider *models.RecruitingProvider
		for _, provider := range providers {
			if provider.Type == boardName {
				matchedProvider = provider
				break
			}
		}

		// Skip if provider not found or not connected
		if matchedProvider == nil || !matchedProvider.IsConnected {
			continue
		}

		// In production, make actual API calls to each provider
		// For now, just create a job board posting record
		posting := &models.JobBoardPosting{
			ID:           uuid.New(),
			JobPostingID: req.JobPostingID,
			BoardName:    boardName,
			PostedAt:     now,
			Status:       "active",
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		err = s.repos.Recruiting.CreateJobBoardPosting(ctx, posting)
		if err != nil {
			continue // Skip failed postings
		}

		// Update provider stats if we found a matching provider
		if matchedProvider != nil {
			s.repos.Recruiting.UpdateProviderStats(ctx, matchedProvider.ID, 1, 0)
		}
	}

	return nil
}

func (s *recruitingService) CreateJobFromUpload(ctx context.Context, req *models.JobUploadRequest, userID uuid.UUID) (*models.JobPosting, error) {
	// Convert JobUploadRequest to CreateJobPostingRequest
	var salaryMin, salaryMax *float64
	
	if req.SalaryMin > 0 {
		f := float64(req.SalaryMin)
		salaryMin = &f
	}
	
	if req.SalaryMax > 0 {
		f := float64(req.SalaryMax)
		salaryMax = &f
	}

	createReq := &models.CreateJobPostingRequest{
		Title:            req.Title,
		Department:       req.Department,
		Location:         req.Location,
		EmploymentType:   req.EmploymentType,
		SalaryMin:        salaryMin,
		SalaryMax:        salaryMax,
		Description:      req.Description,
		Requirements:     req.Requirements,
		Responsibilities: req.Responsibilities,
		Benefits:         req.Benefits,
	}

	// Use existing CreateJobPosting method
	return s.CreateJobPosting(ctx, createReq, userID)
}

// CreateApplicant creates a new applicant record
func (s *recruitingService) CreateApplicant(ctx context.Context, applicant *models.Applicant) error {
	return s.repos.Applicant.Create(ctx, applicant)
}
