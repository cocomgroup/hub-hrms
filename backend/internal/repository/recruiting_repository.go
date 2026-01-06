package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"hub-hrms/backend/internal/models"
)

type RecruitingRepository interface {
	// Job Postings
	CreateJobPosting(ctx context.Context, job *models.JobPosting) error
	GetJobPosting(ctx context.Context, id uuid.UUID) (*models.JobPosting, error)
	ListJobPostings(ctx context.Context, status string) ([]*models.JobPosting, error)
	UpdateJobPosting(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	DeleteJobPosting(ctx context.Context, id uuid.UUID) error
	IncrementApplicationCount(ctx context.Context, jobID uuid.UUID) error

	// Candidates
	CreateCandidate(ctx context.Context, candidate *models.Candidate) error
	GetCandidate(ctx context.Context, id uuid.UUID) (*models.Candidate, error)
	GetCandidatesByJob(ctx context.Context, jobID uuid.UUID, status string) ([]*models.Candidate, error)
	UpdateCandidate(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	DeleteCandidate(ctx context.Context, id uuid.UUID) error

	// Interviews
	CreateInterview(ctx context.Context, interview *models.Interview) error
	GetInterviewsByCandidate(ctx context.Context, candidateID uuid.UUID) ([]*models.Interview, error)
	UpdateInterview(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error

	// Job Board Postings
	CreateJobBoardPosting(ctx context.Context, posting *models.JobBoardPosting) error
	GetJobBoardPostings(ctx context.Context, jobID uuid.UUID) ([]*models.JobBoardPosting, error)
	UpdateJobBoardPosting(ctx context.Context, id uuid.UUID, status string) error

	// Candidate Emails
	CreateCandidateEmail(ctx context.Context, email *models.CandidateEmail) error
	GetCandidateEmails(ctx context.Context, candidateID uuid.UUID) ([]*models.CandidateEmail, error)

	// Providers
	CreateProvider(ctx context.Context, provider *models.RecruitingProvider) error
	GetProvider(ctx context.Context, id uuid.UUID) (*models.RecruitingProvider, error)
	GetAllProviders(ctx context.Context) ([]*models.RecruitingProvider, error)
	UpdateProvider(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	DeleteProvider(ctx context.Context, id uuid.UUID) error
	UpdateProviderStats(ctx context.Context, id uuid.UUID, jobsPosted, applicants int) error

	// Dashboard & Stats
	GetDashboardStats(ctx context.Context) (*models.RecruitingDashboardStats, error)
	GetDashboard(ctx context.Context) (*models.RecruitingDashboard, error)
	GetApplicantLeaderboard(ctx context.Context, limit int) ([]*models.ApplicantLeaderboard, error)
	GetRecentApplications(ctx context.Context, limit int) ([]*models.CandidateWithJob, error)
	GetTopPerformingJobs(ctx context.Context, limit int) ([]*models.JobPosting, error)
	GetUpcomingInterviews(ctx context.Context, limit int) ([]*models.InterviewWithDetails, error)
}

type recruitingRepository struct {
	db *pgxpool.Pool
}

func NewRecruitingRepository(db *pgxpool.Pool) RecruitingRepository {
	return &recruitingRepository{db: db}
}

// Job Postings
func (r *recruitingRepository) CreateJobPosting(ctx context.Context, job *models.JobPosting) error {
	query := `
		INSERT INTO job_postings (
			id, title, department, location, employment_type,
			salary_min, salary_max, description, requirements,
			responsibilities, benefits, status, posted_date,
			created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.Exec(ctx, query,
		job.ID,
		job.Title,
		job.Department,
		job.Location,
		job.EmploymentType,
		job.SalaryMin,
		job.SalaryMax,
		job.Description,
		pq.Array(job.Requirements),
		pq.Array(job.Responsibilities),
		pq.Array(job.Benefits),
		job.Status,
		job.PostedDate,
		job.CreatedBy,
		job.CreatedAt,
		job.UpdatedAt,
	)

	return err
}

func (r *recruitingRepository) GetJobPosting(ctx context.Context, id uuid.UUID) (*models.JobPosting, error) {
	query := `
		SELECT 
			id, title, department, location, employment_type,
			salary_min, salary_max, description, requirements,
			responsibilities, benefits, status, posted_date,
			closed_date, applications_count, created_by,
			created_at, updated_at
		FROM job_postings
		WHERE id = $1
	`

	var job models.JobPosting
	err := r.db.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.Title,
		&job.Department,
		&job.Location,
		&job.EmploymentType,
		&job.SalaryMin,
		&job.SalaryMax,
		&job.Description,
		pq.Array(&job.Requirements),
		pq.Array(&job.Responsibilities),
		pq.Array(&job.Benefits),
		&job.Status,
		&job.PostedDate,
		&job.ClosedDate,
		&job.ApplicationsCount,
		&job.CreatedBy,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job posting not found")
	}
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *recruitingRepository) ListJobPostings(ctx context.Context, status string) ([]*models.JobPosting, error) {
	query := `
		SELECT 
			id, title, department, location, employment_type,
			salary_min, salary_max, description, requirements,
			responsibilities, benefits, status, posted_date,
			closed_date, applications_count, created_by,
			created_at, updated_at
		FROM job_postings
		WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.JobPosting
	for rows.Next() {
		var job models.JobPosting
		err := rows.Scan(
			&job.ID,
			&job.Title,
			&job.Department,
			&job.Location,
			&job.EmploymentType,
			&job.SalaryMin,
			&job.SalaryMax,
			&job.Description,
			pq.Array(&job.Requirements),
			pq.Array(&job.Responsibilities),
			pq.Array(&job.Benefits),
			&job.Status,
			&job.PostedDate,
			&job.ClosedDate,
			&job.ApplicationsCount,
			&job.CreatedBy,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (r *recruitingRepository) UpdateJobPosting(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	query := `
		UPDATE job_postings SET
			title = COALESCE($1, title),
			department = COALESCE($2, department),
			location = COALESCE($3, location),
			employment_type = COALESCE($4, employment_type),
			salary_min = COALESCE($5, salary_min),
			salary_max = COALESCE($6, salary_max),
			description = COALESCE($7, description),
			requirements = COALESCE($8, requirements),
			responsibilities = COALESCE($9, responsibilities),
			benefits = COALESCE($10, benefits),
			status = COALESCE($11, status),
			updated_at = $12
		WHERE id = $13
	`

	_, err := r.db.Exec(ctx, query,
		updates["title"],
		updates["department"],
		updates["location"],
		updates["employment_type"],
		updates["salary_min"],
		updates["salary_max"],
		updates["description"],
		updates["requirements"],
		updates["responsibilities"],
		updates["benefits"],
		updates["status"],
		time.Now(),
		id,
	)

	return err
}

func (r *recruitingRepository) DeleteJobPosting(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM job_postings WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *recruitingRepository) IncrementApplicationCount(ctx context.Context, jobID uuid.UUID) error {
	query := `
		UPDATE job_postings 
		SET applications_count = applications_count + 1,
		    updated_at = $1
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, time.Now(), jobID)
	return err
}

// Candidates
func (r *recruitingRepository) CreateCandidate(ctx context.Context, candidate *models.Candidate) error {
	query := `
		INSERT INTO candidates (
			id, job_posting_id, first_name, last_name, email, phone,
			resume_url, cover_letter, linkedin_url, portfolio_url,
			status, score, ai_summary, strengths, weaknesses,
			experience_years, skills, applied_date, notes,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`

	_, err := r.db.Exec(ctx, query,
		candidate.ID,
		candidate.JobPostingID,
		candidate.FirstName,
		candidate.LastName,
		candidate.Email,
		candidate.Phone,
		candidate.ResumeURL,
		candidate.CoverLetter,
		candidate.LinkedInURL,
		candidate.PortfolioURL,
		candidate.Status,
		candidate.Score,
		candidate.AISummary,
		pq.Array(candidate.Strengths),
		pq.Array(candidate.Weaknesses),
		candidate.ExperienceYears,
		pq.Array(candidate.Skills),
		candidate.AppliedDate,
		candidate.Notes,
		candidate.CreatedAt,
		candidate.UpdatedAt,
	)

	return err
}

func (r *recruitingRepository) GetCandidate(ctx context.Context, id uuid.UUID) (*models.Candidate, error) {
	query := `
		SELECT 
			id, job_posting_id, first_name, last_name, email, phone,
			resume_url, cover_letter, linkedin_url, portfolio_url,
			status, score, ai_summary, strengths, weaknesses,
			experience_years, skills, applied_date, notes,
			created_at, updated_at
		FROM candidates
		WHERE id = $1
	`

	var candidate models.Candidate
	err := r.db.QueryRow(ctx, query, id).Scan(
		&candidate.ID,
		&candidate.JobPostingID,
		&candidate.FirstName,
		&candidate.LastName,
		&candidate.Email,
		&candidate.Phone,
		&candidate.ResumeURL,
		&candidate.CoverLetter,
		&candidate.LinkedInURL,
		&candidate.PortfolioURL,
		&candidate.Status,
		&candidate.Score,
		&candidate.AISummary,
		pq.Array(&candidate.Strengths),
		pq.Array(&candidate.Weaknesses),
		&candidate.ExperienceYears,
		pq.Array(&candidate.Skills),
		&candidate.AppliedDate,
		&candidate.Notes,
		&candidate.CreatedAt,
		&candidate.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("candidate not found")
	}
	if err != nil {
		return nil, err
	}

	return &candidate, nil
}

func (r *recruitingRepository) GetCandidatesByJob(ctx context.Context, jobID uuid.UUID, status string) ([]*models.Candidate, error) {
	query := `
		SELECT 
			id, job_posting_id, first_name, last_name, email, phone,
			resume_url, cover_letter, linkedin_url, portfolio_url,
			status, score, ai_summary, strengths, weaknesses,
			experience_years, skills, applied_date, notes,
			created_at, updated_at
		FROM candidates
		WHERE job_posting_id = $1
		  AND ($2 = '' OR status = $2)
		ORDER BY applied_date DESC
	`

	rows, err := r.db.Query(ctx, query, jobID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []*models.Candidate
	for rows.Next() {
		var candidate models.Candidate
		err := rows.Scan(
			&candidate.ID,
			&candidate.JobPostingID,
			&candidate.FirstName,
			&candidate.LastName,
			&candidate.Email,
			&candidate.Phone,
			&candidate.ResumeURL,
			&candidate.CoverLetter,
			&candidate.LinkedInURL,
			&candidate.PortfolioURL,
			&candidate.Status,
			&candidate.Score,
			&candidate.AISummary,
			pq.Array(&candidate.Strengths),
			pq.Array(&candidate.Weaknesses),
			&candidate.ExperienceYears,
			pq.Array(&candidate.Skills),
			&candidate.AppliedDate,
			&candidate.Notes,
			&candidate.CreatedAt,
			&candidate.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, &candidate)
	}

	return candidates, nil
}

func (r *recruitingRepository) UpdateCandidate(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	query := `
		UPDATE candidates SET
			status = COALESCE($1, status),
			score = COALESCE($2, score),
			ai_summary = COALESCE($3, ai_summary),
			strengths = COALESCE($4, strengths),
			weaknesses = COALESCE($5, weaknesses),
			experience_years = COALESCE($6, experience_years),
			skills = COALESCE($7, skills),
			notes = COALESCE($8, notes),
			updated_at = $9
		WHERE id = $10
	`

	_, err := r.db.Exec(ctx, query,
		updates["status"],
		updates["score"],
		updates["ai_summary"],
		updates["strengths"],
		updates["weaknesses"],
		updates["experience_years"],
		updates["skills"],
		updates["notes"],
		time.Now(),
		id,
	)

	return err
}

func (r *recruitingRepository) DeleteCandidate(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM candidates WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Interviews
func (r *recruitingRepository) CreateInterview(ctx context.Context, interview *models.Interview) error {
	query := `
		INSERT INTO interviews (
			id, candidate_id, interviewer_id, scheduled_at,
			duration, interview_type, location, meeting_url,
			status, feedback, rating, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.Exec(ctx, query,
		interview.ID,
		interview.CandidateID,
		interview.InterviewerID,
		interview.ScheduledAt,
		interview.Duration,
		interview.InterviewType,
		interview.Location,
		interview.MeetingURL,
		interview.Status,
		interview.Feedback,
		interview.Rating,
		interview.CreatedAt,
		interview.UpdatedAt,
	)

	return err
}

func (r *recruitingRepository) GetInterviewsByCandidate(ctx context.Context, candidateID uuid.UUID) ([]*models.Interview, error) {
	query := `
		SELECT 
			id, candidate_id, interviewer_id, scheduled_at,
			duration, interview_type, location, meeting_url,
			status, feedback, rating, created_at, updated_at
		FROM interviews
		WHERE candidate_id = $1
		ORDER BY scheduled_at DESC
	`

	rows, err := r.db.Query(ctx, query, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interviews []*models.Interview
	for rows.Next() {
		var interview models.Interview
		err := rows.Scan(
			&interview.ID,
			&interview.CandidateID,
			&interview.InterviewerID,
			&interview.ScheduledAt,
			&interview.Duration,
			&interview.InterviewType,
			&interview.Location,
			&interview.MeetingURL,
			&interview.Status,
			&interview.Feedback,
			&interview.Rating,
			&interview.CreatedAt,
			&interview.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		interviews = append(interviews, &interview)
	}

	return interviews, nil
}

func (r *recruitingRepository) UpdateInterview(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	query := `
		UPDATE interviews SET
			status = COALESCE($1, status),
			feedback = COALESCE($2, feedback),
			rating = COALESCE($3, rating),
			updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.Exec(ctx, query,
		updates["status"],
		updates["feedback"],
		updates["rating"],
		time.Now(),
		id,
	)

	return err
}

// Job Board Postings
func (r *recruitingRepository) CreateJobBoardPosting(ctx context.Context, posting *models.JobBoardPosting) error {
	query := `
		INSERT INTO job_board_postings (
			id, job_posting_id, board_name, external_id,
			posted_at, expires_at, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		posting.ID,
		posting.JobPostingID,
		posting.BoardName,
		posting.ExternalID,
		posting.PostedAt,
		posting.ExpiresAt,
		posting.Status,
		posting.CreatedAt,
		posting.UpdatedAt,
	)

	return err
}

func (r *recruitingRepository) GetJobBoardPostings(ctx context.Context, jobID uuid.UUID) ([]*models.JobBoardPosting, error) {
	query := `
		SELECT 
			id, job_posting_id, board_name, external_id,
			posted_at, expires_at, status, created_at, updated_at
		FROM job_board_postings
		WHERE job_posting_id = $1
		ORDER BY posted_at DESC
	`

	rows, err := r.db.Query(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var postings []*models.JobBoardPosting
	for rows.Next() {
		var posting models.JobBoardPosting
		err := rows.Scan(
			&posting.ID,
			&posting.JobPostingID,
			&posting.BoardName,
			&posting.ExternalID,
			&posting.PostedAt,
			&posting.ExpiresAt,
			&posting.Status,
			&posting.CreatedAt,
			&posting.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		postings = append(postings, &posting)
	}

	return postings, nil
}

func (r *recruitingRepository) UpdateJobBoardPosting(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE job_board_postings 
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, status, time.Now(), id)
	return err
}

// Candidate Emails
func (r *recruitingRepository) CreateCandidateEmail(ctx context.Context, email *models.CandidateEmail) error {
	query := `
		INSERT INTO candidate_emails (
			id, candidate_id, sent_by, subject, body,
			email_type, sent_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		email.ID,
		email.CandidateID,
		email.SentBy,
		email.Subject,
		email.Body,
		email.EmailType,
		email.SentAt,
		email.CreatedAt,
	)

	return err
}

func (r *recruitingRepository) GetCandidateEmails(ctx context.Context, candidateID uuid.UUID) ([]*models.CandidateEmail, error) {
	query := `
		SELECT 
			id, candidate_id, sent_by, subject, body,
			email_type, sent_at, created_at
		FROM candidate_emails
		WHERE candidate_id = $1
		ORDER BY sent_at DESC
	`

	rows, err := r.db.Query(ctx, query, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []*models.CandidateEmail
	for rows.Next() {
		var email models.CandidateEmail
		err := rows.Scan(
			&email.ID,
			&email.CandidateID,
			&email.SentBy,
			&email.Subject,
			&email.Body,
			&email.EmailType,
			&email.SentAt,
			&email.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		emails = append(emails, &email)
	}

	return emails, nil
}

// Implement Provider methods

func (r *recruitingRepository) CreateProvider(ctx context.Context, provider *models.RecruitingProvider) error {
	configJSON, err := json.Marshal(provider.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO recruiting_providers (
			id, type, name, icon, color, is_connected, config,
			jobs_posted, applicants_total, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.Exec(ctx, query,
		provider.ID,
		provider.Type,
		provider.Name,
		provider.Icon,
		provider.Color,
		provider.IsConnected,
		configJSON,
		provider.JobsPosted,
		provider.ApplicantsTotal,
		provider.CreatedAt,
		provider.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	return nil
}

func (r *recruitingRepository) GetProvider(ctx context.Context, id uuid.UUID) (*models.RecruitingProvider, error) {
	query := `
		SELECT id, type, name, icon, color, is_connected, config,
		       jobs_posted, applicants_total, last_synced_at,
		       created_at, updated_at
		FROM recruiting_providers
		WHERE id = $1
	`

	var provider models.RecruitingProvider
	var configJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&provider.ID,
		&provider.Type,
		&provider.Name,
		&provider.Icon,
		&provider.Color,
		&provider.IsConnected,
		&configJSON,
		&provider.JobsPosted,
		&provider.ApplicantsTotal,
		&provider.LastSyncedAt,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("provider not found")
		}
		return nil, err
	}

	if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &provider, nil
}

func (r *recruitingRepository) GetAllProviders(ctx context.Context) ([]*models.RecruitingProvider, error) {
	query := `
		SELECT id, type, name, icon, color, is_connected, config,
		       jobs_posted, applicants_total, last_synced_at,
		       created_at, updated_at
		FROM recruiting_providers
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*models.RecruitingProvider

	for rows.Next() {
		var provider models.RecruitingProvider
		var configJSON []byte

		err := rows.Scan(
			&provider.ID,
			&provider.Type,
			&provider.Name,
			&provider.Icon,
			&provider.Color,
			&provider.IsConnected,
			&configJSON,
			&provider.JobsPosted,
			&provider.ApplicantsTotal,
			&provider.LastSyncedAt,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		providers = append(providers, &provider)
	}

	return providers, nil
}

func (r *recruitingRepository) UpdateProvider(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	query := `UPDATE recruiting_providers SET `
	params := []interface{}{}
	paramCount := 1

	for key, value := range updates {
		if paramCount > 1 {
			query += ", "
		}
		
		// Special handling for config
		if key == "config" {
			configJSON, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			query += fmt.Sprintf("%s = $%d", key, paramCount)
			params = append(params, configJSON)
		} else {
			query += fmt.Sprintf("%s = $%d", key, paramCount)
			params = append(params, value)
		}
		paramCount++
	}

	query += fmt.Sprintf(", updated_at = $%d WHERE id = $%d", paramCount, paramCount+1)
	params = append(params, time.Now(), id)

	result, err := r.db.Exec(ctx, query, params...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("provider not found")
	}

	return nil
}

func (r *recruitingRepository) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM recruiting_providers WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("provider not found")
	}

	return nil
}

func (r *recruitingRepository) UpdateProviderStats(ctx context.Context, id uuid.UUID, jobsPosted, applicants int) error {
	query := `
		UPDATE recruiting_providers 
		SET jobs_posted = jobs_posted + $1,
		    applicants_total = applicants_total + $2,
		    last_synced_at = $3,
		    updated_at = $4
		WHERE id = $5
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, jobsPosted, applicants, now, now, id)
	return err
}

// Dashboard & Stats methods

func (r *recruitingRepository) GetDashboardStats(ctx context.Context) (*models.RecruitingDashboardStats, error) {
	stats := &models.RecruitingDashboardStats{
		ApplicationsByMonth: make(map[string]int),
	}

	// Active jobs
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM job_postings WHERE status = 'active'
	`).Scan(&stats.ActiveJobs)
	if err != nil {
		return nil, err
	}

	// Total applications
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM candidates
	`).Scan(&stats.TotalApplications)
	if err != nil {
		return nil, err
	}

	// Interviews scheduled
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM interviews 
		WHERE status = 'scheduled' AND scheduled_at > NOW()
	`).Scan(&stats.InterviewsScheduled)
	if err != nil {
		return nil, err
	}

	// Offers extended
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM candidates WHERE status = 'offered'
	`).Scan(&stats.OffersExtended)
	if err != nil {
		return nil, err
	}

	// Average time to hire (simplified)
	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(EXTRACT(DAY FROM (updated_at - applied_date))), 0)
		FROM candidates
		WHERE status = 'hired'
		AND updated_at > NOW() - INTERVAL '90 days'
	`).Scan(&stats.AverageTimeToHire)
	if err != nil {
		return nil, err
	}

	// Applications by month (last 6 months)
	rows, err := r.db.Query(ctx, `
		SELECT 
			TO_CHAR(applied_date, 'YYYY-MM') as month,
			COUNT(*) as count
		FROM candidates
		WHERE applied_date > NOW() - INTERVAL '6 months'
		GROUP BY month
		ORDER BY month DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var month string
		var count int
		if err := rows.Scan(&month, &count); err != nil {
			return nil, err
		}
		stats.ApplicationsByMonth[month] = count
	}

	return stats, nil
}

func (r *recruitingRepository) GetDashboard(ctx context.Context) (*models.RecruitingDashboard, error) {
	dashboard := &models.RecruitingDashboard{}

	// Get stats
	stats, err := r.GetDashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	dashboard.Stats = *stats

	// Get recent applications
	recentApps, err := r.GetRecentApplications(ctx, 10)
	if err != nil {
		return nil, err
	}
	dashboard.RecentApplications = recentApps

	// Get top performing jobs
	topJobs, err := r.GetTopPerformingJobs(ctx, 5)
	if err != nil {
		return nil, err
	}
	dashboard.TopPerformingJobs = topJobs

	// Get upcoming interviews
	interviews, err := r.GetUpcomingInterviews(ctx, 10)
	if err != nil {
		return nil, err
	}
	dashboard.UpcomingInterviews = interviews

	return dashboard, nil
}

func (r *recruitingRepository) GetRecentApplications(ctx context.Context, limit int) ([]*models.CandidateWithJob, error) {
	query := `
		SELECT 
			c.id, c.job_posting_id, c.first_name, c.last_name, c.email, c.phone,
			c.resume_url, c.cover_letter, c.linkedin_url, c.portfolio_url,
			c.status, c.score, c.ai_summary, c.strengths, c.weaknesses,
			c.experience_years, c.skills, c.applied_date, c.notes,
			c.created_at, c.updated_at,
			j.title as job_title, j.department as job_department
		FROM candidates c
		JOIN job_postings j ON c.job_posting_id = j.id
		ORDER BY c.applied_date DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []*models.CandidateWithJob

	for rows.Next() {
		var c models.CandidateWithJob
		err := rows.Scan(
			&c.ID, &c.JobPostingID, &c.FirstName, &c.LastName, &c.Email, &c.Phone,
			&c.ResumeURL, &c.CoverLetter, &c.LinkedInURL, &c.PortfolioURL,
			&c.Status, &c.Score, &c.AISummary, pq.Array(&c.Strengths), pq.Array(&c.Weaknesses),
			&c.ExperienceYears, pq.Array(&c.Skills), &c.AppliedDate, &c.Notes,
			&c.CreatedAt, &c.UpdatedAt,
			&c.JobTitle, &c.JobDepartment,
		)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, &c)
	}

	return candidates, nil
}

func (r *recruitingRepository) GetTopPerformingJobs(ctx context.Context, limit int) ([]*models.JobPosting, error) {
	query := `
		SELECT 
			id, title, department, location, employment_type,
			salary_min, salary_max, description, requirements,
			responsibilities, benefits, status, posted_date,
			closed_date, applications_count, created_by,
			created_at, updated_at
		FROM job_postings
		WHERE status = 'active'
		ORDER BY applications_count DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.JobPosting

	for rows.Next() {
		var job models.JobPosting
		err := rows.Scan(
			&job.ID, &job.Title, &job.Department, &job.Location, &job.EmploymentType,
			&job.SalaryMin, &job.SalaryMax, &job.Description, pq.Array(&job.Requirements),
			pq.Array(&job.Responsibilities), pq.Array(&job.Benefits), &job.Status,
			&job.PostedDate, &job.ClosedDate, &job.ApplicationsCount, &job.CreatedBy,
			&job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

func (r *recruitingRepository) GetUpcomingInterviews(ctx context.Context, limit int) ([]*models.InterviewWithDetails, error) {
	query := `
		SELECT 
			i.id, i.candidate_id, i.interviewer_id, i.scheduled_at,
			i.duration, i.interview_type, i.location, i.meeting_url,
			i.status, i.feedback, i.rating, i.created_at, i.updated_at,
			c.first_name || ' ' || c.last_name as candidate_name,
			j.title as job_title,
			u.first_name || ' ' || u.last_name as interviewer_name
		FROM interviews i
		JOIN candidates c ON i.candidate_id = c.id
		JOIN job_postings j ON c.job_posting_id = j.id
		JOIN users u ON i.interviewer_id = u.id
		WHERE i.status = 'scheduled' AND i.scheduled_at > NOW()
		ORDER BY i.scheduled_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interviews []*models.InterviewWithDetails

	for rows.Next() {
		var interview models.InterviewWithDetails
		err := rows.Scan(
			&interview.ID, &interview.CandidateID, &interview.InterviewerID,
			&interview.ScheduledAt, &interview.Duration, &interview.InterviewType,
			&interview.Location, &interview.MeetingURL, &interview.Status,
			&interview.Feedback, &interview.Rating, &interview.CreatedAt,
			&interview.UpdatedAt, &interview.CandidateName, &interview.JobTitle,
			&interview.InterviewerName,
		)
		if err != nil {
			return nil, err
		}
		interviews = append(interviews, &interview)
	}

	return interviews, nil
}

func (r *recruitingRepository) GetApplicantLeaderboard(ctx context.Context, limit int) ([]*models.ApplicantLeaderboard, error) {
	query := `
		WITH ranked_candidates AS (
			SELECT 
				c.id as candidate_id,
				c.first_name || ' ' || c.last_name as candidate_name,
				j.title as job_title,
				COALESCE(c.score, 0) as score,
				c.status,
				c.applied_date,
				'Direct' as source,
				ROW_NUMBER() OVER (ORDER BY COALESCE(c.score, 0) DESC, c.applied_date DESC) as rank
			FROM candidates c
			JOIN job_postings j ON c.job_posting_id = j.id
			WHERE c.status NOT IN ('rejected', 'withdrawn')
		)
		SELECT 
			rank, candidate_id, candidate_name, job_title, score, status, applied_date, source,
			LEAST(score, 100) as skills_match,
			LEAST(score - 10, 100) as experience_match,
			LEAST(score - 5, 100) as culture_fit
		FROM ranked_candidates
		ORDER BY rank
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaderboard []*models.ApplicantLeaderboard

	for rows.Next() {
		var entry models.ApplicantLeaderboard
		err := rows.Scan(
			&entry.Rank, &entry.CandidateID, &entry.CandidateName,
			&entry.JobTitle, &entry.Score, &entry.Status,
			&entry.AppliedDate, &entry.Source,
			&entry.SkillsMatch, &entry.ExperienceMatch, &entry.CultureFit,
		)
		if err != nil {
			return nil, err
		}
		leaderboard = append(leaderboard, &entry)
	}

	return leaderboard, nil
}
