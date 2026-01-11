<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  
  const dispatch = createEventDispatcher();
  
  interface WorkflowInstance {
    id: string;
    employee_id: string;
    employee_name: string;
    employee_email: string;
    employee_department: string;
    template_name: string;
    workflow_type: string;
    status: string;
    progress: number;
    start_date: string;
    due_date: string;
    completed_date?: string;
    total_steps: number;
    completed_steps: number;
    pending_steps: number;
    overdue_steps: number;
    assigned_to: string;
    last_activity: string;
  }
  
  interface WorkflowTask {
    id: string;
    name: string;
    description: string;
    status: string;
    due_date: string;
    completed_date?: string;
    assignee: string;
  }
  
  let workflows: WorkflowInstance[] = [];
  let selectedWorkflow: WorkflowInstance | null = null;
  let workflowTasks: WorkflowTask[] = [];
  let loading = true;
  let showDetailModal = false;
  
  let filterStatus = '';
  let filterType = '';
  let searchQuery = '';
  let sortBy = 'due_date';
  
  const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'active', label: 'Active' },
    { value: 'completed', label: 'Completed' },
    { value: 'on_hold', label: 'On Hold' },
    { value: 'overdue', label: 'Overdue' }
  ];
  
  const typeOptions = [
    { value: '', label: 'All Types' },
    { value: 'onboarding', label: 'Onboarding' },
    { value: 'offboarding', label: 'Offboarding' },
    { value: 'promotion', label: 'Promotion' },
    { value: 'transfer', label: 'Transfer' }
  ];
  
  const sortOptions = [
    { value: 'due_date', label: 'Due Date' },
    { value: 'progress', label: 'Progress' },
    { value: 'start_date', label: 'Start Date' },
    { value: 'employee_name', label: 'Employee Name' }
  ];
  
  onMount(async () => {
    await loadWorkflows();
  });
  
  async function loadWorkflows() {
    try {
      loading = true;
      const token = localStorage.getItem('token');
      const response = await fetch('/api/onboarding/workflows', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        workflows = await response.json();
      }
    } catch (err) {
      console.error('Failed to load workflows:', err);
    } finally {
      loading = false;
    }
  }
  
  async function viewWorkflowDetails(workflow: WorkflowInstance) {
    selectedWorkflow = workflow;
    await loadWorkflowTasks(workflow.id);
    showDetailModal = true;
  }
  
  async function loadWorkflowTasks(workflowId: string) {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/onboarding/workflows/${workflowId}/tasks`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        workflowTasks = await response.json();
      }
    } catch (err) {
      console.error('Failed to load workflow tasks:', err);
    }
  }
  
  async function updateWorkflowStatus(workflowId: string, newStatus: string) {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/onboarding/workflows/${workflowId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ status: newStatus })
      });
      
      if (response.ok) {
        await loadWorkflows();
        dispatch('updated');
      }
    } catch (err) {
      console.error('Failed to update workflow status:', err);
    }
  }
  
  $: filteredWorkflows = workflows.filter(wf => {
    const matchesStatus = !filterStatus || wf.status === filterStatus;
    const matchesType = !filterType || wf.workflow_type === filterType;
    const matchesSearch = !searchQuery || 
      wf.employee_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      wf.template_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      wf.employee_email.toLowerCase().includes(searchQuery.toLowerCase());
    
    return matchesStatus && matchesType && matchesSearch;
  }).sort((a, b) => {
    switch (sortBy) {
      case 'progress':
        return b.progress - a.progress;
      case 'start_date':
        return new Date(b.start_date).getTime() - new Date(a.start_date).getTime();
      case 'employee_name':
        return a.employee_name.localeCompare(b.employee_name);
      default: // due_date
        return new Date(a.due_date).getTime() - new Date(b.due_date).getTime();
    }
  });
  
  function getStatusColor(status: string) {
    const colors: Record<string, string> = {
      'active': '#48bb78',
      'completed': '#4299e1',
      'on_hold': '#ed8936',
      'overdue': '#f56565'
    };
    return colors[status] || '#a0aec0';
  }
  
  function getStatusIcon(status: string) {
    const icons: Record<string, string> = {
      'active': '🚀',
      'completed': '✅',
      'on_hold': '⏸️',
      'overdue': '⚠️'
    };
    return icons[status] || '📋';
  }
  
  function formatDate(date: string) {
    return new Date(date).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }
  
  function getDaysUntilDue(dueDate: string) {
    const days = Math.ceil((new Date(dueDate).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24));
    return days;
  }
  
  function getProgressClass(progress: number) {
    if (progress < 30) return 'low';
    if (progress < 70) return 'medium';
    return 'high';
  }
</script>

<div class="workflow-monitoring">
  <div class="monitoring-header">
    <div>
      <h2>Workflow Monitoring</h2>
      <p>Track and manage employee workflow progress in real-time</p>
    </div>
    <button class="btn-refresh" on:click={loadWorkflows}>
      🔄 Refresh
    </button>
  </div>

  <!-- Summary Cards -->
  <div class="summary-cards">
    <div class="summary-card">
      <div class="summary-icon">🚀</div>
      <div class="summary-content">
        <div class="summary-value">{workflows.filter(w => w.status === 'active').length}</div>
        <div class="summary-label">Active</div>
      </div>
    </div>
    
    <div class="summary-card">
      <div class="summary-icon">⚠️</div>
      <div class="summary-content">
        <div class="summary-value">{workflows.filter(w => w.overdue_steps > 0).length}</div>
        <div class="summary-label">With Overdue Steps</div>
      </div>
    </div>
    
    <div class="summary-card">
      <div class="summary-icon">✅</div>
      <div class="summary-content">
        <div class="summary-value">{workflows.filter(w => w.status === 'completed').length}</div>
        <div class="summary-label">Completed</div>
      </div>
    </div>
    
    <div class="summary-card">
      <div class="summary-icon">📊</div>
      <div class="summary-content">
        <div class="summary-value">
          {workflows.length > 0 ? Math.round(workflows.reduce((sum, w) => sum + w.progress, 0) / workflows.length) : 0}%
        </div>
        <div class="summary-label">Avg Progress</div>
      </div>
    </div>
  </div>

  <!-- Filters and Sort -->
  <div class="controls">
    <input
      type="text"
      class="search-input"
      placeholder="🔍 Search by employee name, email, or workflow..."
      bind:value={searchQuery}
    />
    
    <select class="filter-select" bind:value={filterStatus}>
      {#each statusOptions as option}
        <option value={option.value}>{option.label}</option>
      {/each}
    </select>
    
    <select class="filter-select" bind:value={filterType}>
      {#each typeOptions as option}
        <option value={option.value}>{option.label}</option>
      {/each}
    </select>
    
    <select class="filter-select" bind:value={sortBy}>
      {#each sortOptions as option}
        <option value={option.value}>Sort: {option.label}</option>
      {/each}
    </select>
  </div>

  <!-- Workflows List -->
  {#if loading}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading workflows...</p>
    </div>
  {:else if filteredWorkflows.length === 0}
    <div class="empty-state">
      <div class="empty-icon">📋</div>
      <h3>No Workflows Found</h3>
      <p>No workflows match your current filters</p>
    </div>
  {:else}
    <div class="workflows-list">
      {#each filteredWorkflows as workflow}
        <div class="workflow-card" on:click={() => viewWorkflowDetails(workflow)}>
          <div class="workflow-header">
            <div class="workflow-employee">
              <div class="employee-avatar">{workflow.employee_name[0]}</div>
              <div>
                <div class="employee-name">{workflow.employee_name}</div>
                <div class="employee-meta">{workflow.employee_department} • {workflow.employee_email}</div>
              </div>
            </div>
            <div class="workflow-status">
              <span class="status-badge" style="background-color: {getStatusColor(workflow.status)}">
                {getStatusIcon(workflow.status)} {workflow.status}
              </span>
            </div>
          </div>

          <div class="workflow-body">
            <div class="workflow-info">
              <h3>{workflow.template_name}</h3>
              <div class="workflow-meta">
                <span class="meta-item">
                  <span class="meta-icon">📅</span>
                  Started: {formatDate(workflow.start_date)}
                </span>
                <span class="meta-item">
                  <span class="meta-icon">⏰</span>
                  Due: {formatDate(workflow.due_date)}
                  {#if getDaysUntilDue(workflow.due_date) < 0}
                    <span class="overdue-badge">Overdue by {Math.abs(getDaysUntilDue(workflow.due_date))}d</span>
                  {:else}
                    <span class="days-remaining">{getDaysUntilDue(workflow.due_date)}d left</span>
                  {/if}
                </span>
                <span class="meta-item">
                  <span class="meta-icon">👤</span>
                  Assigned to: {workflow.assigned_to}
                </span>
              </div>
            </div>

            <div class="workflow-progress">
              <div class="progress-header">
                <span>Progress</span>
                <span class="progress-percentage">{Math.round(workflow.progress)}%</span>
              </div>
              <div class="progress-bar">
                <div 
                  class="progress-fill {getProgressClass(workflow.progress)}"
                  style="width: {workflow.progress}%"
                ></div>
              </div>
              <div class="steps-summary">
                <span class="step-stat complete">✓ {workflow.completed_steps} completed</span>
                <span class="step-stat pending">○ {workflow.pending_steps} pending</span>
                {#if workflow.overdue_steps > 0}
                  <span class="step-stat overdue">⚠ {workflow.overdue_steps} overdue</span>
                {/if}
              </div>
            </div>
          </div>

          <div class="workflow-actions">
            <button class="btn-action" on:click|stopPropagation={() => viewWorkflowDetails(workflow)}>
              View Details
            </button>
            {#if workflow.status === 'active'}
              <button class="btn-action secondary" on:click|stopPropagation={() => updateWorkflowStatus(workflow.id, 'on_hold')}>
                Put On Hold
              </button>
            {:else if workflow.status === 'on_hold'}
              <button class="btn-action secondary" on:click|stopPropagation={() => updateWorkflowStatus(workflow.id, 'active')}>
                Resume
              </button>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Detail Modal -->
{#if showDetailModal && selectedWorkflow}
  <div class="modal-overlay" on:click={() => showDetailModal = false}>
    <div class="modal large" on:click|stopPropagation>
      <div class="modal-header">
        <div>
          <h2>{selectedWorkflow.template_name}</h2>
          <p>{selectedWorkflow.employee_name} • {selectedWorkflow.employee_department}</p>
        </div>
        <button class="close-btn" on:click={() => showDetailModal = false}>✕</button>
      </div>
      
      <div class="modal-body">
        <!-- Progress Overview -->
        <div class="detail-progress">
          <div class="circular-progress">
            <svg viewBox="0 0 120 120">
              <circle cx="60" cy="60" r="54" fill="none" stroke="#e2e8f0" stroke-width="8"/>
              <circle 
                cx="60" cy="60" r="54" 
                fill="none" 
                stroke={getStatusColor(selectedWorkflow.status)}
                stroke-width="8"
                stroke-dasharray="{selectedWorkflow.progress * 3.39}, 339"
                transform="rotate(-90 60 60)"
                stroke-linecap="round"
              />
            </svg>
            <div class="progress-center">
              <div class="progress-value">{Math.round(selectedWorkflow.progress)}%</div>
              <div class="progress-label">Complete</div>
            </div>
          </div>

          <div class="detail-stats">
            <div class="stat-item">
              <div class="stat-label">Total Steps</div>
              <div class="stat-value">{selectedWorkflow.total_steps}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">Completed</div>
              <div class="stat-value complete">{selectedWorkflow.completed_steps}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">Pending</div>
              <div class="stat-value pending">{selectedWorkflow.pending_steps}</div>
            </div>
            {#if selectedWorkflow.overdue_steps > 0}
              <div class="stat-item">
                <div class="stat-label">Overdue</div>
                <div class="stat-value overdue">{selectedWorkflow.overdue_steps}</div>
              </div>
            {/if}
          </div>
        </div>

        <!-- Task List -->
        <div class="tasks-section">
          <h3>Workflow Tasks</h3>
          {#if workflowTasks.length === 0}
            <p>Loading tasks...</p>
          {:else}
            <div class="tasks-list">
              {#each workflowTasks as task}
                <div class="task-item" class:completed={task.status === 'completed'}>
                  <div class="task-status-icon">
                    {#if task.status === 'completed'}
                      ✅
                    {:else}
                      ⭕
                    {/if}
                  </div>
                  <div class="task-details">
                    <div class="task-name">{task.name}</div>
                    <div class="task-meta">
                      {#if task.description}
                        <span>{task.description}</span>
                      {/if}
                      <span>Assigned to: {task.assignee}</span>
                      <span>Due: {formatDate(task.due_date)}</span>
                    </div>
                  </div>
                  <div class="task-status">
                    <span class="status-pill small" style="background-color: {getStatusColor(task.status)}">
                      {task.status}
                    </span>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
      
      <div class="modal-footer">
        <button class="btn-secondary" on:click={() => showDetailModal = false}>Close</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .workflow-monitoring {
    padding: 24px;
  }

  .monitoring-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 32px;
  }

  .monitoring-header h2 {
    font-size: 24px;
    font-weight: 700;
    color: #1a202c;
    margin: 0 0 8px 0;
  }

  .monitoring-header p {
    font-size: 14px;
    color: #718096;
    margin: 0;
  }

  .btn-refresh {
    padding: 10px 20px;
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-refresh:hover {
    background: #f7fafc;
    border-color: #cbd5e0;
  }

  /* Summary Cards */
  .summary-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;
    margin-bottom: 24px;
  }

  .summary-card {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px;
    background: white;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
  }

  .summary-icon {
    font-size: 36px;
  }

  .summary-value {
    font-size: 32px;
    font-weight: 700;
    color: #1a202c;
    line-height: 1;
  }

  .summary-label {
    font-size: 13px;
    color: #718096;
    margin-top: 4px;
  }

  /* Controls */
  .controls {
    display: grid;
    grid-template-columns: 2fr 1fr 1fr 1fr;
    gap: 12px;
    margin-bottom: 24px;
  }

  .search-input,
  .filter-select {
    padding: 10px 16px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    font-size: 14px;
  }

  /* Workflows List */
  .workflows-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .workflow-card {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 24px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .workflow-card:hover {
    border-color: #cbd5e0;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }

  .workflow-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .workflow-employee {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .employee-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 18px;
  }

  .employee-name {
    font-size: 16px;
    font-weight: 600;
    color: #1a202c;
  }

  .employee-meta {
    font-size: 12px;
    color: #718096;
  }

  .status-badge {
    padding: 6px 12px;
    border-radius: 12px;
    color: white;
    font-size: 12px;
    font-weight: 600;
    text-transform: capitalize;
  }

  .status-badge.small {
    padding: 4px 8px;
    font-size: 11px;
  }

  .workflow-body {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 24px;
    margin-bottom: 20px;
  }

  .workflow-info h3 {
    font-size: 18px;
    font-weight: 600;
    color: #1a202c;
    margin: 0 0 12px 0;
  }

  .workflow-meta {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .meta-item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #4a5568;
  }

  .meta-icon {
    font-size: 14px;
  }

  .overdue-badge {
    padding: 2px 8px;
    background: #fed7d7;
    color: #c53030;
    border-radius: 6px;
    font-size: 11px;
    font-weight: 600;
    margin-left: 8px;
  }

  .days-remaining {
    color: #48bb78;
    font-weight: 600;
    margin-left: 8px;
  }

  .workflow-progress {
    min-width: 280px;
  }

  .progress-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
    font-size: 13px;
    color: #4a5568;
  }

  .progress-percentage {
    font-weight: 700;
    color: #1a202c;
  }

  .progress-bar {
    height: 8px;
    background: #e2e8f0;
    border-radius: 4px;
    overflow: hidden;
    margin-bottom: 12px;
  }

  .progress-fill {
    height: 100%;
    transition: width 0.3s;
  }

  .progress-fill.low {
    background: linear-gradient(90deg, #fc8181 0%, #f56565 100%);
  }

  .progress-fill.medium {
    background: linear-gradient(90deg, #f6ad55 0%, #ed8936 100%);
  }

  .progress-fill.high {
    background: linear-gradient(90deg, #68d391 0%, #48bb78 100%);
  }

  .steps-summary {
    display: flex;
    gap: 16px;
    font-size: 12px;
  }

  .step-stat {
    font-weight: 600;
  }

  .step-stat.complete {
    color: #48bb78;
  }

  .step-stat.pending {
    color: #718096;
  }

  .step-stat.overdue {
    color: #f56565;
  }

  .workflow-actions {
    display: flex;
    gap: 12px;
  }

  .btn-action {
    padding: 8px 16px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-action.secondary {
    background: white;
    color: #4a5568;
    border: 1px solid #e2e8f0;
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
    max-width: 900px;
    width: 100%;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
  }

  .modal.large {
    max-width: 1000px;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 24px;
    border-bottom: 1px solid #e2e8f0;
  }

  .modal-header h2 {
    font-size: 20px;
    font-weight: 700;
    margin: 0 0 4px 0;
  }

  .modal-header p {
    font-size: 14px;
    color: #718096;
    margin: 0;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 24px;
    color: #718096;
    cursor: pointer;
  }

  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    padding: 24px;
    border-top: 1px solid #e2e8f0;
  }

  .detail-progress {
    display: flex;
    gap: 32px;
    margin-bottom: 32px;
    padding: 24px;
    background: #f7fafc;
    border-radius: 12px;
  }

  .circular-progress {
    position: relative;
    width: 120px;
    height: 120px;
  }

  .circular-progress svg {
    transform: rotate(0deg);
  }

  .progress-center {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    text-align: center;
  }

  .progress-center .progress-value {
    font-size: 24px;
    font-weight: 700;
    color: #1a202c;
  }

  .progress-center .progress-label {
    font-size: 11px;
    color: #718096;
  }

  .detail-stats {
    flex: 1;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
    gap: 20px;
  }

  .stat-item {
    text-align: center;
  }

  .stat-label {
    font-size: 12px;
    color: #718096;
    margin-bottom: 8px;
  }

  .stat-value {
    font-size: 32px;
    font-weight: 700;
    color: #1a202c;
  }

  .stat-value.complete {
    color: #48bb78;
  }

  .stat-value.pending {
    color: #718096;
  }

  .stat-value.overdue {
    color: #f56565;
  }

  .tasks-section h3 {
    font-size: 18px;
    font-weight: 600;
    margin: 0 0 20px 0;
  }

  .tasks-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .task-item {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px;
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
  }

  .task-item.completed {
    opacity: 0.6;
  }

  .task-status-icon {
    font-size: 20px;
    min-width: 24px;
  }

  .task-details {
    flex: 1;
  }

  .task-name {
    font-weight: 600;
    color: #1a202c;
    margin-bottom: 4px;
  }

  .task-meta {
    display: flex;
    gap: 12px;
    font-size: 12px;
    color: #718096;
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

  .loading,
  .empty-state {
    text-align: center;
    padding: 60px 20px;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid #e2e8f0;
    border-top-color: #667eea;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 16px;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
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
    margin: 0;
  }

  @media (max-width: 768px) {
    .controls {
      grid-template-columns: 1fr;
    }

    .workflow-body {
      grid-template-columns: 1fr;
    }

    .workflow-progress {
      min-width: auto;
    }

    .summary-cards {
      grid-template-columns: repeat(2, 1fr);
    }

    .detail-progress {
      flex-direction: column;
    }

    .circular-progress {
      margin: 0 auto;
    }
  }
</style>
