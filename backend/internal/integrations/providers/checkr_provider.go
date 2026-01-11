package providers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"hub-hrms/backend/internal/models"
)

const (
	checkrAPIBaseURL = "https://api.checkr.com/v1"
)

// CheckrProvider implements the BackgroundCheckProvider interface for Checkr
type CheckrProvider struct {
	apiKey        string
	webhookSecret string
	client        *http.Client
}

// NewCheckrProvider creates a new Checkr provider instance
func NewCheckrProvider(apiKey, webhookSecret string) *CheckrProvider {
	return &CheckrProvider{
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the provider's name
func (p *CheckrProvider) GetName() string {
	return "checkr"
}

// CheckrCandidate represents a Checkr candidate
type CheckrCandidate struct {
	ID            string `json:"id,omitempty"`
	FirstName     string `json:"first_name"`
	MiddleName    string `json:"middle_name,omitempty"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	DOB           string `json:"dob"`
	SSN           string `json:"ssn,omitempty"`
	DriverLicense string `json:"driver_license_number,omitempty"`
	DriverLicenseState string `json:"driver_license_state,omitempty"`
	Zipcode       string `json:"zipcode"`
	WorkLocations string `json:"work_locations,omitempty"`
}

// CheckrReport represents a Checkr report (background check)
type CheckrReport struct {
	ID              string                 `json:"id"`
	Status          string                 `json:"status"`
	Package         string                 `json:"package"`
	CandidateID     string                 `json:"candidate_id"`
	CreatedAt       time.Time              `json:"created_at"`
	CompletedAt     *time.Time             `json:"completed_at"`
	TurnaroundTime  int                    `json:"turnaround_time"`
	Adjudication    string                 `json:"adjudication,omitempty"`
	ReportURL       string                 `json:"report_url,omitempty"`
	Documents       []CheckrDocument       `json:"documents,omitempty"`
	ETA             *time.Time             `json:"eta,omitempty"`
}

// CheckrDocument represents a document in a Checkr report
type CheckrDocument struct {
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Result    string    `json:"result,omitempty"`
}

// CheckrInvitation represents a candidate invitation
type CheckrInvitation struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	InvitationURL string  `json:"invitation_url"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// CheckrWebhook represents a Checkr webhook payload
type CheckrWebhook struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Object    string                 `json:"object"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

// InitiateCheck starts a new background check with Checkr
func (p *CheckrProvider) InitiateCheck(
	ctx context.Context,
	req *models.BackgroundCheckRequest,
) (*models.ProviderResponse, error) {
	// Step 1: Create or get candidate
	candidate := &CheckrCandidate{
		FirstName:  req.CandidateInfo.FirstName,
		MiddleName: req.CandidateInfo.MiddleName,
		LastName:   req.CandidateInfo.LastName,
		Email:      req.CandidateInfo.Email,
		Phone:      req.CandidateInfo.Phone,
		DOB:        req.CandidateInfo.DateOfBirth.Format("2006-01-02"),
		SSN:        req.CandidateInfo.SSN,
		DriverLicense: req.CandidateInfo.DriverLicense,
		Zipcode:    req.CandidateInfo.Address.PostalCode,
	}

	candidateResp, err := p.createCandidate(ctx, candidate)
	if err != nil {
		return nil, fmt.Errorf("failed to create candidate: %w", err)
	}

	// Step 2: Create invitation for candidate to complete their portion
	invitation, err := p.createInvitation(ctx, candidateResp.ID, req.PackageID, req.CallbackURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Step 3: Create report
	report, err := p.createReport(ctx, candidateResp.ID, req.PackageID)
	if err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	// Map Checkr status to our status
	status := mapCheckrStatus(report.Status)

	resp := &models.ProviderResponse{
		ProviderCheckID: report.ID,
		Status:          status,
		EstimatedETA:    report.ETA,
		CandidateURL:    invitation.InvitationURL,
		ReportURL:       report.ReportURL,
	}

	// Store the raw response
	rawData, _ := json.Marshal(map[string]interface{}{
		"report":     report,
		"candidate":  candidateResp,
		"invitation": invitation,
	})
	resp.RawData = rawData

	return resp, nil
}

// GetCheckStatus retrieves the current status of a check from Checkr
func (p *CheckrProvider) GetCheckStatus(
	ctx context.Context,
	providerCheckID string,
) (*models.ProviderStatus, error) {
	report, err := p.getReport(ctx, providerCheckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	status := &models.ProviderStatus{
		Status:      mapCheckrStatus(report.Status),
		Result:      mapCheckrAdjudication(report.Adjudication),
		CompletedAt: report.CompletedAt,
		ReportURL:   report.ReportURL,
		Updates:     []models.StatusUpdate{},
	}

	rawData, _ := json.Marshal(report)
	status.RawData = rawData

	return status, nil
}

// GetReport retrieves the full report for a completed check
func (p *CheckrProvider) GetReport(
	ctx context.Context,
	providerCheckID string,
) (*models.ProviderReport, error) {
	report, err := p.getReport(ctx, providerCheckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	checkResults := make(map[models.BackgroundCheckType]models.CheckDetail)
	
	// Parse documents into check results
	for _, doc := range report.Documents {
		checkType := mapCheckrDocumentType(doc.Type)
		checkResults[checkType] = models.CheckDetail{
			Result:      mapCheckrResult(doc.Result),
			Status:      mapCheckrStatus(doc.Status),
			CompletedAt: doc.CompletedAt,
			Message:     fmt.Sprintf("%s check completed", doc.Type),
		}
	}

	providerReport := &models.ProviderReport{
		Overall:      mapCheckrAdjudication(report.Adjudication),
		CheckResults: checkResults,
		ReportURL:    report.ReportURL,
	}

	rawData, _ := json.Marshal(report)
	providerReport.RawData = rawData

	return providerReport, nil
}

// CancelCheck cancels an in-progress check
func (p *CheckrProvider) CancelCheck(ctx context.Context, providerCheckID string) error {
	url := fmt.Sprintf("%s/reports/%s", checkrAPIBaseURL, providerCheckID)
	
	payload := map[string]string{
		"status": "suspended",
	}
	
	if err := p.makeRequest(ctx, "PUT", url, payload, nil); err != nil {
		return fmt.Errorf("failed to cancel report: %w", err)
	}

	return nil
}

// SubmitDispute submits a dispute for a check result
func (p *CheckrProvider) SubmitDispute(
	ctx context.Context,
	providerCheckID string,
	dispute *models.DisputeRequest,
) error {
	url := fmt.Sprintf("%s/reports/%s/disputes", checkrAPIBaseURL, providerCheckID)
	
	payload := map[string]interface{}{
		"reason":   dispute.Reason,
		"evidence": dispute.Evidence,
	}
	
	if err := p.makeRequest(ctx, "POST", url, payload, nil); err != nil {
		return fmt.Errorf("failed to submit dispute: %w", err)
	}

	return nil
}

// ValidateWebhook validates incoming webhook signatures
func (p *CheckrProvider) ValidateWebhook(payload []byte, signature string) error {
	mac := hmac.New(sha256.New, []byte(p.webhookSecret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
		return fmt.Errorf("invalid webhook signature")
	}

	return nil
}

// ParseWebhook parses webhook data into standard format
func (p *CheckrProvider) ParseWebhook(payload []byte) (*models.WebhookData, error) {
	var webhook CheckrWebhook
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse webhook: %w", err)
	}

	// Extract report ID from data
	reportID, _ := webhook.Data["id"].(string)
	status, _ := webhook.Data["status"].(string)
	adjudication, _ := webhook.Data["adjudication"].(string)

	webhookData := &models.WebhookData{
		ProviderCheckID: reportID,
		Event:           webhook.Type,
		Status:          mapCheckrStatus(status),
		Result:          mapCheckrAdjudication(adjudication),
		Timestamp:       webhook.CreatedAt,
		Data:            json.RawMessage(payload),
	}

	return webhookData, nil
}

// Helper methods for API calls

func (p *CheckrProvider) createCandidate(ctx context.Context, candidate *CheckrCandidate) (*CheckrCandidate, error) {
	url := fmt.Sprintf("%s/candidates", checkrAPIBaseURL)
	
	var resp CheckrCandidate
	if err := p.makeRequest(ctx, "POST", url, candidate, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (p *CheckrProvider) createInvitation(
	ctx context.Context,
	candidateID, packageSlug, callbackURL string,
) (*CheckrInvitation, error) {
	url := fmt.Sprintf("%s/invitations", checkrAPIBaseURL)
	
	payload := map[string]string{
		"candidate_id": candidateID,
		"package":      packageSlug,
		"callback_url": callbackURL,
	}

	var resp CheckrInvitation
	if err := p.makeRequest(ctx, "POST", url, payload, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (p *CheckrProvider) createReport(
	ctx context.Context,
	candidateID, packageSlug string,
) (*CheckrReport, error) {
	url := fmt.Sprintf("%s/reports", checkrAPIBaseURL)
	
	payload := map[string]string{
		"candidate_id": candidateID,
		"package":      packageSlug,
	}

	var resp CheckrReport
	if err := p.makeRequest(ctx, "POST", url, payload, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (p *CheckrProvider) getReport(ctx context.Context, reportID string) (*CheckrReport, error) {
	url := fmt.Sprintf("%s/reports/%s", checkrAPIBaseURL, reportID)
	
	var resp CheckrReport
	if err := p.makeRequest(ctx, "GET", url, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (p *CheckrProvider) makeRequest(
	ctx context.Context,
	method, url string,
	payload interface{},
	result interface{},
) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(p.apiKey, "")
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Mapping functions

func mapCheckrStatus(status string) models.BackgroundCheckStatus {
	switch status {
	case "pending":
		return models.BGCheckStatusPending
	case "complete":
		return models.BGCheckStatusCompleted
	case "suspended":
		return models.BGCheckStatusCancelled
	case "dispute":
		return models.BGCheckStatusInProgress
	default:
		return models.BGCheckStatusInProgress
	}
}

func mapCheckrAdjudication(adjudication string) models.BackgroundCheckResult {
	switch adjudication {
	case "clear":
		return models.ResultClear
	case "consider":
		return models.ResultConsider
	case "suspended":
		return models.ResultSuspended
	case "dispute":
		return models.ResultDispute
	default:
		return models.ResultClear
	}
}

func mapCheckrResult(result string) models.BackgroundCheckResult {
	switch result {
	case "clear":
		return models.ResultClear
	case "consider":
		return models.ResultConsider
	default:
		return models.ResultClear
	}
}

func mapCheckrDocumentType(docType string) models.BackgroundCheckType {
	switch docType {
	case "ssn_trace", "national_criminal_search", "sex_offender_search":
		return models.CheckTypeCriminal
	case "employment_verification":
		return models.CheckTypeEmployment
	case "education_verification":
		return models.CheckTypeEducation
	case "motor_vehicle_report":
		return models.CheckTypeIdentity
	default:
		return models.CheckTypeCriminal
	}
}
