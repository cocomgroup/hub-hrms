package helpers

import (
	"hub-hrms/backend/internal/repository"
	"hub-hrms/backend/tests/mocks"
)

// NewMockRepositories creates a Repositories struct with all mock implementations
func NewMockRepositories() (*repository.Repositories, *RepositoryMocks) {
	mocks := &RepositoryMocks{
		User:     &mocks.MockUserRepository{},
		Employee: &mocks.MockEmployeeRepository{},
		Applicant: &mocks.MockApplicantRepository{},
	}

	repos := &repository.Repositories{
		User:            mocks.User,
		Employee:        mocks.Employee,
		Applicant:       mocks.Applicant,
		Onboarding:      nil, // Add mocks as needed
		Workflow:        nil,
		Timesheet:       nil,
		PTO:             nil,
		Benefits:        nil,
		Payroll:         nil,
		Recruiting:      nil,
		Organization:    nil,
		Project:         nil,
		Compensation:    nil,
		BankInfo:        nil,
		BackgroundCheck: nil,
	}

	return repos, mocks
}

// RepositoryMocks holds references to all mock repositories for easy access in tests
type RepositoryMocks struct {
	User      *mocks.MockUserRepository
	Employee  *mocks.MockEmployeeRepository
	Applicant *mocks.MockApplicantRepository
	// Add more as needed
}

// NewMinimalMockRepositories creates a Repositories struct with only required mocks
// Useful for tests that only need specific repositories
func NewMinimalMockRepositories(needed ...string) *repository.Repositories {
	repos := &repository.Repositories{}

	for _, repo := range needed {
		switch repo {
		case "user":
			repos.User = &mocks.MockUserRepository{}
		case "employee":
			repos.Employee = &mocks.MockEmployeeRepository{}
		case "applicant":
			repos.Applicant = &mocks.MockApplicantRepository{}
		// Add more as needed
		}
	}

	return repos
}
