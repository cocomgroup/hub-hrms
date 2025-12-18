package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents an authenticated user
type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	EmployeeID   *uuid.UUID `json:"employee_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Employee represents an employee record
type Employee struct {
	ID                     uuid.UUID  `json:"id"`
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


// Request/Response DTOs

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	User      User      `json:"user"`
	Employee  *Employee `json:"employee,omitempty"`
}


