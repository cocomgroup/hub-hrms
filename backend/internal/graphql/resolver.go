package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Resolver is the root resolver
type Resolver struct {
	DB *sql.DB
}

// Internal helper struct to store application data from database
type applicationData struct {
	ID                string
	JobID             string
	CandidateID       string
	Status            string
	AppliedDate       time.Time
	LastUpdated       time.Time
	ResumeURL         sql.NullString
	CoverLetter       sql.NullString
	LinkedinURL       sql.NullString
	PortfolioURL      sql.NullString
	YearsOfExperience sql.NullInt64
	CurrentLocation   sql.NullString
	WillingToRelocate sql.NullBool
	ExpectedSalary    sql.NullFloat64
	Availability      sql.NullString
}

// =============================================================================
// QUERY RESOLVERS
// =============================================================================

func (r *queryResolver) Jobs(ctx context.Context, filters *JobFilters, limit *int, offset *int) ([]*Job, error) {
	query := `
		SELECT 
			j.id, j.title, j.department, j.location, j.employment_type,
			j.experience_level, j.salary_min, j.salary_max, j.salary_currency,
			j.description, j.requirements, j.responsibilities, j.benefits,
			j.skills, j.status, j.posted_date, j.closed_date,
			j.applications_count, j.view_count, j.remote_work, j.urgent_hiring,
			j.created_by, j.created_at, j.updated_at,
			COALESCE(u.first_name || ' ' || u.last_name, '') as user_name,
			u.email as user_email
		FROM job_postings j
		LEFT JOIN users u ON j.created_by = u.id
		WHERE true
	`
	args := []interface{}{}
	argIdx := 1

	if filters != nil {
		if filters.Department != nil && *filters.Department != "" {
			query += fmt.Sprintf(" AND j.department = $%d", argIdx)
			args = append(args, *filters.Department)
			argIdx++
		}
		if filters.Status != nil && *filters.Status != "" {
			query += fmt.Sprintf(" AND j.status = $%d", argIdx)
			args = append(args, *filters.Status)
			argIdx++
		}
		if filters.EmploymentType != nil && *filters.EmploymentType != "" {
			query += fmt.Sprintf(" AND j.employment_type = $%d", argIdx)
			args = append(args, *filters.EmploymentType)
			argIdx++
		}
		if filters.RemoteWork != nil {
			query += fmt.Sprintf(" AND j.remote_work = $%d", argIdx)
			args = append(args, *filters.RemoteWork)
			argIdx++
		}
		if filters.ExperienceLevel != nil && *filters.ExperienceLevel != "" {
			query += fmt.Sprintf(" AND j.experience_level = $%d", argIdx)
			args = append(args, *filters.ExperienceLevel)
			argIdx++
		}
		if filters.Location != nil && *filters.Location != "" {
			query += fmt.Sprintf(" AND j.location ILIKE $%d", argIdx)
			args = append(args, "%"+*filters.Location+"%")
			argIdx++
		}
	}

	query += " ORDER BY j.posted_date DESC NULLS LAST, j.created_at DESC"

	limitVal := 50
	if limit != nil && *limit > 0 {
		limitVal = *limit
	}
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limitVal)
	argIdx++

	if offset != nil && *offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, *offset)
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("Error querying jobs: %v", err)
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*Job
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			log.Printf("Error scanning job: %v", err)
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (r *queryResolver) Job(ctx context.Context, id string) (*Job, error) {
	query := `
		SELECT 
			j.id, j.title, j.department, j.location, j.employment_type,
			j.experience_level, j.salary_min, j.salary_max, j.salary_currency,
			j.description, j.requirements, j.responsibilities, j.benefits,
			j.skills, j.status, j.posted_date, j.closed_date,
			j.applications_count, j.view_count, j.remote_work, j.urgent_hiring,
			j.created_by, j.created_at, j.updated_at,
			COALESCE(u.first_name || ' ' || u.last_name, '') as user_name,
			u.email as user_email
		FROM job_postings j
		LEFT JOIN users u ON j.created_by = u.id
		WHERE j.id = $1
	`

	row := r.DB.QueryRowContext(ctx, query, id)
	job, err := scanJob(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Printf("Error querying job %s: %v", id, err)
		return nil, err
	}

	return job, nil
}

func (r *queryResolver) Applications(ctx context.Context, filters *ApplicationFilters, limit *int, offset *int) ([]*Application, error) {
	query := `
		SELECT 
			c.id, 
			c.job_posting_id, 
			c.id, 
			c.status, 
			c.applied_date, 
			c.updated_at, 
			COALESCE(c.resume_url, ''), 
			COALESCE(c.cover_letter, ''), 
			COALESCE(c.linkedin_url, ''),
			COALESCE(c.portfolio_url, ''), 
			COALESCE(c.experience_years, 0), 
			COALESCE(c.location, ''),
			COALESCE(c.willing_to_relocate, false), 
			COALESCE(c.expected_salary, 0), 
			COALESCE(c.availability, '')
		FROM candidates c
		WHERE c.job_posting_id IS NOT NULL
	`
	args := []interface{}{}
	argIdx := 1

	if filters != nil {
		if filters.JobID != nil && *filters.JobID != "" {
			query += fmt.Sprintf(" AND c.job_posting_id = $%d", argIdx)
			args = append(args, *filters.JobID)
			argIdx++
		}
		if filters.Status != nil && *filters.Status != "" {
			query += fmt.Sprintf(" AND c.status = $%d", argIdx)
			args = append(args, *filters.Status)
			argIdx++
		}
		if filters.DateFrom != nil && *filters.DateFrom != "" {
			query += fmt.Sprintf(" AND c.applied_date >= $%d", argIdx)
			args = append(args, *filters.DateFrom)
			argIdx++
		}
		if filters.DateTo != nil && *filters.DateTo != "" {
			query += fmt.Sprintf(" AND c.applied_date <= $%d", argIdx)
			args = append(args, *filters.DateTo)
			argIdx++
		}
	}

	query += " ORDER BY c.applied_date DESC"

	limitVal := 100
	if limit != nil && *limit > 0 {
		limitVal = *limit
	}
	query += fmt.Sprintf(" LIMIT %d", limitVal)

	if offset != nil && *offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", *offset)
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("Error querying applications")
		log.Printf("Query: %s", query)
		log.Printf("Args: %v", args)
		log.Printf("Error: %v", err)
		return nil, fmt.Errorf("failed to query applications: %w", err)
	}
	defer rows.Close()

	var applications []*Application
	for rows.Next() {
		appData := &applicationData{}
		err := rows.Scan(
			&appData.ID, &appData.JobID, &appData.CandidateID, &appData.Status,
			&appData.AppliedDate, &appData.LastUpdated, &appData.ResumeURL,
			&appData.CoverLetter, &appData.LinkedinURL, &appData.PortfolioURL,
			&appData.YearsOfExperience, &appData.CurrentLocation,
			&appData.WillingToRelocate, &appData.ExpectedSalary, &appData.Availability,
		)
		if err != nil {
			log.Printf("Error scanning application: %v", err)
			return nil, err
		}

		app := convertApplicationData(appData)
		applications = append(applications, app)
	}

	return applications, nil
}

func (r *queryResolver) Application(ctx context.Context, id string) (*Application, error) {
	query := `
		SELECT 
			c.id::text, 
			COALESCE(c.job_posting_id::text, ''), 
			c.id::text, 
			c.status, 
			c.applied_date, 
			c.updated_at, 
			COALESCE(c.resume_url, ''), 
			COALESCE(c.cover_letter, ''), 
			COALESCE(c.linkedin_url, ''),
			COALESCE(c.portfolio_url, ''), 
			COALESCE(c.experience_years, 0), 
			COALESCE(c.location, ''),
			COALESCE(c.willing_to_relocate, false), 
			COALESCE(c.expected_salary, 0), 
			COALESCE(c.availability, '')
		FROM candidates c
		WHERE c.id = $1
	`

	appData := &applicationData{}
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&appData.ID, &appData.JobID, &appData.CandidateID, &appData.Status,
		&appData.AppliedDate, &appData.LastUpdated, &appData.ResumeURL,
		&appData.CoverLetter, &appData.LinkedinURL, &appData.PortfolioURL,
		&appData.YearsOfExperience, &appData.CurrentLocation,
		&appData.WillingToRelocate, &appData.ExpectedSalary, &appData.Availability,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Printf("Error querying application %s: %v", id, err)
		return nil, err
	}

	return convertApplicationData(appData), nil
}

func (r *queryResolver) Candidate(ctx context.Context, id string) (*Candidate, error) {
	query := `
		SELECT 
			id, first_name, last_name, email, phone, location,
			headline, summary, resume_url, linkedin_url, portfolio_url,
			github_url, skills, availability, expected_salary,
			preferred_locations, remote_preference, created_at, updated_at
		FROM candidates
		WHERE id = $1
	`

	row := r.DB.QueryRowContext(ctx, query, id)
	candidate, err := scanCandidate(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Printf("Error querying candidate %s: %v", id, err)
		return nil, err
	}

	return candidate, nil
}

func (r *queryResolver) RecruitmentMetrics(ctx context.Context, dateRange DateRangeInput) (*RecruitmentMetrics, error) {
	fromDate, err := time.Parse(time.RFC3339, dateRange.From)
	if err != nil {
		return nil, fmt.Errorf("invalid from date: %w", err)
	}
	toDate, err := time.Parse(time.RFC3339, dateRange.To)
	if err != nil {
		return nil, fmt.Errorf("invalid to date: %w", err)
	}

	var totalJobs, activeJobs, totalApplications int
	var avgApplicationsPerJob float64

	query := `
		SELECT 
			COUNT(DISTINCT j.id) as total_jobs,
			COUNT(DISTINCT CASE WHEN j.status = 'PUBLISHED' THEN j.id END) as active_jobs,
			COUNT(DISTINCT c.id) as total_applications,
			CASE 
				WHEN COUNT(DISTINCT j.id) > 0 
				THEN COUNT(DISTINCT c.id)::float / COUNT(DISTINCT j.id)::float 
				ELSE 0 
			END as avg_applications_per_job
		FROM job_postings j
		LEFT JOIN candidates c ON j.id = c.job_posting_id 
			AND c.applied_date BETWEEN $1 AND $2
		WHERE j.created_at BETWEEN $1 AND $2
	`

	err = r.DB.QueryRowContext(ctx, query, fromDate, toDate).Scan(
		&totalJobs, &activeJobs, &totalApplications, &avgApplicationsPerJob,
	)
	if err != nil {
		log.Printf("Error querying recruitment metrics: %v", err)
		return nil, err
	}

	return &RecruitmentMetrics{
		TotalJobs:            totalJobs,
		ActiveJobs:           activeJobs,
		TotalApplications:    totalApplications,
		AvgApplicationsPerJob: avgApplicationsPerJob,
	}, nil
}

func (r *queryResolver) JobPerformance(ctx context.Context, jobID string) (*JobPerformance, error) {
	query := `
		SELECT j.view_count, j.applications_count,
			CASE WHEN j.view_count > 0 
			THEN (j.applications_count::float / j.view_count::float) * 100 
			ELSE 0 END as conversion_rate
		FROM job_postings j WHERE j.id = $1
	`

	var views, applications int
	var conversionRate float64

	err := r.DB.QueryRowContext(ctx, query, jobID).Scan(&views, &applications, &conversionRate)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found")
	}
	if err != nil {
		return nil, err
	}

	job, err := r.Job(ctx, jobID)
	if err != nil {
		return nil, err
	}

	return &JobPerformance{
		Job:            job,
		Views:          views,
		Applications:   applications,
		ConversionRate: conversionRate,
	}, nil
}

func (r *queryResolver) ApplicationPipeline(ctx context.Context, jobID *string) ([]*PipelineStage, error) {
	query := `SELECT status, COUNT(*) as count FROM candidates WHERE true`
	args := []interface{}{}
	
	if jobID != nil && *jobID != "" {
		query += " AND job_posting_id = $1"
		args = append(args, *jobID)
	}
	
	query += " GROUP BY status ORDER BY count DESC"

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stages []*PipelineStage
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stages = append(stages, &PipelineStage{
			Status:       status,
			Count:        count,
			Applications: []*Application{},
		})
	}

	return stages, nil
}

// =============================================================================
// MUTATION RESOLVERS
// =============================================================================

func (r *mutationResolver) SubmitApplication(ctx context.Context, input ApplicationInput) (*Application, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var candidateID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM candidates WHERE email = $1`, input.Email).Scan(&candidateID)

	if err == sql.ErrNoRows {
		candidateID = uuid.New().String()
		_, err = tx.ExecContext(ctx, `
			INSERT INTO candidates (
				id, first_name, last_name, email, phone, location,
				resume_url, linkedin_url, portfolio_url, github_url,
				availability, expected_salary, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		`, candidateID, input.FirstName, input.LastName, input.Email,
			input.Phone, input.CurrentLocation, input.ResumeURL,
			input.LinkedinURL, input.PortfolioURL, nil,
			input.Availability, input.ExpectedSalary)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	appID := uuid.New().String()
	now := time.Now()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO candidates (
			id, job_posting_id, first_name, last_name, email, phone,
			status, applied_date, 
			resume_url, cover_letter, linkedin_url, portfolio_url,
			experience_years, location,
			expected_salary, availability
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, appID, input.JobID, input.FirstName, input.LastName, input.Email, input.Phone,
		"NEW", now,
		input.ResumeURL, input.CoverLetter, input.LinkedinURL, input.PortfolioURL,
		input.YearsOfExperience, input.CurrentLocation,
		input.ExpectedSalary, input.Availability)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `UPDATE job_postings SET applications_count = applications_count + 1, updated_at = NOW() WHERE id = $1`, input.JobID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	appliedDate := now.Format(time.RFC3339)
	return &Application{
		ID:          appID,
		Status:      "NEW",
		AppliedDate: appliedDate,
		LastUpdated: appliedDate,
	}, nil
}

func (r *mutationResolver) IncrementJobView(ctx context.Context, id string) (*Job, error) {
	_, err := r.DB.ExecContext(ctx, `UPDATE job_postings SET view_count = view_count + 1, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return r.Resolver.Query().Job(ctx, id)
}

func (r *mutationResolver) UpdateApplicationStatus(ctx context.Context, id string, status ApplicationStatus, note *string) (*Application, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE candidates SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	if err != nil {
		return nil, err
	}

	if note != nil && *note != "" {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO application_notes (id, application_id, content, is_internal, created_at)
			VALUES ($1, $2, $3, true, NOW())
		`, uuid.New().String(), id, *note)
		if err != nil {
			log.Printf("Warning: failed to add note: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.Resolver.Query().Application(ctx, id)
}

func (r *mutationResolver) BulkUpdateApplicationStatus(ctx context.Context, ids []string, status ApplicationStatus) ([]*Application, error) {
	_, err := r.DB.ExecContext(ctx, `UPDATE candidates SET status = $1, updated_at = NOW() WHERE id = ANY($2)`, status, pq.Array(ids))
	if err != nil {
		return nil, err
	}

	var apps []*Application
	for _, id := range ids {
		app, err := r.Resolver.Query().Application(ctx, id)
		if err != nil {
			log.Printf("Warning: failed to load application %s: %v", id, err)
			continue
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func (r *mutationResolver) AddApplicationNote(ctx context.Context, applicationID string, content string, isInternal *bool) (*ApplicationNote, error) {
	noteID := uuid.New().String()
	internal := true
	if isInternal != nil {
		internal = *isInternal
	}

	now := time.Now()
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO application_notes (id, application_id, content, is_internal, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, noteID, applicationID, content, internal, now)
	if err != nil {
		return nil, err
	}

	return &ApplicationNote{
		ID:         noteID,
		Content:    content,
		CreatedAt:  now.Format(time.RFC3339),
		IsInternal: internal,
	}, nil
}

func (r *mutationResolver) ScoreApplication(ctx context.Context, applicationID string) (*AIScore, error) {
	now := time.Now()
	return &AIScore{
		Overall:       75.0,
		Recommendation: "PROCEED",
		GeneratedAt:   now.Format(time.RFC3339),
	}, nil
}

func (r *mutationResolver) CreateJob(ctx context.Context, input JobInput) (*Job, error) {
	jobID := uuid.New().String()
	now := time.Now()

	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO job_postings (
			id, title, department, location, employment_type, experience_level,
			salary_min, salary_max, salary_currency, description, requirements,
			responsibilities, benefits, skills, remote_work, urgent_hiring,
			status, applications_count, view_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, 
				  'DRAFT', 0, 0, $17, $18)
	`, jobID, input.Title, input.Department, input.Location, input.EmploymentType,
		input.ExperienceLevel, input.SalaryMin, input.SalaryMax, input.SalaryCurrency,
		input.Description, input.Requirements, input.Responsibilities, input.Benefits,
		pq.Array(input.Skills), input.RemoteWork, input.UrgentHiring, now, now)
	if err != nil {
		return nil, err
	}

	return r.Resolver.Query().Job(ctx, jobID)
}

func (r *mutationResolver) UpdateJob(ctx context.Context, id string, input JobInput) (*Job, error) {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE job_postings SET
			title = $1, department = $2, location = $3, employment_type = $4,
			experience_level = $5, salary_min = $6, salary_max = $7, salary_currency = $8,
			description = $9, requirements = $10, responsibilities = $11, benefits = $12,
			skills = $13, remote_work = $14, urgent_hiring = $15, updated_at = NOW()
		WHERE id = $16
	`, input.Title, input.Department, input.Location, input.EmploymentType,
		input.ExperienceLevel, input.SalaryMin, input.SalaryMax, input.SalaryCurrency,
		input.Description, input.Requirements, input.Responsibilities, input.Benefits,
		pq.Array(input.Skills), input.RemoteWork, input.UrgentHiring, id)
	if err != nil {
		return nil, err
	}

	return r.Resolver.Query().Job(ctx, id)
}

func (r *mutationResolver) PublishJob(ctx context.Context, id string) (*Job, error) {
	_, err := r.DB.ExecContext(ctx, `UPDATE job_postings SET status = 'PUBLISHED', posted_date = NOW(), updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return r.Resolver.Query().Job(ctx, id)
}

func (r *mutationResolver) CloseJob(ctx context.Context, id string) (*Job, error) {
	_, err := r.DB.ExecContext(ctx, `UPDATE job_postings SET status = 'CLOSED', closed_date = NOW(), updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return r.Resolver.Query().Job(ctx, id)
}

func (r *mutationResolver) DeleteJob(ctx context.Context, id string) (bool, error) {
	result, err := r.DB.ExecContext(ctx, `DELETE FROM job_postings WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0, nil
}

func (r *mutationResolver) GenerateJobDescription(ctx context.Context, input JobDescriptionInput) (*JobDescriptionOutput, error) {
	return &JobDescriptionOutput{
		Description: fmt.Sprintf("We are seeking a talented %s to join our %s team.", input.Title, input.Department),
		Requirements: "Bachelor's degree or equivalent experience required.",
		Responsibilities: strings.Join(input.KeyResponsibilities, "\n"),
		SuggestedSkills: input.RequiredSkills,
	}, nil
}

func (r *mutationResolver) UpdateCandidateProfile(ctx context.Context, id string, input CandidateProfileInput) (*Candidate, error) {
	updates := []string{}
	args := []interface{}{}
	argIdx := 1

	if input.FirstName != nil {
		updates = append(updates, fmt.Sprintf("first_name = $%d", argIdx))
		args = append(args, *input.FirstName)
		argIdx++
	}
	if input.LastName != nil {
		updates = append(updates, fmt.Sprintf("last_name = $%d", argIdx))
		args = append(args, *input.LastName)
		argIdx++
	}
	if input.Email != nil {
		updates = append(updates, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *input.Email)
		argIdx++
	}
	if input.Phone != nil {
		updates = append(updates, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, *input.Phone)
		argIdx++
	}
	if input.Location != nil {
		updates = append(updates, fmt.Sprintf("location = $%d", argIdx))
		args = append(args, *input.Location)
		argIdx++
	}
	if input.Skills != nil {
		updates = append(updates, fmt.Sprintf("skills = $%d", argIdx))
		args = append(args, pq.Array(input.Skills))
		argIdx++
	}

	updates = append(updates, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE candidates SET %s WHERE id = $%d", strings.Join(updates, ", "), argIdx)
	_, err := r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return r.Resolver.Query().Candidate(ctx, id)
}

// =============================================================================
// NESTED FIELD RESOLVERS (if gqlgen generates ApplicationResolver)
// =============================================================================

// Helper cache to store job and candidate IDs from application queries
var appDataCache = make(map[string]*applicationData)

// Note: These methods will be called by gqlgen's generated code if it creates
// an ApplicationResolver interface. If you get "undefined: ApplicationResolver",
// it means gqlgen didn't generate it. In that case, Application.job and 
// Application.candidate are probably defined as regular fields in the schema,
// not as nested resolvers.

func (r *applicationResolver) Job(ctx context.Context, obj *Application) (*Job, error) {
	// Try cache first
	appData, ok := appDataCache[obj.ID]
	if !ok {
		// Fetch job_posting_id from database
		var jobID string
		err := r.DB.QueryRowContext(ctx, `SELECT job_posting_id FROM candidates WHERE id = $1`, obj.ID).Scan(&jobID)
		if err != nil {
			return nil, err
		}
		return r.Resolver.Query().Job(ctx, jobID)
	}
	return r.Resolver.Query().Job(ctx, appData.JobID)
}

func (r *applicationResolver) Candidate(ctx context.Context, obj *Application) (*Candidate, error) {
	// Try cache first
	appData, ok := appDataCache[obj.ID]
	if !ok {
		// In our schema, candidate_id is the same as the application id
		// since each candidate record IS an application
		return r.Resolver.Query().Candidate(ctx, obj.ID)
	}
	return r.Resolver.Query().Candidate(ctx, appData.CandidateID)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanJob(s scanner) (*Job, error) {
	var (
		id, title, department, location, employmentType string
		status                                          string
		createdAt, updatedAt                            time.Time
		createdByID                                     sql.NullString
		userName, userEmail                             sql.NullString
		experienceLevel, description                    sql.NullString
		requirements, responsibilities, benefits        sql.NullString
		skillsArray                                     pq.StringArray
		postedDate, closingDate                         sql.NullTime
		salaryMin, salaryMax                            sql.NullFloat64
		salaryCurrency                                  sql.NullString
		applicationCount, viewCount                     int
		remoteWork, urgentHiring                        sql.NullBool
	)

	err := s.Scan(
		&id, &title, &department, &location, &employmentType,
		&experienceLevel, &salaryMin, &salaryMax, &salaryCurrency,
		&description, &requirements, &responsibilities, &benefits,
		&skillsArray, &status, &postedDate, &closingDate,
		&applicationCount, &viewCount, &remoteWork, &urgentHiring,
		&createdByID, &createdAt, &updatedAt, &userName, &userEmail,
	)
	if err != nil {
		return nil, err
	}

	job := &Job{
		ID:               id,
		Title:            title,
		Department:       department,
		Location:         location,
		EmploymentType:   employmentType,
		Status:           status,
		ApplicationCount: applicationCount,
		ViewCount:        viewCount,
		CreatedAt:        createdAt.Format(time.RFC3339),
		UpdatedAt:        updatedAt.Format(time.RFC3339),
	}

	if experienceLevel.Valid {
		job.ExperienceLevel = &experienceLevel.String
	}
	if description.Valid {
		job.Description = &description.String
	}
	if requirements.Valid {
		job.Requirements = &requirements.String
	}
	if responsibilities.Valid {
		job.Responsibilities = &responsibilities.String
	}
	if benefits.Valid {
		job.Benefits = &benefits.String
	}
	if len(skillsArray) > 0 {
		job.Skills = skillsArray
	}
	if postedDate.Valid {
		pd := postedDate.Time.Format(time.RFC3339)
		job.PostedDate = &pd
	}
	if closingDate.Valid {
		cd := closingDate.Time.Format(time.RFC3339)
		job.ClosingDate = &cd
	}
	if salaryMin.Valid || salaryMax.Valid {
		job.SalaryRange = &SalaryRange{}
		if salaryMin.Valid {
			job.SalaryRange.Min = &salaryMin.Float64
		}
		if salaryMax.Valid {
			job.SalaryRange.Max = &salaryMax.Float64
		}
		if salaryCurrency.Valid {
			job.SalaryRange.Currency = &salaryCurrency.String
		}
	}
	if remoteWork.Valid {
		job.RemoteWork = &remoteWork.Bool
	}
	if urgentHiring.Valid {
		job.UrgentHiring = &urgentHiring.Bool
	}
	if createdByID.Valid && userName.Valid {
		job.CreatedBy = &User{
			ID:   createdByID.String,
			Name: userName.String,
		}
		if userEmail.Valid {
			job.CreatedBy.Email = &userEmail.String
		}
	}

	return job, nil
}

func convertApplicationData(appData *applicationData) *Application {
	// Cache the application data for nested resolvers
	appDataCache[appData.ID] = appData

	app := &Application{
		ID:          appData.ID,
		Status:      appData.Status,
		AppliedDate: appData.AppliedDate.Format(time.RFC3339),
		LastUpdated: appData.LastUpdated.Format(time.RFC3339),
	}

	if appData.ResumeURL.Valid {
		app.ResumeURL = &appData.ResumeURL.String
	}
	if appData.CoverLetter.Valid {
		app.CoverLetter = &appData.CoverLetter.String
	}
	if appData.LinkedinURL.Valid {
		app.LinkedinURL = &appData.LinkedinURL.String
	}
	if appData.PortfolioURL.Valid {
		app.PortfolioURL = &appData.PortfolioURL.String
	}
	if appData.YearsOfExperience.Valid {
		years := int(appData.YearsOfExperience.Int64)
		app.YearsOfExperience = &years
	}
	if appData.CurrentLocation.Valid {
		app.CurrentLocation = &appData.CurrentLocation.String
	}
	if appData.WillingToRelocate.Valid {
		app.WillingToRelocate = &appData.WillingToRelocate.Bool
	}
	if appData.ExpectedSalary.Valid {
		app.ExpectedSalary = &appData.ExpectedSalary.Float64
	}
	if appData.Availability.Valid {
		app.Availability = &appData.Availability.String
	}

	return app
}

func scanCandidate(s scanner) (*Candidate, error) {
	var (
		id, firstName, lastName, email string
		phone, location                sql.NullString
		headline, summary              sql.NullString
		resumeURL, linkedinURL         sql.NullString
		portfolioURL, githubURL        sql.NullString
		skillsArray                    pq.StringArray
		availability                   sql.NullString
		expectedSalary                 sql.NullFloat64
		preferredLocations             pq.StringArray
		remotePreference               sql.NullString
		createdAt, updatedAt           time.Time
	)

	err := s.Scan(
		&id, &firstName, &lastName, &email, &phone, &location,
		&headline, &summary, &resumeURL, &linkedinURL, &portfolioURL,
		&githubURL, &skillsArray, &availability, &expectedSalary,
		&preferredLocations, &remotePreference, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	candidate := &Candidate{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		CreatedAt: createdAt.Format(time.RFC3339),
		UpdatedAt: updatedAt.Format(time.RFC3339),
	}

	if phone.Valid {
		candidate.Phone = &phone.String
	}
	if location.Valid {
		candidate.Location = &location.String
	}
	if headline.Valid {
		candidate.Headline = &headline.String
	}
	if summary.Valid {
		candidate.Summary = &summary.String
	}
	if resumeURL.Valid {
		candidate.ResumeURL = &resumeURL.String
	}
	if linkedinURL.Valid {
		candidate.LinkedinURL = &linkedinURL.String
	}
	if portfolioURL.Valid {
		candidate.PortfolioURL = &portfolioURL.String
	}
	if githubURL.Valid {
		candidate.GithubURL = &githubURL.String
	}
	if len(skillsArray) > 0 {
		candidate.Skills = skillsArray
	}
	if availability.Valid {
		candidate.Availability = &availability.String
	}
	if expectedSalary.Valid {
		candidate.ExpectedSalary = &expectedSalary.Float64
	}
	if len(preferredLocations) > 0 {
		candidate.PreferredLocations = preferredLocations
	}
	if remotePreference.Valid {
		candidate.RemotePreference = &remotePreference.String
	}

	return candidate, nil
}

// =============================================================================
// RESOLVER TYPE IMPLEMENTATIONS
// =============================================================================

type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type applicationResolver struct{ *Resolver }

func (r *Resolver) Query() QueryResolver       { return &queryResolver{r} }
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Only implement Application() if ApplicationResolver interface exists
// If you get compile errors, comment out this function - it means your schema
// doesn't define Application.job and Application.candidate as nested resolvers
func (r *Resolver) Application() ApplicationResolver { return &applicationResolver{r} }