package service

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
)

// MockEmailService is a mock implementation for testing
type MockEmailService struct {
	SentEmails []SentEmail
}

type SentEmail struct {
	To           []string
	Subject      string
	Body         string
	TemplateName string
	TemplateData map[string]interface{}
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SentEmails: make([]SentEmail, 0),
	}
}

func (m *MockEmailService) SendEmail(
	ctx context.Context,
	to []string,
	subject string,
	body string,
	templateData map[string]interface{},
) error {
	log.Printf("MOCK: Sending email to %v with subject: %s", to, subject)
	
	m.SentEmails = append(m.SentEmails, SentEmail{
		To:           to,
		Subject:      subject,
		Body:         body,
		TemplateData: templateData,
	})
	
	return nil
}

func (m *MockEmailService) SendTemplatedEmail(
	ctx context.Context,
	to []string,
	templateName string,
	data map[string]interface{},
) error {
	log.Printf("MOCK: Sending templated email (%s) to %v", templateName, to)
	
	m.SentEmails = append(m.SentEmails, SentEmail{
		To:           to,
		TemplateName: templateName,
		TemplateData: data,
		Subject:      fmt.Sprintf("Template: %s", templateName),
	})
	
	return nil
}

// MockInAppNotificationService is a mock implementation for testing
type MockInAppNotificationService struct {
	Notifications []*InAppNotification
}

func NewMockInAppNotificationService() *MockInAppNotificationService {
	return &MockInAppNotificationService{
		Notifications: make([]*InAppNotification, 0),
	}
}

func (m *MockInAppNotificationService) CreateNotification(
	ctx context.Context,
	notification *InAppNotification,
) error {
	log.Printf("MOCK: Creating notification for user: %s", notification.UserID)
	m.Notifications = append(m.Notifications, notification)
	return nil
}

func (m *MockInAppNotificationService) CreateBulkNotifications(
	ctx context.Context,
	notifications []*InAppNotification,
) error {
	log.Printf("MOCK: Creating %d bulk notifications", len(notifications))
	m.Notifications = append(m.Notifications, notifications...)
	return nil
}

// BGCheckMockEmployeeRepository is a mock implementation for background check notification testing
// Named with BGCheck prefix to avoid conflicts with existing MockEmployeeRepository in service_test.go
type BGCheckMockEmployeeRepository struct {
	Employees   map[uuid.UUID]*models.Employee
	HRContacts  []*models.Employee
}

func NewBGCheckMockEmployeeRepository() *BGCheckMockEmployeeRepository {
	return &BGCheckMockEmployeeRepository{
		Employees:  make(map[uuid.UUID]*models.Employee),
		HRContacts: make([]*models.Employee, 0),
	}
}

// GetByID takes uuid.UUID to match the EmployeeRepository interface
func (m *BGCheckMockEmployeeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	employee, exists := m.Employees[id]
	if !exists {
		return nil, fmt.Errorf("employee not found: %s", id.String())
	}
	return employee, nil
}

func (m *BGCheckMockEmployeeRepository) GetHRContacts(ctx context.Context) ([]*models.Employee, error) {
	return m.HRContacts, nil
}

// GetManagerByEmployeeID takes uuid.UUID to match the interface
func (m *BGCheckMockEmployeeRepository) GetManagerByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.Employee, error) {
	employee, exists := m.Employees[employeeID]
	if !exists {
		return nil, fmt.Errorf("employee not found: %s", employeeID.String())
	}
	
	if employee.ManagerID == nil {
		return nil, fmt.Errorf("employee has no manager")
	}
	
	return m.GetByID(ctx, *employee.ManagerID)
}

// Helper function to create a test notification service
func NewTestNotificationService() *BackgroundCheckNotificationService {
	return NewBackgroundCheckNotificationService(
		NewMockEmailService(),
		NewMockInAppNotificationService(),
		NewBGCheckMockEmployeeRepository(),
	)
}