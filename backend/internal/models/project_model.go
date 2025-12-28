package models

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a company project
type Project struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Code        string     `json:"code" db:"code"`
	Description string     `json:"description"`
	ClientName  string     `json:"client_name,omitempty" db:"client_name"`
	Status      string     `json:"status"` // active, on-hold, completed, archived
	Priority    string     `json:"priority"` // low, medium, high, critical
	BudgetHours *float64   `json:"budget_hours,omitempty" db:"budget_hours"`
	ManagerID   *uuid.UUID `json:"manager_id,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Budget      *float64   `json:"budget,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ProjectMember represents an employee assigned to a project
type ProjectMember struct {
	ID         uuid.UUID `json:"id"`
	ProjectID  uuid.UUID `json:"project_id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	Role       string    `json:"role"` // lead, developer, designer, etc.
	CreatedAt  time.Time `json:"created_at"`
}

// ProjectWithDetails includes project info plus manager and member details
type ProjectWithDetails struct {
	Project
	ManagerName  string              `json:"manager_name,omitempty"`
	ManagerEmail string              `json:"manager_email,omitempty"`
	MemberCount  int                 `json:"member_count"`
	Members      []*ProjectMemberInfo `json:"members,omitempty"`
}

// ProjectMemberInfo includes employee details for project members
type ProjectMemberInfo struct {
	ProjectMember
	EmployeeName     string `json:"employee_name"`
	EmployeeEmail    string `json:"employee_email"`
	EmployeePosition string `json:"employee_position"`
}

// CreateProjectRequest for creating new projects
type CreateProjectRequest struct {
	Name        string     `json:"name" validate:"required"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	ManagerID   *uuid.UUID `json:"manager_id,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Budget      *float64   `json:"budget,omitempty"`
}

// UpdateProjectRequest for updating projects
type UpdateProjectRequest struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	Status      *string    `json:"status,omitempty"`
	Priority    *string    `json:"priority,omitempty"`
	ManagerID   *uuid.UUID `json:"manager_id,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Budget      *float64   `json:"budget,omitempty"`
}

// AssignProjectMemberRequest for assigning employees to projects
type AssignProjectMemberRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" validate:"required"`
	Role       string    `json:"role"`
}

// AssignManagerRequest for assigning employees to managers
type AssignManagerRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" validate:"required"`
	ManagerID  uuid.UUID `json:"manager_id" validate:"required"`
}