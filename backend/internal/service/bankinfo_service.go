package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
)

var (
	ErrBankInfoNotFound      = errors.New("bank information not found")
	ErrInvalidRoutingNumber  = errors.New("routing number must be exactly 9 digits")
	ErrInvalidAccountNumber  = errors.New("account number must be between 8 and 17 digits")
	ErrEncryptionFailed      = errors.New("encryption failed")
	ErrDecryptionFailed      = errors.New("decryption failed")
	ErrPrimaryAlreadyExists  = errors.New("primary bank account already exists")
)

// BankInfoService interface defines bank information operations
type BankInfoService interface {
	CreateBankInfo(ctx context.Context, req *models.BankInfoCreateRequest, createdBy uuid.UUID) (*models.BankInfo, error)
	GetBankInfo(ctx context.Context, id uuid.UUID) (*models.BankInfo, error)
	GetBankInfoByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*models.BankInfo, error)
	GetPrimaryBankInfo(ctx context.Context, employeeID uuid.UUID) (*models.BankInfo, error)
	UpdateBankInfo(ctx context.Context, id uuid.UUID, req *models.BankInfoUpdateRequest) (*models.BankInfo, error)
	DeleteBankInfo(ctx context.Context, id uuid.UUID) error
	SetPrimaryBankInfo(ctx context.Context, id uuid.UUID, employeeID uuid.UUID) error
	VerifyBankInfo(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error
	ListBankInfo(ctx context.Context, filters map[string]interface{}) ([]*models.BankInfo, error)
	DecryptSensitiveData(ctx context.Context, bankInfo *models.BankInfo) error
}

// bankInfoService implements BankInfoService
type bankInfoService struct {
	repos         *repository.Repositories
	encryptionKey []byte
}

// NewBankInfoService creates a new bank info service
func NewBankInfoService(repos *repository.Repositories, encryptionKey string) BankInfoService {
	// Ensure key is exactly 32 bytes for AES-256
	key := []byte(encryptionKey)
	if len(key) != 32 {
		paddedKey := make([]byte, 32)
		copy(paddedKey, key)
		key = paddedKey
	}
	
	return &bankInfoService{
		repos:         repos,
		encryptionKey: key,
	}
}

// CreateBankInfo creates new bank information
func (s *bankInfoService) CreateBankInfo(ctx context.Context, req *models.BankInfoCreateRequest, createdBy uuid.UUID) (*models.BankInfo, error) {
	// Validate inputs
	if err := s.validateAccountNumber(req.AccountNumber); err != nil {
		return nil, err
	}
	if err := s.validateRoutingNumber(req.RoutingNumber); err != nil {
		return nil, err
	}
	
	// Encrypt sensitive data
	encryptedAccount, err := s.encrypt(req.AccountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt account number: %w", err)
	}
	
	encryptedRouting, err := s.encrypt(req.RoutingNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt routing number: %w", err)
	}
	
	// Get last 4 digits
	last4 := ""
	if len(req.AccountNumber) >= 4 {
		last4 = req.AccountNumber[len(req.AccountNumber)-4:]
	}
	
	// Create bank info model
	bankInfo := &models.BankInfo{
		EmployeeID:             req.EmployeeID,
		AccountHolderName:      req.AccountHolderName,
		BankName:               req.BankName,
		AccountType:            req.AccountType,
		AccountNumberEncrypted: encryptedAccount,
		RoutingNumberEncrypted: encryptedRouting,
		AccountNumberLast4:     last4,
		SwiftCode:              req.SwiftCode,
		IBAN:                   req.IBAN,
		BankAddress:            req.BankAddress,
		BankCity:               req.BankCity,
		BankState:              req.BankState,
		BankZip:                req.BankZip,
		BankCountry:            req.BankCountry,
		IsPrimary:              req.IsPrimary,
		Status:                 "pending",
		Verified:               false,
		CreatedBy:              &createdBy,
	}
	
	// Create in database
	if err := s.repos.BankInfo.Create(ctx, bankInfo); err != nil {
		return nil, fmt.Errorf("failed to create bank info: %w", err)
	}
	
	return bankInfo, nil
}

// GetBankInfo retrieves bank information by ID
func (s *bankInfoService) GetBankInfo(ctx context.Context, id uuid.UUID) (*models.BankInfo, error) {
	bankInfo, err := s.repos.BankInfo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrBankInfoNotFound
	}
	return bankInfo, nil
}

// GetBankInfoByEmployeeID retrieves all bank information for an employee
func (s *bankInfoService) GetBankInfoByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*models.BankInfo, error) {
	return s.repos.BankInfo.GetByEmployeeID(ctx, employeeID)
}

// GetPrimaryBankInfo retrieves the primary bank account for an employee
func (s *bankInfoService) GetPrimaryBankInfo(ctx context.Context, employeeID uuid.UUID) (*models.BankInfo, error) {
	bankInfo, err := s.repos.BankInfo.GetPrimaryByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, ErrBankInfoNotFound
	}
	return bankInfo, nil
}

// UpdateBankInfo updates bank information
func (s *bankInfoService) UpdateBankInfo(ctx context.Context, id uuid.UUID, req *models.BankInfoUpdateRequest) (*models.BankInfo, error) {
	// Get existing bank info
	bankInfo, err := s.repos.BankInfo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrBankInfoNotFound
	}
	
	// Update fields
	bankInfo.AccountHolderName = req.AccountHolderName
	bankInfo.BankName = req.BankName
	bankInfo.AccountType = req.AccountType
	bankInfo.SwiftCode = req.SwiftCode
	bankInfo.IBAN = req.IBAN
	bankInfo.BankAddress = req.BankAddress
	bankInfo.BankCity = req.BankCity
	bankInfo.BankState = req.BankState
	bankInfo.BankZip = req.BankZip
	bankInfo.BankCountry = req.BankCountry
	
	// Update in database
	if err := s.repos.BankInfo.Update(ctx, bankInfo); err != nil {
		return nil, fmt.Errorf("failed to update bank info: %w", err)
	}
	
	return bankInfo, nil
}

// DeleteBankInfo soft deletes bank information
func (s *bankInfoService) DeleteBankInfo(ctx context.Context, id uuid.UUID) error {
	return s.repos.BankInfo.Delete(ctx, id)
}

// SetPrimaryBankInfo sets a bank account as primary
func (s *bankInfoService) SetPrimaryBankInfo(ctx context.Context, id uuid.UUID, employeeID uuid.UUID) error {
	return s.repos.BankInfo.SetPrimary(ctx, id, employeeID)
}

// VerifyBankInfo marks bank information as verified
func (s *bankInfoService) VerifyBankInfo(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error {
	// Create verification log
	log := &models.BankVerificationLog{
		BankInfoID:         id,
		VerificationMethod: "manual",
		VerificationStatus: "success",
		AttemptCount:       1,
		VerifiedBy:         &verifiedBy,
	}
	
	if err := s.repos.BankInfo.CreateVerificationLog(ctx, log); err != nil {
		return fmt.Errorf("failed to create verification log: %w", err)
	}
	
	// Mark as verified
	return s.repos.BankInfo.Verify(ctx, id, verifiedBy)
}

// ListBankInfo retrieves bank information with filters
func (s *bankInfoService) ListBankInfo(ctx context.Context, filters map[string]interface{}) ([]*models.BankInfo, error) {
	return s.repos.BankInfo.List(ctx, filters)
}

// DecryptSensitiveData decrypts the sensitive data in a BankInfo object
func (s *bankInfoService) DecryptSensitiveData(ctx context.Context, bankInfo *models.BankInfo) error {
	// Decrypt account number
	accountNumber, err := s.decrypt(bankInfo.AccountNumberEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt account number: %w", err)
	}
	bankInfo.AccountNumber = accountNumber
	
	// Decrypt routing number
	routingNumber, err := s.decrypt(bankInfo.RoutingNumberEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt routing number: %w", err)
	}
	bankInfo.RoutingNumber = routingNumber
	
	return nil
}

// validateAccountNumber validates the account number format
func (s *bankInfoService) validateAccountNumber(accountNumber string) error {
	// Account numbers are typically 8-17 digits
	matched, _ := regexp.MatchString(`^\d{8,17}$`, accountNumber)
	if !matched {
		return ErrInvalidAccountNumber
	}
	return nil
}

// validateRoutingNumber validates the routing number format
func (s *bankInfoService) validateRoutingNumber(routingNumber string) error {
	// Routing numbers must be exactly 9 digits
	matched, _ := regexp.MatchString(`^\d{9}$`, routingNumber)
	if !matched {
		return ErrInvalidRoutingNumber
	}
	return nil
}

// encrypt encrypts data using AES-256-GCM
func (s *bankInfoService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", ErrEncryptionFailed
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", ErrEncryptionFailed
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", ErrEncryptionFailed
	}
	
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts data using AES-256-GCM
func (s *bankInfoService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrDecryptionFailed
	}
	
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", ErrDecryptionFailed
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", ErrDecryptionFailed
	}
	
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrDecryptionFailed
	}
	
	nonce, encryptedData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}
	
	return string(plaintext), nil
}