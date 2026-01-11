package models

import (
	"time"
	"github.com/google/uuid"
)

// Applicant represents a job applicant
type Applicant struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Email       string    `json:"email" db:"email"`
	Phone       string    `json:"phone" db:"phone"`
	Position    string    `json:"position" db:"position"`
	Source      string    `json:"source" db:"source"`
	ResumeURL   string    `json:"resume_url" db:"resume_url"`
	AppliedDate time.Time `json:"applied_date" db:"applied_date"`
	Status      string    `json:"status" db:"status"`
	AIScore     float64   `json:"ai_score" db:"ai_score"`
	AIAnalysis  string    `json:"ai_analysis" db:"ai_analysis"`
	Notes       string    `json:"notes" db:"notes"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}