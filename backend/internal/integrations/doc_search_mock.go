package integrations

import (
	"context"
	"fmt"
	"hub-hrms/backend/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DocSearchService handles document search from S3
type DocSearchService interface {
	SearchDocuments(ctx context.Context, req *DocSearchRequest) (*models.DocSearchMockResponse, error)
	GetDocumentURL(ctx context.Context, s3Key string) (string, error)
	GetDocumentMetadata(ctx context.Context, s3Key string) (*models.DocSearchDocument, error)
}

type DocSearchRequest struct {
	Query        string
	DocumentType string // handbook, policy, form, training, etc.
	FileType     string // pdf, docx, etc.
	Limit        int
}

type mockDocSearchService struct{}

// NewMockDocSearchService creates a mock document search service for testing
func NewMockDocSearchService() DocSearchService {
	return &mockDocSearchService{}
}

// Mock document repository
var mockDocuments = []models.DocSearchDocument{
	{
		ID:           uuid.New().String(),
		Name:         "Employee Handbook 2025.pdf",
		DocumentType: "handbook",
		S3Key:        "documents/handbooks/employee-handbook-2025.pdf",
		FileType:     "pdf",
		FileSize:     2048576,
		UploadedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Metadata: map[string]interface{}{
			"version": "2025.1",
			"department": "HR",
		},
	},
	{
		ID:           uuid.New().String(),
		Name:         "I-9 Employment Eligibility Form.pdf",
		DocumentType: "form",
		S3Key:        "documents/forms/i9-form.pdf",
		FileType:     "pdf",
		FileSize:     524288,
		UploadedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Metadata: map[string]interface{}{
			"required": true,
			"category": "onboarding",
		},
	},
	{
		ID:           uuid.New().String(),
		Name:         "W-4 Tax Withholding Form.pdf",
		DocumentType: "form",
		S3Key:        "documents/forms/w4-form.pdf",
		FileType:     "pdf",
		FileSize:     409600,
		UploadedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Metadata: map[string]interface{}{
			"required": true,
			"category": "onboarding",
		},
	},
	{
		ID:           uuid.New().String(),
		Name:         "Benefits Overview 2025.pdf",
		DocumentType: "policy",
		S3Key:        "documents/policies/benefits-overview-2025.pdf",
		FileType:     "pdf",
		FileSize:     1048576,
		UploadedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Metadata: map[string]interface{}{
			"year": 2025,
			"department": "Benefits",
		},
	},
	{
		ID:           uuid.New().String(),
		Name:         "IT Security Training.pdf",
		DocumentType: "training",
		S3Key:        "documents/training/it-security-training.pdf",
		FileType:     "pdf",
		FileSize:     3145728,
		UploadedAt:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Metadata: map[string]interface{}{
			"required": true,
			"duration_minutes": 45,
		},
	},
	{
		ID:           uuid.New().String(),
		Name:         "Company Code of Conduct.pdf",
		DocumentType: "policy",
		S3Key:        "documents/policies/code-of-conduct.pdf",
		FileType:     "pdf",
		FileSize:     819200,
		UploadedAt:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		Metadata: map[string]interface{}{
			"version": "3.0",
			"effective_date": "2024-12-01",
		},
	},
	{
		ID:           uuid.New().String(),
		Name:         "Direct Deposit Authorization Form.pdf",
		DocumentType: "form",
		S3Key:        "documents/forms/direct-deposit-form.pdf",
		FileType:     "pdf",
		FileSize:     204800,
		UploadedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Metadata: map[string]interface{}{
			"required": true,
			"category": "payroll",
		},
	},
}

func (s *mockDocSearchService) SearchDocuments(ctx context.Context, req *DocSearchRequest) (*models.DocSearchMockResponse, error) {
	// Simulate API call delay
	time.Sleep(600 * time.Millisecond)

	var results []models.DocSearchDocument

	// Filter documents based on search criteria
	queryLower := strings.ToLower(req.Query)
	for _, doc := range mockDocuments {
		// Check if document matches query
		nameMatch := strings.Contains(strings.ToLower(doc.Name), queryLower)
		
		// Check document type filter
		typeMatch := req.DocumentType == "" || doc.DocumentType == req.DocumentType
		
		// Check file type filter
		fileTypeMatch := req.FileType == "" || doc.FileType == req.FileType

		if (nameMatch || req.Query == "") && typeMatch && fileTypeMatch {
			results = append(results, doc)
		}
	}

	// Apply limit
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	if len(results) > limit {
		results = results[:limit]
	}

	response := &models.DocSearchMockResponse{
		Documents:  results,
		TotalCount: len(results),
	}

	return response, nil
}

func (s *mockDocSearchService) GetDocumentURL(ctx context.Context, s3Key string) (string, error) {
	// Simulate API call delay
	time.Sleep(200 * time.Millisecond)

	// Generate a mock presigned URL
	// In production, this would be an actual S3 presigned URL
	mockURL := fmt.Sprintf("https://mock-s3-bucket.s3.amazonaws.com/%s?expires=3600", s3Key)
	
	return mockURL, nil
}

func (s *mockDocSearchService) GetDocumentMetadata(ctx context.Context, s3Key string) (*models.DocSearchDocument, error) {
	// Simulate API call delay
	time.Sleep(300 * time.Millisecond)

	// Find document by S3 key
	for _, doc := range mockDocuments {
		if doc.S3Key == s3Key {
			return &doc, nil
		}
	}

	return nil, fmt.Errorf("document not found: %s", s3Key)
}
