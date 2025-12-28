package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"hub-hrms/backend/internal/models"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *models.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error)
	List(ctx context.Context, status string, managerID *uuid.UUID) ([]*models.Project, error)
	Update(ctx context.Context, project *models.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetWithDetails(ctx context.Context, id uuid.UUID) (*models.ProjectWithDetails, error)
	ListWithDetails(ctx context.Context, status string, managerID *uuid.UUID) ([]*models.ProjectWithDetails, error)
	
	// Project members
	AddMember(ctx context.Context, member *models.ProjectMember) error
	RemoveMember(ctx context.Context, projectID, employeeID uuid.UUID) error
	GetMembers(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMemberInfo, error)
	GetEmployeeProjects(ctx context.Context, employeeID uuid.UUID) ([]*models.Project, error)
}

type projectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, project *models.Project) error {
	query := `
		INSERT INTO projects (name, description, status, priority, manager_id, start_date, end_date, budget, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		project.Name, project.Description, project.Status, project.Priority,
		project.ManagerID, project.StartDate, project.EndDate, project.Budget, project.CreatedBy,
	).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)
}

func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	project := &models.Project{}
	query := `
		SELECT id, name, description, status, priority, manager_id, start_date, end_date, budget, created_by, created_at, updated_at
		FROM projects WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&project.ID, &project.Name, &project.Description, &project.Status, &project.Priority,
		&project.ManagerID, &project.StartDate, &project.EndDate, &project.Budget,
		&project.CreatedBy, &project.CreatedAt, &project.UpdatedAt,
	)
	return project, err
}

func (r *projectRepository) List(ctx context.Context, status string, managerID *uuid.UUID) ([]*models.Project, error) {
	query := `
		SELECT id, name, description, status, priority, manager_id, start_date, end_date, budget, created_by, created_at, updated_at
		FROM projects
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if status != "" && status != "all" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	if managerID != nil {
		query += fmt.Sprintf(" AND manager_id = $%d", argPos)
		args = append(args, *managerID)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []*models.Project{}
	for rows.Next() {
		project := &models.Project{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.Status, &project.Priority,
			&project.ManagerID, &project.StartDate, &project.EndDate, &project.Budget,
			&project.CreatedBy, &project.CreatedAt, &project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

func (r *projectRepository) Update(ctx context.Context, project *models.Project) error {
	query := `
		UPDATE projects
		SET name = $1, description = $2, status = $3, priority = $4, manager_id = $5,
		    start_date = $6, end_date = $7, budget = $8, updated_at = NOW()
		WHERE id = $9
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		project.Name, project.Description, project.Status, project.Priority, project.ManagerID,
		project.StartDate, project.EndDate, project.Budget, project.ID,
	).Scan(&project.UpdatedAt)
}

func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *projectRepository) GetWithDetails(ctx context.Context, id uuid.UUID) (*models.ProjectWithDetails, error) {
	project := &models.ProjectWithDetails{}
	query := `
		SELECT 
			p.id, p.name, p.description, p.status, p.priority, p.manager_id, 
			p.start_date, p.end_date, p.budget, p.created_by, p.created_at, p.updated_at,
			COALESCE(e.first_name || ' ' || e.last_name, '') as manager_name,
			COALESCE(u.email, '') as manager_email,
			(SELECT COUNT(*) FROM project_members WHERE project_id = p.id) as member_count
		FROM projects p
		LEFT JOIN employees e ON p.manager_id = e.id
		LEFT JOIN users u ON e.id = u.employee_id
		WHERE p.id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&project.ID, &project.Name, &project.Description, &project.Status, &project.Priority,
		&project.ManagerID, &project.StartDate, &project.EndDate, &project.Budget,
		&project.CreatedBy, &project.CreatedAt, &project.UpdatedAt,
		&project.ManagerName, &project.ManagerEmail, &project.MemberCount,
	)
	
	if err != nil {
		return nil, err
	}

	// Load members
	members, err := r.GetMembers(ctx, id)
	if err == nil {
		project.Members = members
	}

	return project, nil
}

func (r *projectRepository) ListWithDetails(ctx context.Context, status string, managerID *uuid.UUID) ([]*models.ProjectWithDetails, error) {
	query := `
		SELECT 
			p.id, p.name, p.description, p.status, p.priority, p.manager_id, 
			p.start_date, p.end_date, p.budget, p.created_by, p.created_at, p.updated_at,
			COALESCE(e.first_name || ' ' || e.last_name, '') as manager_name,
			COALESCE(u.email, '') as manager_email,
			(SELECT COUNT(*) FROM project_members WHERE project_id = p.id) as member_count
		FROM projects p
		LEFT JOIN employees e ON p.manager_id = e.id
		LEFT JOIN users u ON e.id = u.employee_id
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if status != "" && status != "all" {
		query += fmt.Sprintf(" AND p.status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	if managerID != nil {
		query += fmt.Sprintf(" AND p.manager_id = $%d", argPos)
		args = append(args, *managerID)
		argPos++
	}

	query += " ORDER BY p.created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []*models.ProjectWithDetails{}
	for rows.Next() {
		project := &models.ProjectWithDetails{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.Status, &project.Priority,
			&project.ManagerID, &project.StartDate, &project.EndDate, &project.Budget,
			&project.CreatedBy, &project.CreatedAt, &project.UpdatedAt,
			&project.ManagerName, &project.ManagerEmail, &project.MemberCount,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

func (r *projectRepository) AddMember(ctx context.Context, member *models.ProjectMember) error {
	query := `
		INSERT INTO project_members (project_id, employee_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, member.ProjectID, member.EmployeeID, member.Role).
		Scan(&member.ID, &member.CreatedAt)
}

func (r *projectRepository) RemoveMember(ctx context.Context, projectID, employeeID uuid.UUID) error {
	query := `DELETE FROM project_members WHERE project_id = $1 AND employee_id = $2`
	_, err := r.db.Exec(ctx, query, projectID, employeeID)
	return err
}

func (r *projectRepository) GetMembers(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMemberInfo, error) {
	query := `
		SELECT 
			pm.id, pm.project_id, pm.employee_id, pm.role, pm.created_at,
			e.first_name || ' ' || e.last_name as employee_name,
			u.email as employee_email,
			e.position as employee_position
		FROM project_members pm
		JOIN employees e ON pm.employee_id = e.id
		LEFT JOIN users u ON e.id = u.employee_id
		WHERE pm.project_id = $1
		ORDER BY pm.created_at
	`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []*models.ProjectMemberInfo{}
	for rows.Next() {
		member := &models.ProjectMemberInfo{}
		err := rows.Scan(
			&member.ID, &member.ProjectID, &member.EmployeeID, &member.Role, &member.CreatedAt,
			&member.EmployeeName, &member.EmployeeEmail, &member.EmployeePosition,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

func (r *projectRepository) GetEmployeeProjects(ctx context.Context, employeeID uuid.UUID) ([]*models.Project, error) {
	query := `
		SELECT p.id, p.name, p.description, p.status, p.priority, p.manager_id, 
		       p.start_date, p.end_date, p.budget, p.created_by, p.created_at, p.updated_at
		FROM projects p
		JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.employee_id = $1
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []*models.Project{}
	for rows.Next() {
		project := &models.Project{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.Status, &project.Priority,
			&project.ManagerID, &project.StartDate, &project.EndDate, &project.Budget,
			&project.CreatedBy, &project.CreatedAt, &project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}