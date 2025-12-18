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
		Description:       req.Description,
		Requirements:      req.Requirements,
		Responsibilities:  req.Responsibilities,
		Benefits:          req.Benefits,
		Status:            "draft",
		PostedDate:        nil,
		ApplicationsCount: 0,
		CreatedBy:         userID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	// Initialize empty arrays if nil
	if job.Requirements == nil {
		job.Requirements = []string{}
	}
	if job.Responsibilities == nil {
		job.Responsibilities = []string{}
	}
	if job.Benefits == nil {
		job.Benefits = []string{}
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
	return s.repos.Recruiting.GetCandidate(ctx, id)
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

	return s.repos.Recruiting.GetCandidate(ctx, id)
}

func (s *recruitingService) DeleteCandidate(ctx context.Context, id uuid.UUID) error {
	return s.repos.Recruiting.DeleteCandidate(ctx, id)
}

// AI Services
func (s *recruitingService) AnalyzeResume(ctx context.Context, candidateID uuid.UUID) (*models.ResumeAnalysisResponse, error) {
	// Get candidate
	candidate, err := s.repos.Recruiting.GetCandidate(ctx, candidateID)
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
	candidate, err := s.repos.Recruiting.GetCandidate(ctx, req.CandidateID)
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
	_, err := s.repos.Recruiting.GetCandidate(ctx, req.CandidateID)
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