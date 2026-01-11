package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"hub-hrms/backend/internal/models"
)


// BackgroundCheckProvider defines the interface for background check providers
type BackgroundCheckProvider interface {
	// GetName returns the provider's name
	GetName() string

	// InitiateCheck starts a new background check
	InitiateCheck(ctx context.Context, req *BackgroundCheckRequest) (*ProviderResponse, error)

	// GetCheckStatus retrieves the current status of a check
	GetCheckStatus(ctx context.Context, providerCheckID string) (*ProviderStatus, error)

	// GetReport retrieves the full report for a completed check
	GetReport(ctx context.Context, providerCheckID string) (*ProviderReport, error)

	// CancelCheck cancels an in-progress check
	CancelCheck(ctx context.Context, providerCheckID string) error

	// SubmitDispute submits a dispute for a check result
	SubmitDispute(ctx context.Context, providerCheckID string, dispute *DisputeRequest) error

	// ValidateWebhook validates incoming webhook signatures
	ValidateWebhook(payload []byte, signature string) error

	// ParseWebhook parses webhook data into standard format
	ParseWebhook(payload []byte) (*WebhookData, error)
}

// BackgroundCheckRequest represents a request to initiate a background check
type BackgroundCheckRequest struct {
	PackageID     string
	CandidateInfo models.CandidateInfo
	CheckTypes    []models.BackgroundCheckType
	CallbackURL   string
}

// ProviderResponse contains the provider's response to initiating a check
type ProviderResponse struct {
	ProviderCheckID string
	Status          models.BackgroundCheckStatus
	EstimatedETA    *time.Time
	CandidateURL    string // URL for candidate to complete their portion
	ReportURL       string
	RawData         json.RawMessage
}

// ProviderStatus represents the current status from the provider
type ProviderStatus struct {
	Status       models.BackgroundCheckStatus
	Result       models.BackgroundCheckResult
	CompletedAt  *time.Time
	ReportURL    string
	Updates      []StatusUpdate
	RawData      json.RawMessage
}

// StatusUpdate represents a status change event
type StatusUpdate struct {
	Timestamp time.Time
	Status    models.BackgroundCheckStatus
	Message   string
}

// ProviderReport contains the detailed report data
type ProviderReport struct {
	Overall       models.BackgroundCheckResult
	CheckResults  map[models.BackgroundCheckType]CheckDetail
	ReportURL     string
	PDFReport     []byte
	RawData       json.RawMessage
}

// CheckDetail contains details for a specific check type
type CheckDetail struct {
	Result      models.BackgroundCheckResult
	Status      models.BackgroundCheckStatus
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
	CheckType   models.BackgroundCheckType
	RecordID    string
	Reason      string
	Evidence    []string // URLs to supporting documents
	ContactInfo string
}

// WebhookData represents standardized webhook data
type WebhookData struct {
	ProviderCheckID string
	Event           string
	Status          models.BackgroundCheckStatus
	Result          models.BackgroundCheckResult
	Timestamp       time.Time
	Data            json.RawMessage
}

// BackgroundCheckService manages background checks
type BackgroundCheckService struct {
	providers  map[string]BackgroundCheckProvider
	repository BackgroundCheckRepository
	notifier   NotificationService
}

// BackgroundCheckRepository defines data access methods
type BackgroundCheckRepository interface {
	Create(ctx context.Context, check *models.BackgroundCheck) error
	GetByID(ctx context.Context, id string) (*models.BackgroundCheck, error)
	GetByEmployeeID(ctx context.Context, employeeID string) ([]*models.BackgroundCheck, error)
	Update(ctx context.Context, check *models.BackgroundCheck) error
	List(ctx context.Context, filters map[string]interface{}) ([]*models.BackgroundCheck, error)
	
	// Package methods
	CreatePackage(ctx context.Context, pkg *models.BackgroundCheckPackage) error
	GetPackage(ctx context.Context, id string) (*models.BackgroundCheckPackage, error)
	ListPackages(ctx context.Context) ([]*models.BackgroundCheckPackage, error)
}

// NewBackgroundCheckService creates a new background check service
func NewBackgroundCheckService(
	repository BackgroundCheckRepository,
	notifier NotificationService,
) *BackgroundCheckService {
	return &BackgroundCheckService{
		providers:  make(map[string]BackgroundCheckProvider),
		repository: repository,
		notifier:   notifier,
	}
}

// RegisterProvider registers a background check provider
func (s *BackgroundCheckService) RegisterProvider(provider BackgroundCheckProvider) {
	s.providers[provider.GetName()] = provider
}

// InitiateBackgroundCheck starts a new background check
func (s *BackgroundCheckService) InitiateBackgroundCheck(
	ctx context.Context,
	employeeID string,
	packageID string,
	candidateInfo models.CandidateInfo,
	initiatedBy string,
) (*models.BackgroundCheck, error) {
	// Get package details
	pkg, err := s.repository.GetPackage(ctx, packageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}

	if !pkg.Active {
		return nil, fmt.Errorf("package %s is not active", packageID)
	}

	// Get the provider
	provider, ok := s.providers[pkg.ProviderID]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", pkg.ProviderID)
	}

	// Note: FCRA compliance should be validated before calling this method
	// The handler should ensure consent is obtained before initiating the check

	// Create background check record
	check := &models.BackgroundCheck{
		ID:            uuid.New().String(),
		EmployeeID:    employeeID,
		PackageID:     packageID,
		ProviderID:    pkg.ProviderID,
		Status:        models.BGCheckStatusPending,
		CheckTypes:    pkg.CheckTypes,
		CandidateInfo: candidateInfo,
		InitiatedBy:   initiatedBy,
		InitiatedAt:   time.Now(),
		ComplianceData: models.ComplianceData{
			FCRADisclosureProvided: true,
			FCRAConsentObtained:    true,
			FCRAConsentDate:        timePtr(time.Now()),
		},
	}

	// Initiate check with provider
	req := &BackgroundCheckRequest{
		PackageID:     packageID,
		CandidateInfo: candidateInfo,
		CheckTypes:    pkg.CheckTypes,
		CallbackURL:   fmt.Sprintf("https://api.yourdomain.com/webhooks/background-checks/%s", check.ID),
	}

	providerResp, err := provider.InitiateCheck(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate check with provider: %w", err)
	}

	// Update check with provider response
	check.ProviderData = providerResp.RawData
	check.Status = providerResp.Status
	check.EstimatedETA = providerResp.EstimatedETA
	check.ReportURL = providerResp.ReportURL

	// Save to repository
	if err := s.repository.Create(ctx, check); err != nil {
		return nil, fmt.Errorf("failed to save background check: %w", err)
	}

	// Send notifications
	if err := s.notifier.NotifyCheckInitiated(ctx, check); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("failed to send notification: %v\n", err)
	}

	return check, nil
}

// GetCheckStatus retrieves the current status of a background check
func (s *BackgroundCheckService) GetCheckStatus(ctx context.Context, checkID string) (*models.BackgroundCheck, error) {
	check, err := s.repository.GetByID(ctx, checkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get check: %w", err)
	}

	// If check is not completed, fetch latest status from provider
	if check.Status != models.BGCheckStatusCompleted && check.Status != models.BGCheckStatusFailed && check.Status != models.BGCheckStatusCancelled {
		provider, ok := s.providers[check.ProviderID]
		if !ok {
			return nil, fmt.Errorf("provider %s not found", check.ProviderID)
		}

		var providerCheckID string
		if err := json.Unmarshal(check.ProviderData, &providerCheckID); err == nil {
			status, err := provider.GetCheckStatus(ctx, providerCheckID)
			if err == nil {
				// Update check with latest status
				check.Status = status.Status
				check.Result = status.Result
				check.CompletedAt = status.CompletedAt
				check.ReportURL = status.ReportURL

				if err := s.repository.Update(ctx, check); err != nil {
					fmt.Printf("failed to update check status: %v\n", err)
				}

				// Send completion notification if newly completed
				if status.Status == models.BGCheckStatusCompleted {
					if err := s.notifier.NotifyCheckCompleted(ctx, check); err != nil {
						fmt.Printf("failed to send completion notification: %v\n", err)
					}
				}
			}
		}
	}

	return check, nil
}

// HandleWebhook processes webhook callbacks from providers
func (s *BackgroundCheckService) HandleWebhook(
	ctx context.Context,
	providerName string,
	payload []byte,
	signature string,
) error {
	provider, ok := s.providers[providerName]
	if !ok {
		return fmt.Errorf("provider %s not found", providerName)
	}

	// Validate webhook signature
	if err := provider.ValidateWebhook(payload, signature); err != nil {
		return fmt.Errorf("invalid webhook signature: %w", err)
	}

	// Parse webhook data
	webhookData, err := provider.ParseWebhook(payload)
	if err != nil {
		return fmt.Errorf("failed to parse webhook: %w", err)
	}

	// Find the check by provider check ID
	// Note: This is a simplified implementation. In production, you would need
	// a repository method to find checks by provider check ID
	
	// For now, log the webhook data for debugging
	fmt.Printf("Received webhook: event=%s, status=%s, provider_check_id=%s\n",
		webhookData.Event, webhookData.Status, webhookData.ProviderCheckID)
	
	// TODO: Implement check lookup and update based on webhook data
	// Example:
	// check, err := s.repository.GetByProviderCheckID(ctx, webhookData.ProviderCheckID)
	// if err != nil {
	//     return fmt.Errorf("failed to find check: %w", err)
	// }
	// check.Status = webhookData.Status
	// check.Result = webhookData.Result
	// if err := s.repository.Update(ctx, check); err != nil {
	//     return fmt.Errorf("failed to update check: %w", err)
	// }

	return nil
}

// GetEmployeeChecks retrieves all background checks for an employee
func (s *BackgroundCheckService) GetEmployeeChecks(
	ctx context.Context,
	employeeID string,
) ([]*models.BackgroundCheck, error) {
	return s.repository.GetByEmployeeID(ctx, employeeID)
}

// CancelCheck cancels an in-progress background check
func (s *BackgroundCheckService) CancelCheck(ctx context.Context, checkID string) error {
	check, err := s.repository.GetByID(ctx, checkID)
	if err != nil {
		return fmt.Errorf("failed to get check: %w", err)
	}

	if check.Status == models.BGCheckStatusCompleted || check.Status == models.BGCheckStatusCancelled {
		return fmt.Errorf("cannot cancel check in status: %s", check.Status)
	}

	provider, ok := s.providers[check.ProviderID]
	if !ok {
		return fmt.Errorf("provider %s not found", check.ProviderID)
	}

	var providerCheckID string
	if err := json.Unmarshal(check.ProviderData, &providerCheckID); err != nil {
		return fmt.Errorf("failed to get provider check ID: %w", err)
	}

	if err := provider.CancelCheck(ctx, providerCheckID); err != nil {
		return fmt.Errorf("failed to cancel with provider: %w", err)
	}

	check.Status = models.BGCheckStatusCancelled
	if err := s.repository.Update(ctx, check); err != nil {
		return fmt.Errorf("failed to update check: %w", err)
	}

	return nil
}

// CreatePackage creates a new background check package
func (s *BackgroundCheckService) CreatePackage(ctx context.Context, pkg *models.BackgroundCheckPackage) error {
	// Validate the package
	if err := pkg.Validate(); err != nil {
		return fmt.Errorf("invalid package: %w", err)
	}

	// Verify the provider exists
	if _, ok := s.providers[pkg.ProviderID]; !ok {
		return fmt.Errorf("provider %s not found", pkg.ProviderID)
	}

	// Set timestamps
	now := time.Now()
	pkg.CreatedAt = now
	pkg.UpdatedAt = now

	// Save to repository
	if err := s.repository.CreatePackage(ctx, pkg); err != nil {
		return fmt.Errorf("failed to create package: %w", err)
	}

	return nil
}

// ListPackages retrieves all background check packages
func (s *BackgroundCheckService) ListPackages(ctx context.Context) ([]*models.BackgroundCheckPackage, error) {
	packages, err := s.repository.ListPackages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	return packages, nil
}
