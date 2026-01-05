package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	User         UserRepository
	Employee     EmployeeRepository
	Onboarding   OnboardingRepository
	Workflow     WorkflowRepository
	Timesheet    TimesheetRepository
	PTO          PTORepository
	Benefits     BenefitsRepository
	Payroll      PayrollRepository
	Recruiting   RecruitingRepository
	Organization OrganizationRepository
	Project	     ProjectRepository
	Compensation CompensationRepository
	BankInfo     BankInfoRepository
}

func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		User:         NewUserRepository(db),
		Employee:     NewEmployeeRepository(db),
		Onboarding:   NewOnboardingRepository(db),
		Workflow:     NewWorkflowRepository(db),
		Timesheet:    NewTimesheetRepository(db),
		PTO:          NewPTORepository(db),
		Benefits:     NewBenefitsRepository(db),
		Payroll:      NewPayrollRepository(db),
		Recruiting:   NewRecruitingRepository(db),
		Organization: NewOrganizationRepository(db),
		Project:      NewProjectRepository(db),
		Compensation: NewCompensationRepository(db),
		BankInfo:     NewBankInfoRepository(db),
	}
}
