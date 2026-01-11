package fixtures

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"hub-hrms/backend/internal/models"
)

// User Fixtures

func NewUser() *models.User {
	return &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  MustHashPassword("password123"),
		FirstName: "Test",
		LastName:  "User",
		Role:      "employee",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func NewAdminUser() *models.User {
	user := NewUser()
	user.Email = "admin@example.com"
	user.Role = "admin"
	return user
}

func NewManagerUser() *models.User {
	user := NewUser()
	user.Email = "manager@example.com"
	user.Role = "manager"
	return user
}

// Employee Fixtures

func NewEmployee() *models.Employee {
	return &models.Employee{
		ID:         uuid.New(),
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john.doe@example.com",
		Phone:      "555-1234",
		Department: "Engineering",
		Position:   "Software Engineer",
		HireDate:   time.Now().AddDate(0, -6, 0), // 6 months ago
		Status:     "active",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func NewManagerEmployee() *models.Employee {
	emp := NewEmployee()
	emp.Email = "manager@example.com"
	emp.Position = "Engineering Manager"
	return emp
}

// Applicant Fixtures

func NewApplicant() *models.Applicant {
	return &models.Applicant{
		ID:          uuid.New(),
		Name:        "Sarah Martinez",
		Email:       "sarah.martinez@example.com",
		Phone:       "555-0147",
		Position:    "Marketing Coordinator",
		Source:      "manual",
		ResumeURL:   "/uploads/resumes/test-resume.pdf",
		AppliedDate: time.Now(),
		Status:      "new",
		AIScore:     0.0,
		Notes:       "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func NewApplicantWithScore(score float64) *models.Applicant {
	applicant := NewApplicant()
	applicant.AIScore = score
	applicant.Status = "reviewing"
	return applicant
}

// Helper Functions

func MustHashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

// Batch Fixtures

func NewEmployeeList(count int) []*models.Employee {
	employees := make([]*models.Employee, count)
	for i := 0; i < count; i++ {
		employees[i] = NewEmployee()
		employees[i].ID = uuid.New()
		employees[i].Email = uuid.New().String() + "@example.com"
	}
	return employees
}

func NewApplicantList(count int) []*models.Applicant {
	applicants := make([]*models.Applicant, count)
	for i := 0; i < count; i++ {
		applicants[i] = NewApplicant()
		applicants[i].ID = uuid.New()
		applicants[i].Email = uuid.New().String() + "@example.com"
	}
	return applicants
}
