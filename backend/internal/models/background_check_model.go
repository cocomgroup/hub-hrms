package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// BackgroundCheckStatus represents the current status of a background check
type BackgroundCheckStatus string

const (
	BGCheckStatusPending    BackgroundCheckStatus = "pending"
	BGCheckStatusInProgress BackgroundCheckStatus = "in_progress"
	BGCheckStatusCompleted  BackgroundCheckStatus = "completed"
	BGCheckStatusFailed     BackgroundCheckStatus = "failed"
	BGCheckStatusCancelled  BackgroundCheckStatus = "cancelled"
)

// BackgroundCheckType represents different types of checks
type BackgroundCheckType string

const (
	CheckTypeCriminal   BackgroundCheckType = "criminal"
	CheckTypeEmployment BackgroundCheckType = "employment"
	CheckTypeEducation  BackgroundCheckType = "education"
	CheckTypeCredit     BackgroundCheckType = "credit"
	CheckTypeDrugScreen BackgroundCheckType = "drug_screen"
	CheckTypeReference  BackgroundCheckType = "reference"
	CheckTypeIdentity   BackgroundCheckType = "identity"
)

// BackgroundCheckResult represents the outcome of a check
type BackgroundCheckResult string

const (
	ResultClear     BackgroundCheckResult = "clear"
	ResultConsider  BackgroundCheckResult = "consider"
	ResultSuspended BackgroundCheckResult = "suspended"
	ResultDispute   BackgroundCheckResult = "dispute"
)

// BackgroundCheck represents a background check record
type BackgroundCheck struct {
	ID             string                `json:"id" dynamodbav:"ID"`
	EmployeeID     string                `json:"employee_id" dynamodbav:"EmployeeID"`
	PackageID      string                `json:"package_id" dynamodbav:"PackageID"`
	ProviderID     string                `json:"provider_id" dynamodbav:"ProviderID"`
	Status         BackgroundCheckStatus `json:"status" dynamodbav:"Status"`
	Result         BackgroundCheckResult `json:"result,omitempty" dynamodbav:"Result,omitempty"`
	CheckTypes     []BackgroundCheckType `json:"check_types" dynamodbav:"CheckTypes"`
	CandidateInfo  CandidateInfo         `json:"candidate_info" dynamodbav:"CandidateInfo"`
	ReportURL      string                `json:"report_url,omitempty" dynamodbav:"ReportURL,omitempty"`
	ProviderData   json.RawMessage       `json:"provider_data,omitempty" dynamodbav:"ProviderData,omitempty"`
	InitiatedBy    string                `json:"initiated_by" dynamodbav:"InitiatedBy"`
	InitiatedAt    time.Time             `json:"initiated_at" dynamodbav:"InitiatedAt"`
	CompletedAt    *time.Time            `json:"completed_at,omitempty" dynamodbav:"CompletedAt,omitempty"`
	EstimatedETA   *time.Time            `json:"estimated_eta,omitempty" dynamodbav:"EstimatedETA,omitempty"`
	Notes          string                `json:"notes,omitempty" dynamodbav:"Notes,omitempty"`
	ComplianceData ComplianceData        `json:"compliance_data" dynamodbav:"ComplianceData"`
	CreatedAt      time.Time             `json:"created_at" dynamodbav:"CreatedAt"`
	UpdatedAt      time.Time             `json:"updated_at" dynamodbav:"UpdatedAt"`
}

// CandidateInfo contains information about the candidate being checked
type CandidateInfo struct {
	FirstName     string    `json:"first_name" dynamodbav:"FirstName"`
	MiddleName    string    `json:"middle_name,omitempty" dynamodbav:"MiddleName,omitempty"`
	LastName      string    `json:"last_name" dynamodbav:"LastName"`
	Email         string    `json:"email" dynamodbav:"Email"`
	Phone         string    `json:"phone" dynamodbav:"Phone"`
	DateOfBirth   time.Time `json:"date_of_birth" dynamodbav:"DateOfBirth"`
	SSN           string    `json:"ssn,omitempty" dynamodbav:"SSN,omitempty"` // Encrypted in storage
	DriverLicense string    `json:"driver_license,omitempty" dynamodbav:"DriverLicense,omitempty"`
	Address       Address   `json:"address" dynamodbav:"Address"`
}

// Address represents a physical address
type Address struct {
	Street1    string `json:"street1" dynamodbav:"Street1"`
	Street2    string `json:"street2,omitempty" dynamodbav:"Street2,omitempty"`
	City       string `json:"city" dynamodbav:"City"`
	State      string `json:"state" dynamodbav:"State"`
	PostalCode string `json:"postal_code" dynamodbav:"PostalCode"`
	Country    string `json:"country" dynamodbav:"Country"`
}

// ComplianceData tracks FCRA and compliance requirements
type ComplianceData struct {
	FCRADisclosureProvided bool       `json:"fcra_disclosure_provided" dynamodbav:"FCRADisclosureProvided"`
	FCRAConsentObtained    bool       `json:"fcra_consent_obtained" dynamodbav:"FCRAConsentObtained"`
	FCRAConsentDate        *time.Time `json:"fcra_consent_date,omitempty" dynamodbav:"FCRAConsentDate,omitempty"`
	AdverseActionNotice    bool       `json:"adverse_action_notice,omitempty" dynamodbav:"AdverseActionNotice,omitempty"`
	AdverseActionDate      *time.Time `json:"adverse_action_date,omitempty" dynamodbav:"AdverseActionDate,omitempty"`
	DisputeRaised          bool       `json:"dispute_raised,omitempty" dynamodbav:"DisputeRaised,omitempty"`
	DisputeResolved        bool       `json:"dispute_resolved,omitempty" dynamodbav:"DisputeResolved,omitempty"`
}

// BackgroundCheckPackage represents a predefined set of checks
type BackgroundCheckPackage struct {
	ID             string                  `json:"id" dynamodbav:"ID"`
	Name           string                  `json:"name" dynamodbav:"Name"`
	Description    string                  `json:"description" dynamodbav:"Description"`
	CheckTypes     []BackgroundCheckType   `json:"check_types" dynamodbav:"CheckTypes"`
	ProviderID     string                  `json:"provider_id" dynamodbav:"ProviderID"`
	TurnaroundDays int                     `json:"turnaround_days" dynamodbav:"TurnaroundDays"`
	Cost           float64                 `json:"cost" dynamodbav:"Cost"`
	Active         bool                    `json:"active" dynamodbav:"Active"`
	CreatedAt      time.Time               `json:"created_at" dynamodbav:"CreatedAt"`
	UpdatedAt      time.Time               `json:"updated_at" dynamodbav:"UpdatedAt"`
}

// BackgroundCheckRequest represents a request to initiate a background check
type BackgroundCheckRequest struct {
	PackageID     string
	CandidateInfo CandidateInfo
	CheckTypes    []BackgroundCheckType
	CallbackURL   string
}

// ProviderResponse contains the provider's response to initiating a check
type ProviderResponse struct {
	ProviderCheckID string
	Status          BackgroundCheckStatus
	EstimatedETA    *time.Time
	CandidateURL    string // URL for candidate to complete their portion
	ReportURL       string
	RawData         json.RawMessage
}

// ProviderStatus represents the current status from the provider
type ProviderStatus struct {
	Status      BackgroundCheckStatus
	Result      BackgroundCheckResult
	CompletedAt *time.Time
	ReportURL   string
	Updates     []StatusUpdate
	RawData     json.RawMessage
}

// StatusUpdate represents a status change event
type StatusUpdate struct {
	Timestamp time.Time
	Status    BackgroundCheckStatus
	Message   string
}

// ProviderReport contains the detailed report data
type ProviderReport struct {
	Overall      BackgroundCheckResult
	CheckResults map[BackgroundCheckType]CheckDetail
	ReportURL    string
	PDFReport    []byte
	RawData      json.RawMessage
}

// CheckDetail contains details for a specific check type
type CheckDetail struct {
	Result      BackgroundCheckResult
	Status      BackgroundCheckStatus
	Records     []Record
	Message     string
	CompletedAt *time.Time
}

// Record represents a single finding in a background check
type Record struct {
	Type        string
	Description string
	Date        *time.Time
	Location    string
	Severity    string
	Details     map[string]interface{}
}

// DisputeRequest represents a dispute submission
type DisputeRequest struct {
	CheckType   BackgroundCheckType
	RecordID    string
	Reason      string
	Evidence    []string // URLs to supporting documents
	ContactInfo string
}

// WebhookData represents standardized webhook data
type WebhookData struct {
	ProviderCheckID string
	Event           string
	Status          BackgroundCheckStatus
	Result          BackgroundCheckResult
	Timestamp       time.Time
	Data            json.RawMessage
}

// DynamoDB item structures for persistence

// BackgroundCheckItem is the DynamoDB representation
type BackgroundCheckItem struct {
	PK             string    `dynamodbav:"PK"`          // BGCHECK#{id}
	SK             string    `dynamodbav:"SK"`          // METADATA
	GSI1PK         string    `dynamodbav:"GSI1PK"`      // EMPLOYEE#{employee_id}
	GSI1SK         string    `dynamodbav:"GSI1SK"`      // BGCHECK#{initiated_at}
	EntityType     string    `dynamodbav:"EntityType"`  // BACKGROUND_CHECK
	ID             string    `dynamodbav:"ID"`
	EmployeeID     string    `dynamodbav:"EmployeeID"`
	PackageID      string    `dynamodbav:"PackageID"`
	ProviderID     string    `dynamodbav:"ProviderID"`
	Status         string    `dynamodbav:"Status"`
	Result         string    `dynamodbav:"Result,omitempty"`
	CheckTypes     []string  `dynamodbav:"CheckTypes"`
	CandidateInfo  string    `dynamodbav:"CandidateInfo"` // JSON
	ReportURL      string    `dynamodbav:"ReportURL,omitempty"`
	ProviderData   string    `dynamodbav:"ProviderData,omitempty"` // JSON
	InitiatedBy    string    `dynamodbav:"InitiatedBy"`
	InitiatedAt    string    `dynamodbav:"InitiatedAt"`
	CompletedAt    string    `dynamodbav:"CompletedAt,omitempty"`
	EstimatedETA   string    `dynamodbav:"EstimatedETA,omitempty"`
	Notes          string    `dynamodbav:"Notes,omitempty"`
	ComplianceData string    `dynamodbav:"ComplianceData"` // JSON
	CreatedAt      string    `dynamodbav:"CreatedAt"`
	UpdatedAt      string    `dynamodbav:"UpdatedAt"`
}

// PackageItem is the DynamoDB representation of a package
type PackageItem struct {
	PK             string   `dynamodbav:"PK"` // PACKAGE#{id}
	SK             string   `dynamodbav:"SK"` // METADATA
	EntityType     string   `dynamodbav:"EntityType"`
	ID             string   `dynamodbav:"ID"`
	Name           string   `dynamodbav:"Name"`
	Description    string   `dynamodbav:"Description"`
	CheckTypes     []string `dynamodbav:"CheckTypes"`
	ProviderID     string   `dynamodbav:"ProviderID"`
	TurnaroundDays int      `dynamodbav:"TurnaroundDays"`
	Cost           float64  `dynamodbav:"Cost"`
	Active         bool     `dynamodbav:"Active"`
	CreatedAt      string   `dynamodbav:"CreatedAt"`
	UpdatedAt      string   `dynamodbav:"UpdatedAt"`
}

// Helper methods

// NewBackgroundCheck creates a new background check with defaults
func NewBackgroundCheck(employeeID, packageID, providerID, initiatedBy string) *BackgroundCheck {
	now := time.Now()
	return &BackgroundCheck{
		ID:          uuid.New().String(),
		EmployeeID:  employeeID,
		PackageID:   packageID,
		ProviderID:  providerID,
		Status:      BGCheckStatusPending,
		InitiatedBy: initiatedBy,
		InitiatedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
		ComplianceData: ComplianceData{
			FCRADisclosureProvided: false,
			FCRAConsentObtained:    false,
		},
	}
}

// IsComplete returns true if the check is in a terminal state
func (bc *BackgroundCheck) IsComplete() bool {
	return bc.Status == BGCheckStatusCompleted ||
		bc.Status == BGCheckStatusFailed ||
		bc.Status == BGCheckStatusCancelled
}

// CanBeCancelled returns true if the check can be cancelled
func (bc *BackgroundCheck) CanBeCancelled() bool {
	return bc.Status == BGCheckStatusPending ||
		bc.Status == BGCheckStatusInProgress
}

// SetCompleted marks the check as completed
func (bc *BackgroundCheck) SetCompleted(result BackgroundCheckResult) {
	now := time.Now()
	bc.Status = BGCheckStatusCompleted
	bc.Result = result
	bc.CompletedAt = &now
	bc.UpdatedAt = now
}

// SetFailed marks the check as failed
func (bc *BackgroundCheck) SetFailed(notes string) {
	bc.Status = BGCheckStatusFailed
	bc.Notes = notes
	bc.UpdatedAt = time.Now()
}

// SetCancelled marks the check as cancelled
func (bc *BackgroundCheck) SetCancelled(notes string) {
	bc.Status = BGCheckStatusCancelled
	bc.Notes = notes
	bc.UpdatedAt = time.Now()
}

// UpdateFromProvider updates the check with data from provider
func (bc *BackgroundCheck) UpdateFromProvider(status ProviderStatus) {
	bc.Status = status.Status
	bc.Result = status.Result
	bc.CompletedAt = status.CompletedAt
	bc.ReportURL = status.ReportURL
	bc.ProviderData = status.RawData
	bc.UpdatedAt = time.Now()
}

// NewBackgroundCheckPackage creates a new package with defaults
func NewBackgroundCheckPackage(name, providerID string) *BackgroundCheckPackage {
	now := time.Now()
	return &BackgroundCheckPackage{
		ID:         uuid.New().String(),
		Name:       name,
		ProviderID: providerID,
		Active:     true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// IsActive returns true if the package is active
func (pkg *BackgroundCheckPackage) IsActive() bool {
	return pkg.Active
}

// Activate activates the package
func (pkg *BackgroundCheckPackage) Activate() {
	pkg.Active = true
	pkg.UpdatedAt = time.Now()
}

// Deactivate deactivates the package
func (pkg *BackgroundCheckPackage) Deactivate() {
	pkg.Active = false
	pkg.UpdatedAt = time.Now()
}

// Validation methods

// Validate validates the background check
func (bc *BackgroundCheck) Validate() error {
	if bc.EmployeeID == "" {
		return ErrEmployeeIDRequired
	}
	if bc.PackageID == "" {
		return ErrPackageIDRequired
	}
	if bc.ProviderID == "" {
		return ErrProviderIDRequired
	}
	if !bc.ComplianceData.FCRAConsentObtained {
		return ErrFCRAConsentRequired
	}
	return nil
}

// Validate validates the candidate info
func (ci *CandidateInfo) Validate() error {
	if ci.FirstName == "" || ci.LastName == "" {
		return ErrNameRequired
	}
	if ci.Email == "" {
		return ErrEmailRequired
	}
	if ci.Phone == "" {
		return ErrPhoneRequired
	}
	if ci.Address.Street1 == "" || ci.Address.City == "" ||
		ci.Address.State == "" || ci.Address.PostalCode == "" {
		return ErrAddressRequired
	}
	return nil
}

// Validate validates the package
func (pkg *BackgroundCheckPackage) Validate() error {
	if pkg.Name == "" {
		return ErrPackageNameRequired
	}
	if pkg.ProviderID == "" {
		return ErrProviderIDRequired
	}
	if len(pkg.CheckTypes) == 0 {
		return ErrCheckTypesRequired
	}
	return nil
}

// Custom errors
var (
	ErrEmployeeIDRequired   = NewValidationError("employee_id is required")
	ErrPackageIDRequired    = NewValidationError("package_id is required")
	ErrProviderIDRequired   = NewValidationError("provider_id is required")
	ErrFCRAConsentRequired  = NewValidationError("FCRA consent is required")
	ErrNameRequired         = NewValidationError("first name and last name are required")
	ErrEmailRequired        = NewValidationError("email is required")
	ErrPhoneRequired        = NewValidationError("phone is required")
	ErrAddressRequired      = NewValidationError("complete address is required")
	ErrPackageNameRequired  = NewValidationError("package name is required")
	ErrCheckTypesRequired   = NewValidationError("at least one check type is required")
	ErrCheckNotFound        = NewNotFoundError("background check not found")
	ErrPackageNotFound      = NewNotFoundError("package not found")
	ErrCheckAlreadyComplete = NewValidationError("check is already complete")
	ErrCheckCannotCancel    = NewValidationError("check cannot be cancelled in current state")
)

// ValidationError represents a validation error
type ValidationError struct {
	message string
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{message: message}
}

func (e *ValidationError) Error() string {
	return e.message
}

// NotFoundError represents a not found error
type NotFoundError struct {
	message string
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{message: message}
}

func (e *NotFoundError) Error() string {
	return e.message
}