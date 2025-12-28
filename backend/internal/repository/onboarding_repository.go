package repository

import (
	"context"

	"hub-hrms/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OnboardingRepository interface
type OnboardingRepository interface {
	CreateTask(ctx context.Context, task *models.OnboardingTask) error
	GetTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.OnboardingTask, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error)
	UpdateTask(ctx context.Context, task *models.OnboardingTask) error
	DeleteTask(ctx context.Context, id uuid.UUID) error
}

func NewOnboardingRepository(db *pgxpool.Pool) OnboardingRepository {
	return &onboardingRepository{db: db}
}

func (r *onboardingRepository) CreateTask(ctx context.Context, task *models.OnboardingTask) error {
	query := `
		INSERT INTO onboarding_tasks (
			employee_id, task_name, description, category, status, due_date,
			assigned_to, documents_required
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		task.EmployeeID, task.TaskName, task.Description, task.Category,
		task.Status, task.DueDate, task.AssignedTo, task.DocumentsRequired,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *onboardingRepository) GetTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.OnboardingTask, error) {
	query := `
		SELECT id, employee_id, task_name, description, category, status, due_date,
			completed_at, assigned_to, documents_required, document_url, created_at, updated_at
		FROM onboarding_tasks
		WHERE employee_id = $1
		ORDER BY due_date NULLS LAST, created_at
	`
	
	rows, err := r.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.OnboardingTask
	for rows.Next() {
		task := &models.OnboardingTask{}
		err := rows.Scan(
			&task.ID, &task.EmployeeID, &task.TaskName, &task.Description,
			&task.Category, &task.Status, &task.DueDate, &task.CompletedAt,
			&task.AssignedTo, &task.DocumentsRequired, &task.DocumentURL,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r *onboardingRepository) GetTaskByID(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error) {
	task := &models.OnboardingTask{}
	query := `
		SELECT id, employee_id, task_name, description, category, status, due_date,
			completed_at, assigned_to, documents_required, document_url, created_at, updated_at
		FROM onboarding_tasks WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.EmployeeID, &task.TaskName, &task.Description,
		&task.Category, &task.Status, &task.DueDate, &task.CompletedAt,
		&task.AssignedTo, &task.DocumentsRequired, &task.DocumentURL,
		&task.CreatedAt, &task.UpdatedAt,
	)
	return task, err
}

func (r *onboardingRepository) UpdateTask(ctx context.Context, task *models.OnboardingTask) error {
	query := `
		UPDATE onboarding_tasks SET
			task_name = $1, description = $2, category = $3, status = $4,
			due_date = $5, completed_at = $6, assigned_to = $7,
			documents_required = $8, document_url = $9, updated_at = NOW()
		WHERE id = $10
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		task.TaskName, task.Description, task.Category, task.Status,
		task.DueDate, task.CompletedAt, task.AssignedTo,
		task.DocumentsRequired, task.DocumentURL, task.ID,
	).Scan(&task.UpdatedAt)
}

func (r *onboardingRepository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM onboarding_tasks WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
