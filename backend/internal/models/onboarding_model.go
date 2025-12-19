package models

import (
	"time"

	"github.com/google/uuid"
)

// OnboardingTask represents a task in the onboarding process
type OnboardingTask struct {
	ID                uuid.UUID  `json:"id"`
	EmployeeID        uuid.UUID  `json:"employee_id"`
	TaskName          string     `json:"task_name"`
	Description       *string    `json:"description,omitempty"`
	Category          *string    `json:"category,omitempty"`
	Status            string     `json:"status"`
	DueDate           *time.Time `json:"due_date,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	AssignedTo        *uuid.UUID `json:"assigned_to,omitempty"`
	DocumentsRequired bool       `json:"documents_required"`
	DocumentURL       *string    `json:"document_url,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}