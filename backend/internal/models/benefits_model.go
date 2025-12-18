package models

import (
	"time"

	"github.com/google/uuid"
)

// BenefitCategory represents the category of benefit
type BenefitCategory string

const (
	BenefitCategoryHealth      BenefitCategory = "health"
	BenefitCategoryDental      BenefitCategory = "dental"
	BenefitCategoryVision      BenefitCategory = "vision"
	BenefitCategoryLife        BenefitCategory = "life"
	BenefitCategoryDisability  BenefitCategory = "disability"
	BenefitCategoryRetirement  BenefitCategory = "retirement"
	BenefitCategoryFSA         BenefitCategory = "fsa"
	BenefitCategoryHSA         BenefitCategory = "hsa"
	BenefitCategoryCommuter    BenefitCategory = "commuter"
	BenefitCategoryWellness    BenefitCategory = "wellness"
	BenefitCategoryOther       BenefitCategory = "other"
)

// BenefitPlanType represents the type of plan
type BenefitPlanType string

const (
	BenefitPlanTypeHMO         BenefitPlanType = "hmo"
	BenefitPlanTypePPO         BenefitPlanType = "ppo"
	BenefitPlanTypeEPO         BenefitPlanType = "epo"
	BenefitPlanTypePOS         BenefitPlanType = "pos"
	BenefitPlanTypeHDHP        BenefitPlanType = "hdhp"
	BenefitPlanTypeTraditional BenefitPlanType = "traditional"
)

// CoverageLevel represents the level of coverage
type CoverageLevel string

const (
	CoverageLevelEmployee       CoverageLevel = "employee"
	CoverageLevelEmployeeSpouse CoverageLevel = "employee_spouse"
	CoverageLevelEmployeeChild  CoverageLevel = "employee_child"
	CoverageLevelFamily         CoverageLevel = "family"
)

// EnrollmentStatus represents the status of an enrollment
type EnrollmentStatus string

const (
	EnrollmentStatusActive    EnrollmentStatus = "active"
	EnrollmentStatusPending   EnrollmentStatus = "pending"
	EnrollmentStatusCancelled EnrollmentStatus = "cancelled"
	EnrollmentStatusExpired   EnrollmentStatus = "expired"
)

// BenefitPlan represents a benefit plan offered by the company
type BenefitPlan struct {
	ID                  uuid.UUID       `json:"id" db:"id"`
	Name                string          `json:"name" db:"name"`
	Category            BenefitCategory `json:"category" db:"category"`
	PlanType            BenefitPlanType `json:"plan_type" db:"plan_type"`
	Provider            string          `json:"provider" db:"provider"`
	Description         string          `json:"description" db:"description"`
	EmployeeCost        float64         `json:"employee_cost" db:"employee_cost"`
	EmployerCost        float64         `json:"employer_cost" db:"employer_cost"`
	DeductibleSingle    float64         `json:"deductible_single" db:"deductible_single"`
	DeductibleFamily    float64         `json:"deductible_family" db:"deductible_family"`
	OutOfPocketMaxSingle float64        `json:"out_of_pocket_max_single" db:"out_of_pocket_max_single"`
	OutOfPocketMaxFamily float64        `json:"out_of_pocket_max_family" db:"out_of_pocket_max_family"`
	CopayPrimaryCare    float64         `json:"copay_primary_care" db:"copay_primary_care"`
	CopaySpecialist     float64         `json:"copay_specialist" db:"copay_specialist"`
	CopayEmergency      float64         `json:"copay_emergency" db:"copay_emergency"`
	CoinsuranceRate     float64         `json:"coinsurance_rate" db:"coinsurance_rate"`
	Active              bool            `json:"active" db:"active"`
	EnrollmentStartDate time.Time       `json:"enrollment_start_date" db:"enrollment_start_date"`
	EnrollmentEndDate   time.Time       `json:"enrollment_end_date" db:"enrollment_end_date"`
	EffectiveDate       time.Time       `json:"effective_date" db:"effective_date"`
	TerminationDate     *time.Time      `json:"termination_date" db:"termination_date"`
	Documents           []string        `json:"documents" db:"documents"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// BenefitEnrollment represents an employee's enrollment in a benefit plan
type BenefitEnrollment struct {
	ID                uuid.UUID        `json:"id" db:"id"`
	EmployeeID        uuid.UUID        `json:"employee_id" db:"employee_id"`
	EmployeeName      string           `json:"employee_name,omitempty" db:"employee_name"`
	PlanID            uuid.UUID        `json:"plan_id" db:"plan_id"`
	PlanName          string           `json:"plan_name,omitempty" db:"plan_name"`
	PlanCategory      BenefitCategory  `json:"plan_category,omitempty" db:"plan_category"`
	CoverageLevel     CoverageLevel    `json:"coverage_level" db:"coverage_level"`
	Status            EnrollmentStatus `json:"status" db:"status"`
	EnrollmentDate    time.Time        `json:"enrollment_date" db:"enrollment_date"`
	EffectiveDate     time.Time        `json:"effective_date" db:"effective_date"`
	TerminationDate   *time.Time       `json:"termination_date" db:"termination_date"`
	EmployeeCost      float64          `json:"employee_cost" db:"employee_cost"`
	EmployerCost      float64          `json:"employer_cost" db:"employer_cost"`
	TotalCost         float64          `json:"total_cost" db:"total_cost"`
	PayrollDeduction  float64          `json:"payroll_deduction" db:"payroll_deduction"`
	Dependents        []Dependent      `json:"dependents,omitempty"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" db:"updated_at"`
}

// Dependent represents a dependent for benefits coverage
type Dependent struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	EnrollmentID uuid.UUID  `json:"enrollment_id" db:"enrollment_id"`
	FirstName    string     `json:"first_name" db:"first_name"`
	LastName     string     `json:"last_name" db:"last_name"`
	Relationship string     `json:"relationship" db:"relationship"`
	DateOfBirth  time.Time  `json:"date_of_birth" db:"date_of_birth"`
	SSN          string     `json:"ssn,omitempty" db:"ssn"`
	Active       bool       `json:"active" db:"active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// BenefitClaim represents a benefit claim
type BenefitClaim struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	EnrollmentID    uuid.UUID    `json:"enrollment_id" db:"enrollment_id"`
	EmployeeID      uuid.UUID    `json:"employee_id" db:"employee_id"`
	EmployeeName    string       `json:"employee_name,omitempty" db:"employee_name"`
	PlanID          uuid.UUID    `json:"plan_id" db:"plan_id"`
	PlanName        string       `json:"plan_name,omitempty" db:"plan_name"`
	ClaimNumber     string       `json:"claim_number" db:"claim_number"`
	ServiceDate     time.Time    `json:"service_date" db:"service_date"`
	Provider        string       `json:"provider" db:"provider"`
	ServiceType     string       `json:"service_type" db:"service_type"`
	ClaimAmount     float64      `json:"claim_amount" db:"claim_amount"`
	ApprovedAmount  float64      `json:"approved_amount" db:"approved_amount"`
	PaidAmount      float64      `json:"paid_amount" db:"paid_amount"`
	EmployeePortion float64      `json:"employee_portion" db:"employee_portion"`
	Status          string       `json:"status" db:"status"`
	SubmittedDate   time.Time    `json:"submitted_date" db:"submitted_date"`
	ProcessedDate   *time.Time   `json:"processed_date" db:"processed_date"`
	Notes           string       `json:"notes" db:"notes"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at" db:"updated_at"`
}

// BenefitDocument represents a benefit-related document
type BenefitDocument struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	PlanID       *uuid.UUID      `json:"plan_id" db:"plan_id"`
	EnrollmentID *uuid.UUID      `json:"enrollment_id" db:"enrollment_id"`
	Category     BenefitCategory `json:"category" db:"category"`
	Name         string          `json:"name" db:"name"`
	Description  string          `json:"description" db:"description"`
	FileURL      string          `json:"file_url" db:"file_url"`
	FileType     string          `json:"file_type" db:"file_type"`
	FileSize     int64           `json:"file_size" db:"file_size"`
	UploadedBy   uuid.UUID       `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// CreateBenefitPlanInput represents input for creating a benefit plan
type CreateBenefitPlanInput struct {
	Name                 string          `json:"name" binding:"required"`
	Category             BenefitCategory `json:"category" binding:"required"`
	PlanType             BenefitPlanType `json:"plan_type"`
	Provider             string          `json:"provider" binding:"required"`
	Description          string          `json:"description"`
	EmployeeCost         float64         `json:"employee_cost" binding:"min=0"`
	EmployerCost         float64         `json:"employer_cost" binding:"min=0"`
	DeductibleSingle     float64         `json:"deductible_single" binding:"min=0"`
	DeductibleFamily     float64         `json:"deductible_family" binding:"min=0"`
	OutOfPocketMaxSingle float64         `json:"out_of_pocket_max_single" binding:"min=0"`
	OutOfPocketMaxFamily float64         `json:"out_of_pocket_max_family" binding:"min=0"`
	CopayPrimaryCare     float64         `json:"copay_primary_care" binding:"min=0"`
	CopaySpecialist      float64         `json:"copay_specialist" binding:"min=0"`
	CopayEmergency       float64         `json:"copay_emergency" binding:"min=0"`
	CoinsuranceRate      float64         `json:"coinsurance_rate" binding:"min=0,max=100"`
	EnrollmentStartDate  string          `json:"enrollment_start_date" binding:"required"`
	EnrollmentEndDate    string          `json:"enrollment_end_date" binding:"required"`
	EffectiveDate        string          `json:"effective_date" binding:"required"`
}

// UpdateBenefitPlanInput represents input for updating a benefit plan
type UpdateBenefitPlanInput struct {
	Name                 string          `json:"name"`
	Category             BenefitCategory `json:"category"`
	PlanType             BenefitPlanType `json:"plan_type"`
	Provider             string          `json:"provider"`
	Description          string          `json:"description"`
	EmployeeCost         *float64        `json:"employee_cost"`
	EmployerCost         *float64        `json:"employer_cost"`
	DeductibleSingle     *float64        `json:"deductible_single"`
	DeductibleFamily     *float64        `json:"deductible_family"`
	OutOfPocketMaxSingle *float64        `json:"out_of_pocket_max_single"`
	OutOfPocketMaxFamily *float64        `json:"out_of_pocket_max_family"`
	CopayPrimaryCare     *float64        `json:"copay_primary_care"`
	CopaySpecialist      *float64        `json:"copay_specialist"`
	CopayEmergency       *float64        `json:"copay_emergency"`
	CoinsuranceRate      *float64        `json:"coinsurance_rate"`
	Active               *bool           `json:"active"`
}

// CreateEnrollmentInput represents input for enrolling in a benefit
type CreateEnrollmentInput struct {
	PlanID        uuid.UUID     `json:"plan_id" binding:"required"`
	CoverageLevel CoverageLevel `json:"coverage_level" binding:"required"`
	EffectiveDate string        `json:"effective_date" binding:"required"`
	Dependents    []DependentInput `json:"dependents"`
}

// DependentInput represents input for a dependent
type DependentInput struct {
	FirstName    string `json:"first_name" binding:"required"`
	LastName     string `json:"last_name" binding:"required"`
	Relationship string `json:"relationship" binding:"required"`
	DateOfBirth  string `json:"date_of_birth" binding:"required"`
	SSN          string `json:"ssn"`
}

// BenefitsSummary represents a summary of employee benefits
type BenefitsSummary struct {
	EmployeeID         uuid.UUID            `json:"employee_id"`
	EmployeeName       string               `json:"employee_name"`
	ActiveEnrollments  int                  `json:"active_enrollments"`
	TotalEmployeeCost  float64              `json:"total_employee_cost"`
	TotalEmployerCost  float64              `json:"total_employer_cost"`
	MonthlyDeduction   float64              `json:"monthly_deduction"`
	Enrollments        []BenefitEnrollment  `json:"enrollments"`
	AvailablePlans     []BenefitPlan        `json:"available_plans"`
}

// EnrollmentPeriod represents an open enrollment period
type EnrollmentPeriod struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	StartDate   time.Time  `json:"start_date" db:"start_date"`
	EndDate     time.Time  `json:"end_date" db:"end_date"`
	PlanYear    int        `json:"plan_year" db:"plan_year"`
	Active      bool       `json:"active" db:"active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type EnrollmentCreate struct {
	PlanID         uuid.UUID              `json:"plan_id"`
	EffectiveDate  time.Time              `json:"effective_date"`
	Dependents     map[string]interface{} `json:"dependents,omitempty"`
}
