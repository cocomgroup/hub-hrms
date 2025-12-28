package repository

import (
	"context"
	"database/sql"

	"hub-hrms/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EmployeeRepository interface
type EmployeeRepository interface {
	Create(ctx context.Context, employee *models.Employee) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error)
	GetByEmail(ctx context.Context, email string) (*models.Employee, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*models.Employee, error)
	Update(ctx context.Context, employee *models.Employee) error
	Delete(ctx context.Context, id uuid.UUID) error
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
