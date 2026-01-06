package repository

import (
	"context"
	"fmt"
	"time"

	"hub-hrms/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OnboardingRepository interface for new hire onboarding operations
type OnboardingRepository interface {
	// Workflows
	CreateOnboarding(ctx context.Context, onboarding *models.NewHireOnboarding) error
	GetOnboarding(ctx context.Context, id uuid.UUID) (*models.NewHireOnboarding, error)
	GetOnboardingByEmployee(ctx context.Context, employeeID uuid.UUID) (*models.NewHireOnboarding, error)
	ListOnboardings(ctx context.Context, filters map[string]interface{}) ([]*models.NewHireOnboarding, error)
	UpdateOnboarding(ctx context.Context, onboarding *models.NewHireOnboarding) error
	DeleteOnboarding(ctx context.Context, id uuid.UUID) error
	
	// Tasks
	CreateTask(ctx context.Context, task *models.OnboardingTask) error
	GetTask(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error)
	GetTasksByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*models.OnboardingTask, error)
	GetTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.OnboardingTask, error)
	UpdateTask(ctx context.Context, task *models.OnboardingTask) error
	CompleteTask(ctx context.Context, taskID, completedBy uuid.UUID, actualHours *float64) error
	DeleteTask(ctx context.Context, id uuid.UUID) error
	
	// Interactions
	CreateInteraction(ctx context.Context, interaction *models.OnboardingInteraction) error
	ListInteractionsByWorkflow(ctx context.Context, workflowID uuid.UUID, limit int) ([]*models.OnboardingInteraction, error)
	
	// Milestones
	CreateMilestone(ctx context.Context, milestone *models.OnboardingMilestone) error
	GetMilestone(ctx context.Context, id uuid.UUID) (*models.OnboardingMilestone, error)
	ListMilestonesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*models.OnboardingMilestone, error)
	UpdateMilestone(ctx context.Context, milestone *models.OnboardingMilestone) error
	
	// Templates
	GetTemplate(ctx context.Context, id uuid.UUID) (*models.OnboardingChecklistTemplate, error)
	ListTemplates(ctx context.Context, department, roleType string) ([]*models.OnboardingChecklistTemplate, error)
	CreateTasksFromTemplate(ctx context.Context, workflowID uuid.UUID, templateID uuid.UUID, startDate time.Time) error
	
	// Statistics
	GetStatistics(ctx context.Context, workflowID uuid.UUID) (*models.OnboardingStatistics, error)
	GetDashboardData(ctx context.Context, filters map[string]interface{}) (*models.OnboardingDashboardResponse, error)
}

// onboardingRepository implements OnboardingRepository interface
type onboardingRepository struct {
	db *pgxpool.Pool
}

// NewOnboardingRepository creates a new onboarding repository instance
func NewOnboardingRepository(db *pgxpool.Pool) OnboardingRepository {
	return &onboardingRepository{db: db}
}

// ============================================================================
// WORKFLOW OPERATIONS
// ============================================================================

func (r *onboardingRepository) CreateOnboarding(ctx context.Context, onboarding *models.NewHireOnboarding) error {
	query := `
		INSERT INTO new_hire_onboardings (
			id, employee_id, workflow_template_id, employee_workflow_id,
			start_date, expected_completion_date, status, overall_progress,
			assigned_buddy_id, assigned_manager_id, notes, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at
	`
	
	return r.db.QueryRow(ctx, query,
		onboarding.ID,
		onboarding.EmployeeID,
		onboarding.WorkflowTemplateID,
		onboarding.EmployeeWorkflowID,
		onboarding.StartDate,
		onboarding.ExpectedCompletionDate,
		onboarding.Status,
		onboarding.OverallProgress,
		onboarding.AssignedBuddyID,
		onboarding.AssignedManagerID,
		onboarding.Notes,
		onboarding.CreatedBy,
	).Scan(&onboarding.CreatedAt, &onboarding.UpdatedAt)
}

func (r *onboardingRepository) GetOnboarding(ctx context.Context, id uuid.UUID) (*models.NewHireOnboarding, error) {
	query := `
		SELECT 
			nho.id, nho.employee_id, nho.workflow_template_id, nho.employee_workflow_id,
			nho.start_date, nho.expected_completion_date, nho.actual_completion_date,
			nho.status, nho.overall_progress,
			nho.assigned_buddy_id, nho.assigned_manager_id, nho.notes,
			nho.created_at, nho.updated_at, nho.created_by,
			e.first_name || ' ' || e.last_name as employee_name,
			e.email as employee_email,
			buddy.first_name || ' ' || buddy.last_name as buddy_name,
			mgr.first_name || ' ' || mgr.last_name as manager_name
		FROM new_hire_onboardings nho
		JOIN employees e ON e.id = nho.employee_id
		LEFT JOIN employees buddy ON buddy.id = nho.assigned_buddy_id
		LEFT JOIN employees mgr ON mgr.id = nho.assigned_manager_id
		WHERE nho.id = $1
	`
	
	onboarding := &models.NewHireOnboarding{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&onboarding.ID,
		&onboarding.EmployeeID,
		&onboarding.WorkflowTemplateID,
		&onboarding.EmployeeWorkflowID,
		&onboarding.StartDate,
		&onboarding.ExpectedCompletionDate,
		&onboarding.ActualCompletionDate,
		&onboarding.Status,
		&onboarding.OverallProgress,
		&onboarding.AssignedBuddyID,
		&onboarding.AssignedManagerID,
		&onboarding.Notes,
		&onboarding.CreatedAt,
		&onboarding.UpdatedAt,
		&onboarding.CreatedBy,
		&onboarding.EmployeeName,
		&onboarding.EmployeeEmail,
		&onboarding.BuddyName,
		&onboarding.ManagerName,
	)
	
	if err != nil {
		return nil, err
	}
	
	// Load tasks
	tasks, _ := r.GetTasksByWorkflow(ctx, id)
	onboarding.Tasks = tasks
	
	// Load milestones
	milestones, _ := r.ListMilestonesByWorkflow(ctx, id)
	onboarding.Milestones = milestones
	
	// Load recent interactions
	interactions, _ := r.ListInteractionsByWorkflow(ctx, id, 10)
	onboarding.RecentInteractions = interactions
	
	// Load statistics
	stats, _ := r.GetStatistics(ctx, id)
	onboarding.Statistics = stats
	
	return onboarding, nil
}

func (r *onboardingRepository) GetOnboardingByEmployee(ctx context.Context, employeeID uuid.UUID) (*models.NewHireOnboarding, error) {
	query := `SELECT id FROM new_hire_onboardings WHERE employee_id = $1 ORDER BY created_at DESC LIMIT 1`
	
	var id uuid.UUID
	err := r.db.QueryRow(ctx, query, employeeID).Scan(&id)
	if err != nil {
		return nil, err
	}
	
	return r.GetOnboarding(ctx, id)
}

func (r *onboardingRepository) ListOnboardings(ctx context.Context, filters map[string]interface{}) ([]*models.NewHireOnboarding, error) {
	query := `
		SELECT 
			nho.id, nho.employee_id, nho.start_date, nho.expected_completion_date,
			nho.actual_completion_date, nho.status, nho.overall_progress,
			nho.assigned_buddy_id, nho.assigned_manager_id,
			nho.created_at, nho.updated_at,
			e.first_name || ' ' || e.last_name as employee_name,
			e.email as employee_email,
			COALESCE(buddy.first_name || ' ' || buddy.last_name, '') as buddy_name,
			COALESCE(mgr.first_name || ' ' || mgr.last_name, '') as manager_name
		FROM new_hire_onboardings nho
		JOIN employees e ON e.id = nho.employee_id
		LEFT JOIN employees buddy ON buddy.id = nho.assigned_buddy_id
		LEFT JOIN employees mgr ON mgr.id = nho.assigned_manager_id
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 1
	
	if status, ok := filters["status"].(string); ok {
		query += fmt.Sprintf(" AND nho.status = $%d", argCount)
		args = append(args, status)
		argCount++
	}
	
	if managerID, ok := filters["manager_id"].(uuid.UUID); ok {
		query += fmt.Sprintf(" AND nho.assigned_manager_id = $%d", argCount)
		args = append(args, managerID)
		argCount++
	}
	
	query += " ORDER BY nho.created_at DESC LIMIT 100"
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var onboardings []*models.NewHireOnboarding
	for rows.Next() {
		o := &models.NewHireOnboarding{}
		err := rows.Scan(
			&o.ID, &o.EmployeeID, &o.StartDate,
			&o.ExpectedCompletionDate, &o.ActualCompletionDate,
			&o.Status, &o.OverallProgress,
			&o.AssignedBuddyID, &o.AssignedManagerID,
			&o.CreatedAt, &o.UpdatedAt,
			&o.EmployeeName, &o.EmployeeEmail,
			&o.BuddyName, &o.ManagerName,
		)
		if err != nil {
			return nil, err
		}
		onboardings = append(onboardings, o)
	}
	
	return onboardings, rows.Err()
}

func (r *onboardingRepository) UpdateOnboarding(ctx context.Context, onboarding *models.NewHireOnboarding) error {
	query := `
		UPDATE new_hire_onboardings
		SET 
			expected_completion_date = $2,
			actual_completion_date = $3,
			status = $4,
			overall_progress = $5,
			assigned_buddy_id = $6,
			assigned_manager_id = $7,
			notes = $8,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	
	return r.db.QueryRow(ctx, query,
		onboarding.ID,
		onboarding.ExpectedCompletionDate,
		onboarding.ActualCompletionDate,
		onboarding.Status,
		onboarding.OverallProgress,
		onboarding.AssignedBuddyID,
		onboarding.AssignedManagerID,
		onboarding.Notes,
	).Scan(&onboarding.UpdatedAt)
}

func (r *onboardingRepository) DeleteOnboarding(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM new_hire_onboardings WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ============================================================================
// TASK OPERATIONS
// ============================================================================

func (r *onboardingRepository) CreateTask(ctx context.Context, task *models.OnboardingTask) error {
	query := `
		INSERT INTO onboarding_tasks (
			id, workflow_id, title, description, category, priority,
			status, assigned_to, due_date, order_index, is_mandatory,
			estimated_hours, ai_generated, ai_suggestions
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING created_at, updated_at
	`
	
	return r.db.QueryRow(ctx, query,
		task.ID, task.WorkflowID, task.Title, task.Description,
		task.Category, task.Priority, task.Status, task.AssignedTo,
		task.DueDate, task.OrderIndex, task.IsMandatory,
		task.EstimatedHours, task.AIGenerated, task.AISuggestions,
	).Scan(&task.CreatedAt, &task.UpdatedAt)
}

func (r *onboardingRepository) GetTask(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error) {
	query := `
		SELECT 
			ot.id, ot.workflow_id, ot.title, ot.description,
			ot.category, ot.priority, ot.status, ot.assigned_to,
			ot.due_date, ot.completed_at, ot.completed_by,
			ot.order_index, ot.is_mandatory, ot.estimated_hours,
			ot.actual_hours, ot.ai_generated, ot.ai_suggestions,
			ot.created_at, ot.updated_at,
			COALESCE(assigned.first_name || ' ' || assigned.last_name, '') as assigned_to_name,
			COALESCE(completed.first_name || ' ' || completed.last_name, '') as completed_by_name
		FROM onboarding_tasks ot
		LEFT JOIN employees assigned ON assigned.id = ot.assigned_to
		LEFT JOIN employees completed ON completed.id = ot.completed_by
		WHERE ot.id = $1
	`
	
	task := &models.OnboardingTask{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.WorkflowID, &task.Title, &task.Description,
		&task.Category, &task.Priority, &task.Status, &task.AssignedTo,
		&task.DueDate, &task.CompletedAt, &task.CompletedBy,
		&task.OrderIndex, &task.IsMandatory, &task.EstimatedHours,
		&task.ActualHours, &task.AIGenerated, &task.AISuggestions,
		&task.CreatedAt, &task.UpdatedAt,
		&task.AssignedToName, &task.CompletedByName,
	)
	
	return task, err
}

func (r *onboardingRepository) GetTasksByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*models.OnboardingTask, error) {
	query := `
		SELECT 
			ot.id, ot.workflow_id, ot.title, ot.description,
			ot.category, ot.priority, ot.status, ot.assigned_to,
			ot.due_date, ot.completed_at, ot.completed_by,
			ot.order_index, ot.is_mandatory, ot.estimated_hours,
			ot.actual_hours, ot.ai_generated, ot.ai_suggestions,
			ot.created_at, ot.updated_at,
			COALESCE(assigned.first_name || ' ' || assigned.last_name, '') as assigned_to_name,
			COALESCE(completed.first_name || ' ' || completed.last_name, '') as completed_by_name
		FROM onboarding_tasks ot
		LEFT JOIN employees assigned ON assigned.id = ot.assigned_to
		LEFT JOIN employees completed ON completed.id = ot.completed_by
		WHERE ot.workflow_id = $1
		ORDER BY ot.order_index, ot.due_date NULLS LAST
	`
	
	rows, err := r.db.Query(ctx, query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tasks []*models.OnboardingTask
	for rows.Next() {
		task := &models.OnboardingTask{}
		err := rows.Scan(
			&task.ID, &task.WorkflowID, &task.Title, &task.Description,
			&task.Category, &task.Priority, &task.Status, &task.AssignedTo,
			&task.DueDate, &task.CompletedAt, &task.CompletedBy,
			&task.OrderIndex, &task.IsMandatory, &task.EstimatedHours,
			&task.ActualHours, &task.AIGenerated, &task.AISuggestions,
			&task.CreatedAt, &task.UpdatedAt,
			&task.AssignedToName, &task.CompletedByName,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	
	return tasks, rows.Err()
}

func (r *onboardingRepository) GetTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.OnboardingTask, error) {
	query := `
		SELECT 
			ot.id, ot.workflow_id, ot.title, ot.description,
			ot.category, ot.priority, ot.status, ot.assigned_to,
			ot.due_date, ot.completed_at, ot.completed_by,
			ot.order_index, ot.is_mandatory, ot.estimated_hours,
			ot.actual_hours, ot.ai_generated, ot.ai_suggestions,
			ot.created_at, ot.updated_at,
			COALESCE(assigned.first_name || ' ' || assigned.last_name, '') as assigned_to_name,
			COALESCE(completed.first_name || ' ' || completed.last_name, '') as completed_by_name
		FROM onboarding_tasks ot
		JOIN new_hire_onboardings nho ON nho.id = ot.workflow_id
		LEFT JOIN employees assigned ON assigned.id = ot.assigned_to
		LEFT JOIN employees completed ON completed.id = ot.completed_by
		WHERE nho.employee_id = $1
		ORDER BY ot.order_index, ot.due_date NULLS LAST
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
			&task.ID, &task.WorkflowID, &task.Title, &task.Description,
			&task.Category, &task.Priority, &task.Status, &task.AssignedTo,
			&task.DueDate, &task.CompletedAt, &task.CompletedBy,
			&task.OrderIndex, &task.IsMandatory, &task.EstimatedHours,
			&task.ActualHours, &task.AIGenerated, &task.AISuggestions,
			&task.CreatedAt, &task.UpdatedAt,
			&task.AssignedToName, &task.CompletedByName,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	
	return tasks, rows.Err()
}

func (r *onboardingRepository) UpdateTask(ctx context.Context, task *models.OnboardingTask) error {
	query := `
		UPDATE onboarding_tasks
		SET 
			title = $2,
			description = $3,
			category = $4,
			priority = $5,
			status = $6,
			assigned_to = $7,
			due_date = $8,
			actual_hours = $9,
			ai_suggestions = $10,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	
	return r.db.QueryRow(ctx, query,
		task.ID, task.Title, task.Description, task.Category,
		task.Priority, task.Status, task.AssignedTo, task.DueDate,
		task.ActualHours, task.AISuggestions,
	).Scan(&task.UpdatedAt)
}

func (r *onboardingRepository) CompleteTask(ctx context.Context, taskID, completedBy uuid.UUID, actualHours *float64) error {
	query := `
		UPDATE onboarding_tasks
		SET 
			status = 'completed',
			completed_at = NOW(),
			completed_by = $2,
			actual_hours = COALESCE($3, actual_hours),
			updated_at = NOW()
		WHERE id = $1
	`
	
	_, err := r.db.Exec(ctx, query, taskID, completedBy, actualHours)
	return err
}

func (r *onboardingRepository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM onboarding_tasks WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// ============================================================================
// INTERACTION OPERATIONS
// ============================================================================

func (r *onboardingRepository) CreateInteraction(ctx context.Context, interaction *models.OnboardingInteraction) error {
	query := `
		INSERT INTO onboarding_interactions (
			id, workflow_id, employee_id, interaction_type,
			message, ai_response, sentiment, requires_action, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at
	`
	
	return r.db.QueryRow(ctx, query,
		interaction.ID, interaction.WorkflowID, interaction.EmployeeID,
		interaction.InteractionType, interaction.Message, interaction.AIResponse,
		interaction.Sentiment, interaction.RequiresAction, interaction.Metadata,
	).Scan(&interaction.CreatedAt)
}

func (r *onboardingRepository) ListInteractionsByWorkflow(ctx context.Context, workflowID uuid.UUID, limit int) ([]*models.OnboardingInteraction, error) {
	query := `
		SELECT 
			oi.id, oi.workflow_id, oi.employee_id, oi.interaction_type,
			oi.message, oi.ai_response, oi.sentiment, oi.requires_action,
			oi.action_taken, oi.metadata, oi.created_at,
			e.first_name || ' ' || e.last_name as employee_name
		FROM onboarding_interactions oi
		JOIN employees e ON e.id = oi.employee_id
		WHERE oi.workflow_id = $1
		ORDER BY oi.created_at DESC
		LIMIT $2
	`
	
	rows, err := r.db.Query(ctx, query, workflowID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var interactions []*models.OnboardingInteraction
	for rows.Next() {
		i := &models.OnboardingInteraction{}
		err := rows.Scan(
			&i.ID, &i.WorkflowID, &i.EmployeeID, &i.InteractionType,
			&i.Message, &i.AIResponse, &i.Sentiment, &i.RequiresAction,
			&i.ActionTaken, &i.Metadata, &i.CreatedAt, &i.EmployeeName,
		)
		if err != nil {
			return nil, err
		}
		interactions = append(interactions, i)
	}
	
	return interactions, rows.Err()
}

// ============================================================================
// MILESTONE OPERATIONS
// ============================================================================

func (r *onboardingRepository) CreateMilestone(ctx context.Context, milestone *models.OnboardingMilestone) error {
	query := `
		INSERT INTO onboarding_milestones (
			id, workflow_id, name, description, target_date, status
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`
	
	return r.db.QueryRow(ctx, query,
		milestone.ID, milestone.WorkflowID, milestone.Name,
		milestone.Description, milestone.TargetDate, milestone.Status,
	).Scan(&milestone.CreatedAt)
}

func (r *onboardingRepository) GetMilestone(ctx context.Context, id uuid.UUID) (*models.OnboardingMilestone, error) {
	query := `
		SELECT 
			id, workflow_id, name, description, target_date,
			completed_date, status, celebration_sent, created_at
		FROM onboarding_milestones
		WHERE id = $1
	`
	
	milestone := &models.OnboardingMilestone{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&milestone.ID, &milestone.WorkflowID, &milestone.Name, &milestone.Description,
		&milestone.TargetDate, &milestone.CompletedDate, &milestone.Status,
		&milestone.CelebrationSent, &milestone.CreatedAt,
	)
	
	return milestone, err
}

func (r *onboardingRepository) ListMilestonesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*models.OnboardingMilestone, error) {
	query := `
		SELECT 
			id, workflow_id, name, description, target_date,
			completed_date, status, celebration_sent, created_at
		FROM onboarding_milestones
		WHERE workflow_id = $1
		ORDER BY target_date NULLS LAST
	`
	
	rows, err := r.db.Query(ctx, query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var milestones []*models.OnboardingMilestone
	for rows.Next() {
		m := &models.OnboardingMilestone{}
		err := rows.Scan(
			&m.ID, &m.WorkflowID, &m.Name, &m.Description,
			&m.TargetDate, &m.CompletedDate, &m.Status,
			&m.CelebrationSent, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		milestones = append(milestones, m)
	}
	
	return milestones, rows.Err()
}

func (r *onboardingRepository) UpdateMilestone(ctx context.Context, milestone *models.OnboardingMilestone) error {
	query := `
		UPDATE onboarding_milestones
		SET 
			completed_date = $2,
			status = $3,
			celebration_sent = $4
		WHERE id = $1
	`
	
	_, err := r.db.Exec(ctx, query,
		milestone.ID, milestone.CompletedDate,
		milestone.Status, milestone.CelebrationSent,
	)
	
	return err
}

// ============================================================================
// TEMPLATE OPERATIONS
// ============================================================================

func (r *onboardingRepository) GetTemplate(ctx context.Context, id uuid.UUID) (*models.OnboardingChecklistTemplate, error) {
	query := `
		SELECT id, name, description, department, role_type, is_active, created_at, updated_at
		FROM onboarding_checklist_templates
		WHERE id = $1
	`
	
	template := &models.OnboardingChecklistTemplate{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&template.ID, &template.Name, &template.Description,
		&template.Department, &template.RoleType, &template.IsActive,
		&template.CreatedAt, &template.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	// Load items
	itemsQuery := `
		SELECT 
			id, template_id, title, description, category,
			priority, day_offset, is_mandatory, estimated_hours,
			order_index, assigned_to_role, created_at
		FROM onboarding_checklist_template_items
		WHERE template_id = $1
		ORDER BY order_index
	`
	
	rows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return template, nil
	}
	defer rows.Close()
	
	var items []*models.OnboardingChecklistTemplateItem
	for rows.Next() {
		item := &models.OnboardingChecklistTemplateItem{}
		err := rows.Scan(
			&item.ID, &item.TemplateID, &item.Title, &item.Description,
			&item.Category, &item.Priority, &item.DayOffset, &item.IsMandatory,
			&item.EstimatedHours, &item.OrderIndex, &item.AssignedToRole,
			&item.CreatedAt,
		)
		if err != nil {
			continue
		}
		items = append(items, item)
	}
	
	template.Items = items
	return template, nil
}

func (r *onboardingRepository) ListTemplates(ctx context.Context, department, roleType string) ([]*models.OnboardingChecklistTemplate, error) {
	query := `
		SELECT id, name, description, department, role_type, is_active, created_at, updated_at
		FROM onboarding_checklist_templates
		WHERE is_active = true
	`
	
	args := []interface{}{}
	argCount := 1
	
	if department != "" {
		query += fmt.Sprintf(" AND (department = $%d OR department IS NULL)", argCount)
		args = append(args, department)
		argCount++
	}
	
	if roleType != "" {
		query += fmt.Sprintf(" AND (role_type = $%d OR role_type IS NULL)", argCount)
		args = append(args, roleType)
	}
	
	query += " ORDER BY name"
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var templates []*models.OnboardingChecklistTemplate
	for rows.Next() {
		t := &models.OnboardingChecklistTemplate{}
		err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.Department,
			&t.RoleType, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	
	return templates, rows.Err()
}

func (r *onboardingRepository) CreateTasksFromTemplate(ctx context.Context, workflowID uuid.UUID, templateID uuid.UUID, startDate time.Time) error {
	// This needs to be implemented - left as placeholder
	return fmt.Errorf("CreateTasksFromTemplate not yet implemented")
}

// ============================================================================
// STATISTICS AND DASHBOARD
// ============================================================================

func (r *onboardingRepository) GetStatistics(ctx context.Context, workflowID uuid.UUID) (*models.OnboardingStatistics, error) {
	// Placeholder implementation
	return &models.OnboardingStatistics{}, nil
}

func (r *onboardingRepository) GetDashboardData(ctx context.Context, filters map[string]interface{}) (*models.OnboardingDashboardResponse, error) {
	// Placeholder implementation
	return &models.OnboardingDashboardResponse{}, nil
}