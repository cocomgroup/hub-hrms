package repository

import (
	"context"
	"hub-hrms/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	User        UserRepository
	Employee    EmployeeRepository
	Onboarding  OnboardingRepository
	Workflow    WorkflowRepository
	Timesheet   TimesheetRepository
	PTO         PTORepository
	Benefits    BenefitsRepository
	Payroll     PayrollRepository
	Recruiting  RecruitingRepository
	Organization OrganizationRepository
}

func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		User:        NewUserRepository(db),
		Employee:    NewEmployeeRepository(db),
		Onboarding:  NewOnboardingRepository(db),
		Workflow:    NewWorkflowRepository(db),
		Timesheet:   NewTimesheetRepository(db),
		PTO:         NewPTORepository(db),
		Benefits:    NewBenefitsRepository(db),
		Payroll:     NewPayrollRepository(db),
		Recruiting:  NewRecruitingRepository(db),
		Organization: NewOrganizationRepository(db),
	}
}

// UserRepository interface
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

// EmployeeRepository interface
type EmployeeRepository interface {
	Create(ctx context.Context, employee *models.Employee) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error)
	GetByEmail(ctx context.Context, email string) (*models.Employee, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*models.Employee, error)
	Update(ctx context.Context, employee *models.Employee) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// OnboardingRepository interface
type OnboardingRepository interface {
	CreateTask(ctx context.Context, task *models.OnboardingTask) error
	GetTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.OnboardingTask, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error)
	UpdateTask(ctx context.Context, task *models.OnboardingTask) error
	DeleteTask(ctx context.Context, id uuid.UUID) error
}

// TimesheetRepository interface
type TimesheetRepository interface {
	Create(ctx context.Context, timesheet *models.Timesheet) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Timesheet, error)
	GetByEmployee(ctx context.Context, employeeID uuid.UUID, filters map[string]interface{}) ([]*models.Timesheet, error)
	GetActiveTimesheet(ctx context.Context, employeeID uuid.UUID) (*models.Timesheet, error)
	Update(ctx context.Context, timesheet *models.Timesheet) error
	List(ctx context.Context, filters map[string]interface{}) ([]*models.Timesheet, error)
}

// PTORepository interface
type PTORepository interface {
	GetBalance(ctx context.Context, employeeID uuid.UUID) (*models.PTOBalance, error)
	CreateBalance(ctx context.Context, balance *models.PTOBalance) error
	UpdateBalance(ctx context.Context, balance *models.PTOBalance) error
	CreateRequest(ctx context.Context, request *models.PTORequest) error
	GetRequestByID(ctx context.Context, id uuid.UUID) (*models.PTORequest, error)
	GetRequestsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.PTORequest, error)
	UpdateRequest(ctx context.Context, request *models.PTORequest) error
	ListRequests(ctx context.Context, filters map[string]interface{}) ([]*models.PTORequest, error)
}


// PayrollRepository interface defines payroll operations
type PayrollRepository interface {
	// Compensation
	CreateCompensation(ctx context.Context, comp *models.EmployeeCompensation) error
	GetCompensationByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.EmployeeCompensation, error)
	UpdateCompensation(ctx context.Context, comp *models.EmployeeCompensation) error

	// Tax Withholding
	CreateTaxWithholding(ctx context.Context, tax *models.W2TaxWithholding) error
	GetTaxWithholdingByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*models.W2TaxWithholding, error)
	UpdateTaxWithholding(ctx context.Context, tax *models.W2TaxWithholding) error

	// Payroll Periods
	CreatePeriod(ctx context.Context, period *models.PayrollPeriod) error
	GetPeriodByID(ctx context.Context, id uuid.UUID) (*models.PayrollPeriod, error)
	ListPeriods(ctx context.Context, filters map[string]interface{}) ([]*models.PayrollPeriod, error)
	UpdatePeriod(ctx context.Context, period *models.PayrollPeriod) error

	// Pay Stubs
	CreatePayStub(ctx context.Context, stub *models.PayStub) error
	GetPayStubByID(ctx context.Context, id uuid.UUID) (*models.PayStub, error)
	ListPayStubsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.PayStub, error)
	ListPayStubsByPeriod(ctx context.Context, periodID uuid.UUID) ([]*models.PayStub, error)

	// Pay Stub Details
	CreatePayStubEarning(ctx context.Context, earning *models.PayStubEarning) error
	CreatePayStubDeduction(ctx context.Context, deduction *models.PayStubDeduction) error
	CreatePayStubTax(ctx context.Context, tax *models.PayStubTax) error
	GetPayStubEarnings(ctx context.Context, payStubID uuid.UUID) ([]models.PayStubEarning, error)
	GetPayStubDeductions(ctx context.Context, payStubID uuid.UUID) ([]models.PayStubDeduction, error)
	GetPayStubTaxes(ctx context.Context, payStubID uuid.UUID) ([]models.PayStubTax, error)

	// 1099 Forms
	Create1099(ctx context.Context, form *models.Form1099) error
	Get1099ByEmployeeAndYear(ctx context.Context, employeeID uuid.UUID, year int) (*models.Form1099, error)
	List1099ByYear(ctx context.Context, year int) ([]*models.Form1099, error)
	Update1099(ctx context.Context, form *models.Form1099) error

	// YTD Calculations
	GetYTDEarnings(ctx context.Context, employeeID uuid.UUID, year int) (float64, error)
	GetYTDTaxes(ctx context.Context, employeeID uuid.UUID, year int) (float64, error)
}

// UserRepository implementation
type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, role, employee_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		user.Email, user.PasswordHash, user.Role, user.EmployeeID,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, password_hash, role, employee_id, created_at, updated_at
		FROM users WHERE email = $1
	`
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role,
		&user.EmployeeID, &user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, password_hash, role, employee_id, created_at, updated_at
		FROM users WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role,
		&user.EmployeeID, &user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, role = $3, employee_id = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		user.Email, user.PasswordHash, user.Role, user.EmployeeID, user.ID,
	).Scan(&user.UpdatedAt)
}


func NewOnboardingRepository(db *pgxpool.Pool) OnboardingRepository {
	return &onboardingRepository{db: db}
}

func (r *onboardingRepository) CreateTask(ctx context.Context, task *models.OnboardingTask) error {
	query := `
		INSERT INTO onboarding_tasks (
			employee_id, task_name, description, category, status, due_date,
			assigned_to, documents_required
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		task.EmployeeID, task.TaskName, task.Description, task.Category,
		task.Status, task.DueDate, task.AssignedTo, task.DocumentsRequired,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *onboardingRepository) GetTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.OnboardingTask, error) {
	query := `
		SELECT id, employee_id, task_name, description, category, status, due_date,
			completed_at, assigned_to, documents_required, document_url, created_at, updated_at
		FROM onboarding_tasks
		WHERE employee_id = $1
		ORDER BY due_date NULLS LAST, created_at
	`
	
	rows, err := r.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.OnboardingTask
	for rows.Next() {
		task := &models.OnboardingTask{}
		err := rows.Scan(
			&task.ID, &task.EmployeeID, &task.TaskName, &task.Description,
			&task.Category, &task.Status, &task.DueDate, &task.CompletedAt,
			&task.AssignedTo, &task.DocumentsRequired, &task.DocumentURL,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r *onboardingRepository) GetTaskByID(ctx context.Context, id uuid.UUID) (*models.OnboardingTask, error) {
	task := &models.OnboardingTask{}
	query := `
		SELECT id, employee_id, task_name, description, category, status, due_date,
			completed_at, assigned_to, documents_required, document_url, created_at, updated_at
		FROM onboarding_tasks WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.EmployeeID, &task.TaskName, &task.Description,
		&task.Category, &task.Status, &task.DueDate, &task.CompletedAt,
		&task.AssignedTo, &task.DocumentsRequired, &task.DocumentURL,
		&task.CreatedAt, &task.UpdatedAt,
	)
	return task, err
}

func (r *onboardingRepository) UpdateTask(ctx context.Context, task *models.OnboardingTask) error {
	query := `
		UPDATE onboarding_tasks SET
			task_name = $1, description = $2, category = $3, status = $4,
			due_date = $5, completed_at = $6, assigned_to = $7,
			documents_required = $8, document_url = $9, updated_at = NOW()
		WHERE id = $10
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		task.TaskName, task.Description, task.Category, task.Status,
		task.DueDate, task.CompletedAt, task.AssignedTo,
		task.DocumentsRequired, task.DocumentURL, task.ID,
	).Scan(&task.UpdatedAt)
}

func (r *onboardingRepository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM onboarding_tasks WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
