package repository

import (
	"context"

	"hub-hrms/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ApplicantRepository interface
type ApplicantRepository interface {
	Create(ctx context.Context, applicant *models.Applicant) error
	GetAll(ctx context.Context) ([]*models.Applicant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Applicant, error)
	Update(ctx context.Context, applicant *models.Applicant) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ApplicantRepository implementation
type applicantRepository struct {
	db *pgxpool.Pool
}

func NewApplicantRepository(db *pgxpool.Pool) ApplicantRepository {
	return &applicantRepository{db: db}
}

func (r *applicantRepository) Create(ctx context.Context, applicant *models.Applicant) error {
	query := `
		INSERT INTO applicants (
			id, name, email, phone, position, source, 
			resume_url, applied_date, status, ai_score, 
			ai_analysis, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.Exec(ctx, query,
		applicant.ID,
		applicant.Name,
		applicant.Email,
		applicant.Phone,
		applicant.Position,
		applicant.Source,
		applicant.ResumeURL,
		applicant.AppliedDate,
		applicant.Status,
		applicant.AIScore,
		applicant.AIAnalysis,
		applicant.Notes,
		applicant.CreatedAt,
		applicant.UpdatedAt,
	)

	return err
}

func (r *applicantRepository) GetAll(ctx context.Context) ([]*models.Applicant, error) {
	query := `
		SELECT id, name, email, phone, position, source, 
		       resume_url, applied_date, status, ai_score, 
		       ai_analysis, notes, created_at, updated_at
		FROM applicants
		ORDER BY applied_date DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applicants []*models.Applicant
	for rows.Next() {
		var a models.Applicant
		err := rows.Scan(
			&a.ID,
			&a.Name,
			&a.Email,
			&a.Phone,
			&a.Position,
			&a.Source,
			&a.ResumeURL,
			&a.AppliedDate,
			&a.Status,
			&a.AIScore,
			&a.AIAnalysis,
			&a.Notes,
			&a.CreatedAt,
			&a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		applicants = append(applicants, &a)
	}

	return applicants, rows.Err()
}

func (r *applicantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Applicant, error) {
	query := `
		SELECT id, name, email, phone, position, source, 
		       resume_url, applied_date, status, ai_score, 
		       ai_analysis, notes, created_at, updated_at
		FROM applicants
		WHERE id = $1
	`

	var applicant models.Applicant
	err := r.db.QueryRow(ctx, query, id).Scan(
		&applicant.ID,
		&applicant.Name,
		&applicant.Email,
		&applicant.Phone,
		&applicant.Position,
		&applicant.Source,
		&applicant.ResumeURL,
		&applicant.AppliedDate,
		&applicant.Status,
		&applicant.AIScore,
		&applicant.AIAnalysis,
		&applicant.Notes,
		&applicant.CreatedAt,
		&applicant.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &applicant, nil
}

func (r *applicantRepository) Update(ctx context.Context, applicant *models.Applicant) error {
	query := `
		UPDATE applicants
		SET name = $2, email = $3, phone = $4, position = $5, 
		    source = $6, status = $7, ai_score = $8, 
		    ai_analysis = $9, notes = $10, updated_at = $11
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		applicant.ID,
		applicant.Name,
		applicant.Email,
		applicant.Phone,
		applicant.Position,
		applicant.Source,
		applicant.Status,
		applicant.AIScore,
		applicant.AIAnalysis,
		applicant.Notes,
		applicant.UpdatedAt,
	)

	return err
}

func (r *applicantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM applicants WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
