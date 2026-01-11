package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"hub-hrms/backend/internal/repository"
	"hub-hrms/backend/internal/service"
	"hub-hrms/backend/tests/fixtures"
	"hub-hrms/backend/tests/mocks"
)

func TestApplicantService_Create_Success(t *testing.T) {
	// Arrange
	mockApplicantRepo := &mocks.MockApplicantRepository{}
	repos := &repository.Repositories{
		Applicant: mockApplicantRepo,
	}

	testApplicant := fixtures.NewApplicant()

	mockApplicantRepo.On("Create", mock.Anything, testApplicant).
		Return(nil)

	applicantService := service.NewApplicantService(repos)

	// Act
	err := applicantService.Create(context.Background(), testApplicant)

	// Assert
	assert.NoError(t, err)
	mockApplicantRepo.AssertExpectations(t)
}

func TestApplicantService_Create_RepositoryError(t *testing.T) {
	// Arrange
	mockApplicantRepo := &mocks.MockApplicantRepository{}
	repos := &repository.Repositories{
		Applicant: mockApplicantRepo,
	}

	testApplicant := fixtures.NewApplicant()

	mockApplicantRepo.On("Create", mock.Anything, testApplicant).
		Return(errors.New("database error"))

	applicantService := service.NewApplicantService(repos)

	// Act
	err := applicantService.Create(context.Background(), testApplicant)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockApplicantRepo.AssertExpectations(t)
}

func TestApplicantService_GetByID_Success(t *testing.T) {
	// Arrange
	mockApplicantRepo := &mocks.MockApplicantRepository{}
	repos := &repository.Repositories{
		Applicant: mockApplicantRepo,
	}

	testApplicant := fixtures.NewApplicant()
	applicantID := testApplicant.ID

	mockApplicantRepo.On("GetByID", mock.Anything, applicantID).
		Return(testApplicant, nil)

	applicantService := service.NewApplicantService(repos)

	// Act
	applicant, err := applicantService.GetByID(context.Background(), applicantID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, applicant)
	assert.Equal(t, applicantID, applicant.ID)
	assert.Equal(t, testApplicant.Email, applicant.Email)
	mockApplicantRepo.AssertExpectations(t)
}

func TestApplicantService_GetByID_NotFound(t *testing.T) {
	// Arrange
	mockApplicantRepo := &mocks.MockApplicantRepository{}
	repos := &repository.Repositories{
		Applicant: mockApplicantRepo,
	}

	applicantID := uuid.New()

	mockApplicantRepo.On("GetByID", mock.Anything, applicantID).
		Return(nil, errors.New("applicant not found"))

	applicantService := service.NewApplicantService(repos)

	// Act
	applicant, err := applicantService.GetByID(context.Background(), applicantID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, applicant)
	mockApplicantRepo.AssertExpectations(t)
}

func TestApplicantService_Update_Success(t *testing.T) {
	// Arrange
	mockApplicantRepo := &mocks.MockApplicantRepository{}
	repos := &repository.Repositories{
		Applicant: mockApplicantRepo,
	}

	testApplicant := fixtures.NewApplicantWithScore(85.5)
	testApplicant.Status = "interview"

	mockApplicantRepo.On("Update", mock.Anything, testApplicant).
		Return(nil)

	applicantService := service.NewApplicantService(repos)

	// Act
	err := applicantService.Update(context.Background(), testApplicant)

	// Assert
	assert.NoError(t, err)
	mockApplicantRepo.AssertExpectations(t)
}

func TestApplicantService_Delete_Success(t *testing.T) {
	// Arrange
	mockApplicantRepo := &mocks.MockApplicantRepository{}
	repos := &repository.Repositories{
		Applicant: mockApplicantRepo,
	}

	applicantID := uuid.New()

	mockApplicantRepo.On("Delete", mock.Anything, applicantID).
		Return(nil)

	applicantService := service.NewApplicantService(repos)

	// Act
	err := applicantService.Delete(context.Background(), applicantID)

	// Assert
	assert.NoError(t, err)
	mockApplicantRepo.AssertExpectations(t)
}
