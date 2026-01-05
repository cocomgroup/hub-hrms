<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../../../stores/auth';
  import { getApiBaseUrl } from '../../../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();

  interface WorkflowStep {
    id?: string;
    step_order: number;
    step_type: string;
    step_name: string;
    description: string;
    required: boolean;
    auto_trigger: boolean;
    assigned_role?: string;
    due_days?: number;
  }

  interface Workflow {
    id: string;
    name: string;
    description: string;
    workflow_type: string;
    status: string;
    steps: WorkflowStep[];
    created_at: string;
    updated_at: string;
  }

  let workflows: Workflow[] = [];
  let loading = false;
  let error = '';
  let showWorkflowModal = false;
  let showStepModal = false;
  let editingWorkflow: Workflow | null = null;
  let editingStepIndex: number | null = null;

  // New workflow form
  let newWorkflow = {
    name: '',
    description: '',
    workflow_type: 'onboarding',
    status: 'active'
  };

  // New step form
  let newStep: WorkflowStep = {
    step_order: 1,
    step_type: 'document',
    step_name: '',
    description: '',
    required: true,
    auto_trigger: false,
    assigned_role: 'hr',
    due_days: 3
  };

  let currentWorkflowSteps: WorkflowStep[] = [];

  const workflowTypes = [
    { value: 'onboarding', label: 'Onboarding' },
    { value: 'offboarding', label: 'Offboarding' },
    { value: 'promotion', label: 'Promotion' },
    { value: 'transfer', label: 'Transfer' },
    { value: 'performance_review', label: 'Performance Review' },
    { value: 'custom', label: 'Custom' }
  ];

  const stepTypes = [
    { value: 'document', label: 'Document Submission' },
    { value: 'approval', label: 'Approval Required' },
    { value: 'background_check', label: 'Background Check' },
    { value: 'i9_verification', label: 'I-9 Verification' },
    { value: 'equipment_setup', label: 'Equipment Setup' },
    { value: 'training', label: 'Training' },
    { value: 'system_access', label: 'System Access' },
    { value: 'meeting', label: 'Meeting/Orientation' },
    { value: 'notification', label: 'Notification' },
    { value: 'custom', label: 'Custom Task' }
  ];

  const roles = [
    { value: 'hr', label: 'HR' },
    { value: 'manager', label: 'Manager' },
    { value: 'it', label: 'IT' },
    { value: 'legal', label: 'Legal' },
    { value: 'employee', label: 'Employee' },
    { value: 'admin', label: 'Admin' }
  ];

  onMount(() => {
    loadWorkflows();
  });

  async function loadWorkflows() {
    try {
      loading = true;
      error = '';

      const response = await fetch(`${API_BASE_URL}/workflows/templates`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });

      if (!response.ok) {
        const errorText = await response.text();
        console.error('Failed to load workflows:', response.status, errorText);
        throw new Error(`Failed to load workflows: ${response.status}`);
      }
      
      const data = await response.json();
      
      // Ensure workflows is always an array
      workflows = Array.isArray(data) ? data : [];
      
    } catch (err: any) {
      console.error('Error loading workflows:', err);
      error = err.message;
      workflows = [];  // Reset to empty array on error
    } finally {
      loading = false;
    }
  }

  function openWorkflowModal(workflow?: Workflow) {
    if (workflow) {
      editingWorkflow = workflow;
      newWorkflow = {
        name: workflow.name,
        description: workflow.description,
        workflow_type: workflow.workflow_type,
        status: workflow.status
      };
      currentWorkflowSteps = [...workflow.steps].sort((a, b) => a.step_order - b.step_order);
    } else {
      editingWorkflow = null;
      newWorkflow = {
        name: '',
        description: '',
        workflow_type: 'onboarding',
        status: 'active'
      };
      currentWorkflowSteps = [];
    }
    showWorkflowModal = true;
  }

  function closeWorkflowModal() {
    showWorkflowModal = false;
    editingWorkflow = null;
    currentWorkflowSteps = [];
  }

  function openStepModal(stepIndex?: number) {
    if (stepIndex !== undefined && stepIndex !== null) {
      editingStepIndex = stepIndex;
      newStep = { ...currentWorkflowSteps[stepIndex] };
    } else {
      editingStepIndex = null;
      newStep = {
        step_order: currentWorkflowSteps.length + 1,
        step_type: 'document',
        step_name: '',
        description: '',
        required: true,
        auto_trigger: false,
        assigned_role: 'hr',
        due_days: 3
      };
    }
    showStepModal = true;
  }

  function closeStepModal() {
    showStepModal = false;
    editingStepIndex = null;
  }

  function saveStep() {
    if (editingStepIndex !== null) {
      currentWorkflowSteps[editingStepIndex] = { ...newStep };
    } else {
      currentWorkflowSteps = [...currentWorkflowSteps, { ...newStep }];
    }
    closeStepModal();
  }

  function deleteStep(index: number) {
    if (confirm('Delete this step?')) {
      currentWorkflowSteps = currentWorkflowSteps.filter((_, i) => i !== index);
      // Reorder steps
      currentWorkflowSteps = currentWorkflowSteps.map((step, i) => ({
        ...step,
        step_order: i + 1
      }));
    }
  }

  function moveStep(index: number, direction: 'up' | 'down') {
    const newIndex = direction === 'up' ? index - 1 : index + 1;
    if (newIndex < 0 || newIndex >= currentWorkflowSteps.length) return;

    const temp = currentWorkflowSteps[index];
    currentWorkflowSteps[index] = currentWorkflowSteps[newIndex];
    currentWorkflowSteps[newIndex] = temp;

    // Update step_order
    currentWorkflowSteps = currentWorkflowSteps.map((step, i) => ({
      ...step,
      step_order: i + 1
    }));
  }

  async function saveWorkflow() {
    if (!newWorkflow.name.trim()) {
      error = 'Workflow name is required';
      return;
    }

    if (currentWorkflowSteps.length === 0) {
      error = 'At least one step is required';
      return;
    }

    try {
      loading = true;
      error = '';

      const workflowData = {
        ...newWorkflow,
        steps: currentWorkflowSteps
      };

      const url = editingWorkflow
        ? `${API_BASE_URL}/workflows/templates/${editingWorkflow.id}`
        : `${API_BASE_URL}/workflows/templates`;

      const response = await fetch(url, {
        method: editingWorkflow ? 'PUT' : 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(workflowData)
      });

      if (!response.ok) throw new Error('Failed to save workflow');

      await loadWorkflows();
      closeWorkflowModal();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function deleteWorkflow(id: string) {
    if (!confirm('Delete this workflow? This cannot be undone.')) return;

    try {
      loading = true;
      error = '';

      const response = await fetch(`${API_BASE_URL}/workflows/templates/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });

      if (!response.ok) throw new Error('Failed to delete workflow');

      await loadWorkflows();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function getStepTypeLabel(type: string): string {
    return stepTypes.find(t => t.value === type)?.label || type;
  }

  function getWorkflowTypeLabel(type: string): string {
    return workflowTypes.find(t => t.value === type)?.label || type;
  }

  function getRoleLabel(role: string): string {
    return roles.find(r => r.value === role)?.label || role;
  }

  // Quick action templates
  function loadTemplate(templateType: string) {
    switch (templateType) {
      case 'onboarding':
        newWorkflow = {
          name: 'New Employee Onboarding',
          description: 'Complete onboarding process for new hires',
          workflow_type: 'onboarding',
          status: 'active'
        };
        currentWorkflowSteps = [
          {
            step_order: 1,
            step_type: 'document',
            step_name: 'Submit Personal Information',
            description: 'Employee submits personal details, emergency contacts',
            required: true,
            auto_trigger: true,
            assigned_role: 'employee',
            due_days: 1
          },
          {
            step_order: 2,
            step_type: 'i9_verification',
            step_name: 'I-9 Verification',
            description: 'Complete I-9 employment eligibility verification',
            required: true,
            auto_trigger: false,
            assigned_role: 'hr',
            due_days: 3
          },
          {
            step_order: 3,
            step_type: 'background_check',
            step_name: 'Background Check',
            description: 'Initiate and complete background check',
            required: true,
            auto_trigger: true,
            assigned_role: 'hr',
            due_days: 7
          },
          {
            step_order: 4,
            step_type: 'equipment_setup',
            step_name: 'IT Equipment Setup',
            description: 'Provision laptop, phone, and accessories',
            required: true,
            auto_trigger: false,
            assigned_role: 'it',
            due_days: 2
          },
          {
            step_order: 5,
            step_type: 'system_access',
            step_name: 'System Access Setup',
            description: 'Create accounts and grant system access',
            required: true,
            auto_trigger: false,
            assigned_role: 'it',
            due_days: 1
          },
          {
            step_order: 6,
            step_type: 'meeting',
            step_name: 'First Day Orientation',
            description: 'Welcome meeting and company orientation',
            required: true,
            auto_trigger: false,
            assigned_role: 'hr',
            due_days: 0
          },
          {
            step_order: 7,
            step_type: 'training',
            step_name: 'Required Training Modules',
            description: 'Complete mandatory training courses',
            required: true,
            auto_trigger: true,
            assigned_role: 'employee',
            due_days: 5
          }
        ];
        break;

      case 'offboarding':
        newWorkflow = {
          name: 'Employee Offboarding',
          description: 'Process for departing employees',
          workflow_type: 'offboarding',
          status: 'active'
        };
        currentWorkflowSteps = [
          {
            step_order: 1,
            step_type: 'notification',
            step_name: 'Notify Stakeholders',
            description: 'Inform IT, manager, and relevant teams',
            required: true,
            auto_trigger: true,
            assigned_role: 'hr',
            due_days: 0
          },
          {
            step_order: 2,
            step_type: 'meeting',
            step_name: 'Exit Interview',
            description: 'Conduct exit interview with employee',
            required: false,
            auto_trigger: false,
            assigned_role: 'hr',
            due_days: 1
          },
          {
            step_order: 3,
            step_type: 'system_access',
            step_name: 'Revoke System Access',
            description: 'Remove access to all systems and accounts',
            required: true,
            auto_trigger: false,
            assigned_role: 'it',
            due_days: 0
          },
          {
            step_order: 4,
            step_type: 'equipment_setup',
            step_name: 'Collect Equipment',
            description: 'Retrieve laptop, phone, badge, keys',
            required: true,
            auto_trigger: false,
            assigned_role: 'it',
            due_days: 0
          },
          {
            step_order: 5,
            step_type: 'document',
            step_name: 'Final Paperwork',
            description: 'Complete final payroll, benefits termination',
            required: true,
            auto_trigger: false,
            assigned_role: 'hr',
            due_days: 3
          }
        ];
        break;

      case 'performance_review':
        newWorkflow = {
          name: 'Performance Review Process',
          description: 'Annual/quarterly performance review cycle',
          workflow_type: 'performance_review',
          status: 'active'
        };
        currentWorkflowSteps = [
          {
            step_order: 1,
            step_type: 'notification',
            step_name: 'Launch Review Cycle',
            description: 'Notify all participants about review period',
            required: true,
            auto_trigger: true,
            assigned_role: 'hr',
            due_days: 0
          },
          {
            step_order: 2,
            step_type: 'document',
            step_name: 'Self-Assessment',
            description: 'Employee completes self-assessment',
            required: true,
            auto_trigger: false,
            assigned_role: 'employee',
            due_days: 7
          },
          {
            step_order: 3,
            step_type: 'document',
            step_name: 'Manager Review',
            description: 'Manager completes performance evaluation',
            required: true,
            auto_trigger: false,
            assigned_role: 'manager',
            due_days: 7
          },
          {
            step_order: 4,
            step_type: 'meeting',
            step_name: 'Review Meeting',
            description: 'One-on-one discussion of performance',
            required: true,
            auto_trigger: false,
            assigned_role: 'manager',
            due_days: 3
          },
          {
            step_order: 5,
            step_type: 'approval',
            step_name: 'HR Approval',
            description: 'HR reviews and approves final rating',
            required: true,
            auto_trigger: false,
            assigned_role: 'hr',
            due_days: 2
          }
        ];
        break;
    }
    
    openWorkflowModal();
  }
</script>

<div class="workflow-manager">
  {#if loading && workflows.length === 0}
    <div class="loading">Loading workflows...</div>
  {:else if error && workflows.length === 0}
    <div class="error">{error}</div>
  {:else}
    <!-- Quick Actions -->
    <div class="quick-actions">
      <h3>Workflow Templates</h3>
      <div class="template-buttons">
        <button class="btn-template onboarding" onclick={() => loadTemplate('onboarding')}>
          <span class="icon">👋</span>
          <span>New Hire Onboarding</span>
        </button>
        <button class="btn-template offboarding" onclick={() => loadTemplate('offboarding')}>
          <span class="icon">👋</span>
          <span>Employee Offboarding</span>
        </button>
        <button class="btn-template performance" onclick={() => loadTemplate('performance_review')}>
          <span class="icon">📊</span>
          <span>Performance Review</span>
        </button>
        <button class="btn-template custom" onclick={() => openWorkflowModal()}>
          <span class="icon">➕</span>
          <span>Custom Workflow</span>
        </button>
      </div>
    </div>

    <!-- Workflows List -->
    <div class="workflows-list">
      <h3>Existing Workflows ({workflows.length})</h3>
      {#if workflows.length === 0}
        <div class="empty-state">
          No workflows created yet. Use a template above to get started!
        </div>
      {:else}
        <div class="workflows-grid">
          {#each workflows as workflow}
            <div class="workflow-card">
              <div class="workflow-header">
                <div>
                  <h4>{workflow.name}</h4>
                  <span class="workflow-type">{getWorkflowTypeLabel(workflow.workflow_type)}</span>
                </div>
                <span class="status-badge status-{workflow.status}">{workflow.status}</span>
              </div>
              <p class="workflow-description">{workflow.description}</p>
              <div class="workflow-meta">
                <span>{workflow.steps?.length || 0} steps</span>
                <span>•</span>
                <span>Created {new Date(workflow.created_at).toLocaleDateString()}</span>
              </div>
              <div class="workflow-actions">
                <button class="btn-secondary" onclick={() => openWorkflowModal(workflow)}>
                  Edit
                </button>
                <button class="btn-danger" onclick={() => deleteWorkflow(workflow.id)}>
                  Delete
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Workflow Modal -->
{#if showWorkflowModal}
  <div class="modal-overlay" role="button" tabindex="0" onclick={closeWorkflowModal} onkeydown={(e) => e.key === "Escape" && closeWorkflowModal()}>
    <div class="modal large" role="dialog" aria-modal="true" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2>{editingWorkflow ? 'Edit Workflow' : 'Create Workflow'}</h2>
        <button class="close-btn" onclick={closeWorkflowModal}>×</button>
      </div>
      
      <div class="modal-body">
        {#if error}
          <div class="error-banner">{error}</div>
        {/if}

        <div class="form-section">
          <h3>Workflow Details</h3>
          <div class="form-grid">
            <div class="form-group">
              <label for="workflow-name">Workflow Name *</label>
              <input
                id="workflow-name"
                type="text"
                bind:value={newWorkflow.name}
                placeholder="e.g., New Employee Onboarding"
                required
              />
            </div>

            <div class="form-group">
              <label for="workflow-type">Workflow Type *</label>
              <select id="workflow-type" bind:value={newWorkflow.workflow_type}>
                {#each workflowTypes as type}
                  <option value={type.value}>{type.label}</option>
                {/each}
              </select>
            </div>

            <div class="form-group full-width">
              <label for="workflow-description">Description</label>
              <textarea
                id="workflow-description"
                bind:value={newWorkflow.description}
                placeholder="Describe the purpose and scope of this workflow"
                rows="3"
              ></textarea>
            </div>

            <div class="form-group">
              <label for="workflow-status">Status</label>
              <select id="workflow-status" bind:value={newWorkflow.status}>
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
                <option value="draft">Draft</option>
              </select>
            </div>
          </div>
        </div>

        <div class="form-section">
          <div class="section-header">
            <h3>Workflow Steps ({currentWorkflowSteps.length})</h3>
            <button class="btn-primary btn-sm" onclick={() => openStepModal()}>
              + Add Step
            </button>
          </div>

          {#if currentWorkflowSteps.length === 0}
            <div class="empty-state">
              No steps added yet. Click "Add Step" to create your first step.
            </div>
          {:else}
            <div class="steps-list">
              {#each currentWorkflowSteps as step, index}
                <div class="step-card">
                  <div class="step-order">
                    <span>{step.step_order}</span>
                    <div class="step-controls">
                      <button 
                        class="btn-icon" 
                        onclick={() => moveStep(index, 'up')}
                        disabled={index === 0}
                      >
                        ▲
                      </button>
                      <button 
                        class="btn-icon" 
                        onclick={() => moveStep(index, 'down')}
                        disabled={index === currentWorkflowSteps.length - 1}
                      >
                        ▼
                      </button>
                    </div>
                  </div>
                  
                  <div class="step-content">
                    <div class="step-header">
                      <h4>{step.step_name}</h4>
                      <div class="step-badges">
                        <span class="badge badge-type">{getStepTypeLabel(step.step_type)}</span>
                        {#if step.required}
                          <span class="badge badge-required">Required</span>
                        {/if}
                        {#if step.auto_trigger}
                          <span class="badge badge-auto">Auto</span>
                        {/if}
                      </div>
                    </div>
                    <p class="step-description">{step.description}</p>
                    <div class="step-meta">
                      <span>Assigned: {getRoleLabel(step.assigned_role || 'hr')}</span>
                      {#if step.due_days !== undefined}
                        <span>•</span>
                        <span>Due: {step.due_days} days</span>
                      {/if}
                    </div>
                  </div>

                  <div class="step-actions">
                    <button class="btn-icon" onclick={() => openStepModal(index)}>✏️</button>
                    <button class="btn-icon" onclick={() => deleteStep(index)}>🗑️</button>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-secondary" onclick={closeWorkflowModal}>Cancel</button>
        <button class="btn-primary" onclick={saveWorkflow} disabled={loading}>
          {loading ? 'Saving...' : editingWorkflow ? 'Update Workflow' : 'Create Workflow'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Step Modal -->
{#if showStepModal}
  <div class="modal-overlay" role="button" tabindex="0" onclick={closeStepModal} onkeydown={(e) => e.key === "Escape" && closeStepModal()}>
    <div class="modal" role="dialog" aria-modal="true" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2>{editingStepIndex !== null ? 'Edit Step' : 'Add Step'}</h2>
        <button class="close-btn" onclick={closeStepModal}>×</button>
      </div>

      <div class="modal-body">
        <div class="form-grid">
          <div class="form-group">
            <label for="step-order">Step Order</label>
            <input
              id="step-order"
              type="number"
              bind:value={newStep.step_order}
              min="1"
              required
            />
          </div>

          <div class="form-group">
            <label for="step-type">Step Type *</label>
            <select id="step-type" bind:value={newStep.step_type} required>
              {#each stepTypes as type}
                <option value={type.value}>{type.label}</option>
              {/each}
            </select>
          </div>

          <div class="form-group full-width">
            <label for="step-name">Step Name *</label>
            <input
              id="step-name"
              type="text"
              bind:value={newStep.step_name}
              placeholder="e.g., Complete I-9 Form"
              required
            />
          </div>

          <div class="form-group full-width">
            <label for="step-description">Description</label>
            <textarea
              id="step-description"
              bind:value={newStep.description}
              placeholder="Detailed description of what needs to be done"
              rows="3"
            ></textarea>
          </div>

          <div class="form-group">
            <label for="assigned-role">Assigned To *</label>
            <select id="assigned-role" bind:value={newStep.assigned_role} required>
              {#each roles as role}
                <option value={role.value}>{role.label}</option>
              {/each}
            </select>
          </div>

          <div class="form-group">
            <label for="due-days">Due Days</label>
            <input
              id="due-days"
              type="number"
              bind:value={newStep.due_days}
              min="0"
              placeholder="Days to complete"
            />
          </div>

          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" bind:checked={newStep.required} />
              <span>Required Step</span>
            </label>
          </div>

          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" bind:checked={newStep.auto_trigger} />
              <span>Auto-trigger</span>
            </label>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-secondary" onclick={closeStepModal}>Cancel</button>
        <button class="btn-primary" onclick={saveStep}>
          {editingStepIndex !== null ? 'Update Step' : 'Add Step'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .workflow-manager {
    padding: 24px;
  }

  .loading, .error {
    text-align: center;
    padding: 48px;
    color: #666;
  }

  .error {
    color: #dc3545;
  }

  .error-banner {
    background: #f8d7da;
    color: #721c24;
    padding: 12px;
    border-radius: 4px;
    margin-bottom: 16px;
  }

  /* Quick Actions */
  .quick-actions {
    margin-bottom: 32px;
  }

  .quick-actions h3 {
    margin: 0 0 16px 0;
    font-size: 18px;
    font-weight: 600;
  }

  .template-buttons {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 12px;
  }

  .btn-template {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 20px;
    background: white;
    border: 2px solid #e9ecef;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-template:hover {
    border-color: #007bff;
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0,0,0,0.1);
  }

  .btn-template .icon {
    font-size: 32px;
  }

  .btn-template span:last-child {
    font-weight: 500;
    text-align: center;
  }

  .btn-template.onboarding:hover { border-color: #28a745; }
  .btn-template.offboarding:hover { border-color: #dc3545; }
  .btn-template.performance:hover { border-color: #ffc107; }
  .btn-template.custom:hover { border-color: #6c757d; }

  /* Workflows List */
  .workflows-list h3 {
    margin: 0 0 16px 0;
    font-size: 18px;
    font-weight: 600;
  }

  .workflows-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 16px;
  }

  .workflow-card {
    background: white;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    padding: 20px;
  }

  .workflow-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 12px;
  }

  .workflow-card h4 {
    margin: 0 0 4px 0;
    font-size: 16px;
    font-weight: 600;
  }

  .workflow-type {
    display: inline-block;
    padding: 2px 8px;
    background: #f8f9fa;
    border-radius: 4px;
    font-size: 12px;
    color: #666;
  }

  .status-badge {
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .status-active { background: #d4edda; color: #155724; }
  .status-inactive { background: #f8d7da; color: #721c24; }
  .status-draft { background: #fff3cd; color: #856404; }

  .workflow-description {
    color: #666;
    font-size: 14px;
    margin: 0 0 12px 0;
  }

  .workflow-meta {
    display: flex;
    gap: 8px;
    font-size: 13px;
    color: #999;
    margin-bottom: 16px;
  }

  .workflow-actions {
    display: flex;
    gap: 8px;
  }

  /* Modal */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: white;
    border-radius: 8px;
    width: 90%;
    max-width: 600px;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }

  .modal.large {
    max-width: 900px;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid #e9ecef;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 20px;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 32px;
    line-height: 1;
    color: #999;
    cursor: pointer;
    padding: 0;
    width: 32px;
    height: 32px;
  }

  .close-btn:hover {
    color: #333;
  }

  .modal-body {
    padding: 24px;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 16px 24px;
    border-top: 1px solid #e9ecef;
  }

  /* Form */
  .form-section {
    margin-bottom: 32px;
  }

  .form-section h3 {
    margin: 0 0 16px 0;
    font-size: 16px;
    font-weight: 600;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
  }

  .form-group.full-width {
    grid-column: 1 / -1;
  }

  .form-group label {
    font-size: 13px;
    font-weight: 600;
    margin-bottom: 4px;
    color: #495057;
  }

  .form-group input,
  .form-group select,
  .form-group textarea {
    padding: 8px 12px;
    border: 1px solid #dee2e6;
    border-radius: 4px;
    font-size: 14px;
  }

  .form-group textarea {
    resize: vertical;
    font-family: inherit;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
  }

  .checkbox-label input[type="checkbox"] {
    width: auto;
    cursor: pointer;
  }

  /* Steps List */
  .steps-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .step-card {
    display: flex;
    gap: 16px;
    padding: 16px;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    background: #f8f9fa;
  }

  .step-order {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .step-order > span {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #007bff;
    color: white;
    border-radius: 50%;
    font-weight: 600;
    font-size: 14px;
  }

  .step-controls {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .step-content {
    flex: 1;
  }

  .step-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 8px;
  }

  .step-header h4 {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
  }

  .step-badges {
    display: flex;
    gap: 4px;
  }

  .badge {
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
  }

  .badge-type { background: #e7f3ff; color: #004085; }
  .badge-required { background: #f8d7da; color: #721c24; }
  .badge-auto { background: #d4edda; color: #155724; }

  .step-description {
    margin: 0 0 8px 0;
    font-size: 13px;
    color: #666;
  }

  .step-meta {
    display: flex;
    gap: 8px;
    font-size: 12px;
    color: #999;
  }

  .step-actions {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .btn-icon {
    background: none;
    border: none;
    font-size: 16px;
    cursor: pointer;
    padding: 4px;
    opacity: 0.7;
  }

  .btn-icon:hover {
    opacity: 1;
  }

  .btn-icon:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  /* Buttons */
  .btn-primary, .btn-secondary, .btn-danger {
    padding: 10px 20px;
    border: none;
    border-radius: 4px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary {
    background: #007bff;
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background: #0056b3;
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-secondary {
    background: #6c757d;
    color: white;
  }

  .btn-secondary:hover {
    background: #545b62;
  }

  .btn-danger {
    background: #dc3545;
    color: white;
  }

  .btn-danger:hover {
    background: #bd2130;
  }

  .btn-sm {
    padding: 6px 12px;
    font-size: 13px;
  }

  .empty-state {
    text-align: center;
    padding: 48px;
    color: #999;
    font-style: italic;
  }
</style>