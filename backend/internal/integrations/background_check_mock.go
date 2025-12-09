package integrations

import (
	"context"
	"fmt"
	"hub-hrms/backend/internal/models"
	"time"

	"github.com/google/uuid"
)

// BackgroundCheckService handles background check API integration
type BackgroundCheckService interface {
	InitiateCheck(ctx context.Context, req *BackgroundCheckRequest) (*models.BackgroundCheckMockResponse, error)
	GetCheckStatus(ctx context.Context, checkID string) (*models.BackgroundCheckMockResponse, error)
	CancelCheck(ctx context.Context, checkID string) error
}

type BackgroundCheckRequest struct {
	FirstName  string
	LastName   string
	Email      string
	DateOfBirth string
	CheckTypes []string // criminal, employment, education
	EmployeeID uuid.UUID
}

type mockBackgroundCheckService struct{}

// NewMockBackgroundCheckService creates a mock background check service for testing
func NewMockBackgroundCheckService() BackgroundCheckService {
	return &mockBackgroundCheckService{}
}

func (s *mockBackgroundCheckService) InitiateCheck(ctx context.Context, req *BackgroundCheckRequest) (*models.BackgroundCheckMockResponse, error) {
	// Simulate API call delay
	time.Sleep(800 * time.Millisecond)

	// Generate mock check ID
	checkID := fmt.Sprintf("mock-check-%s", uuid.New().String()[:8])

	candidateName := fmt.Sprintf("%s %s", req.FirstName, req.LastName)

	response := &models.BackgroundCheckMockResponse{
		CheckID:     checkID,
		Status:      "in-progress",
		Candidate:   candidateName,
		CheckTypes:  req.CheckTypes,
		InitiatedAt: time.Now(),
	}

	return response, nil
}

func (s *mockBackgroundCheckService) GetCheckStatus(ctx context.Context, checkID string) (*models.BackgroundCheckMockResponse, error) {
	// Simulate API call delay
	time.Sleep(400 * time.Millisecond)

	// Mock different statuses based on check ID (for testing)
	var status string
	var completedAt *time.Time
	var result string

	// Simulate progression: in-progress -> completed
	if len(checkID) > 20 {
		// Most checks are still in progress
		status = "in-progress"
	} else {
		// Some are completed
		status = "completed"
		now := time.Now()
		completedAt = &now
		result = "clear" // Most checks pass
	}

	response := &models.BackgroundCheckMockResponse{
		CheckID:     checkID,
		Status:      status,
		Candidate:   "John Doe",
		CheckTypes:  []string{"criminal", "employment"},
		InitiatedAt: time.Now().Add(-7 * 24 * time.Hour), // Initiated 7 days ago
		CompletedAt: completedAt,
		Result:      result,
	}

	return response, nil
}

func (s *mockBackgroundCheckService) CancelCheck(ctx context.Context, checkID string) error {
	// Simulate API call delay
	time.Sleep(300 * time.Millisecond)
	return nil
}
