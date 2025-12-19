package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hub-hrms/backend/internal/api"
    "hub-hrms/backend/internal/db"
	"hub-hrms/backend/internal/config"
	"hub-hrms/backend/internal/repository"
	"hub-hrms/backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {

    cfg := config.Load()

	log.Printf("Starting backend server...")
	log.Printf("Database URL: %s", maskPassword(cfg.DatabaseURL))

	// Connect to database with retries
	var database *db.Postgres
	var err error

	for i := 0; i < 30; i++ {
		database, err = db.NewPostgres(cfg.DatabaseURL)
		if err == nil {
			log.Printf("Connected to database")
			break
		}
		log.Printf("Database connection attempt %d/30: %v", i+1, err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	log.Printf("Running database migrations...")
   
	if err := runMigrations(database); err != nil {
		log.Printf("WARNING: Migration error: %v", err)
	} else {
		log.Printf("Migrations completed")
	}


	// Initialize repositories
    var dbPool = database.GetPool() 
	repos := repository.NewRepositories(dbPool)

	// Initialize services
	services := service.NewServices(repos, cfg)

	// Initialize HTTP server
	router := setupRouter(services, cfg)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRouter(services *service.Services, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))


	// API routes
	r.Route("/api", func(r chi.Router) {
        
		// Health check
        r.Route("/health", func(r chi.Router) {
            r.Get("/", healthCheckHandler)
            r.Head("/", healthCheckHandler)
        })
		
		api.RegisterAuthRoutes(r, services)
		api.RegisterEmployeeRoutes(r, services)
		api.RegisterOnboardingRoutes(r, services)
		api.RegisterWorkflowRoutes(r, services)
		api.RegisterTimesheetRoutes(r, services)
		api.RegisterPTORoutes(r, services)
		api.RegisterBenefitsRoutes(r, services)
		api.RegisterPayrollRoutes(r, services)
        api.RegisterRecruitingRoutes(r, services)
	})

	return r
}

// Health function:
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    // Only send body for GET (HEAD must have no body per HTTP spec)
    if r.Method == http.MethodGet {
        w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
    }
}
type Config struct {
	DatabaseURL    string
	ServerAddr     string
    Port           string
    JWTSecret      string
    FrontendURL    string
    Environment    string
	S3Bucket       string
}

type DatabaseSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	DBName   string `json:"dbname"`
}


func maskPassword(connStr string) string {
	result := connStr
	if idx := len("postgres://"); idx < len(connStr) {
		afterUser := connStr[idx:]
		if colonIdx := indexOf(afterUser, ":"); colonIdx > 0 {
			if atIdx := indexOf(afterUser, "@"); atIdx > colonIdx {
				result = connStr[:idx+colonIdx+1] + "****" + connStr[idx+atIdx:]
			}
		}
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func runMigrations(database *db.Postgres) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Run migrations

	migrations := []string{
		//enableExtensions,
		//createUsersTable,
		//createEmployeesTable,
		//createOnboardingTasksTable,
		//createTimesheetsTable,
		//createPTOTables,
		//createBenefitsTables,
		//createPayrollTables,
		//createWorkflowTables,
		//seedInitialData,
	}

	for i, sql := range migrations {
		log.Printf("Migration %d/%d...", i+1, len(migrations))
		if err := database.Exec(ctx, sql); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	return nil
}

// Migration SQL statements
const enableExtensions = `
-- Enable required PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
`

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    employee_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_employee_id ON users(employee_id);
`

const createEmployeesTable = `
CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    date_of_birth DATE,
    hire_date DATE NOT NULL,
    department VARCHAR(100),
    position VARCHAR(100),
    manager_id UUID,
    employment_type VARCHAR(50),
    status VARCHAR(50) DEFAULT 'active',
    street_address VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(50),
    zip_code VARCHAR(20),
    country VARCHAR(100),
    emergency_contact_name VARCHAR(200),
    emergency_contact_phone VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (manager_id) REFERENCES employees(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);
CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department);
CREATE INDEX IF NOT EXISTS idx_employees_manager_id ON employees(manager_id);
`

const createOnboardingTasksTable = `
CREATE TABLE IF NOT EXISTS onboarding_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    task_name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    status VARCHAR(50) DEFAULT 'pending',
    due_date DATE,
    completed_at TIMESTAMP WITH TIME ZONE,
    assigned_to UUID,
    documents_required BOOLEAN DEFAULT false,
    document_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_to) REFERENCES employees(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_onboarding_employee_id ON onboarding_tasks(employee_id);
CREATE INDEX IF NOT EXISTS idx_onboarding_status ON onboarding_tasks(status);
`

const createTimesheetsTable = `
CREATE TABLE IF NOT EXISTS timesheets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    clock_in TIMESTAMP WITH TIME ZONE NOT NULL,
    clock_out TIMESTAMP WITH TIME ZONE,
    break_minutes INTEGER DEFAULT 0,
    total_hours DECIMAL(5,2),
    project_code VARCHAR(50),
    notes TEXT,
    status VARCHAR(50) DEFAULT 'draft',
    approved_by UUID,
    approved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (approved_by) REFERENCES employees(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_timesheets_employee_id ON timesheets(employee_id);
CREATE INDEX IF NOT EXISTS idx_timesheets_clock_in ON timesheets(clock_in);
CREATE INDEX IF NOT EXISTS idx_timesheets_status ON timesheets(status);
`

const createPTOTables = `
CREATE TABLE IF NOT EXISTS pto_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL UNIQUE,
    vacation_days DECIMAL(5,2) DEFAULT 0,
    sick_days DECIMAL(5,2) DEFAULT 0,
    personal_days DECIMAL(5,2) DEFAULT 0,
    accrual_rate_vacation DECIMAL(5,2) DEFAULT 0,
    accrual_rate_sick DECIMAL(5,2) DEFAULT 0,
    last_accrual_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS pto_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    pto_type VARCHAR(50) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    days_requested DECIMAL(5,2) NOT NULL,
    reason TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    reviewed_by UUID,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewed_by) REFERENCES employees(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_pto_requests_employee_id ON pto_requests(employee_id);
CREATE INDEX IF NOT EXISTS idx_pto_requests_status ON pto_requests(status);
CREATE INDEX IF NOT EXISTS idx_pto_requests_dates ON pto_requests(start_date, end_date);
`

const createBenefitsTables = `
CREATE TABLE IF NOT EXISTS benefit_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_name VARCHAR(255) NOT NULL,
    plan_type VARCHAR(100) NOT NULL,
    provider VARCHAR(255),
    description TEXT,
    employee_cost DECIMAL(10,2),
    employer_cost DECIMAL(10,2),
    coverage_details JSONB,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS benefit_enrollments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    plan_id UUID NOT NULL,
    enrollment_date DATE NOT NULL,
    effective_date DATE NOT NULL,
    termination_date DATE,
    status VARCHAR(50) DEFAULT 'active',
    dependents JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (plan_id) REFERENCES benefit_plans(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_benefit_enrollments_employee_id ON benefit_enrollments(employee_id);
CREATE INDEX IF NOT EXISTS idx_benefit_enrollments_plan_id ON benefit_enrollments(plan_id);
`

const createPayrollTables = `
CREATE TABLE IF NOT EXISTS payroll_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    pay_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'open',
    processed_by UUID,
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (processed_by) REFERENCES employees(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS pay_stubs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    payroll_period_id UUID NOT NULL,
    gross_pay DECIMAL(10,2) NOT NULL,
    federal_tax DECIMAL(10,2) DEFAULT 0,
    state_tax DECIMAL(10,2) DEFAULT 0,
    social_security DECIMAL(10,2) DEFAULT 0,
    medicare DECIMAL(10,2) DEFAULT 0,
    other_deductions DECIMAL(10,2) DEFAULT 0,
    net_pay DECIMAL(10,2) NOT NULL,
    hours_worked DECIMAL(5,2),
    overtime_hours DECIMAL(5,2) DEFAULT 0,
    hourly_rate DECIMAL(10,2),
    benefits_deductions DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (payroll_period_id) REFERENCES payroll_periods(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_pay_stubs_employee_id ON pay_stubs(employee_id);
CREATE INDEX IF NOT EXISTS idx_pay_stubs_period_id ON pay_stubs(payroll_period_id);
`

const createWorkflowTables = `
-- Workflow system tables
CREATE TABLE IF NOT EXISTS onboarding_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    template_name VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    current_stage VARCHAR(100) NOT NULL DEFAULT 'pre-boarding',
    progress_percentage INTEGER DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expected_completion TIMESTAMP WITH TIME ZONE,
    actual_completion TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workflow_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    step_order INTEGER NOT NULL,
    step_name VARCHAR(255) NOT NULL,
    step_type VARCHAR(50) NOT NULL,
    stage VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    description TEXT,
    dependencies JSONB DEFAULT '[]',
    assigned_to UUID REFERENCES employees(id),
    integration_type VARCHAR(50),
    integration_config JSONB,
    due_date TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    completed_by UUID REFERENCES employees(id),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workflow_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    step_id UUID NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    integration_type VARCHAR(50) NOT NULL,
    external_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    request_payload JSONB,
    response_payload JSONB,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workflow_exceptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    step_id UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,
    exception_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    resolution_status VARCHAR(50) NOT NULL DEFAULT 'open',
    assigned_to UUID REFERENCES employees(id),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID REFERENCES employees(id),
    resolution_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workflow_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES onboarding_workflows(id) ON DELETE CASCADE,
    step_id UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,
    document_name VARCHAR(255) NOT NULL,
    document_type VARCHAR(50) NOT NULL,
    s3_key VARCHAR(500),
    file_type VARCHAR(20),
    file_size INTEGER,
    status VARCHAR(50) DEFAULT 'pending',
    uploaded_by UUID REFERENCES employees(id),
    uploaded_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflows_employee ON onboarding_workflows(employee_id);
CREATE INDEX IF NOT EXISTS idx_workflows_status ON onboarding_workflows(status);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_workflow ON workflow_steps(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_status ON workflow_steps(status);
CREATE INDEX IF NOT EXISTS idx_workflow_integrations_workflow ON workflow_integrations(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_exceptions_workflow ON workflow_exceptions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_documents_workflow ON workflow_documents(workflow_id);
`

const seedInitialData = `
-- Insert admin user (password: admin123)
INSERT INTO users (id, email, password_hash, role, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'admin@cocomgroup.com',
    '$2a$10$rQZN5YhJXKYQX5ZQXQXQXO5YhJXKYQX5ZQXQXQXO5YhJXKYQX5ZQ',
    'admin',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO NOTHING;

-- Insert sample employees
INSERT INTO employees (id, first_name, last_name, email, phone, hire_date, department, position, status, created_at, updated_at)
VALUES 
(
    gen_random_uuid(),
    'Evan',
    'Hunt',
    'evan.hunt@cocomgroup.com',
    '555-0101',
    '2020-01-15',
    'Engineering',
    'CTO',
    'active',
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'Bob',
    'Johnson',
    'bob.johnson@cocomgroup.com',
    '555-0102',
    '2021-03-20',
    'Sales',
    'Sales Representative',
    'active',
    NOW(),
    NOW()
),
(
    gen_random_uuid(),
    'Jane',
    'Smith',
    'jane.smith@cocomgroup.com',
    '555-0103',
    '2021-06-10',
    'Human Resources',
    'HR Manager',
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO NOTHING;
`