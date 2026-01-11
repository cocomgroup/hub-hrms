package service

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	Auth                AuthService
	User                UserService
	Employee            EmployeeService
	Onboarding          OnboardingService
	Workflow            WorkflowService
	Timesheet           TimesheetService
	PTO                 PTOService
	Benefits            BenefitsService
	Payroll             PayrollService
	Recruiting          RecruitingService
	Organization        OrganizationService
	Project             ProjectService
	Compensation        CompensationService
	BankInfo            BankInfoService
	BackgroundCheck     *BackgroundCheckService
	// Notification services
	Email               EmailService
	InAppNotification   InAppNotificationService
	BGCheckNotification NotificationService
}

func NewServices(repos *repository.Repositories, cfg *config.Config) *Services {
	// Get encryption key from config/environment
	encryptionKey := os.Getenv("BANK_INFO_ENCRYPTION_KEY")
	if encryptionKey == "" {
		log.Fatal("BANK_INFO_ENCRYPTION_KEY must be set in environment")
	}

	// Initialize notification services
	// For now, using mock services. In production, you would check environment
	// and initialize AWS SES and DynamoDB services accordingly.
	emailService := initializeEmailService()
	inAppService := initializeInAppNotificationService()
	
	// Create employee repository adapter for notification service
	employeeRepoAdapter := newEmployeeRepositoryAdapter(repos.Employee)
	
	// Create background check notification service
	bgCheckNotificationService := NewBackgroundCheckNotificationService(
		emailService,
		inAppService,
		employeeRepoAdapter,
	)
	
	// Create background check service with notification support
	bgCheckService := NewBackgroundCheckService(
		repos.BackgroundCheck,
		bgCheckNotificationService,
	)
	
	return &Services{
		Auth:                NewAuthService(repos, cfg),
		User:                NewUserService(repos),
		Employee:            NewEmployeeService(repos),
		Onboarding:          NewOnboardingService(repos),
		Workflow:            NewWorkflowService(repos),
		Timesheet:           NewTimesheetService(repos.Timesheet, repos.Project),
		PTO:                 NewPTOService(repos),
		Benefits:            NewBenefitsService(repos),
		Payroll:             NewPayrollService(repos),
		Recruiting:          NewRecruitingService(repos),
		Organization:        NewOrganizationService(repos),
		Project:             NewProjectService(repos),
		Compensation:        NewCompensationService(repos),
		BankInfo:            NewBankInfoService(repos, encryptionKey),
		BackgroundCheck:     bgCheckService,
		Email:               emailService,
		InAppNotification:   inAppService,
		BGCheckNotification: bgCheckNotificationService,
	}
}

// initializeEmailService creates the email service based on environment
func initializeEmailService() EmailService {
	// Check if we should use AWS SES or mock
	useMock := os.Getenv("USE_MOCK_EMAIL") == "true"
	
	if useMock {
		log.Println("Email service: Using mock implementation (development mode)")
		return NewMockEmailService()
	}
	
	// TODO: When you're ready for production, uncomment and configure AWS SES:
	/*
	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("WARNING: Failed to load AWS config: %v. Falling back to mock email service.", err)
		return NewMockEmailService()
	}
	
	sesClient := ses.NewFromConfig(awsConfig)
	fromEmail := os.Getenv("HRMS_FROM_EMAIL")
	replyToEmail := os.Getenv("HRMS_REPLY_TO_EMAIL")
	
	if fromEmail == "" || replyToEmail == "" {
		log.Println("WARNING: Email addresses not configured. Using mock email service.")
		return NewMockEmailService()
	}
	
	log.Printf("Email service: Using AWS SES (from: %s, reply-to: %s)", fromEmail, replyToEmail)
	return NewSESEmailService(sesClient, fromEmail, replyToEmail)
	*/
	
	// For now, default to mock
	log.Println("Email service: Using mock implementation (AWS SES not configured)")
	return NewMockEmailService()
}

// initializeInAppNotificationService creates the in-app notification service
func initializeInAppNotificationService() InAppNotificationService {
	// Check if we should use DynamoDB or mock
	useMock := os.Getenv("USE_MOCK_NOTIFICATIONS") == "true"
	
	if useMock {
		log.Println("In-app notification service: Using mock implementation (development mode)")
		return NewMockInAppNotificationService()
	}
	
	// TODO: When you're ready for production, uncomment and configure DynamoDB:
	/*
	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("WARNING: Failed to load AWS config: %v. Falling back to mock notification service.", err)
		return NewMockInAppNotificationService()
	}
	
	dynamoClient := dynamodb.NewFromConfig(awsConfig)
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		tableName = "hrms-notifications" // default table name
	}
	
	log.Printf("In-app notification service: Using DynamoDB (table: %s)", tableName)
	return NewDynamoDBInAppNotificationService(dynamoClient, tableName)
	*/
	
	// For now, default to mock
	log.Println("In-app notification service: Using mock implementation (DynamoDB not configured)")
	return NewMockInAppNotificationService()
}

// AuthService handles authentication and authorization
type AuthService interface {
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	HashPassword(password string) (string, error)
	CheckPassword(hashedPassword, password string) error
	GenerateToken(userID uuid.UUID, email, role string, employeeID *uuid.UUID) (string, error)
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

	// Get employee_id to include in JWT
	var employeeID *uuid.UUID
	if user.EmployeeID != nil {
		employeeID = user.EmployeeID
		log.Printf("Including employee_id in JWT: %s", employeeID.String())
	} else {
		log.Printf("WARNING: User has no employee_id, JWT will not include employee_id")
	}
	log.Printf("=== LOGIN DEBUG END ===")
	
	token, err := s.GenerateToken(user.ID, user.Email, user.Role, employeeID)
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

func (s *authService) GenerateToken(userID uuid.UUID, email, role string, employeeID *uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	// Include employee_id in JWT claims
	if employeeID != nil {
		claims["employee_id"] = employeeID.String()
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

// Helper functions
func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func float64Ptr(f float64) *float64 {
	return &f
}