<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  
  const dispatch = createEventDispatcher();
  
  interface WorkflowTemplate {
    id: string;
    name: string;
    description: string;
    type: string;
    steps: WorkflowStep[];
    estimated_days: number;
    is_active: boolean;
    created_at: string;
  }
  
  interface WorkflowStep {
    id?: string;
    name: string;
    description: string;
    order: number;
    estimated_days: number;
    required: boolean;
    assignee_role?: string;
  }
  
  let templates: WorkflowTemplate[] = [];
  let loading = true;
  let showCreateModal = false;
  let editingTemplate: WorkflowTemplate | null = null;
  
  // Form state
  let formData = {
    name: '',
    description: '',
    type: 'onboarding',
    estimated_days: 14,
    steps: [] as WorkflowStep[]
  };
  
  let newStep: WorkflowStep = {
    name: '',
    description: '',
    order: 1,
    estimated_days: 1,
    required: true,
    assignee_role: 'HR'
  };
  
  const workflowTypes = [
    { value: 'onboarding', label: 'New Hire Onboarding' },
    { value: 'offboarding', label: 'Employee Offboarding' },
    { value: 'promotion', label: 'Promotion Process' },
    { value: 'transfer', label: 'Department Transfer' },
    { value: 'performance', label: 'Performance Review' },
    { value: 'custom', label: 'Custom Workflow' }
  ];
  
  const roles = ['HR', 'Manager', 'IT', 'Admin', 'Finance', 'Employee'];
  
  onMount(async () => {
    await loadTemplates();
  });
  
  async function loadTemplates() {
    try {
      loading = true;
      const token = localStorage.getItem('token');
      const response = await fetch('/api/workflows/templates', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        templates = await response.json();
      }
    } catch (err) {
      console.error('Failed to load templates:', err);
    } finally {
      loading = false;
    }
  }
  
  function openCreateModal() {
    editingTemplate = null;
    resetForm();
    showCreateModal = true;
  }
  
  function openEditModal(template: WorkflowTemplate) {
    editingTemplate = template;
    formData = {
      name: template.name,
      description: template.description,
      type: template.type,
      estimated_days: template.estimated_days,
      steps: [...template.steps]
    };
    showCreateModal = true;
  }
  
  function resetForm() {
    formData = {
      name: '',
      description: '',
      type: 'onboarding',
      estimated_days: 14,
      steps: []
    };
    newStep = {
      name: '',
      description: '',
      order: 1,
      estimated_days: 1,
      required: true,
      assignee_role: 'HR'
    };
  }
  
  function addStep() {
    if (!newStep.name) return;
    
    formData.steps = [...formData.steps, {
      ...newStep,
      order: formData.steps.length + 1
    }];
    
    newStep = {
      name: '',
      description: '',
      order: formData.steps.length + 2,
      estimated_days: 1,
      required: true,
      assignee_role: 'HR'
    };
  }
  
  function removeStep(index: number) {
    formData.steps = formData.steps.filter((_, i) => i !== index);
    // Reorder remaining steps
    formData.steps.forEach((step, i) => step.order = i + 1);
  }
  
  function moveStepUp(index: number) {
    if (index === 0) return;
    const steps = [...formData.steps];
    [steps[index - 1], steps[index]] = [steps[index], steps[index - 1]];
    steps.forEach((step, i) => step.order = i + 1);
    formData.steps = steps;
  }
  
  function moveStepDown(index: number) {
    if (index === formData.steps.length - 1) return;
    const steps = [...formData.steps];
    [steps[index], steps[index + 1]] = [steps[index + 1], steps[index]];
    steps.forEach((step, i) => step.order = i + 1);
    formData.steps = steps;
  }
  
  async function saveTemplate() {
    if (!formData.name || formData.steps.length === 0) {
      alert('Please provide a name and at least one step');
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      const method = editingTemplate ? 'PUT' : 'POST';
      const url = editingTemplate 
        ? `/api/workflows/templates/${editingTemplate.id}`
        : '/api/workflows/templates';
      
      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData)
      });
      
      if (response.ok) {
        showCreateModal = false;
        await loadTemplates();
        dispatch('updated');
      } else {
        const error = await response.text();
        alert(`Failed to save template: ${error}`);
      }
    } catch (err) {
      console.error('Failed to save template:', err);
      alert('Failed to save template');
    }
  }
  
  async function toggleTemplate(template: WorkflowTemplate) {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/workflows/templates/${template.id}/toggle`, {
        method: 'PUT',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        await loadTemplates();
        dispatch('updated');
      }
    } catch (err) {
      console.error('Failed to toggle template:', err);
    }
  }
  
  async function deleteTemplate(id: string) {
    if (!confirm('Are you sure you want to delete this template?')) return;
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/workflows/templates/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        await loadTemplates();
        dispatch('updated');
      }
    } catch (err) {
      console.error('Failed to delete template:', err);
    }
  }
</script>

<div class="template-manager">
  <div class="manager-header">
    <div>
      <h2>Workflow Templates</h2>
      <p>Create and manage reusable workflow templates for common processes</p>
    </div>
    <button class="btn-primary" on:click={openCreateModal}>
      <span>➕</span>
      Create Template
    </button>
  </div>

  {#if loading}
    <div class="loading">Loading templates...</div>
  {:else if templates.length === 0}
    <div class="empty-state">
      <div class="empty-icon">📋</div>
      <h3>No Templates Yet</h3>
      <p>Create your first workflow template to get started</p>
      <button class="btn-primary" on:click={openCreateModal}>
        Create First Template
      </button>
    </div>
  {:else}
    <div class="templates-grid">
      {#each templates as template}
        <div class="template-card" class:inactive={!template.is_active}>
          <div class="template-header">
            <div class="template-type">
              <span class="type-badge">{template.type}</span>
              {#if !template.is_active}
                <span class="status-badge inactive">Inactive</span>
              {/if}
            </div>
            <div class="template-actions">
              <button class="icon-btn" on:click={() => toggleTemplate(template)} title={template.is_active ? 'Deactivate' : 'Activate'}>
                {template.is_active ? '⏸️' : '▶️'}
              </button>
              <button class="icon-btn" on:click={() => openEditModal(template)} title="Edit">
                ✏️
              </button>
              <button class="icon-btn danger" on:click={() => deleteTemplate(template.id)} title="Delete">
                🗑️
              </button>
            </div>
          </div>
          
          <h3 class="template-name">{template.name}</h3>
          <p class="template-description">{template.description}</p>
          
          <div class="template-meta">
            <div class="meta-item">
              <span class="meta-icon">📝</span>
              <span>{template.steps.length} steps</span>
            </div>
            <div class="meta-item">
              <span class="meta-icon">⏱️</span>
              <span>~{template.estimated_days} days</span>
            </div>
          </div>
          
          <div class="template-steps">
            {#each template.steps.slice(0, 3) as step, i}
              <div class="step-preview">
                <span class="step-number">{i + 1}</span>
                <span class="step-name">{step.name}</span>
              </div>
            {/each}
            {#if template.steps.length > 3}
              <div class="step-preview more">
                <span>+{template.steps.length - 3} more steps</span>
              </div>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Create/Edit Modal -->
{#if showCreateModal}
  <div class="modal-overlay" on:click={() => showCreateModal = false}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <h2>{editingTemplate ? 'Edit Template' : 'Create Workflow Template'}</h2>
        <button class="close-btn" on:click={() => showCreateModal = false}>✕</button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>Template Name *</label>
          <input type="text" bind:value={formData.name} placeholder="e.g., New Hire Onboarding">
        </div>
        
        <div class="form-group">
          <label>Description</label>
          <textarea bind:value={formData.description} placeholder="Describe the workflow purpose and process"></textarea>
        </div>
        
        <div class="form-row">
          <div class="form-group">
            <label>Workflow Type *</label>
            <select bind:value={formData.type}>
              {#each workflowTypes as type}
                <option value={type.value}>{type.label}</option>
              {/each}
            </select>
          </div>
          
          <div class="form-group">
            <label>Estimated Days</label>
            <input type="number" bind:value={formData.estimated_days} min="1" max="365">
          </div>
        </div>
        
        <div class="steps-section">
          <h3>Workflow Steps</h3>
          
          {#if formData.steps.length > 0}
            <div class="steps-list">
              {#each formData.steps as step, index}
                <div class="step-item">
                  <div class="step-order">
                    <button class="move-btn" on:click={() => moveStepUp(index)} disabled={index === 0}>▲</button>
                    <span class="step-num">{step.order}</span>
                    <button class="move-btn" on:click={() => moveStepDown(index)} disabled={index === formData.steps.length - 1}>▼</button>
                  </div>
                  <div class="step-details">
                    <div class="step-info">
                      <strong>{step.name}</strong>
                      <span class="step-meta">
                        {step.description} • {step.estimated_days}d • {step.assignee_role}
                        {#if step.required}<span class="required-badge">Required</span>{/if}
                      </span>
                    </div>
                    <button class="remove-btn" on:click={() => removeStep(index)}>🗑️</button>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
          
          <div class="add-step-form">
            <h4>Add Step</h4>
            <div class="form-row">
              <div class="form-group flex-2">
                <input type="text" bind:value={newStep.name} placeholder="Step name *">
              </div>
              <div class="form-group flex-1">
                <input type="number" bind:value={newStep.estimated_days} placeholder="Days" min="1">
              </div>
            </div>
            <div class="form-row">
              <div class="form-group flex-2">
                <input type="text" bind:value={newStep.description} placeholder="Step description">
              </div>
              <div class="form-group flex-1">
                <select bind:value={newStep.assignee_role}>
                  {#each roles as role}
                    <option value={role}>{role}</option>
                  {/each}
                </select>
              </div>
            </div>
            <div class="form-row">
              <label class="checkbox-label">
                <input type="checkbox" bind:checked={newStep.required}>
                Required step
              </label>
              <button class="btn-secondary" on:click={addStep}>Add Step</button>
            </div>
          </div>
        </div>
      </div>
      
      <div class="modal-footer">
        <button class="btn-secondary" on:click={() => showCreateModal = false}>Cancel</button>
        <button class="btn-primary" on:click={saveTemplate}>
          {editingTemplate ? 'Update Template' : 'Create Template'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .template-manager {
    padding: 24px;
  }

  .manager-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 32px;
  }

  .manager-header h2 {
    font-size: 24px;
    font-weight: 700;
    color: #1a202c;
    margin: 0 0 8px 0;
  }

  .manager-header p {
    font-size: 14px;
    color: #718096;
    margin: 0;
  }

  /* Templates Grid */
  .templates-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 20px;
  }

  .template-card {
    background: white;
    border: 2px solid #e2e8f0;
    border-radius: 12px;
    padding: 20px;
    transition: all 0.2s;
  }

  .template-card:hover {
    border-color: #667eea;
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.1);
  }

  .template-card.inactive {
    opacity: 0.6;
    background: #f7fafc;
  }

  .template-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .template-type {
    display: flex;
    gap: 8px;
  }

  .type-badge {
    padding: 4px 12px;
    background: #667eea;
    color: white;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .status-badge {
    padding: 4px 12px;
    background: #cbd5e0;
    color: #4a5568;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
  }

  .template-actions {
    display: flex;
    gap: 8px;
  }

  .template-name {
    font-size: 18px;
    font-weight: 700;
    color: #1a202c;
    margin: 0 0 8px 0;
  }

  .template-description {
    font-size: 14px;
    color: #718096;
    margin: 0 0 16px 0;
    line-height: 1.5;
  }

  .template-meta {
    display: flex;
    gap: 16px;
    margin-bottom: 16px;
    padding-bottom: 16px;
    border-bottom: 1px solid #e2e8f0;
  }

  .meta-item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #4a5568;
  }

  .meta-icon {
    font-size: 16px;
  }

  .template-steps {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .step-preview {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
  }

  .step-number {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    background: #edf2f7;
    border-radius: 50%;
    font-weight: 600;
    color: #4a5568;
    flex-shrink: 0;
  }

  .step-name {
    color: #2d3748;
  }

  .step-preview.more {
    color: #718096;
    font-style: italic;
    margin-left: 32px;
  }

  /* Empty State */
  .empty-state {
    text-align: center;
    padding: 60px 20px;
  }

  .empty-icon {
    font-size: 64px;
    margin-bottom: 16px;
  }

  .empty-state h3 {
    font-size: 20px;
    font-weight: 600;
    color: #2d3748;
    margin: 0 0 8px 0;
  }

  .empty-state p {
    font-size: 14px;
    color: #718096;
    margin: 0 0 24px 0;
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
    padding: 20px;
  }

  .modal {
    background: white;
    border-radius: 12px;
    max-width: 800px;
    width: 100%;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px;
    border-bottom: 1px solid #e2e8f0;
  }

  .modal-header h2 {
    font-size: 20px;
    font-weight: 700;
    margin: 0;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 24px;
    color: #718096;
    cursor: pointer;
    padding: 0;
    width: 32px;
    height: 32px;
  }

  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 24px;
    border-top: 1px solid #e2e8f0;
  }

  /* Form Styles */
  .form-group {
    margin-bottom: 20px;
  }

  .form-group label {
    display: block;
    font-size: 14px;
    font-weight: 600;
    color: #2d3748;
    margin-bottom: 8px;
  }

  .form-group input,
  .form-group textarea,
  .form-group select {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid #e2e8f0;
    border-radius: 6px;
    font-size: 14px;
  }

  .form-group textarea {
    min-height: 80px;
    resize: vertical;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }

  .form-group.flex-1 {
    flex: 1;
  }

  .form-group.flex-2 {
    flex: 2;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
  }

  /* Steps Section */
  .steps-section {
    margin-top: 32px;
    padding-top: 24px;
    border-top: 2px solid #e2e8f0;
  }

  .steps-section h3 {
    font-size: 18px;
    font-weight: 700;
    margin: 0 0 20px 0;
  }

  .steps-list {
    margin-bottom: 24px;
  }

  .step-item {
    display: flex;
    gap: 12px;
    padding: 12px;
    background: #f7fafc;
    border-radius: 8px;
    margin-bottom: 8px;
  }

  .step-order {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }

  .move-btn {
    background: white;
    border: 1px solid #e2e8f0;
    width: 24px;
    height: 24px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 10px;
  }

  .move-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .step-num {
    font-weight: 700;
    color: #667eea;
  }

  .step-details {
    flex: 1;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .step-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .step-meta {
    font-size: 12px;
    color: #718096;
  }

  .required-badge {
    padding: 2px 6px;
    background: #f56565;
    color: white;
    border-radius: 4px;
    font-size: 10px;
    font-weight: 600;
    margin-left: 8px;
  }

  .remove-btn {
    background: none;
    border: none;
    font-size: 18px;
    cursor: pointer;
    color: #f56565;
  }

  .add-step-form {
    padding: 20px;
    background: #f7fafc;
    border-radius: 8px;
  }

  .add-step-form h4 {
    font-size: 14px;
    font-weight: 600;
    margin: 0 0 16px 0;
  }

  /* Buttons */
  .btn-primary {
    padding: 10px 20px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .btn-primary:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
  }

  .btn-secondary {
    padding: 10px 20px;
    background: white;
    color: #4a5568;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
  }

  .icon-btn {
    background: white;
    border: 1px solid #e2e8f0;
    width: 32px;
    height: 32px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 14px;
    transition: all 0.2s;
  }

  .icon-btn:hover {
    background: #f7fafc;
    border-color: #cbd5e0;
  }

  .icon-btn.danger:hover {
    background: #fed7d7;
    border-color: #fc8181;
  }

  .loading {
    text-align: center;
    padding: 60px 20px;
    color: #718096;
  }

  @media (max-width: 768px) {
    .templates-grid {
      grid-template-columns: 1fr;
    }

    .manager-header {
      flex-direction: column;
      gap: 16px;
    }

    .form-row {
      grid-template-columns: 1fr;
    }
  }
</style>
