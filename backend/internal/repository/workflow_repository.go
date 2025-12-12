package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hub-hrms/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkflowRepository handles workflow database operations
type WorkflowRepository interface {
	// Workflow operations
	CreateWorkflow(ctx context.Context, workflow *models.OnboardingWorkflow) error
	GetWorkflow(ctx context.Context, id uuid.UUID) (*models.OnboardingWorkflow, error)
	GetWorkflowWithDetails(ctx context.Context, id uuid.UUID) (*models.WorkflowWithDetails, error)
	ListWorkflows(ctx context.Context, filters map[string]interface{}) ([]*models.OnboardingWorkflow, error)
	UpdateWorkflowStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateWorkflowStage(ctx context.Context, id uuid.UUID, stage string) error
	
	// Step operations
	CreateStep(ctx context.Context, step *models.WorkflowStep) error
	GetSteps(ctx context.Context, workflowID uuid.UUID) ([]*models.WorkflowStep, error)
	GetStepByID(ctx context.Context, id uuid.UUID) (*models.WorkflowStep, error)
	UpdateStep(ctx context.Context, step *models.WorkflowStep) error
	CompleteStep(ctx context.Context, stepID, completedBy uuid.UUID) error
	
	// Integration operations
	CreateIntegration(ctx context.Context, integration *models.WorkflowIntegration) error
	GetIntegration(ctx context.Context, id uuid.UUID) (*models.WorkflowIntegration, error)
	GetIntegrations(ctx context.Context, workflowID uuid.UUID) ([]*models.WorkflowIntegration, error)
	UpdateIntegration(ctx context.Context, integration *models.WorkflowIntegration) error
	
	// Exception operations
	CreateException(ctx context.Context, exception *models.WorkflowException) error
	GetExceptions(ctx context.Context, workflowID uuid.UUID) ([]*models.WorkflowException, error)
	ResolveException(ctx context.Context, exceptionID, resolvedBy uuid.UUID, notes string) error
	
	// Document operations
	CreateDocument(ctx context.Context, doc *models.WorkflowDocument) error
	GetDocuments(ctx context.Context, workflowID uuid.UUID) ([]*models.WorkflowDocument, error)
	UpdateDocument(ctx context.Context, doc *models.WorkflowDocument) error
}

type workflowRepository struct {
	db *pgxpool.Pool
}

func NewWorkflowRepository(db *pgxpool.Pool) WorkflowRepository {
	return &workflowRepository{db: db}
}

// Workflow operations

func (r *workflowRepository) CreateWorkflow(ctx context.Context, workflow *models.OnboardingWorkflow) error {
	query := `
		INSERT INTO onboarding_workflows (
			employee_id, template_name, status, current_stage, 
			expected_completion, created_by
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, progress_percentage, started_at, created_at, updated_at
	`
	
	return r.db.QueryRow(ctx, query,
		workflow.EmployeeID,
		workflow.TemplateName,
		workflow.Status,
		workflow.CurrentStage,
		workflow.ExpectedCompletion,
		workflow.CreatedBy,
	).Scan(
		&workflow.ID,
		&workflow.ProgressPercentage,
		&workflow.StartedAt,
		&workflow.CreatedAt,
		&workflow.UpdatedAt,
	)
}

func (r *workflowRepository) GetWorkflow(ctx context.Context, id uuid.UUID) (*models.OnboardingWorkflow, error) {
	query := `
		SELECT id, employee_id, template_name, status, current_stage,
			progress_percentage, started_at, expected_completion,
			actual_completion, created_by, created_at, updated_at
		FROM onboarding_workflows
		WHERE id = $1
	`
	
	workflow := &models.OnboardingWorkflow{}
	var createdBy sql.NullString
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&workflow.ID,
		&workflow.EmployeeID,
		&workflow.TemplateName,
		&workflow.Status,
		&workflow.CurrentStage,
		&workflow.ProgressPercentage,
		&workflow.StartedAt,
		&workflow.ExpectedCompletion,
		&workflow.ActualCompletion,
		&createdBy,
		&workflow.CreatedAt,
		&workflow.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if createdBy.Valid {
		id, _ := uuid.Parse(createdBy.String)
		workflow.CreatedBy = &id
	}
	
	return workflow, nil
}

func (r *workflowRepository) GetWorkflowWithDetails(ctx context.Context, id uuid.UUID) (*models.WorkflowWithDetails, error) {
	// Get workflow
	workflow, err := r.GetWorkflow(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Get employee
	empQuery := `SELECT id, first_name, last_name, email, department, position, status FROM employees WHERE id = $1`
	employee := models.Employee{}
	err = r.db.QueryRow(ctx, empQuery, workflow.EmployeeID).Scan(
		&employee.ID,
		&employee.FirstName,
		&employee.LastName,
		&employee.Email,
		&employee.Department,
		&employee.Position,
		&employee.Status,
	)
	if err != nil {
		return nil, err
	}
	
	// Get steps
	stepsPtr, err := r.GetSteps(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Convert []*WorkflowStep to []WorkflowStep
	steps := make([]models.WorkflowStep, len(stepsPtr))
	for i, s := range stepsPtr {
		steps[i] = *s
	}
	
	// Get exceptions
	exceptionsPtr, err := r.GetExceptions(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Convert []*WorkflowException to []WorkflowException
	exceptions := make([]models.WorkflowException, len(exceptionsPtr))
	for i, e := range exceptionsPtr {
		exceptions[i] = *e
	}
	
	// Get documents
	documentsPtr, err := r.GetDocuments(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Convert []*WorkflowDocument to []WorkflowDocument
	documents := make([]models.WorkflowDocument, len(documentsPtr))
	for i, d := range documentsPtr {
		documents[i] = *d
	}
	
	return &models.WorkflowWithDetails{
		Workflow:   *workflow,
		Steps:      steps,
		Exceptions: exceptions,
		Documents:  documents,
		Employee:   employee,
	}, nil
}

func (r *workflowRepository) ListWorkflows(ctx context.Context, filters map[string]interface{}) ([]*models.OnboardingWorkflow, error) {
	query := `
		SELECT id, employee_id, template_name, status, current_stage,
			progress_percentage, started_at, expected_completion,
			actual_completion, created_by, created_at, updated_at
		FROM onboarding_workflows
		WHERE 1=1
	`
	
	args := []interface{}{}
	argPos := 1
	
	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}
	
	if employeeID, ok := filters["employee_id"].(uuid.UUID); ok {
		query += fmt.Sprintf(" AND employee_id = $%d", argPos)
		args = append(args, employeeID)
		argPos++
	}
	
	query += " ORDER BY created_at DESC"
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	workflows := []*models.OnboardingWorkflow{}
	for rows.Next() {
		workflow := &models.OnboardingWorkflow{}
		var createdBy sql.NullString
		
		err := rows.Scan(
			&workflow.ID,
			&workflow.EmployeeID,
			&workflow.TemplateName,
			&workflow.Status,
			&workflow.CurrentStage,
			&workflow.ProgressPercentage,
			&workflow.StartedAt,
			&workflow.ExpectedCompletion,
			&workflow.ActualCompletion,
			&createdBy,
			&workflow.CreatedAt,
			&workflow.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		if createdBy.Valid {
			id, _ := uuid.Parse(createdBy.String)
			workflow.CreatedBy = &id
		}
		
		workflows = append(workflows, workflow)
	}
	
	return workflows, nil
}

func (r *workflowRepository) UpdateWorkflowStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE onboarding_workflows SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *workflowRepository) UpdateWorkflowStage(ctx context.Context, id uuid.UUID, stage string) error {
	query := `UPDATE onboarding_workflows SET current_stage = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, stage, id)
	return err
}

// Step operations

func (r *workflowRepository) CreateStep(ctx context.Context, step *models.WorkflowStep) error {
	query := `
		INSERT INTO workflow_steps (
			workflow_id, step_order, step_name, step_type, stage, status,
			description, dependencies, assigned_to, integration_type,
			integration_config, due_date, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`
	
	// Convert dependencies to JSON
	depsJSON, _ := json.Marshal(step.Dependencies)
	
	// Convert metadata to JSON
	metaJSON, _ := json.Marshal(step.Metadata)
	
	// Convert integration config to JSON
	configJSON, _ := json.Marshal(step.IntegrationConfig)
	
	return r.db.QueryRow(ctx, query,
		step.WorkflowID,
		step.StepOrder,
		step.StepName,
		step.StepType,
		step.Stage,
		step.Status,
		step.Description,
		depsJSON,
		step.AssignedTo,
		step.IntegrationType,
		configJSON,
		step.DueDate,
		metaJSON,
	).Scan(&step.ID, &step.CreatedAt, &step.UpdatedAt)
}

func (r *workflowRepository) GetSteps(ctx context.Context, workflowID uuid.UUID) ([]*models.WorkflowStep, error) {
	query := `
		SELECT id, workflow_id, step_order, step_name, step_type, stage, status,
			description, dependencies, assigned_to, integration_type,
			integration_config, due_date, started_at, completed_at,
			completed_by, metadata, created_at, updated_at
		FROM workflow_steps
		WHERE workflow_id = $1
		ORDER BY step_order ASC
	`
	
	rows, err := r.db.Query(ctx, query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	steps := []*models.WorkflowStep{}
	for rows.Next() {
		step, err := r.scanStep(rows)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	
	return steps, nil
}

func (r *workflowRepository) GetStepByID(ctx context.Context, id uuid.UUID) (*models.WorkflowStep, error) {
	query := `
		SELECT id, workflow_id, step_order, step_name, step_type, stage, status,
			description, dependencies, assigned_to, integration_type,
			integration_config, due_date, started_at, completed_at,
			completed_by, metadata, created_at, updated_at
		FROM workflow_steps
		WHERE id = $1
	`
	
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	if rows.Next() {
		return r.scanStep(rows)
	}
	
	return nil, fmt.Errorf("step not found")
}

func (r *workflowRepository) scanStep(rows interface{ Scan(...interface{}) error }) (*models.WorkflowStep, error) {
	step := &models.WorkflowStep{}
	var depsJSON []byte
	var metaJSON []byte
	var configJSON []byte
	var description, integrationType sql.NullString
	var assignedTo, completedBy sql.NullString
	
	err := rows.Scan(
		&step.ID,
		&step.WorkflowID,
		&step.StepOrder,
		&step.StepName,
		&step.StepType,
		&step.Stage,
		&step.Status,
		&description,
		&depsJSON,
		&assignedTo,
		&integrationType,
		&configJSON,
		&step.DueDate,
		&step.StartedAt,
		&step.CompletedAt,
		&completedBy,
		&metaJSON,
		&step.CreatedAt,
		&step.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if description.Valid {
		step.Description = description.String
	}
	
	if integrationType.Valid {
		step.IntegrationType = integrationType.String
	}
	
	if assignedTo.Valid {
		id, _ := uuid.Parse(assignedTo.String)
		step.AssignedTo = &id
	}
	
	if completedBy.Valid {
		id, _ := uuid.Parse(completedBy.String)
		step.CompletedBy = &id
	}
	
	// Parse JSON fields
	if len(depsJSON) > 0 {
		json.Unmarshal(depsJSON, &step.Dependencies)
	}
	
	if len(metaJSON) > 0 {
		json.Unmarshal(metaJSON, &step.Metadata)
	}
	
	if len(configJSON) > 0 {
		json.Unmarshal(configJSON, &step.IntegrationConfig)
	}
	
	return step, nil
}

func (r *workflowRepository) UpdateStep(ctx context.Context, step *models.WorkflowStep) error {
	query := `
		UPDATE workflow_steps
		SET status = $1, started_at = $2, completed_at = $3,
			completed_by = $4, updated_at = NOW()
		WHERE id = $5
	`
	
	_, err := r.db.Exec(ctx, query,
		step.Status,
		step.StartedAt,
		step.CompletedAt,
		step.CompletedBy,
		step.ID,
	)
	
	return err
}

func (r *workflowRepository) CompleteStep(ctx context.Context, stepID, completedBy uuid.UUID) error {
	query := `
		UPDATE workflow_steps
		SET status = 'completed', completed_at = NOW(), completed_by = $1, updated_at = NOW()
		WHERE id = $2
	`
	
	_, err := r.db.Exec(ctx, query, completedBy, stepID)
	return err
}

// Integration operations

func (r *workflowRepository) CreateIntegration(ctx context.Context, integration *models.WorkflowIntegration) error {
	query := `
		INSERT INTO workflow_integrations (
			workflow_id, step_id, integration_type, external_id, status,
			request_payload, response_payload, max_retries
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, retry_count, created_at, updated_at
	`
	
	reqJSON, _ := json.Marshal(integration.RequestPayload)
	respJSON, _ := json.Marshal(integration.ResponsePayload)
	
	return r.db.QueryRow(ctx, query,
		integration.WorkflowID,
		integration.StepID,
		integration.IntegrationType,
		integration.ExternalID,
		integration.Status,
		reqJSON,
		respJSON,
		integration.MaxRetries,
	).Scan(
		&integration.ID,
		&integration.RetryCount,
		&integration.CreatedAt,
		&integration.UpdatedAt,
	)
}

func (r *workflowRepository) GetIntegration(ctx context.Context, id uuid.UUID) (*models.WorkflowIntegration, error) {
	query := `
		SELECT id, workflow_id, step_id, integration_type, external_id, status,
			request_payload, response_payload, error_message, retry_count,
			max_retries, last_attempt_at, created_at, updated_at
		FROM workflow_integrations
		WHERE id = $1
	`
	
	integration := &models.WorkflowIntegration{}
	var reqJSON, respJSON []byte
	var externalID, errorMsg sql.NullString
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&integration.ID,
		&integration.WorkflowID,
		&integration.StepID,
		&integration.IntegrationType,
		&externalID,
		&integration.Status,
		&reqJSON,
		&respJSON,
		&errorMsg,
		&integration.RetryCount,
		&integration.MaxRetries,
		&integration.LastAttemptAt,
		&integration.CreatedAt,
		&integration.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if externalID.Valid {
		integration.ExternalID = externalID.String
	}
	
	if errorMsg.Valid {
		integration.ErrorMessage = errorMsg.String
	}
	
	if len(reqJSON) > 0 {
		json.Unmarshal(reqJSON, &integration.RequestPayload)
	}
	
	if len(respJSON) > 0 {
		json.Unmarshal(respJSON, &integration.ResponsePayload)
	}
	
	return integration, nil
}

func (r *workflowRepository) GetIntegrations(ctx context.Context, workflowID uuid.UUID) ([]*models.WorkflowIntegration, error) {
	query := `
		SELECT id, workflow_id, step_id, integration_type, external_id, status,
			request_payload, response_payload, error_message, retry_count,
			max_retries, last_attempt_at, created_at, updated_at
		FROM workflow_integrations
		WHERE workflow_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(ctx, query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	integrations := []*models.WorkflowIntegration{}
	for rows.Next() {
		integration := &models.WorkflowIntegration{}
		var reqJSON, respJSON []byte
		var externalID, errorMsg sql.NullString
		
		err := rows.Scan(
			&integration.ID,
			&integration.WorkflowID,
			&integration.StepID,
			&integration.IntegrationType,
			&externalID,
			&integration.Status,
			&reqJSON,
			&respJSON,
			&errorMsg,
			&integration.RetryCount,
			&integration.MaxRetries,
			&integration.LastAttemptAt,
			&integration.CreatedAt,
			&integration.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if externalID.Valid {
			integration.ExternalID = externalID.String
		}
		
		if errorMsg.Valid {
			integration.ErrorMessage = errorMsg.String
		}
		
		if len(reqJSON) > 0 {
			json.Unmarshal(reqJSON, &integration.RequestPayload)
		}
		
		if len(respJSON) > 0 {
			json.Unmarshal(respJSON, &integration.ResponsePayload)
		}
		
		integrations = append(integrations, integration)
	}
	
	return integrations, nil
}

func (r *workflowRepository) UpdateIntegration(ctx context.Context, integration *models.WorkflowIntegration) error {
	query := `
		UPDATE workflow_integrations
		SET status = $1, external_id = $2, response_payload = $3,
			error_message = $4, retry_count = $5, last_attempt_at = $6,
			updated_at = NOW()
		WHERE id = $7
	`
	
	respJSON, _ := json.Marshal(integration.ResponsePayload)
	
	_, err := r.db.Exec(ctx, query,
		integration.Status,
		integration.ExternalID,
		respJSON,
		integration.ErrorMessage,
		integration.RetryCount,
		integration.LastAttemptAt,
		integration.ID,
	)
	
	return err
}

// Exception operations

func (r *workflowRepository) CreateException(ctx context.Context, exception *models.WorkflowException) error {
	query := `
		INSERT INTO workflow_exceptions (
			workflow_id, step_id, exception_type, severity, title,
			description, resolution_status, assigned_to
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	
	return r.db.QueryRow(ctx, query,
		exception.WorkflowID,
		exception.StepID,
		exception.ExceptionType,
		exception.Severity,
		exception.Title,
		exception.Description,
		exception.ResolutionStatus,
		exception.AssignedTo,
	).Scan(&exception.ID, &exception.CreatedAt, &exception.UpdatedAt)
}

func (r *workflowRepository) GetExceptions(ctx context.Context, workflowID uuid.UUID) ([]*models.WorkflowException, error) {
	query := `
		SELECT id, workflow_id, step_id, exception_type, severity, title,
			description, resolution_status, assigned_to, resolved_at,
			resolved_by, resolution_notes, created_at, updated_at
		FROM workflow_exceptions
		WHERE workflow_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(ctx, query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	exceptions := []*models.WorkflowException{}
	for rows.Next() {
		exception := &models.WorkflowException{}
		var stepID, assignedTo, resolvedBy sql.NullString
		var description, resolutionNotes sql.NullString
		
		err := rows.Scan(
			&exception.ID,
			&exception.WorkflowID,
			&stepID,
			&exception.ExceptionType,
			&exception.Severity,
			&exception.Title,
			&description,
			&exception.ResolutionStatus,
			&assignedTo,
			&exception.ResolvedAt,
			&resolvedBy,
			&resolutionNotes,
			&exception.CreatedAt,
			&exception.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if stepID.Valid {
			id, _ := uuid.Parse(stepID.String)
			exception.StepID = &id
		}
		
		if assignedTo.Valid {
			id, _ := uuid.Parse(assignedTo.String)
			exception.AssignedTo = &id
		}
		
		if resolvedBy.Valid {
			id, _ := uuid.Parse(resolvedBy.String)
			exception.ResolvedBy = &id
		}
		
		if description.Valid {
			exception.Description = description.String
		}
		
		if resolutionNotes.Valid {
			exception.ResolutionNotes = resolutionNotes.String
		}
		
		exceptions = append(exceptions, exception)
	}
	
	return exceptions, nil
}

func (r *workflowRepository) ResolveException(ctx context.Context, exceptionID, resolvedBy uuid.UUID, notes string) error {
	query := `
		UPDATE workflow_exceptions
		SET resolution_status = 'resolved', resolved_at = NOW(),
			resolved_by = $1, resolution_notes = $2, updated_at = NOW()
		WHERE id = $3
	`
	
	_, err := r.db.Exec(ctx, query, resolvedBy, notes, exceptionID)
	return err
}

// Document operations

func (r *workflowRepository) CreateDocument(ctx context.Context, doc *models.WorkflowDocument) error {
	query := `
		INSERT INTO workflow_documents (
			workflow_id, step_id, document_name, document_type, s3_key,
			file_type, file_size, status, uploaded_by, uploaded_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	
	metaJSON, _ := json.Marshal(doc.Metadata)
	
	return r.db.QueryRow(ctx, query,
		doc.WorkflowID,
		doc.StepID,
		doc.DocumentName,
		doc.DocumentType,
		doc.S3Key,
		doc.FileType,
		doc.FileSize,
		doc.Status,
		doc.UploadedBy,
		doc.UploadedAt,
		metaJSON,
	).Scan(&doc.ID, &doc.CreatedAt, &doc.UpdatedAt)
}

func (r *workflowRepository) GetDocuments(ctx context.Context, workflowID uuid.UUID) ([]*models.WorkflowDocument, error) {
	query := `
		SELECT id, workflow_id, step_id, document_name, document_type, s3_key,
			file_type, file_size, status, uploaded_by, uploaded_at, metadata,
			created_at, updated_at
		FROM workflow_documents
		WHERE workflow_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(ctx, query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	documents := []*models.WorkflowDocument{}
	for rows.Next() {
		doc := &models.WorkflowDocument{}
		var stepID, uploadedBy sql.NullString
		var s3Key sql.NullString
		var fileSize sql.NullInt64
		var metaJSON []byte
		
		err := rows.Scan(
			&doc.ID,
			&doc.WorkflowID,
			&stepID,
			&doc.DocumentName,
			&doc.DocumentType,
			&s3Key,
			&doc.FileType,
			&fileSize,
			&doc.Status,
			&uploadedBy,
			&doc.UploadedAt,
			&metaJSON,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		if stepID.Valid {
			id, _ := uuid.Parse(stepID.String)
			doc.StepID = &id
		}
		
		if uploadedBy.Valid {
			id, _ := uuid.Parse(uploadedBy.String)
			doc.UploadedBy = &id
		}
		
		if s3Key.Valid {
			doc.S3Key = s3Key.String
		}
		
		if fileSize.Valid {
			doc.FileSize = int(fileSize.Int64)
		}
		
		if len(metaJSON) > 0 {
			json.Unmarshal(metaJSON, &doc.Metadata)
		}
		
		documents = append(documents, doc)
	}
	
	return documents, nil
}

func (r *workflowRepository) UpdateDocument(ctx context.Context, doc *models.WorkflowDocument) error {
	query := `
		UPDATE workflow_documents
		SET status = $1, s3_key = $2, file_size = $3,
			uploaded_by = $4, uploaded_at = $5, updated_at = NOW()
		WHERE id = $6
	`
	
	_, err := r.db.Exec(ctx, query,
		doc.Status,
		doc.S3Key,
		doc.FileSize,
		doc.UploadedBy,
		doc.UploadedAt,
		doc.ID,
	)
	
	return err
}
