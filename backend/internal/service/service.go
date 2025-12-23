package service

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	User 	   UserService
	Employee   EmployeeService
	Onboarding OnboardingService
	Workflow   WorkflowService
	Timesheet  TimesheetService
	PTO        PTOService
	Benefits   BenefitsService
	Payroll    PayrollService
	Recruiting RecruitingService
	Organization OrganizationService
}

func NewServices(repos *repository.Repositories, cfg *config.Config) *Services {
	return &Services{
		Auth:       NewAuthService(repos, cfg),
		User:   	NewUserService(repos),
		Employee:   NewEmployeeService(repos),
		Onboarding: NewOnboardingService(repos),
		Workflow:   NewWorkflowService(repos),
		Timesheet:  NewTimesheetService(repos),
		PTO:        NewPTOService(repos),
		Benefits:   NewBenefitsService(repos),
		Payroll:    NewPayrollService(repos),
		Recruiting: NewRecruitingService(repos),
		Organization: NewOrganizationService(repos),
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
	log.Printf("=== LOGIN DEBUG START ===")
	log.Printf("Email from request: '%s'", req.Email)
	log.Printf("Password from request: '%s'", req.Password)
	log.Printf("Password length: %d", len(req.Password))

	user, err := s.repos.User.GetByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("GetByEmail ERROR: %v", err)
		return nil, ErrInvalidCredentials
	}

	log.Printf("User found in DB:")
	log.Printf("  - Email: '%s'", user.Email)
	log.Printf("  - Role: '%s'", user.Role)
	log.Printf("  - Hash: '%s'", user.PasswordHash)
	log.Printf("  - Hash length: %d", len(user.PasswordHash))

	if err := s.CheckPassword(user.PasswordHash, req.Password); err != nil {
		log.Printf("CheckPassword ERROR: %v", err)
		return nil, ErrInvalidCredentials
	}

	log.Printf("CheckPassword SUCCESS!")
	log.Printf("=== LOGIN DEBUG END ===")
	
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



// Helper functions
func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
