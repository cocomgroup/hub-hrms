package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
)

// PTOService handles PTO operations
type PTOService interface {
	GetBalance(ctx context.Context, employeeID uuid.UUID) (*models.PTOBalance, error)
	CreateRequest(ctx context.Context, employeeID uuid.UUID, req *models.PTORequestCreate) (*models.PTORequest, error)
	ReviewRequest(ctx context.Context, requestID, reviewerID uuid.UUID, review *models.PTORequestReview) error
	GetRequestsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.PTORequest, error)
}

type ptoService struct {
	repos *repository.Repositories
}

func NewPTOService(repos *repository.Repositories) PTOService {
	return &ptoService{repos: repos}
}

func (s *ptoService) GetBalance(ctx context.Context, employeeID uuid.UUID) (*models.PTOBalance, error) {
	return s.repos.PTO.GetBalance(ctx, employeeID)
}

func (s *ptoService) CreateRequest(ctx context.Context, employeeID uuid.UUID, req *models.PTORequestCreate) (*models.PTORequest, error) {
	request := &models.PTORequest{
		EmployeeID:    employeeID,
		PTOType:       req.PTOType,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		DaysRequested: req.DaysRequested,
		Reason:        req.Reason,
		Status:        "pending",
	}

	if err := s.repos.PTO.CreateRequest(ctx, request); err != nil {
		return nil, err
	}

	return request, nil
}

func (s *ptoService) ReviewRequest(ctx context.Context, requestID, reviewerID uuid.UUID, review *models.PTORequestReview) error {
	request, err := s.repos.PTO.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}

	request.Status = review.Status
	request.ReviewedBy = &reviewerID
	now := time.Now()
	request.ReviewedAt = &now
	request.ReviewNotes = review.ReviewNotes

	if err := s.repos.PTO.UpdateRequest(ctx, request); err != nil {
		return err
	}

	// If approved, deduct from balance
	if review.Status == "approved" {
		balance, err := s.repos.PTO.GetBalance(ctx, request.EmployeeID)
		if err != nil {
			return err
		}

		switch request.PTOType {
		case "vacation":
			balance.VacationDays -= request.DaysRequested
		case "sick":
			balance.SickDays -= request.DaysRequested
		case "personal":
			balance.PersonalDays -= request.DaysRequested
		}

		return s.repos.PTO.UpdateBalance(ctx, balance)
	}

	return nil
}

func (s *ptoService) GetRequestsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.PTORequest, error) {
	return s.repos.PTO.GetRequestsByEmployee(ctx, employeeID)
}
