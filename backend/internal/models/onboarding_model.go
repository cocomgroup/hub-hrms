// File: internal/models/onboarding_model.go
package models

import (
	"time"
	"github.com/google/uuid"
)

// NewHireOnboarding represents a specialized onboarding journey for new employees
// This is SEPARATE from EmployeeWorkflow and adds AI assistance, tasks, and milestones
type NewHireOnboarding struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	EmployeeID             uuid.UUID  `json:"employee_id" db:"employee_id"`
	EmployeeName           string     `json:"employee_name,omitempty" db:"employee_name"`
	EmployeeEmail          string     `json:"employee_email,omitempty" db:"employee_email"`
	
	// Link to generic workflow engine (optional)
	WorkflowTemplateID     *uuid.UUID `json:"workflow_template_id,omitempty" db:"workflow_template_id"` // WorkflowTemplate being used
	EmployeeWorkflowID     *uuid.UUID `json:"employee_workflow_id,omitempty" db:"employee_workflow_id"` // Linked EmployeeWorkflow instance
	
	StartDate              time.Time  `json:"start_date" db:"start_date"`
	ExpectedCompletionDate *time.Time `json:"expected_completion_date,omitempty" db:"expected_completion_date"`
	ActualCompletionDate   *time.Time `json:"actual_completion_date,omitempty" db:"actual_completion_date"`
	Status                 string     `json:"status" db:"status"` // not_started, in_progress, completed, overdue
	OverallProgress        int        `json:"overall_progress" db:"overall_progress"` // 0-100
	AssignedBuddyID        *uuid.UUID `json:"assigned_buddy_id,omitempty" db:"assigned_buddy_id"`
	BuddyName              string     `json:"buddy_name,omitempty" db:"buddy_name"`
	AssignedManagerID      *uuid.UUID `json:"assigned_manager_id,omitempty" db:"assigned_manager_id"`
	ManagerName            string     `json:"manager_name,omitempty" db:"manager_name"`
	Notes                  string     `json:"notes,omitempty" db:"notes"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy              *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	
	// Nested data - AI-powered features unique to onboarding
	Tasks                  []*OnboardingTask        `json:"tasks,omitempty"`
	Milestones             []*OnboardingMilestone   `json:"milestones,omitempty"`
	RecentInteractions     []*OnboardingInteraction `json:"recent_interactions,omitempty"`
	Statistics             *OnboardingStatistics    `json:"statistics,omitempty"`
}

// OnboardingTask represents a task in the onboarding workflow
// Enhanced with AI suggestions and assistance
type OnboardingTask struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	WorkflowID      uuid.UUID  `json:"workflow_id" db:"workflow_id"` // References new_hire_onboardings.id
	Title           string     `json:"title" db:"title"`
	Description     string     `json:"description,omitempty" db:"description"`
	Category        string     `json:"category,omitempty" db:"category"` // documentation, equipment, training, access, administrative, social
	Priority        string     `json:"priority" db:"priority"` // low, medium, high, critical
	Status          string     `json:"status" db:"status"` // pending, in_progress, completed, blocked, skipped
	AssignedTo      *uuid.UUID `json:"assigned_to,omitempty" db:"assigned_to"`
	AssignedToName  string     `json:"assigned_to_name,omitempty" db:"assigned_to_name"`
	DueDate         *time.Time `json:"due_date,omitempty" db:"due_date"`
	CompletedAt     *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CompletedBy     *uuid.UUID `json:"completed_by,omitempty" db:"completed_by"`
	CompletedByName string     `json:"completed_by_name,omitempty" db:"completed_by_name"`
	OrderIndex      int        `json:"order_index" db:"order_index"`
	IsMandatory     bool       `json:"is_mandatory" db:"is_mandatory"`
	EstimatedHours  *float64   `json:"estimated_hours,omitempty" db:"estimated_hours"`
	ActualHours     *float64   `json:"actual_hours,omitempty" db:"actual_hours"`
	Dependencies    []byte     `json:"dependencies,omitempty" db:"dependencies"` // JSONB
	Attachments     []byte     `json:"attachments,omitempty" db:"attachments"` // JSONB
	AIGenerated     bool       `json:"ai_generated" db:"ai_generated"` // AI-specific feature
	AISuggestions   string     `json:"ai_suggestions,omitempty" db:"ai_suggestions"` // AI-specific feature
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// OnboardingInteraction represents AI agent interaction with new hire
// This is unique to the onboarding system - not part of generic workflows
type OnboardingInteraction struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	WorkflowID      uuid.UUID  `json:"workflow_id" db:"workflow_id"` // References new_hire_onboardings.id
	EmployeeID      uuid.UUID  `json:"employee_id" db:"employee_id"`
	EmployeeName    string     `json:"employee_name,omitempty" db:"employee_name"`
	InteractionType string     `json:"interaction_type" db:"interaction_type"` // chat, reminder, suggestion, check_in, escalation
	Message         string     `json:"message" db:"message"`
	AIResponse      string     `json:"ai_response,omitempty" db:"ai_response"`
	Sentiment       string     `json:"sentiment,omitempty" db:"sentiment"` // positive, neutral, negative, concerned
	RequiresAction  bool       `json:"requires_action" db:"requires_action"`
	ActionTaken     bool       `json:"action_taken" db:"action_taken"`
	Metadata        []byte     `json:"metadata,omitempty" db:"metadata"` // JSONB
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// OnboardingMilestone represents important milestones in onboarding
type OnboardingMilestone struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	WorkflowID      uuid.UUID  `json:"workflow_id" db:"workflow_id"` // References new_hire_onboardings.id
	Name            string     `json:"name" db:"name"`
	Description     string     `json:"description,omitempty" db:"description"`
	TargetDate      *time.Time `json:"target_date,omitempty" db:"target_date"`
	CompletedDate   *time.Time `json:"completed_date,omitempty" db:"completed_date"`
	Status          string     `json:"status" db:"status"` // pending, completed, missed
	CelebrationSent bool       `json:"celebration_sent" db:"celebration_sent"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// OnboardingChecklistTemplate for reusable onboarding task templates
// These are used to generate OnboardingTasks (separate from WorkflowTemplate)
type OnboardingChecklistTemplate struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description,omitempty" db:"description"`
	Department  string     `json:"department,omitempty" db:"department"`
	RoleType    string     `json:"role_type,omitempty" db:"role_type"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	
	Items       []*OnboardingChecklistTemplateItem `json:"items,omitempty"`
}

// OnboardingChecklistTemplateItem for template tasks
type OnboardingChecklistTemplateItem struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	TemplateID      uuid.UUID  `json:"template_id" db:"template_id"`
	Title           string     `json:"title" db:"title"`
	Description     string     `json:"description,omitempty" db:"description"`
	Category        string     `json:"category,omitempty" db:"category"`
	Priority        string     `json:"priority" db:"priority"`
	DayOffset       int        `json:"day_offset" db:"day_offset"` // Days from start date
	IsMandatory     bool       `json:"is_mandatory" db:"is_mandatory"`
	EstimatedHours  *float64   `json:"estimated_hours,omitempty" db:"estimated_hours"`
	OrderIndex      int        `json:"order_index" db:"order_index"`
	AssignedToRole  string     `json:"assigned_to_role,omitempty" db:"assigned_to_role"` // new_hire, manager, hr, it, buddy
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// OnboardingStatistics for analytics
type OnboardingStatistics struct {
	TotalTasks         int        `json:"total_tasks"`
	CompletedTasks     int        `json:"completed_tasks"`
	PendingTasks       int        `json:"pending_tasks"`
	OverdueTasks       int        `json:"overdue_tasks"`
	BlockedTasks       int        `json:"blocked_tasks"`
	CompletionRate     float64    `json:"completion_rate"`
	DaysActive         int        `json:"days_active"`
	EstimatedHoursLeft float64    `json:"estimated_hours_left"`
	AIInteractions     int        `json:"ai_interactions"`
	LastInteraction    *time.Time `json:"last_interaction,omitempty"`
}

// ============================================================================
// Request/Response DTOs
// ============================================================================

// CreateOnboardingRequest for creating a new hire onboarding
type CreateOnboardingRequest struct {
	EmployeeID             uuid.UUID  `json:"employee_id" validate:"required"`
	StartDate              time.Time  `json:"start_date" validate:"required"`
	ExpectedCompletionDate *time.Time `json:"expected_completion_date,omitempty"`
	AssignedBuddyID        *uuid.UUID `json:"assigned_buddy_id,omitempty"`
	AssignedManagerID      *uuid.UUID `json:"assigned_manager_id,omitempty"`
	WorkflowTemplateID     *uuid.UUID `json:"workflow_template_id,omitempty"` // Link to WorkflowTemplate
	ChecklistTemplateID    *uuid.UUID `json:"checklist_template_id,omitempty"` // Use OnboardingChecklistTemplate
	Notes                  string     `json:"notes,omitempty"`
}

// UpdateOnboardingRequest for updating onboarding details
type UpdateOnboardingRequest struct {
	ExpectedCompletionDate *time.Time `json:"expected_completion_date,omitempty"`
	AssignedBuddyID        *uuid.UUID `json:"assigned_buddy_id,omitempty"`
	AssignedManagerID      *uuid.UUID `json:"assigned_manager_id,omitempty"`
	Notes                  string     `json:"notes,omitempty"`
	Status                 string     `json:"status,omitempty"`
}

// CreateTaskRequest for creating onboarding tasks
type CreateTaskRequest struct {
	WorkflowID      uuid.UUID  `json:"workflow_id" validate:"required"` // new_hire_onboardings.id
	Title           string     `json:"title" validate:"required"`
	Description     string     `json:"description,omitempty"`
	Category        string     `json:"category,omitempty"`
	Priority        string     `json:"priority,omitempty"`
	AssignedTo      *uuid.UUID `json:"assigned_to,omitempty"`
	DueDate         *time.Time `json:"due_date,omitempty"`
	IsMandatory     bool       `json:"is_mandatory"`
	EstimatedHours  *float64   `json:"estimated_hours,omitempty"`
	OrderIndex      int        `json:"order_index,omitempty"`
}

// UpdateTaskRequest for updating tasks
type UpdateTaskRequest struct {
	Title          string     `json:"title,omitempty"`
	Description    string     `json:"description,omitempty"`
	Status         string     `json:"status,omitempty"`
	Priority       string     `json:"priority,omitempty"`
	AssignedTo     *uuid.UUID `json:"assigned_to,omitempty"`
	DueDate        *time.Time `json:"due_date,omitempty"`
	ActualHours    *float64   `json:"actual_hours,omitempty"`
	AISuggestions  string     `json:"ai_suggestions,omitempty"`
}

// CompleteTaskRequest for completing tasks
type CompleteTaskRequest struct {
	ActualHours *float64 `json:"actual_hours,omitempty"`
	Notes       string   `json:"notes,omitempty"`
}

// AIInteractionRequest for AI chat interactions
type AIInteractionRequest struct {
	WorkflowID uuid.UUID `json:"workflow_id" validate:"required"` // new_hire_onboardings.id
	Message    string    `json:"message" validate:"required"`
	Context    string    `json:"context,omitempty"` // Additional context for AI
}

// AIInteractionResponse for AI responses
type AIInteractionResponse struct {
	InteractionID uuid.UUID           `json:"interaction_id"`
	Response      string              `json:"response"`
	Suggestions   []string            `json:"suggestions,omitempty"`
	Actions       []AISuggestedAction `json:"actions,omitempty"`
	Sentiment     string              `json:"sentiment"`
}

// AISuggestedAction from AI assistant
type AISuggestedAction struct {
	Type        string                 `json:"type"` // create_task, update_task, schedule_meeting, send_reminder
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// OnboardingDashboardResponse for dashboard data
type OnboardingDashboardResponse struct {
	TotalOnboardings     int                  `json:"total_onboardings"`
	ActiveOnboardings    int                  `json:"active_onboardings"`
	CompletedOnboardings int                  `json:"completed_onboardings"`
	OverdueOnboardings   int                  `json:"overdue_onboardings"`
	RecentOnboardings    []*NewHireOnboarding `json:"recent_onboardings"`
	UpcomingTasks        []*OnboardingTask    `json:"upcoming_tasks"`
	AIInsights           []string             `json:"ai_insights"`
}

// CreateMilestoneRequest for creating milestones
type CreateMilestoneRequest struct {
	WorkflowID  uuid.UUID  `json:"workflow_id" validate:"required"` // new_hire_onboardings.id
	Name        string     `json:"name" validate:"required"`
	Description string     `json:"description,omitempty"`
	TargetDate  *time.Time `json:"target_date,omitempty"`
}

// ============================================================================
// BACKWARD COMPATIBILITY (temporary aliases)
// These will be removed in a future version
// ============================================================================

// OnboardingWorkflow is deprecated, use NewHireOnboarding instead
// Kept for backward compatibility with existing code
type OnboardingWorkflow = NewHireOnboarding

// CreateWorkflowRequest is deprecated, use CreateOnboardingRequest instead
type CreateWorkflowRequest = CreateOnboardingRequest

// UpdateWorkflowRequest is deprecated, use UpdateOnboardingRequest instead  
type UpdateWorkflowRequest = UpdateOnboardingRequest