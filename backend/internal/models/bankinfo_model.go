package models

import (
	"time"
	"github.com/google/uuid"
)

// ============================================================================
// Bank Information Models
// ============================================================================

// BankInfo represents banking information for an employee/vendor
type BankInfo struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	EmployeeID            uuid.UUID  `json:"employee_id" db:"employee_id"`
	AccountHolderName     string     `json:"account_holder_name" db:"account_holder_name"`
	BankName              string     `json:"bank_name" db:"bank_name"`
	AccountType           string     `json:"account_type" db:"account_type"` // checking, savings
	AccountNumberLast4    string     `json:"account_number_last4" db:"account_number_last4"`
	AccountNumberEncrypted string    `json:"-" db:"account_number_encrypted"` // Never expose in JSON
	RoutingNumberEncrypted string    `json:"-" db:"routing_number_encrypted"` // Never expose in JSON
	SwiftCode             *string    `json:"swift_code,omitempty" db:"swift_code"`
	IBAN                  *string    `json:"iban,omitempty" db:"iban"`
	BankAddress           *string    `json:"bank_address,omitempty" db:"bank_address"`
	BankCity              *string    `json:"bank_city,omitempty" db:"bank_city"`
	BankState             *string    `json:"bank_state,omitempty" db:"bank_state"`
	BankZip               *string    `json:"bank_zip,omitempty" db:"bank_zip"`
	BankCountry           string     `json:"bank_country" db:"bank_country"`
	IsPrimary             bool       `json:"is_primary" db:"is_primary"`
	Status                string     `json:"status" db:"status"` // pending, active, inactive, suspended
	Verified              bool       `json:"verified" db:"verified"`
	VerifiedAt            *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	VerifiedBy            *uuid.UUID `json:"verified_by,omitempty" db:"verified_by"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy             *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	
	// Decrypted fields (only populated when explicitly requested)
	AccountNumber         string     `json:"account_number,omitempty" db:"-"`
	RoutingNumber         string     `json:"routing_number,omitempty" db:"-"`
}

// BankInfoCreateRequest for creating bank information
type BankInfoCreateRequest struct {
	EmployeeID        uuid.UUID `json:"employee_id"`
	AccountHolderName string    `json:"account_holder_name"`
	BankName          string    `json:"bank_name"`
	AccountType       string    `json:"account_type"`
	AccountNumber     string    `json:"account_number"` // Will be encrypted before storage
	RoutingNumber     string    `json:"routing_number"` // Will be encrypted before storage
	SwiftCode         *string   `json:"swift_code,omitempty"`
	IBAN              *string   `json:"iban,omitempty"`
	BankAddress       *string   `json:"bank_address,omitempty"`
	BankCity          *string   `json:"bank_city,omitempty"`
	BankState         *string   `json:"bank_state,omitempty"`
	BankZip           *string   `json:"bank_zip,omitempty"`
	BankCountry       string    `json:"bank_country"`
	IsPrimary         bool      `json:"is_primary"`
}

// BankInfoUpdateRequest for updating bank information
type BankInfoUpdateRequest struct {
	AccountHolderName string  `json:"account_holder_name"`
	BankName          string  `json:"bank_name"`
	AccountType       string  `json:"account_type"`
	SwiftCode         *string `json:"swift_code,omitempty"`
	IBAN              *string `json:"iban,omitempty"`
	BankAddress       *string `json:"bank_address,omitempty"`
	BankCity          *string `json:"bank_city,omitempty"`
	BankState         *string `json:"bank_state,omitempty"`
	BankZip           *string `json:"bank_zip,omitempty"`
	BankCountry       string  `json:"bank_country"`
}

// BankVerificationLog tracks verification attempts
type BankVerificationLog struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	BankInfoID         uuid.UUID  `json:"bank_info_id" db:"bank_info_id"`
	VerificationMethod string     `json:"verification_method" db:"verification_method"` // micro-deposit, instant, manual, third-party
	VerificationStatus string     `json:"verification_status" db:"verification_status"` // initiated, pending, success, failed, expired
	VerificationCode   *string    `json:"verification_code,omitempty" db:"verification_code"`
	AttemptCount       int        `json:"attempt_count" db:"attempt_count"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	VerifiedBy         *uuid.UUID `json:"verified_by,omitempty" db:"verified_by"`
	FailureReason      *string    `json:"failure_reason,omitempty" db:"failure_reason"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

// BankInfoResponse for returning bank info to client (with masked sensitive data)
type BankInfoResponse struct {
	ID                 uuid.UUID  `json:"id"`
	EmployeeID         uuid.UUID  `json:"employee_id"`
	AccountHolderName  string     `json:"account_holder_name"`
	BankName           string     `json:"bank_name"`
	AccountType        string     `json:"account_type"`
	AccountNumberLast4 string     `json:"account_number_last4"`
	SwiftCode          *string    `json:"swift_code,omitempty"`
	IBAN               *string    `json:"iban,omitempty"`
	BankAddress        *string    `json:"bank_address,omitempty"`
	BankCity           *string    `json:"bank_city,omitempty"`
	BankState          *string    `json:"bank_state,omitempty"`
	BankZip            *string    `json:"bank_zip,omitempty"`
	BankCountry        string     `json:"bank_country"`
	IsPrimary          bool       `json:"is_primary"`
	Status             string     `json:"status"`
	Verified           bool       `json:"verified"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ToBankInfoResponse converts BankInfo to BankInfoResponse (masks sensitive data)
func (b *BankInfo) ToBankInfoResponse() *BankInfoResponse {
	return &BankInfoResponse{
		ID:                 b.ID,
		EmployeeID:         b.EmployeeID,
		AccountHolderName:  b.AccountHolderName,
		BankName:           b.BankName,
		AccountType:        b.AccountType,
		AccountNumberLast4: b.AccountNumberLast4,
		SwiftCode:          b.SwiftCode,
		IBAN:               b.IBAN,
		BankAddress:        b.BankAddress,
		BankCity:           b.BankCity,
		BankState:          b.BankState,
		BankZip:            b.BankZip,
		BankCountry:        b.BankCountry,
		IsPrimary:          b.IsPrimary,
		Status:             b.Status,
		Verified:           b.Verified,
		VerifiedAt:         b.VerifiedAt,
		CreatedAt:          b.CreatedAt,
		UpdatedAt:          b.UpdatedAt,
	}
}