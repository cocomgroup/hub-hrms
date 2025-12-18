package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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