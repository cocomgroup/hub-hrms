package repository

import (
	"context"
	"fmt"
	"time"
	"hub-hrms/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrganizationRepository interface
type OrganizationRepository interface {
	Create(ctx context.Context, org *models.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.OrganizationWithDetails, error)
	GetByCode(ctx context.Context, code string) (*models.Organization, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*models.OrganizationWithDetails, error)
	Update(ctx context.Context, org *models.Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetHierarchy(ctx context.Context, rootID *uuid.UUID) (*models.OrganizationHierarchy, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*models.Organization, error)
	
	// Employee assignments
	AssignEmployee(ctx context.Context, assignment *models.OrganizationEmployee) error
	UnassignEmployee(ctx context.Context, orgID, employeeID uuid.UUID) error
	GetEmployees(ctx context.Context, orgID uuid.UUID) ([]*models.Employee, error)
	GetEmployeeOrganizations(ctx context.Context, employeeID uuid.UUID) ([]*models.Organization, error)
	BulkAssignEmployees(ctx context.Context, orgID uuid.UUID, employeeIDs []uuid.UUID, req *models.BulkAssignEmployeesRequest) error
	
	// Stats
	GetStats(ctx context.Context, orgID uuid.UUID) (*models.OrganizationStats, error)
}


type organizationRepository struct {
	db *pgxpool.Pool
}

func NewOrganizationRepository(db *pgxpool.Pool) OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) Create(ctx context.Context, org *models.Organization) error {
	query := `
		INSERT INTO organizations (
			id, name, code, description, parent_id, manager_id, type, level,
			cost_center, location, is_active, employee_count, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)`
	
	_, err := r.db.Exec(ctx, query,
		org.ID, org.Name, org.Code, org.Description, org.ParentID, org.ManagerID,
		org.Type, org.Level, org.CostCenter, org.Location, org.IsActive,
		org.EmployeeCount, org.CreatedBy, org.CreatedAt, org.UpdatedAt,
	)
	return err
}

func (r *organizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.OrganizationWithDetails, error) {
	query := `
		SELECT 
			o.id, o.name, o.code, o.description, o.parent_id, o.manager_id, o.type, o.level,
			o.cost_center, o.location, o.is_active, o.employee_count, o.created_by, o.created_at, o.updated_at,
			m.first_name || ' ' || m.last_name as manager_name,
			m.email as manager_email,
			p.name as parent_name
		FROM organizations o
		LEFT JOIN employees m ON o.manager_id = m.id
		LEFT JOIN organizations p ON o.parent_id = p.id
		WHERE o.id = $1`
	
	org := &models.OrganizationWithDetails{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.Name, &org.Code, &org.Description, &org.ParentID, &org.ManagerID,
		&org.Type, &org.Level, &org.CostCenter, &org.Location, &org.IsActive,
		&org.EmployeeCount, &org.CreatedBy, &org.CreatedAt, &org.UpdatedAt,
		&org.ManagerName, &org.ManagerEmail, &org.ParentName,
	)
	if err != nil {
		return nil, err
	}
	
	// Get employees
	employees, err := r.GetEmployees(ctx, id)
	if err == nil {
		for _, emp := range employees {
			org.Employees = append(org.Employees, *emp)
		}
	}
	
	return org, nil
}

func (r *organizationRepository) GetByCode(ctx context.Context, code string) (*models.Organization, error) {
	query := `
		SELECT id, name, code, description, parent_id, manager_id, type, level,
			   cost_center, location, is_active, employee_count, created_by, created_at, updated_at
		FROM organizations WHERE code = $1`
	
	org := &models.Organization{}
	err := r.db.QueryRow(ctx, query, code).Scan(
		&org.ID, &org.Name, &org.Code, &org.Description, &org.ParentID, &org.ManagerID,
		&org.Type, &org.Level, &org.CostCenter, &org.Location, &org.IsActive,
		&org.EmployeeCount, &org.CreatedBy, &org.CreatedAt, &org.UpdatedAt,
	)
	return org, err
}

func (r *organizationRepository) List(ctx context.Context, filters map[string]interface{}) ([]*models.OrganizationWithDetails, error) {
	query := `
		SELECT 
			o.id, o.name, o.code, 
			COALESCE(o.description, '') as description, 
			o.parent_id, o.manager_id, o.type, o.level,
			COALESCE(o.cost_center, '') as cost_center, 
			COALESCE(o.location, '') as location, 
			o.is_active, o.employee_count, o.created_by, o.created_at, o.updated_at,
			COALESCE(m.first_name || ' ' || m.last_name, '') as manager_name,
			COALESCE(m.email, '') as manager_email,
			COALESCE(p.name, '') as parent_name
		FROM organizations o
		LEFT JOIN employees m ON o.manager_id = m.id
		LEFT JOIN organizations p ON o.parent_id = p.id
		WHERE 1=1`
	
	args := []interface{}{}
	argCount := 1
	
	if parentID, ok := filters["parent_id"].(uuid.UUID); ok {
		query += fmt.Sprintf(" AND o.parent_id = $%d", argCount)
		args = append(args, parentID)
		argCount++
	}
	
	if isActive, ok := filters["is_active"].(bool); ok {
		query += fmt.Sprintf(" AND o.is_active = $%d", argCount)
		args = append(args, isActive)
		argCount++
	}
	
	if orgType, ok := filters["type"].(string); ok {
		query += fmt.Sprintf(" AND o.type = $%d", argCount)
		args = append(args, orgType)
		argCount++
	}
	
	query += " ORDER BY o.level, o.name"
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	orgs := []*models.OrganizationWithDetails{}
	for rows.Next() {
		org := &models.OrganizationWithDetails{}
		err := rows.Scan(
			&org.ID, &org.Name, &org.Code, &org.Description, &org.ParentID, &org.ManagerID,
			&org.Type, &org.Level, &org.CostCenter, &org.Location, &org.IsActive,
			&org.EmployeeCount, &org.CreatedBy, &org.CreatedAt, &org.UpdatedAt,
			&org.ManagerName, &org.ManagerEmail, &org.ParentName,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	
	return orgs, nil
}

func (r *organizationRepository) Update(ctx context.Context, org *models.Organization) error {
	query := `
		UPDATE organizations SET
			name = $2, description = $3, parent_id = $4, manager_id = $5,
			type = $6, level = $7, cost_center = $8, location = $9,
			is_active = $10, updated_at = $11
		WHERE id = $1`
	
	_, err := r.db.Exec(ctx, query,
		org.ID, org.Name, org.Description, org.ParentID, org.ManagerID,
		org.Type, org.Level, org.CostCenter, org.Location, org.IsActive, org.UpdatedAt,
	)
	return err
}

func (r *organizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM organizations WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *organizationRepository) GetHierarchy(ctx context.Context, rootID *uuid.UUID) (*models.OrganizationHierarchy, error) {
	// Recursive CTE to build the hierarchy
	query := `
		WITH RECURSIVE org_hierarchy AS (
			SELECT id, name, code, description, parent_id, manager_id, type, level,
				   cost_center, location, is_active, employee_count, created_by, created_at, updated_at
			FROM organizations
			WHERE ($1::uuid IS NULL AND parent_id IS NULL) OR id = $1
			UNION ALL
			SELECT o.id, o.name, o.code, o.description, o.parent_id, o.manager_id, o.type, o.level,
				   o.cost_center, o.location, o.is_active, o.employee_count, o.created_by, o.created_at, o.updated_at
			FROM organizations o
			INNER JOIN org_hierarchy oh ON o.parent_id = oh.id
		)
		SELECT * FROM org_hierarchy ORDER BY level, name`
	
	rows, err := r.db.Query(ctx, query, rootID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	orgs := make(map[uuid.UUID]*models.OrganizationHierarchy)
	var root *models.OrganizationHierarchy
	
	for rows.Next() {
		org := &models.OrganizationHierarchy{}
		err := rows.Scan(
			&org.ID, &org.Name, &org.Code, &org.Description, &org.ParentID, &org.ManagerID,
			&org.Type, &org.Level, &org.CostCenter, &org.Location, &org.IsActive,
			&org.EmployeeCount, &org.CreatedBy, &org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		orgs[org.ID] = org
		
		if org.ParentID == nil || (rootID != nil && org.ID == *rootID) {
			root = org
		} else if parent, ok := orgs[*org.ParentID]; ok {
			parent.Children = append(parent.Children, *org)
		}
	}
	
	return root, nil
}

func (r *organizationRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*models.Organization, error) {
	query := `
		SELECT id, name, code, description, parent_id, manager_id, type, level,
			   cost_center, location, is_active, employee_count, created_by, created_at, updated_at
		FROM organizations
		WHERE parent_id = $1
		ORDER BY name`
	
	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	orgs := []*models.Organization{}
	for rows.Next() {
		org := &models.Organization{}
		err := rows.Scan(
			&org.ID, &org.Name, &org.Code, &org.Description, &org.ParentID, &org.ManagerID,
			&org.Type, &org.Level, &org.CostCenter, &org.Location, &org.IsActive,
			&org.EmployeeCount, &org.CreatedBy, &org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	
	return orgs, nil
}

func (r *organizationRepository) AssignEmployee(ctx context.Context, assignment *models.OrganizationEmployee) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	
	// If this is a primary assignment, unset other primary assignments
	if assignment.IsPrimary {
		_, err = tx.Exec(ctx, `
			UPDATE organization_employees
			SET is_primary = false
			WHERE employee_id = $1 AND end_date IS NULL`,
			assignment.EmployeeID,
		)
		if err != nil {
			return err
		}
	}
	
	// Insert the new assignment
	query := `
		INSERT INTO organization_employees (
			id, organization_id, employee_id, role, is_primary, start_date, assigned_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	_, err = tx.Exec(ctx, query,
		assignment.ID, assignment.OrganizationID, assignment.EmployeeID, assignment.Role,
		assignment.IsPrimary, assignment.StartDate, assignment.AssignedBy, assignment.CreatedAt, assignment.UpdatedAt,
	)
	if err != nil {
		return err
	}
	
	// Update employee count
	_, err = tx.Exec(ctx, `
		UPDATE organizations
		SET employee_count = (
			SELECT COUNT(DISTINCT employee_id)
			FROM organization_employees
			WHERE organization_id = $1 AND end_date IS NULL
		)
		WHERE id = $1`,
		assignment.OrganizationID,
	)
	if err != nil {
		return err
	}
	
	return tx.Commit(ctx)
}

func (r *organizationRepository) UnassignEmployee(ctx context.Context, orgID, employeeID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	
	// Set end date on the assignment
	_, err = tx.Exec(ctx, `
		UPDATE organization_employees
		SET end_date = NOW()
		WHERE organization_id = $1 AND employee_id = $2 AND end_date IS NULL`,
		orgID, employeeID,
	)
	if err != nil {
		return err
	}
	
	// Update employee count
	_, err = tx.Exec(ctx, `
		UPDATE organizations
		SET employee_count = (
			SELECT COUNT(DISTINCT employee_id)
			FROM organization_employees
			WHERE organization_id = $1 AND end_date IS NULL
		)
		WHERE id = $1`,
		orgID,
	)
	if err != nil {
		return err
	}
	
	return tx.Commit(ctx)
}

func (r *organizationRepository) GetEmployees(ctx context.Context, orgID uuid.UUID) ([]*models.Employee, error) {
	query := `
		SELECT DISTINCT e.id, e.first_name, e.last_name, e.email, e.phone, e.hire_date,
			   e.department, e.position, e.employment_status, e.manager_id, e.created_at, e.updated_at
		FROM employees e
		INNER JOIN organization_employees oe ON e.id = oe.employee_id
		WHERE oe.organization_id = $1 AND oe.end_date IS NULL
		ORDER BY e.last_name, e.first_name`
	
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	employees := []*models.Employee{}
	for rows.Next() {
		emp := &models.Employee{}
		err := rows.Scan(
			&emp.ID, &emp.FirstName, &emp.LastName, &emp.Email, &emp.Phone, &emp.HireDate,
			&emp.Department, &emp.Position, &emp.Status, &emp.ManagerID,
			&emp.CreatedAt, &emp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}
	
	return employees, nil
}

func (r *organizationRepository) GetEmployeeOrganizations(ctx context.Context, employeeID uuid.UUID) ([]*models.Organization, error) {
	query := `
		SELECT DISTINCT o.id, o.name, o.code, o.description, o.parent_id, o.manager_id, o.type, o.level,
			   o.cost_center, o.location, o.is_active, o.employee_count, o.created_by, o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN organization_employees oe ON o.id = oe.organization_id
		WHERE oe.employee_id = $1 AND oe.end_date IS NULL
		ORDER BY oe.is_primary DESC, o.name`
	
	rows, err := r.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	orgs := []*models.Organization{}
	for rows.Next() {
		org := &models.Organization{}
		err := rows.Scan(
			&org.ID, &org.Name, &org.Code, &org.Description, &org.ParentID, &org.ManagerID,
			&org.Type, &org.Level, &org.CostCenter, &org.Location, &org.IsActive,
			&org.EmployeeCount, &org.CreatedBy, &org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	
	return orgs, nil
}

func (r *organizationRepository) BulkAssignEmployees(ctx context.Context, orgID uuid.UUID, employeeIDs []uuid.UUID, req *models.BulkAssignEmployeesRequest) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	
	for _, empID := range employeeIDs {
		assignment := &models.OrganizationEmployee{
			ID:             uuid.New(),
			OrganizationID: orgID,
			EmployeeID:     empID,
			Role:           req.Role,
			IsPrimary:      req.IsPrimary,
			StartDate:      req.StartDate,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		
		query := `
			INSERT INTO organization_employees (
				id, organization_id, employee_id, role, is_primary, start_date, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (organization_id, employee_id) 
			WHERE end_date IS NULL
			DO UPDATE SET role = $4, is_primary = $5, updated_at = $8`
		
		_, err = tx.Exec(ctx, query,
			assignment.ID, assignment.OrganizationID, assignment.EmployeeID, assignment.Role,
			assignment.IsPrimary, assignment.StartDate, assignment.CreatedAt, assignment.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}
	
	// Update employee count
	_, err = tx.Exec(ctx, `
		UPDATE organizations
		SET employee_count = (
			SELECT COUNT(DISTINCT employee_id)
			FROM organization_employees
			WHERE organization_id = $1 AND end_date IS NULL
		)
		WHERE id = $1`,
		orgID,
	)
	if err != nil {
		return err
	}
	
	return tx.Commit(ctx)
}

func (r *organizationRepository) GetStats(ctx context.Context, orgID uuid.UUID) (*models.OrganizationStats, error) {
	query := `
		SELECT 
			$1 as organization_id,
			COUNT(DISTINCT oe.employee_id) as total_employees,
			COUNT(DISTINCT CASE WHEN e.employment_status = 'active' THEN oe.employee_id END) as active_employees,
			COUNT(DISTINCT CASE WHEN o.type = 'department' AND o.parent_id = $1 THEN o.id END) as departments,
			COUNT(DISTINCT CASE WHEN o.type = 'team' AND o.parent_id = $1 THEN o.id END) as teams
		FROM organizations o
		LEFT JOIN organization_employees oe ON o.id = oe.organization_id AND oe.end_date IS NULL
		LEFT JOIN employees e ON oe.employee_id = e.id
		WHERE o.id = $1 OR o.parent_id = $1`
	
	stats := &models.OrganizationStats{}
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&stats.OrganizationID, &stats.TotalEmployees, &stats.ActiveEmployees,
		&stats.Departments, &stats.Teams,
	)
	return stats, err
}
