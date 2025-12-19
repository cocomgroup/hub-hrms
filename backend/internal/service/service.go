package service

import (
	"context"
	"errors"
	"fmt"
	"hub-hrms/backend/internal/config"
	"hub-hrms/backend/internal/models"
	"hub-hrms/backend/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmployeeNotFound   = errors.New("employee not found")
	ErrUnauthorized       = errors.New("unauthorized")
)

type Services struct {
	Auth       AuthService
	Employee   EmployeeService
	Onboarding OnboardingService
	Workflow   WorkflowService
	Timesheet  TimesheetService
	PTO        PTOService
	Benefits   BenefitsService
	Payroll    PayrollService
	Recruiting RecruitingService
}

func NewServices(repos *repository.Repositories, cfg *config.Config) *Services {
	return &Services{
		Auth:       NewAuthService(repos, cfg),
		Employee:   NewEmployeeService(repos),
		Onboarding: NewOnboardingService(repos),
		Workflow:   NewWorkflowService(repos),
		Timesheet:  NewTimesheetService(repos),
		PTO:        NewPTOService(repos),
		Benefits:   NewBenefitsService(repos),
		Payroll:    NewPayrollService(repos),
	}
}

// AuthService handles authentication and authorization
type AuthService interface {
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	HashPassword(password string) (string, error)
	CheckPassword(hashedPassword, password string) error
	GenerateToken(userID uuid.UUID, email, role string) (string, error)
}

type authService struct {
	repos *repository.Repositories
	cfg   *config.Config
}

func NewAuthService(repos *repository.Repositories, cfg *config.Config) AuthService {
	return &authService{repos: repos, cfg: cfg}
}

func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.repos.User.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := s.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	response := &models.LoginResponse{
		Token: token,
		User:  *user,
	}

	if user.EmployeeID != nil {
		employee, err := s.repos.Employee.GetByID(ctx, *user.EmployeeID)
		if err == nil {
			response.Employee = employee
		}
	}

	return response, nil
}

func (s *authService) GenerateToken(userID uuid.UUID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *authService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
}

func (s *authService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *authService) CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}


// OnboardingService handles onboarding operations
type OnboardingService interface {
	CreateTask(ctx context.Context, task *models.OnboardingTask) error
	GetTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.OnboardingTask, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error)
	UpdateTask(ctx context.Context, task *models.OnboardingTask) error
	CreateOnboardingPlan(ctx context.Context, employeeID uuid.UUID, department string) error
}

type onboardingService struct {
	repos *repository.Repositories
}

func NewOnboardingService(repos *repository.Repositories) OnboardingService {
	return &onboardingService{repos: repos}
}

func (s *onboardingService) CreateTask(ctx context.Context, task *models.OnboardingTask) error {
	return s.repos.Onboarding.CreateTask(ctx, task)
}

func (s *onboardingService) GetTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.OnboardingTask, error) {
	return s.repos.Onboarding.GetTasksByEmployee(ctx, employeeID)
}

func (s *onboardingService) GetTaskByID(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error) {
	return s.repos.Onboarding.GetTaskByID(ctx, id)
}

func (s *onboardingService) UpdateTask(ctx context.Context, task *models.OnboardingTask) error {
	if task.Status == "completed" && task.CompletedAt == nil {
		now := time.Now()
		task.CompletedAt = &now
	}
	return s.repos.Onboarding.UpdateTask(ctx, task)
}

func (s *onboardingService) CreateOnboardingPlan(ctx context.Context, employeeID uuid.UUID, department string) error {
	// Default onboarding tasks
	tasks := []models.OnboardingTask{
		{
			EmployeeID:        employeeID,
			TaskName:          "Complete I-9 Form",
			Description:       strPtr("Complete employment eligibility verification"),
			Category:          strPtr("HR Documents"),
			Status:            "pending",
			DueDate:           timePtr(time.Now().AddDate(0, 0, 3)),
			DocumentsRequired: true,
		},
		{
			EmployeeID:  employeeID,
			TaskName:    "Setup Direct Deposit",
			Description: strPtr("Provide bank account information for payroll"),
			Category:    strPtr("Payroll"),
			Status:      "pending",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 7)),
		},
		{
			EmployeeID:  employeeID,
			TaskName:    "Complete Benefits Enrollment",
			Description: strPtr("Select health insurance and other benefits"),
			Category:    strPtr("Benefits"),
			Status:      "pending",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 30)),
		},
		{
			EmployeeID:  employeeID,
			TaskName:    "IT Account Setup",
			Description: strPtr("Receive email, system access credentials"),
			Category:    strPtr("IT"),
			Status:      "pending",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 1)),
		},
		{
			EmployeeID:  employeeID,
			TaskName:    "Review Employee Handbook",
			Description: strPtr("Read and acknowledge company policies"),
			Category:    strPtr("HR Documents"),
			Status:      "pending",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 7)),
		},
	}

	for _, task := range tasks {
		if err := s.repos.Onboarding.CreateTask(ctx, &task); err != nil {
			return err
		}
	}

	return nil
}


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

// Helper functions
func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
