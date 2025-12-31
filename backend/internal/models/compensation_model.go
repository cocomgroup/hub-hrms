package models

import (
	"time"

	"github.com/google/uuid"
)

// CompensationPlan represents an employee's compensation plan
type CompensationPlan struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	EmployeeID       uuid.UUID  `json:"employee_id" db:"employee_id"`
	EmployeeName     string     `json:"employee_name,omitempty" db:"employee_name"`
	CompensationType string     `json:"compensation_type" db:"compensation_type"` // salary, hourly, contract
	BaseAmount       float64    `json:"base_amount" db:"base_amount"`
	Currency         string     `json:"currency" db:"currency"`
	PayFrequency     string     `json:"pay_frequency" db:"pay_frequency"` // hourly, weekly, biweekly, monthly, annually
	EffectiveDate    time.Time  `json:"effective_date" db:"effective_date"`
	EndDate          *time.Time `json:"end_date,omitempty" db:"end_date"`
	Status           string     `json:"status" db:"status"` // active, pending, expired
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// Bonus represents an employee bonus
type Bonus struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	EmployeeID   uuid.UUID  `json:"employee_id" db:"employee_id"`
	EmployeeName string     `json:"employee_name,omitempty" db:"employee_name"`
	BonusType    string     `json:"bonus_type" db:"bonus_type"` // monthly, quarterly, annual, performance, signing, retention
	Amount       float64    `json:"amount" db:"amount"`
	Currency     string     `json:"currency" db:"currency"`
	Description  string     `json:"description" db:"description"`
	PaymentDate  time.Time  `json:"payment_date" db:"payment_date"`
	Status       string     `json:"status" db:"status"` // pending, approved, paid, cancelled
	ApprovedBy   *uuid.UUID `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty" db:"approved_at"`
	PaidAt       *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateCompensationPlanRequest represents the request to create a compensation plan
type CreateCompensationPlanRequest struct {
	EmployeeID       uuid.UUID  `json:"employee_id" binding:"required"`
	CompensationType string     `json:"compensation_type" binding:"required"`
	BaseAmount       float64    `json:"base_amount" binding:"required,gt=0"`
	Currency         string     `json:"currency"`
	PayFrequency     string     `json:"pay_frequency" binding:"required"`
	EffectiveDate    time.Time  `json:"effective_date" binding:"required"`
	EndDate          *time.Time `json:"end_date,omitempty"`
	Status           string     `json:"status"`
}

// UpdateCompensationPlanRequest represents the request to update a compensation plan
type UpdateCompensationPlanRequest struct {
	CompensationType string     `json:"compensation_type"`
	BaseAmount       float64    `json:"base_amount" binding:"gt=0"`
	Currency         string     `json:"currency"`
	PayFrequency     string     `json:"pay_frequency"`
	EffectiveDate    time.Time  `json:"effective_date"`
	EndDate          *time.Time `json:"end_date,omitempty"`
	Status           string     `json:"status"`
}

// CreateBonusRequest represents the request to create a bonus
type CreateBonusRequest struct {
	EmployeeID  uuid.UUID `json:"employee_id" binding:"required"`
	BonusType   string    `json:"bonus_type" binding:"required"`
	Amount      float64   `json:"amount" binding:"required,gt=0"`
	Currency    string    `json:"currency"`
	Description string    `json:"description" binding:"required"`
	PaymentDate time.Time `json:"payment_date" binding:"required"`
	Status      string    `json:"status"`
}

// UpdateBonusRequest represents the request to update a bonus
type UpdateBonusRequest struct {
	BonusType   string    `json:"bonus_type"`
	Amount      float64   `json:"amount" binding:"gt=0"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	PaymentDate time.Time `json:"payment_date"`
	Status      string    `json:"status"`
}