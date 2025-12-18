package models

import (
	"time"

	"github.com/google/uuid"
)

// JobPosting represents a job opening
type JobPosting struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Title             string    `json:"title" db:"title"`
	Department        string    `json:"department" db:"department"`
	Location          string    `json:"location" db:"location"`
	EmploymentType    string    `json:"employment_type" db:"employment_type"` // full-time, part-time, contract, internship
	SalaryMin         *float64  `json:"salary_min,omitempty" db:"salary_min"`
	SalaryMax         *float64  `json:"salary_max,omitempty" db:"salary_max"`
	Description       string    `json:"description" db:"description"`
	Requirements      []string  `json:"requirements" db:"requirements"`
	Responsibilities  []string  `json:"responsibilities" db:"responsibilities"`
	Benefits          []string  `json:"benefits" db:"benefits"`
	Status            string    `json:"status" db:"status"` // draft, active, closed, filled
	PostedDate        *time.Time `json:"posted_date,omitempty" db:"posted_date"`
	ClosedDate        *time.Time `json:"closed_date,omitempty" db:"closed_date"`
	ApplicationsCount int       `json:"applications_count" db:"applications_count"`
	CreatedBy         uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Candidate represents a job applicant
type Candidate struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	JobPostingID   uuid.UUID  `json:"job_posting_id" db:"job_posting_id"`
	FirstName      string     `json:"first_name" db:"first_name"`
	LastName       string     `json:"last_name" db:"last_name"`
	Email          string     `json:"email" db:"email"`
	Phone          string     `json:"phone" db:"phone"`
	ResumeURL      *string    `json:"resume_url,omitempty" db:"resume_url"`
	CoverLetter    *string    `json:"cover_letter,omitempty" db:"cover_letter"`
	LinkedInURL    *string    `json:"linkedin_url,omitempty" db:"linkedin_url"`
	PortfolioURL   *string    `json:"portfolio_url,omitempty" db:"portfolio_url"`
	Status         string     `json:"status" db:"status"` // new, screening, interview, offered, rejected, hired
	Score          *int       `json:"score,omitempty" db:"score"` // AI match score 0-100
	AISummary      *string    `json:"ai_summary,omitempty" db:"ai_summary"`
	Strengths      []string   `json:"strengths,omitempty" db:"strengths"`
	Weaknesses     []string   `json:"weaknesses,omitempty" db:"weaknesses"`
	ExperienceYears *int      `json:"experience_years,omitempty" db:"experience_years"`
	Skills         []string   `json:"skills" db:"skills"`
	AppliedDate    time.Time  `json:"applied_date" db:"applied_date"`
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// CandidateWithJob includes job details with candidate
type CandidateWithJob struct {
	Candidate
	JobTitle      string `json:"job_title" db:"job_title"`
	JobDepartment string `json:"job_department" db:"job_department"`
}

// Interview represents an interview scheduled for a candidate
type Interview struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	CandidateID  uuid.UUID  `json:"candidate_id" db:"candidate_id"`
	InterviewerID uuid.UUID `json:"interviewer_id" db:"interviewer_id"`
	ScheduledAt  time.Time  `json:"scheduled_at" db:"scheduled_at"`
	Duration     int        `json:"duration" db:"duration"` // minutes
	InterviewType string    `json:"interview_type" db:"interview_type"` // phone, video, onsite
	Location     *string    `json:"location,omitempty" db:"location"`
	MeetingURL   *string    `json:"meeting_url,omitempty" db:"meeting_url"`
	Status       string     `json:"status" db:"status"` // scheduled, completed, cancelled, no_show
	Feedback     *string    `json:"feedback,omitempty" db:"feedback"`
	Rating       *int       `json:"rating,omitempty" db:"rating"` // 1-5
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// JobBoardPosting tracks which job boards a job is posted to
type JobBoardPosting struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	JobPostingID uuid.UUID  `json:"job_posting_id" db:"job_posting_id"`
	BoardName    string     `json:"board_name" db:"board_name"` // linkedin, indeed, glassdoor, etc.
	ExternalID   *string    `json:"external_id,omitempty" db:"external_id"` // ID on the job board platform
	PostedAt     time.Time  `json:"posted_at" db:"posted_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Status       string     `json:"status" db:"status"` // active, expired, removed
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// CandidateEmail represents an email sent to a candidate
type CandidateEmail struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CandidateID uuid.UUID `json:"candidate_id" db:"candidate_id"`
	SentBy      uuid.UUID `json:"sent_by" db:"sent_by"`
	Subject     string    `json:"subject" db:"subject"`
	Body        string    `json:"body" db:"body"`
	EmailType   string    `json:"email_type" db:"email_type"` // screening, interview, offer, rejection
	SentAt      time.Time `json:"sent_at" db:"sent_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ResumeAnalysisRequest for AI analysis
type ResumeAnalysisRequest struct {
	CandidateID uuid.UUID `json:"candidate_id"`
	ResumeText  string    `json:"resume_text"`
	JobDescription string `json:"job_description"`
}

// ResumeAnalysisResponse from AI analysis
type ResumeAnalysisResponse struct {
	Score          int      `json:"score"`
	Summary        string   `json:"summary"`
	Strengths      []string `json:"strengths"`
	Weaknesses     []string `json:"weaknesses"`
	ExperienceYears int     `json:"experience_years"`
	Skills         []string `json:"skills"`
	KeyQualifications []string `json:"key_qualifications"`
	RedFlags       []string `json:"red_flags"`
}

// EmailGenerationRequest for AI email generation
type EmailGenerationRequest struct {
	CandidateID uuid.UUID `json:"candidate_id"`
	JobID       uuid.UUID `json:"job_id"`
	Context     string    `json:"context"` // screening, interview, offer, rejection, custom
	Tone        string    `json:"tone"`    // professional, friendly, formal
	CustomPrompt *string  `json:"custom_prompt,omitempty"`
}

// EmailGenerationResponse from AI
type EmailGenerationResponse struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// SendEmailRequest to send email to candidate
type SendEmailRequest struct {
	CandidateID uuid.UUID `json:"candidate_id"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	EmailType   *string   `json:"email_type,omitempty"`
}

// CreateJobPostingRequest for creating a new job
type CreateJobPostingRequest struct {
	Title            string    `json:"title" validate:"required"`
	Department       string    `json:"department" validate:"required"`
	Location         string    `json:"location" validate:"required"`
	EmploymentType   string    `json:"employment_type" validate:"required,oneof=full-time part-time contract internship"`
	SalaryMin        *float64  `json:"salary_min,omitempty"`
	SalaryMax        *float64  `json:"salary_max,omitempty"`
	Description      string    `json:"description" validate:"required"`
	Requirements     []string  `json:"requirements"`
	Responsibilities []string  `json:"responsibilities"`
	Benefits         []string  `json:"benefits"`
}

// UpdateJobPostingRequest for updating a job
type UpdateJobPostingRequest struct {
	Title            *string   `json:"title,omitempty"`
	Department       *string   `json:"department,omitempty"`
	Location         *string   `json:"location,omitempty"`
	EmploymentType   *string   `json:"employment_type,omitempty"`
	SalaryMin        *float64  `json:"salary_min,omitempty"`
	SalaryMax        *float64  `json:"salary_max,omitempty"`
	Description      *string   `json:"description,omitempty"`
	Requirements     []string  `json:"requirements,omitempty"`
	Responsibilities []string  `json:"responsibilities,omitempty"`
	Benefits         []string  `json:"benefits,omitempty"`
	Status           *string   `json:"status,omitempty"`
}

// CreateCandidateRequest for adding a candidate
type CreateCandidateRequest struct {
	JobPostingID uuid.UUID `json:"job_posting_id" validate:"required"`
	FirstName    string    `json:"first_name" validate:"required"`
	LastName     string    `json:"last_name" validate:"required"`
	Email        string    `json:"email" validate:"required,email"`
	Phone        string    `json:"phone"`
	ResumeURL    *string   `json:"resume_url,omitempty"`
	CoverLetter  *string   `json:"cover_letter,omitempty"`
	LinkedInURL  *string   `json:"linkedin_url,omitempty"`
	PortfolioURL *string   `json:"portfolio_url,omitempty"`
	Skills       []string  `json:"skills"`
}

// UpdateCandidateRequest for updating candidate info
type UpdateCandidateRequest struct {
	Status    *string  `json:"status,omitempty"`
	Score     *int     `json:"score,omitempty"`
	AISummary *string  `json:"ai_summary,omitempty"`
	Strengths []string `json:"strengths,omitempty"`
	Weaknesses []string `json:"weaknesses,omitempty"`
	ExperienceYears *int `json:"experience_years,omitempty"`
	Skills    []string `json:"skills,omitempty"`
	Notes     *string  `json:"notes,omitempty"`
}

// PostToJobBoardsRequest for posting to multiple boards
type PostToJobBoardsRequest struct {
	JobPostingID uuid.UUID `json:"job_posting_id" validate:"required"`
	Boards       []string  `json:"boards" validate:"required"` // linkedin, indeed, glassdoor, etc.
}