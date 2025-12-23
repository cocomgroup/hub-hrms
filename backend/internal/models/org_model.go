package models

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents a business unit or department in the company
type Organization struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Name            string     `json:"name" db:"name"`
	Code            string     `json:"code" db:"code"` // Unique org code (e.g., "FIN", "IT", "LEGAL")
	Description     string     `json:"description" db:"description"`
	ParentID        *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"` // For hierarchical structure
	ManagerID       *uuid.UUID `json:"manager_id,omitempty" db:"manager_id"` // Organization manager
	Type            string     `json:"type" db:"type"` // e.g., "division", "department", "team"
	Level           int        `json:"level" db:"level"` // Hierarchy level (0 = top level)
	CostCenter      string     `json:"cost_center,omitempty" db:"cost_center"`
	Location        string     `json:"location,omitempty" db:"location"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	EmployeeCount   int        `json:"employee_count" db:"employee_count"`
	CreatedBy       uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// OrganizationEmployee represents the assignment of an employee to an organization
type OrganizationEmployee struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	EmployeeID     uuid.UUID  `json:"employee_id" db:"employee_id"`
	Role           string     `json:"role,omitempty" db:"role"` // Role within the org
	IsPrimary      bool       `json:"is_primary" db:"is_primary"` // Primary org assignment
	StartDate      time.Time  `json:"start_date" db:"start_date"`
	EndDate        *time.Time `json:"end_date,omitempty" db:"end_date"`
	AssignedBy     uuid.UUID  `json:"assigned_by" db:"assigned_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// OrganizationHierarchy represents the complete organizational tree
type OrganizationHierarchy struct {
	Organization
	Parent   *OrganizationHierarchy   `json:"parent,omitempty"`
	Children []OrganizationHierarchy  `json:"children,omitempty"`
	Manager  *Employee                `json:"manager,omitempty"`
}

// OrganizationWithDetails includes full organization information
type OrganizationWithDetails struct {
	Organization
	ManagerName    *string    `json:"manager_name,omitempty"`
	ManagerEmail   *string    `json:"manager_email,omitempty"`
	ParentName     *string    `json:"parent_name,omitempty"`
	Employees      []Employee `json:"employees,omitempty"`
}

// CreateOrganizationRequest is the request to create a new organization
type CreateOrganizationRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=255"`
	Code        string     `json:"code" validate:"required,min=2,max=20,alphanum"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	ManagerID   *uuid.UUID `json:"manager_id,omitempty"`
	Type        string     `json:"type" validate:"required,oneof=division department team unit group"`
	CostCenter  string     `json:"cost_center,omitempty"`
	Location    string     `json:"location,omitempty"`
}

// UpdateOrganizationRequest is the request to update an organization
type UpdateOrganizationRequest struct {
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	ManagerID   *uuid.UUID `json:"manager_id,omitempty"`
	Type        *string    `json:"type,omitempty" validate:"omitempty,oneof=division department team unit group"`
	CostCenter  *string    `json:"cost_center,omitempty"`
	Location    *string    `json:"location,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
}

// AssignEmployeeRequest is the request to assign an employee to an organization
type AssignEmployeeRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" validate:"required"`
	Role       string    `json:"role,omitempty"`
	IsPrimary  bool      `json:"is_primary"`
	StartDate  time.Time `json:"start_date" validate:"required"`
}

// BulkAssignEmployeesRequest is the request to assign multiple employees
type BulkAssignEmployeesRequest struct {
	EmployeeIDs []uuid.UUID `json:"employee_ids" validate:"required,min=1"`
	Role        string      `json:"role,omitempty"`
	IsPrimary   bool        `json:"is_primary"`
	StartDate   time.Time   `json:"start_date" validate:"required"`
}

// OrganizationStats represents statistics about an organization
type OrganizationStats struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	TotalEmployees int       `json:"total_employees"`
	ActiveEmployees int      `json:"active_employees"`
	Departments    int       `json:"departments"`
	Teams          int       `json:"teams"`
}
