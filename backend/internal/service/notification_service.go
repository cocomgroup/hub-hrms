package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"hub-hrms/backend/internal/models"
)

// NotificationService handles notifications for background checks
type NotificationService interface {
	NotifyCheckInitiated(ctx context.Context, check *models.BackgroundCheck) error
	NotifyCheckCompleted(ctx context.Context, check *models.BackgroundCheck) error
	NotifyAdverseAction(ctx context.Context, check *models.BackgroundCheck) error
	NotifyHR(ctx context.Context, subject, message string, check *models.BackgroundCheck) error
}

// BackgroundCheckNotificationService implements NotificationService
type BackgroundCheckNotificationService struct {
	emailService EmailService
	inAppService InAppNotificationService
	employeeRepo EmployeeRepository
}

// EmailService defines the interface for sending emails
type EmailService interface {
	SendEmail(ctx context.Context, to []string, subject, body string, templateData map[string]interface{}) error
	SendTemplatedEmail(ctx context.Context, to []string, templateName string, data map[string]interface{}) error
}

// InAppNotificationService defines the interface for in-app notifications
type InAppNotificationService interface {
	CreateNotification(ctx context.Context, notification *InAppNotification) error
	CreateBulkNotifications(ctx context.Context, notifications []*InAppNotification) error
}

// EmployeeRepository defines methods for retrieving employee information
// This is an adapter interface that wraps repository.EmployeeRepository
type EmployeeRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error)
	GetHRContacts(ctx context.Context) ([]*models.Employee, error)
	GetManagerByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.Employee, error)
}

// employeeRepositoryAdapter adapts repository.EmployeeRepository to include HR-specific methods
type employeeRepositoryAdapter struct {
	repo interface {
		GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error)
		List(ctx context.Context, filters map[string]interface{}) ([]*models.Employee, error)
	}
}

// newEmployeeRepositoryAdapter creates an adapter for the employee repository
func newEmployeeRepositoryAdapter(repo interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*models.Employee, error)
}) EmployeeRepository {
	return &employeeRepositoryAdapter{repo: repo}
}

func (a *employeeRepositoryAdapter) GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	return a.repo.GetByID(ctx, id)
}

func (a *employeeRepositoryAdapter) GetHRContacts(ctx context.Context) ([]*models.Employee, error) {
	// Get employees with HR role
	// Adjust the filter based on your actual employee model structure
	filters := map[string]interface{}{
		"department": "Human Resources",
	}
	
	employees, err := a.repo.List(ctx, filters)
	if err != nil {
		return nil, err
	}
	
	// If no HR department employees found, try to find by position
	if len(employees) == 0 {
		filters = map[string]interface{}{
			"position": "HR Manager",
		}
		employees, err = a.repo.List(ctx, filters)
		if err != nil {
			return nil, err
		}
	}
	
	return employees, nil
}

func (a *employeeRepositoryAdapter) GetManagerByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.Employee, error) {
	employee, err := a.repo.GetByID(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	
	if employee.ManagerID == nil {
		return nil, fmt.Errorf("employee has no manager")
	}
	
	return a.repo.GetByID(ctx, *employee.ManagerID)
}

// InAppNotification represents an in-app notification
type InAppNotification struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Type         string                 `json:"type"`
	Title        string                 `json:"title"`
	Message      string                 `json:"message"`
	Data         map[string]interface{} `json:"data"`
	Read         bool                   `json:"read"`
	ActionURL    string                 `json:"action_url,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
}

// NotificationTemplate names
const (
	TemplateCheckInitiated      = "background_check_initiated"
	TemplateCheckCompleted      = "background_check_completed"
	TemplateAdverseActionPre    = "background_check_adverse_action_pre"
	TemplateAdverseActionFinal  = "background_check_adverse_action_final"
	TemplateCheckFailed         = "background_check_failed"
)

// NewBackgroundCheckNotificationService creates a new notification service
func NewBackgroundCheckNotificationService(
	emailService EmailService,
	inAppService InAppNotificationService,
	employeeRepo EmployeeRepository,
) *BackgroundCheckNotificationService {
	return &BackgroundCheckNotificationService{
		emailService: emailService,
		inAppService: inAppService,
		employeeRepo: employeeRepo,
	}
}

// NotifyCheckInitiated sends notifications when a background check is initiated
func (s *BackgroundCheckNotificationService) NotifyCheckInitiated(
	ctx context.Context,
	check *models.BackgroundCheck,
) error {
	log.Printf("Sending check initiated notifications for check: %s", check.ID)

	// Parse employee ID to UUID
	employeeID, err := uuid.Parse(check.EmployeeID)
	if err != nil {
		log.Printf("ERROR: Invalid employee ID: %v", err)
		return fmt.Errorf("invalid employee ID: %w", err)
	}

	// Get employee details
	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		log.Printf("ERROR: Failed to get employee: %v", err)
		return fmt.Errorf("failed to get employee: %w", err)
	}

	// Send email to candidate
	if err := s.sendCandidateInitiatedEmail(ctx, check, employee); err != nil {
		log.Printf("ERROR: Failed to send candidate email: %v", err)
		// Don't fail the operation if email fails
	}

	// Create in-app notification for HR
	if err := s.createHRNotification(
		ctx,
		"Background Check Initiated",
		fmt.Sprintf("Background check initiated for %s %s", 
			employee.FirstName, employee.LastName),
		check,
	); err != nil {
		log.Printf("ERROR: Failed to create HR notification: %v", err)
	}

	// Notify manager if exists
	if employee.ManagerID != nil {
		if err := s.notifyManager(
			ctx,
			employee.ManagerID.String(),
			"Background Check Initiated",
			fmt.Sprintf("Background check initiated for your team member %s %s",
				employee.FirstName, employee.LastName),
			check,
		); err != nil {
			log.Printf("ERROR: Failed to notify manager: %v", err)
		}
	}

	log.Printf("Check initiated notifications sent successfully")
	return nil
}

// NotifyCheckCompleted sends notifications when a background check is completed
func (s *BackgroundCheckNotificationService) NotifyCheckCompleted(
	ctx context.Context,
	check *models.BackgroundCheck,
) error {
	log.Printf("Sending check completed notifications for check: %s", check.ID)

	// Parse employee ID to UUID
	employeeID, err := uuid.Parse(check.EmployeeID)
	if err != nil {
		log.Printf("ERROR: Invalid employee ID: %v", err)
		return fmt.Errorf("invalid employee ID: %w", err)
	}

	// Get employee details
	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		log.Printf("ERROR: Failed to get employee: %v", err)
		return fmt.Errorf("failed to get employee: %w", err)
	}

	// Send email based on result
	if err := s.sendCompletionEmail(ctx, check, employee); err != nil {
		log.Printf("ERROR: Failed to send completion email: %v", err)
	}

	// Create in-app notifications
	if err := s.createCompletionNotifications(ctx, check, employee); err != nil {
		log.Printf("ERROR: Failed to create completion notifications: %v", err)
	}

	// If result is not clear, trigger adverse action process
	if check.Result == models.ResultConsider || check.Result == models.ResultSuspended {
		log.Printf("Result requires consideration - triggering adverse action process")
		if err := s.NotifyAdverseAction(ctx, check); err != nil {
			log.Printf("ERROR: Failed to send adverse action notification: %v", err)
		}
	}

	log.Printf("Check completed notifications sent successfully")
	return nil
}

// NotifyAdverseAction sends adverse action notifications (required by FCRA)
func (s *BackgroundCheckNotificationService) NotifyAdverseAction(
	ctx context.Context,
	check *models.BackgroundCheck,
) error {
	log.Printf("Sending adverse action notification for check: %s", check.ID)

	// Parse employee ID to UUID
	employeeID, err := uuid.Parse(check.EmployeeID)
	if err != nil {
		return fmt.Errorf("invalid employee ID: %w", err)
	}

	// Get employee details
	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		return fmt.Errorf("failed to get employee: %w", err)
	}

	// FCRA requires pre-adverse action notice
	// This gives the candidate time to dispute the findings
	if err := s.sendPreAdverseActionNotice(ctx, check, employee); err != nil {
		log.Printf("ERROR: Failed to send pre-adverse action notice: %v", err)
		return fmt.Errorf("failed to send pre-adverse action notice: %w", err)
	}

	// Notify HR about adverse action
	if err := s.createHRNotification(
		ctx,
		"Adverse Action Required",
		fmt.Sprintf("Background check for %s %s requires adverse action consideration",
			employee.FirstName, employee.LastName),
		check,
	); err != nil {
		log.Printf("ERROR: Failed to notify HR: %v", err)
	}

	log.Printf("Adverse action notifications sent successfully")
	return nil
}

// NotifyHR sends a notification to HR team
func (s *BackgroundCheckNotificationService) NotifyHR(
	ctx context.Context,
	subject string,
	message string,
	check *models.BackgroundCheck,
) error {
	log.Printf("Sending HR notification: %s", subject)

	// Get HR contacts
	hrContacts, err := s.employeeRepo.GetHRContacts(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to get HR contacts: %v", err)
		return fmt.Errorf("failed to get HR contacts: %w", err)
	}

	if len(hrContacts) == 0 {
		log.Printf("WARNING: No HR contacts found")
		return nil
	}

	// Send emails to HR
	hrEmails := make([]string, 0, len(hrContacts))
	for _, contact := range hrContacts {
		hrEmails = append(hrEmails, contact.Email)
	}

	templateData := map[string]interface{}{
		"subject":        subject,
		"message":        message,
		"check_id":       check.ID,
		"employee_id":    check.EmployeeID,
		"status":         check.Status,
		"result":         check.Result,
		"dashboard_url":  fmt.Sprintf("https://yourdomain.com/hr/background-checks/%s", check.ID),
	}

	if err := s.emailService.SendEmail(ctx, hrEmails, subject, message, templateData); err != nil {
		log.Printf("ERROR: Failed to send HR emails: %v", err)
		return fmt.Errorf("failed to send HR emails: %w", err)
	}

	// Create in-app notifications for HR
	if err := s.createHRNotification(ctx, subject, message, check); err != nil {
		log.Printf("ERROR: Failed to create HR in-app notifications: %v", err)
	}

	return nil
}

// Helper methods

func (s *BackgroundCheckNotificationService) sendCandidateInitiatedEmail(
	ctx context.Context,
	check *models.BackgroundCheck,
	employee *models.Employee,
) error {
	templateData := map[string]interface{}{
		"first_name":      check.CandidateInfo.FirstName,
		"last_name":       check.CandidateInfo.LastName,
		"company_name":    "Your Company", // Get from config
		"check_types":     formatCheckTypes(check.CheckTypes),
		"estimated_eta":   formatETA(check.EstimatedETA),
		"candidate_url":   "", // Provider-specific URL if available
		"support_email":   "hr@yourcompany.com",
		"initiated_date":  check.InitiatedAt.Format("January 2, 2006"),
	}

	return s.emailService.SendTemplatedEmail(
		ctx,
		[]string{check.CandidateInfo.Email},
		TemplateCheckInitiated,
		templateData,
	)
}

func (s *BackgroundCheckNotificationService) sendCompletionEmail(
	ctx context.Context,
	check *models.BackgroundCheck,
	employee *models.Employee,
) error {
	var templateName string

	switch check.Result {
	case models.ResultClear:
		templateName = TemplateCheckCompleted
	case models.ResultConsider, models.ResultSuspended:
		templateName = TemplateCheckCompleted
	default:
		templateName = TemplateCheckFailed
	}

	templateData := map[string]interface{}{
		"first_name":     check.CandidateInfo.FirstName,
		"last_name":      check.CandidateInfo.LastName,
		"result":         string(check.Result),
		"completed_date": formatCompletedDate(check.CompletedAt),
		"report_url":     check.ReportURL,
		"support_email":  "hr@yourcompany.com",
	}

	return s.emailService.SendTemplatedEmail(
		ctx,
		[]string{check.CandidateInfo.Email},
		templateName,
		templateData,
	)
}

func (s *BackgroundCheckNotificationService) sendPreAdverseActionNotice(
	ctx context.Context,
	check *models.BackgroundCheck,
	employee *models.Employee,
) error {
	// FCRA requires specific information in pre-adverse action notices
	templateData := map[string]interface{}{
		"first_name":              check.CandidateInfo.FirstName,
		"last_name":               check.CandidateInfo.LastName,
		"company_name":            "Your Company",
		"report_url":              check.ReportURL,
		"provider_name":           "Checkr", // Get from config
		"provider_address":        "1 Montgomery St., Suite 2400, San Francisco, CA 94104",
		"provider_phone":          "1-844-824-3257",
		"provider_website":        "https://checkr.com",
		"dispute_deadline":        time.Now().Add(5 * 24 * time.Hour).Format("January 2, 2006"),
		"support_email":           "hr@yourcompany.com",
		"fcra_summary_of_rights":  getFCRASummary(),
	}

	return s.emailService.SendTemplatedEmail(
		ctx,
		[]string{check.CandidateInfo.Email},
		TemplateAdverseActionPre,
		templateData,
	)
}

func (s *BackgroundCheckNotificationService) createCompletionNotifications(
	ctx context.Context,
	check *models.BackgroundCheck,
	employee *models.Employee,
) error {
	// Notify HR
	title := fmt.Sprintf("Background Check Completed - %s", check.Result)
	message := fmt.Sprintf("Background check for %s %s has been completed with result: %s",
		employee.FirstName, employee.LastName, check.Result)

	if err := s.createHRNotification(ctx, title, message, check); err != nil {
		return err
	}

	// Notify manager
	if employee.ManagerID != nil {
		if err := s.notifyManager(ctx, employee.ManagerID.String(), title, message, check); err != nil {
			log.Printf("ERROR: Failed to notify manager: %v", err)
		}
	}

	return nil
}

func (s *BackgroundCheckNotificationService) createHRNotification(
	ctx context.Context,
	title string,
	message string,
	check *models.BackgroundCheck,
) error {
	// Get HR contacts
	hrContacts, err := s.employeeRepo.GetHRContacts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get HR contacts: %w", err)
	}

	if len(hrContacts) == 0 {
		log.Printf("WARNING: No HR contacts found for notification")
		return nil
	}

	// Create notifications for each HR contact
	notifications := make([]*InAppNotification, 0, len(hrContacts))
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days

	for _, contact := range hrContacts {
		notification := &InAppNotification{
			ID:        fmt.Sprintf("bgcheck_%s_%s", check.ID, contact.ID.String()),
			UserID:    contact.ID.String(),
			Type:      "background_check",
			Title:     title,
			Message:   message,
			Data: map[string]interface{}{
				"check_id":    check.ID,
				"employee_id": check.EmployeeID,
				"status":      check.Status,
				"result":      check.Result,
			},
			Read:      false,
			ActionURL: fmt.Sprintf("/hr/background-checks/%s", check.ID),
			CreatedAt: time.Now(),
			ExpiresAt: &expiresAt,
		}
		notifications = append(notifications, notification)
	}

	return s.inAppService.CreateBulkNotifications(ctx, notifications)
}

func (s *BackgroundCheckNotificationService) notifyManager(
	ctx context.Context,
	managerID string,
	title string,
	message string,
	check *models.BackgroundCheck,
) error {
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	notification := &InAppNotification{
		ID:        fmt.Sprintf("bgcheck_%s_mgr_%s", check.ID, managerID),
		UserID:    managerID,
		Type:      "background_check",
		Title:     title,
		Message:   message,
		Data: map[string]interface{}{
			"check_id":    check.ID,
			"employee_id": check.EmployeeID,
			"status":      check.Status,
		},
		Read:      false,
		ActionURL: fmt.Sprintf("/team/background-checks/%s", check.ID),
		CreatedAt: time.Now(),
		ExpiresAt: &expiresAt,
	}

	return s.inAppService.CreateNotification(ctx, notification)
}

// Utility functions

func formatCheckTypes(checkTypes []models.BackgroundCheckType) string {
	if len(checkTypes) == 0 {
		return "Standard background check"
	}

	formatted := make([]string, len(checkTypes))
	for i, ct := range checkTypes {
		formatted[i] = formatCheckType(ct)
	}

	if len(formatted) == 1 {
		return formatted[0]
	}

	// Join all but last with comma, then add "and" before last
	allButLast := formatted[:len(formatted)-1]
	last := formatted[len(formatted)-1]
	return fmt.Sprintf("%s, and %s", joinStrings(allButLast, ", "), last)
}

func formatCheckType(ct models.BackgroundCheckType) string {
	switch ct {
	case models.CheckTypeCriminal:
		return "Criminal record check"
	case models.CheckTypeEmployment:
		return "Employment verification"
	case models.CheckTypeEducation:
		return "Education verification"
	case models.CheckTypeCredit:
		return "Credit history check"
	case models.CheckTypeDrugScreen:
		return "Drug screening"
	case models.CheckTypeReference:
		return "Reference check"
	case models.CheckTypeIdentity:
		return "Identity verification"
	default:
		return string(ct)
	}
}

func formatETA(eta *time.Time) string {
	if eta == nil {
		return "3-5 business days"
	}
	return eta.Format("January 2, 2006")
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func getFCRASummary() string {
	return `Summary of Your Rights Under the Fair Credit Reporting Act

The federal Fair Credit Reporting Act (FCRA) promotes the accuracy, fairness, and privacy of information in the files of consumer reporting agencies. There are many types of consumer reporting agencies, including credit bureaus and specialty agencies (such as agencies that sell information about check writing histories, medical records, and rental history records). Here is a summary of your major rights under FCRA.

You must be told if information in your file has been used against you.
You have the right to know what is in your file.
You have the right to ask for a credit score.
You have the right to dispute incomplete or inaccurate information.
Consumer reporting agencies must correct or delete inaccurate, incomplete, or unverifiable information.
Consumer reporting agencies may not report outdated negative information.
Access to your file is limited.
You must give your consent for reports to be provided to employers.
You may limit "prescreened" offers of credit and insurance you get based on information in your credit report.
You may seek damages from violators.`
}

// Helper function to safely format completed date
func formatCompletedDate(completedAt *time.Time) string {
	if completedAt == nil {
		return time.Now().Format("January 2, 2006")
	}
	return completedAt.Format("January 2, 2006")
}