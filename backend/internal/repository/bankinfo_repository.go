package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"hub-hrms/backend/internal/models"
)

// BankInfoRepository interface
type BankInfoRepository interface {
	Create(ctx context.Context, bankInfo *models.BankInfo) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.BankInfo, error)
	GetByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*models.BankInfo, error)
	GetPrimaryByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.BankInfo, error)
	Update(ctx context.Context, bankInfo *models.BankInfo) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetPrimary(ctx context.Context, id uuid.UUID, employeeID uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Verify(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error
	List(ctx context.Context, filters map[string]interface{}) ([]*models.BankInfo, error)
	CreateVerificationLog(ctx context.Context, log *models.BankVerificationLog) error
	GetVerificationLogs(ctx context.Context, bankInfoID uuid.UUID) ([]*models.BankVerificationLog, error)
}

type bankInfoRepository struct {
	db *pgxpool.Pool
}

// NewBankInfoRepository creates a new bank info repository
func NewBankInfoRepository(db *pgxpool.Pool) BankInfoRepository {
	return &bankInfoRepository{db: db}
}

// Create creates a new bank information record
func (r *bankInfoRepository) Create(ctx context.Context, bankInfo *models.BankInfo) error {
	query := `
		INSERT INTO bank_information (
			id, employee_id, account_holder_name, bank_name, account_type,
			account_number_encrypted, routing_number_encrypted, account_number_last4,
			swift_code, iban, bank_address, bank_city, bank_state, bank_zip, bank_country,
			is_primary, status, verified, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
		RETURNING created_at, updated_at
	`
	
	if bankInfo.ID == uuid.Nil {
		bankInfo.ID = uuid.New()
	}
	
	return r.db.QueryRow(
		ctx, query,
		bankInfo.ID,
		bankInfo.EmployeeID,
		bankInfo.AccountHolderName,
		bankInfo.BankName,
		bankInfo.AccountType,
		bankInfo.AccountNumberEncrypted,
		bankInfo.RoutingNumberEncrypted,
		bankInfo.AccountNumberLast4,
		bankInfo.SwiftCode,
		bankInfo.IBAN,
		bankInfo.BankAddress,
		bankInfo.BankCity,
		bankInfo.BankState,
		bankInfo.BankZip,
		bankInfo.BankCountry,
		bankInfo.IsPrimary,
		bankInfo.Status,
		bankInfo.Verified,
		bankInfo.CreatedBy,
	).Scan(&bankInfo.CreatedAt, &bankInfo.UpdatedAt)
}

// GetByID retrieves bank information by ID
func (r *bankInfoRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.BankInfo, error) {
	query := `
		SELECT 
			id, employee_id, account_holder_name, bank_name, account_type,
			account_number_encrypted, routing_number_encrypted, account_number_last4,
			swift_code, iban, bank_address, bank_city, bank_state, bank_zip, bank_country,
			is_primary, status, verified, verified_at, verified_by,
			created_at, updated_at, created_by
		FROM bank_information
		WHERE id = $1 AND status != 'deleted'
	`
	
	bankInfo := &models.BankInfo{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&bankInfo.ID,
		&bankInfo.EmployeeID,
		&bankInfo.AccountHolderName,
		&bankInfo.BankName,
		&bankInfo.AccountType,
		&bankInfo.AccountNumberEncrypted,
		&bankInfo.RoutingNumberEncrypted,
		&bankInfo.AccountNumberLast4,
		&bankInfo.SwiftCode,
		&bankInfo.IBAN,
		&bankInfo.BankAddress,
		&bankInfo.BankCity,
		&bankInfo.BankState,
		&bankInfo.BankZip,
		&bankInfo.BankCountry,
		&bankInfo.IsPrimary,
		&bankInfo.Status,
		&bankInfo.Verified,
		&bankInfo.VerifiedAt,
		&bankInfo.VerifiedBy,
		&bankInfo.CreatedAt,
		&bankInfo.UpdatedAt,
		&bankInfo.CreatedBy,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bank information not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bank info: %w", err)
	}
	
	return bankInfo, nil
}

// GetByEmployeeID retrieves all bank information for an employee
func (r *bankInfoRepository) GetByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*models.BankInfo, error) {
	query := `
		SELECT 
			id, employee_id, account_holder_name, bank_name, account_type,
			account_number_encrypted, routing_number_encrypted, account_number_last4,
			swift_code, iban, bank_address, bank_city, bank_state, bank_zip, bank_country,
			is_primary, status, verified, verified_at, verified_by,
			created_at, updated_at, created_by
		FROM bank_information
		WHERE employee_id = $1 AND status != 'deleted'
		ORDER BY is_primary DESC, created_at DESC
	`
	
	rows, err := r.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query bank info: %w", err)
	}
	defer rows.Close()
	
	var bankInfos []*models.BankInfo
	for rows.Next() {
		bankInfo := &models.BankInfo{}
		err := rows.Scan(
			&bankInfo.ID,
			&bankInfo.EmployeeID,
			&bankInfo.AccountHolderName,
			&bankInfo.BankName,
			&bankInfo.AccountType,
			&bankInfo.AccountNumberEncrypted,
			&bankInfo.RoutingNumberEncrypted,
			&bankInfo.AccountNumberLast4,
			&bankInfo.SwiftCode,
			&bankInfo.IBAN,
			&bankInfo.BankAddress,
			&bankInfo.BankCity,
			&bankInfo.BankState,
			&bankInfo.BankZip,
			&bankInfo.BankCountry,
			&bankInfo.IsPrimary,
			&bankInfo.Status,
			&bankInfo.Verified,
			&bankInfo.VerifiedAt,
			&bankInfo.VerifiedBy,
			&bankInfo.CreatedAt,
			&bankInfo.UpdatedAt,
			&bankInfo.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bank info: %w", err)
		}
		bankInfos = append(bankInfos, bankInfo)
	}
	
	return bankInfos, nil
}

// GetPrimaryByEmployeeID retrieves the primary bank account for an employee
func (r *bankInfoRepository) GetPrimaryByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.BankInfo, error) {
	query := `
		SELECT 
			id, employee_id, account_holder_name, bank_name, account_type,
			account_number_encrypted, routing_number_encrypted, account_number_last4,
			swift_code, iban, bank_address, bank_city, bank_state, bank_zip, bank_country,
			is_primary, status, verified, verified_at, verified_by,
			created_at, updated_at, created_by
		FROM bank_information
		WHERE employee_id = $1 AND is_primary = true AND status = 'active'
		LIMIT 1
	`
	
	bankInfo := &models.BankInfo{}
	err := r.db.QueryRow(ctx, query, employeeID).Scan(
		&bankInfo.ID,
		&bankInfo.EmployeeID,
		&bankInfo.AccountHolderName,
		&bankInfo.BankName,
		&bankInfo.AccountType,
		&bankInfo.AccountNumberEncrypted,
		&bankInfo.RoutingNumberEncrypted,
		&bankInfo.AccountNumberLast4,
		&bankInfo.SwiftCode,
		&bankInfo.IBAN,
		&bankInfo.BankAddress,
		&bankInfo.BankCity,
		&bankInfo.BankState,
		&bankInfo.BankZip,
		&bankInfo.BankCountry,
		&bankInfo.IsPrimary,
		&bankInfo.Status,
		&bankInfo.Verified,
		&bankInfo.VerifiedAt,
		&bankInfo.VerifiedBy,
		&bankInfo.CreatedAt,
		&bankInfo.UpdatedAt,
		&bankInfo.CreatedBy,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no primary bank account found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get primary bank info: %w", err)
	}
	
	return bankInfo, nil
}

// Update updates bank information
func (r *bankInfoRepository) Update(ctx context.Context, bankInfo *models.BankInfo) error {
	query := `
		UPDATE bank_information
		SET 
			account_holder_name = $1,
			bank_name = $2,
			account_type = $3,
			swift_code = $4,
			iban = $5,
			bank_address = $6,
			bank_city = $7,
			bank_state = $8,
			bank_zip = $9,
			bank_country = $10,
			updated_at = NOW()
		WHERE id = $11
		RETURNING updated_at
	`
	
	return r.db.QueryRow(
		ctx, query,
		bankInfo.AccountHolderName,
		bankInfo.BankName,
		bankInfo.AccountType,
		bankInfo.SwiftCode,
		bankInfo.IBAN,
		bankInfo.BankAddress,
		bankInfo.BankCity,
		bankInfo.BankState,
		bankInfo.BankZip,
		bankInfo.BankCountry,
		bankInfo.ID,
	).Scan(&bankInfo.UpdatedAt)
}

// Delete soft deletes bank information
func (r *bankInfoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE bank_information SET status = 'deleted', updated_at = NOW() WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete bank info: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("bank information not found")
	}
	
	return nil
}

// SetPrimary sets a bank account as primary (and unsets others)
func (r *bankInfoRepository) SetPrimary(ctx context.Context, id uuid.UUID, employeeID uuid.UUID) error {
	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	
	// Unset all primary flags for this employee
	_, err = tx.Exec(ctx, `
		UPDATE bank_information 
		SET is_primary = false, updated_at = NOW() 
		WHERE employee_id = $1
	`, employeeID)
	if err != nil {
		return fmt.Errorf("failed to unset primary flags: %w", err)
	}
	
	// Set new primary
	result, err := tx.Exec(ctx, `
		UPDATE bank_information 
		SET is_primary = true, updated_at = NOW() 
		WHERE id = $1 AND employee_id = $2
	`, id, employeeID)
	if err != nil {
		return fmt.Errorf("failed to set primary: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("bank information not found")
	}
	
	return tx.Commit(ctx)
}

// UpdateStatus updates the status of bank information
func (r *bankInfoRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE bank_information SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("bank information not found")
	}
	
	return nil
}

// Verify marks bank information as verified
func (r *bankInfoRepository) Verify(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error {
	query := `
		UPDATE bank_information 
		SET verified = true, status = 'active', verified_at = NOW(), verified_by = $1, updated_at = NOW() 
		WHERE id = $2
	`
	result, err := r.db.Exec(ctx, query, verifiedBy, id)
	if err != nil {
		return fmt.Errorf("failed to verify bank info: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("bank information not found")
	}
	
	return nil
}

// List retrieves bank information with filters
func (r *bankInfoRepository) List(ctx context.Context, filters map[string]interface{}) ([]*models.BankInfo, error) {
	query := `
		SELECT 
			id, employee_id, account_holder_name, bank_name, account_type,
			account_number_encrypted, routing_number_encrypted, account_number_last4,
			swift_code, iban, bank_address, bank_city, bank_state, bank_zip, bank_country,
			is_primary, status, verified, verified_at, verified_by,
			created_at, updated_at, created_by
		FROM bank_information
		WHERE status != 'deleted'
	`
	
	args := []interface{}{}
	argCount := 0
	
	// Add filters
	if status, ok := filters["status"].(string); ok {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}
	
	if verified, ok := filters["verified"].(bool); ok {
		argCount++
		query += fmt.Sprintf(" AND verified = $%d", argCount)
		args = append(args, verified)
	}
	
	query += " ORDER BY created_at DESC"
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list bank info: %w", err)
	}
	defer rows.Close()
	
	var bankInfos []*models.BankInfo
	for rows.Next() {
		bankInfo := &models.BankInfo{}
		err := rows.Scan(
			&bankInfo.ID,
			&bankInfo.EmployeeID,
			&bankInfo.AccountHolderName,
			&bankInfo.BankName,
			&bankInfo.AccountType,
			&bankInfo.AccountNumberEncrypted,
			&bankInfo.RoutingNumberEncrypted,
			&bankInfo.AccountNumberLast4,
			&bankInfo.SwiftCode,
			&bankInfo.IBAN,
			&bankInfo.BankAddress,
			&bankInfo.BankCity,
			&bankInfo.BankState,
			&bankInfo.BankZip,
			&bankInfo.BankCountry,
			&bankInfo.IsPrimary,
			&bankInfo.Status,
			&bankInfo.Verified,
			&bankInfo.VerifiedAt,
			&bankInfo.VerifiedBy,
			&bankInfo.CreatedAt,
			&bankInfo.UpdatedAt,
			&bankInfo.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bank info: %w", err)
		}
		bankInfos = append(bankInfos, bankInfo)
	}
	
	return bankInfos, nil
}

// CreateVerificationLog creates a verification log entry
func (r *bankInfoRepository) CreateVerificationLog(ctx context.Context, log *models.BankVerificationLog) error {
	query := `
		INSERT INTO bank_verification_log (
			id, bank_info_id, verification_method, verification_status,
			verification_code, attempt_count, verified_at, verified_by, failure_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at
	`
	
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	
	return r.db.QueryRow(
		ctx, query,
		log.ID,
		log.BankInfoID,
		log.VerificationMethod,
		log.VerificationStatus,
		log.VerificationCode,
		log.AttemptCount,
		log.VerifiedAt,
		log.VerifiedBy,
		log.FailureReason,
	).Scan(&log.CreatedAt)
}

// GetVerificationLogs retrieves verification logs for a bank account
func (r *bankInfoRepository) GetVerificationLogs(ctx context.Context, bankInfoID uuid.UUID) ([]*models.BankVerificationLog, error) {
	query := `
		SELECT 
			id, bank_info_id, verification_method, verification_status,
			verification_code, attempt_count, verified_at, verified_by,
			failure_reason, created_at
		FROM bank_verification_log
		WHERE bank_info_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(ctx, query, bankInfoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification logs: %w", err)
	}
	defer rows.Close()
	
	var logs []*models.BankVerificationLog
	for rows.Next() {
		log := &models.BankVerificationLog{}
		err := rows.Scan(
			&log.ID,
			&log.BankInfoID,
			&log.VerificationMethod,
			&log.VerificationStatus,
			&log.VerificationCode,
			&log.AttemptCount,
			&log.VerifiedAt,
			&log.VerifiedBy,
			&log.FailureReason,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan verification log: %w", err)
		}
		logs = append(logs, log)
	}
	
	return logs, nil
}