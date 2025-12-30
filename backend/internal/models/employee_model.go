package models

import (
	"time"

	"github.com/google/uuid"
)

// Employee represents an employee record
type Employee struct {
	ID                     uuid.UUID  `json:"id"`
	UserID                 *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	FirstName              string     `json:"first_name"`
	LastName               string     `json:"last_name"`
	Email                  string     `json:"email"`
	Phone                  string     `json:"phone"`
	DateOfBirth            *time.Time `json:"date_of_birth,omitempty"`
	HireDate               time.Time  `json:"hire_date"`
	Department             string     `json:"department"`
	Position               string     `json:"position"`
	ManagerID              *uuid.UUID `json:"manager_id,omitempty"`
	EmploymentType         string     `json:"employment_type"`
	Status                 string     `json:"status"`
	StreetAddress          string     `json:"street_address"`
	City                   string     `json:"city"`
	State                  string     `json:"state"`
	ZipCode                string     `json:"zip_code"`
	Country                string     `json:"country"`
	EmergencyContactName   string     `json:"emergency_contact_name"`
	EmergencyContactPhone  string     `json:"emergency_contact_phone"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}
