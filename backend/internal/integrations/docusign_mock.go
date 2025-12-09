package integrations

import (
	"context"
	"fmt"
	"hub-hrms/backend/internal/models"
	"time"

	"github.com/google/uuid"
)

// DocuSignService handles DocuSign API integration
type DocuSignService interface {
	SendEnvelope(ctx context.Context, req *DocuSignEnvelopeRequest) (*models.DocuSignMockResponse, error)
	GetEnvelopeStatus(ctx context.Context, envelopeID string) (*models.DocuSignMockResponse, error)
	VoidEnvelope(ctx context.Context, envelopeID string, reason string) error
}

type DocuSignEnvelopeRequest struct {
	DocumentType string
	SignerEmail  string
	SignerName   string
	EmployeeID   uuid.UUID
	Metadata     map[string]interface{}
}

type mockDocuSignService struct{}

// NewMockDocuSignService creates a mock DocuSign service for testing
func NewMockDocuSignService() DocuSignService {
	return &mockDocuSignService{}
}

func (s *mockDocuSignService) SendEnvelope(ctx context.Context, req *DocuSignEnvelopeRequest) (*models.DocuSignMockResponse, error) {
	// Simulate API call delay
	time.Sleep(500 * time.Millisecond)

	// Generate mock envelope ID
	envelopeID := fmt.Sprintf("mock-env-%s", uuid.New().String()[:8])

	response := &models.DocuSignMockResponse{
		EnvelopeID:  envelopeID,
		Status:      "sent",
		SentAt:      time.Now(),
		SignerEmail: req.SignerEmail,
	}

	return response, nil
}

func (s *mockDocuSignService) GetEnvelopeStatus(ctx context.Context, envelopeID string) (*models.DocuSignMockResponse, error) {
	// Simulate API call delay
	time.Sleep(300 * time.Millisecond)

	// Mock different statuses based on envelope ID (for testing)
	var status string
	var signedAt *time.Time

	// Simulate some envelopes being signed
	if len(envelopeID) > 15 {
		// Most envelopes are still pending
		status = "sent"
	} else {
		// Some are signed
		status = "completed"
		now := time.Now()
		signedAt = &now
	}

	response := &models.DocuSignMockResponse{
		EnvelopeID:  envelopeID,
		Status:      status,
		SentAt:      time.Now().Add(-24 * time.Hour), // Sent yesterday
		SignedAt:    signedAt,
		SignerEmail: "employee@example.com",
	}

	return response, nil
}

func (s *mockDocuSignService) VoidEnvelope(ctx context.Context, envelopeID string, reason string) error {
	// Simulate API call delay
	time.Sleep(300 * time.Millisecond)
	return nil
}
