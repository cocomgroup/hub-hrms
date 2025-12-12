package repository

import (
	"context"
	"database/sql"
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


// PayrollRepository interface
type PayrollRepository interface {
	CreatePeriod(ctx context.Context, period *models.PayrollPeriod) error
	GetPeriodByID(ctx context.Context, id uuid.UUID) (*models.PayrollPeriod, error)
	ListPeriods(ctx context.Context, filters map[string]interface{}) ([]*models.PayrollPeriod, error)
	UpdatePeriod(ctx context.Context, period *models.PayrollPeriod) error
	CreatePayStub(ctx context.Context, stub *models.PayStub) error
	GetPayStubByID(ctx context.Context, id uuid.UUID) (*models.PayStub, error)
	GetPayStubsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.PayStub, error)
	GetPayStubsByPeriod(ctx context.Context, periodID uuid.UUID) ([]*models.PayStub, error)
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

// EmployeeRepository implementation
type employeeRepository struct {
	db *pgxpool.Pool
}

func NewEmployeeRepository(db *pgxpool.Pool) EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(ctx context.Context, employee *models.Employee) error {
	query := `
		INSERT INTO employees (
			first_name, last_name, email, phone, date_of_birth, hire_date,
			department, position, manager_id, employment_type, status,
			street_address, city, state, zip_code, country,
			emergency_contact_name, emergency_contact_phone
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, updated_at
	`
	
	// Convert empty strings to NULL for nullable fields
	var phone, dept, pos, empType, addr, city, state, zip, country, ecName, ecPhone interface{}
	if employee.Phone != "" {
		phone = employee.Phone
	}
	if employee.Department != "" {
		dept = employee.Department
	}
	if employee.Position != "" {
		pos = employee.Position
	}
	if employee.EmploymentType != "" {
		empType = employee.EmploymentType
	}
	if employee.StreetAddress != "" {
		addr = employee.StreetAddress
	}
	if employee.City != "" {
		city = employee.City
	}
	if employee.State != "" {
		state = employee.State
	}
	if employee.ZipCode != "" {
		zip = employee.ZipCode
	}
	if employee.Country != "" {
		country = employee.Country
	}
	if employee.EmergencyContactName != "" {
		ecName = employee.EmergencyContactName
	}
	if employee.EmergencyContactPhone != "" {
		ecPhone = employee.EmergencyContactPhone
	}
	
	return r.db.QueryRow(ctx, query,
		employee.FirstName, employee.LastName, employee.Email, phone,
		employee.DateOfBirth, employee.HireDate, dept, pos,
		employee.ManagerID, empType, employee.Status, addr,
		city, state, zip, country,
		ecName, ecPhone,
	).Scan(&employee.ID, &employee.CreatedAt, &employee.UpdatedAt)
}

func (r *employeeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	employee := &models.Employee{}
	var phone, dept, pos, empType, addr, city, state, zip, country, ecName, ecPhone sql.NullString
	
	query := `
		SELECT id, first_name, last_name, email, phone, date_of_birth, hire_date,
			department, position, manager_id, employment_type, status,
			street_address, city, state, zip_code, country,
			emergency_contact_name, emergency_contact_phone, created_at, updated_at
		FROM employees WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&employee.ID, &employee.FirstName, &employee.LastName, &employee.Email,
		&phone, &employee.DateOfBirth, &employee.HireDate, &dept,
		&pos, &employee.ManagerID, &empType, &employee.Status,
		&addr, &city, &state, &zip,
		&country, &ecName, &ecPhone,
		&employee.CreatedAt, &employee.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	// Convert NullString to string
	employee.Phone = phone.String
	employee.Department = dept.String
	employee.Position = pos.String
	employee.EmploymentType = empType.String
	employee.StreetAddress = addr.String
	employee.City = city.String
	employee.State = state.String
	employee.ZipCode = zip.String
	employee.Country = country.String
	employee.EmergencyContactName = ecName.String
	employee.EmergencyContactPhone = ecPhone.String
	
	return employee, nil
}

func (r *employeeRepository) GetByEmail(ctx context.Context, email string) (*models.Employee, error) {
	employee := &models.Employee{}
	var phone, dept, pos, empType, addr, city, state, zip, country, ecName, ecPhone sql.NullString
	
	query := `
		SELECT id, first_name, last_name, email, phone, date_of_birth, hire_date,
			department, position, manager_id, employment_type, status,
			street_address, city, state, zip_code, country,
			emergency_contact_name, emergency_contact_phone, created_at, updated_at
		FROM employees WHERE email = $1
	`
	err := r.db.QueryRow(ctx, query, email).Scan(
		&employee.ID, &employee.FirstName, &employee.LastName, &employee.Email,
		&phone, &employee.DateOfBirth, &employee.HireDate, &dept,
		&pos, &employee.ManagerID, &empType, &employee.Status,
		&addr, &city, &state, &zip,
		&country, &ecName, &ecPhone,
		&employee.CreatedAt, &employee.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	// Convert NullString to string
	employee.Phone = phone.String
	employee.Department = dept.String
	employee.Position = pos.String
	employee.EmploymentType = empType.String
	employee.StreetAddress = addr.String
	employee.City = city.String
	employee.State = state.String
	employee.ZipCode = zip.String
	employee.Country = country.String
	employee.EmergencyContactName = ecName.String
	employee.EmergencyContactPhone = ecPhone.String
	
	return employee, nil
}

func (r *employeeRepository) List(ctx context.Context, filters map[string]interface{}) ([]*models.Employee, error) {
	query := `
		SELECT id, first_name, last_name, email, phone, date_of_birth, hire_date,
			department, position, manager_id, employment_type, status,
			street_address, city, state, zip_code, country,
			emergency_contact_name, emergency_contact_phone, created_at, updated_at
		FROM employees
		WHERE status = COALESCE($1, status)
		ORDER BY last_name, first_name
	`
	
	status, _ := filters["status"].(string)
	if status == "" {
		status = "active"
	}

	rows, err := r.db.Query(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []*models.Employee
	for rows.Next() {
		employee := &models.Employee{}
		var phone, dept, pos, empType, addr, city, state, zip, country, ecName, ecPhone sql.NullString
		
		err := rows.Scan(
			&employee.ID, &employee.FirstName, &employee.LastName, &employee.Email,
			&phone, &employee.DateOfBirth, &employee.HireDate, &dept,
			&pos, &employee.ManagerID, &empType, &employee.Status,
			&addr, &city, &state, &zip,
			&country, &ecName, &ecPhone,
			&employee.CreatedAt, &employee.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		// Convert NullString to string
		employee.Phone = phone.String
		employee.Department = dept.String
		employee.Position = pos.String
		employee.EmploymentType = empType.String
		employee.StreetAddress = addr.String
		employee.City = city.String
		employee.State = state.String
		employee.ZipCode = zip.String
		employee.Country = country.String
		employee.EmergencyContactName = ecName.String
		employee.EmergencyContactPhone = ecPhone.String
		
		employees = append(employees, employee)
	}

	return employees, rows.Err()
}

func (r *employeeRepository) Update(ctx context.Context, employee *models.Employee) error {
	query := `
		UPDATE employees SET
			first_name = $1, last_name = $2, email = $3, phone = $4, date_of_birth = $5,
			hire_date = $6, department = $7, position = $8, manager_id = $9,
			employment_type = $10, status = $11, street_address = $12, city = $13,
			state = $14, zip_code = $15, country = $16, emergency_contact_name = $17,
			emergency_contact_phone = $18, updated_at = NOW()
		WHERE id = $19
		RETURNING updated_at
	`
	
	// Convert empty strings to NULL for nullable fields
	var phone, dept, pos, empType, addr, city, state, zip, country, ecName, ecPhone interface{}
	if employee.Phone != "" {
		phone = employee.Phone
	}
	if employee.Department != "" {
		dept = employee.Department
	}
	if employee.Position != "" {
		pos = employee.Position
	}
	if employee.EmploymentType != "" {
		empType = employee.EmploymentType
	}
	if employee.StreetAddress != "" {
		addr = employee.StreetAddress
	}
	if employee.City != "" {
		city = employee.City
	}
	if employee.State != "" {
		state = employee.State
	}
	if employee.ZipCode != "" {
		zip = employee.ZipCode
	}
	if employee.Country != "" {
		country = employee.Country
	}
	if employee.EmergencyContactName != "" {
		ecName = employee.EmergencyContactName
	}
	if employee.EmergencyContactPhone != "" {
		ecPhone = employee.EmergencyContactPhone
	}
	
	return r.db.QueryRow(ctx, query,
		employee.FirstName, employee.LastName, employee.Email, phone,
		employee.DateOfBirth, employee.HireDate, dept, pos,
		employee.ManagerID, empType, employee.Status, addr,
		city, state, zip, country,
		ecName, ecPhone, employee.ID,
	).Scan(&employee.UpdatedAt)
}

func (r *employeeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE employees SET status = 'terminated', updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Additional repository implementations would follow similar patterns
// For brevity, showing structure for remaining repositories

type onboardingRepository struct {
	db *pgxpool.Pool
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


type ptoRepository struct{ db *pgxpool.Pool }
func NewPTORepository(db *pgxpool.Pool) PTORepository { return &ptoRepository{db: db} }
func (r *ptoRepository) GetBalance(ctx context.Context, employeeID uuid.UUID) (*models.PTOBalance, error) { return nil, nil }
func (r *ptoRepository) CreateBalance(ctx context.Context, balance *models.PTOBalance) error { return nil }
func (r *ptoRepository) UpdateBalance(ctx context.Context, balance *models.PTOBalance) error { return nil }
func (r *ptoRepository) CreateRequest(ctx context.Context, request *models.PTORequest) error { return nil }
func (r *ptoRepository) GetRequestByID(ctx context.Context, id uuid.UUID) (*models.PTORequest, error) { return nil, nil }
func (r *ptoRepository) GetRequestsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.PTORequest, error) { return nil, nil }
func (r *ptoRepository) UpdateRequest(ctx context.Context, request *models.PTORequest) error { return nil }
func (r *ptoRepository) ListRequests(ctx context.Context, filters map[string]interface{}) ([]*models.PTORequest, error) { return nil, nil }


type payrollRepository struct{ db *pgxpool.Pool }
func NewPayrollRepository(db *pgxpool.Pool) PayrollRepository { return &payrollRepository{db: db} }
func (r *payrollRepository) CreatePeriod(ctx context.Context, period *models.PayrollPeriod) error { return nil }
func (r *payrollRepository) GetPeriodByID(ctx context.Context, id uuid.UUID) (*models.PayrollPeriod, error) { return nil, nil }
func (r *payrollRepository) ListPeriods(ctx context.Context, filters map[string]interface{}) ([]*models.PayrollPeriod, error) { return nil, nil }
func (r *payrollRepository) UpdatePeriod(ctx context.Context, period *models.PayrollPeriod) error { return nil }
func (r *payrollRepository) CreatePayStub(ctx context.Context, stub *models.PayStub) error { return nil }
func (r *payrollRepository) GetPayStubByID(ctx context.Context, id uuid.UUID) (*models.PayStub, error) { return nil, nil }
func (r *payrollRepository) GetPayStubsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*models.PayStub, error) { return nil, nil }
func (r *payrollRepository) GetPayStubsByPeriod(ctx context.Context, periodID uuid.UUID) ([]*models.PayStub, error) { return nil, nil }
