<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';

  // Types
  interface Employee {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    department: string;
    position: string;
    status: string;
    hire_date: string;
  }

  interface OnboardingTask {
    id: string;
    employee_id: string;
    task_name: string;
    description?: string;
    category?: string;
    status: string;
    due_date?: string;
    completed_at?: string;
    assigned_to?: string;
    documents_required: boolean;
    document_url?: string;
    created_at: string;
    updated_at: string;
  }

  interface Workflow {
    id: string;
    employee_id: string;
    employee_name?: string;
    template_name: string;
    status: string;
    current_stage: string;
    progress_percentage: number;
    started_at: string;
    expected_completion: string;
    actual_completion?: string;
    created_at: string;
  }

  interface WorkflowStep {
    id: string;
    workflow_id: string;
    step_name: string;
    step_order: number;
    status: string;
    assigned_to?: string;
    due_date?: string;
    completed_at?: string;
    notes?: string;
  }

  // State
  let activeTab: 'tasks' | 'workflows' = 'tasks';
  let loading = false;
  let error = '';
  let success = '';

  // Tasks State
  let employees: Employee[] = [];
  let selectedEmployee: Employee | null = null;
  let tasks: OnboardingTask[] = [];
  let tasksLoading = false;
  let showAddTaskModal = false;
  let filterStatus = 'all';
  let filterCategory = 'all';

  let newTask = {
    task_name: '',
    description: '',
    category: 'documentation',
    due_date: '',
    documents_required: false,
    assigned_to: ''
  };

  // Workflows State
  let workflows: Workflow[] = [];
  let workflowsLoading = false;
  let showCreateWorkflowModal = false;
  let showWorkflowDetailModal = false;
  let selectedWorkflow: Workflow | null = null;
  let workflowSteps: WorkflowStep[] = [];
  let workflowStatusFilter = 'all';

  let newWorkflow = {
    employee_id: '',
    template_name: 'software-engineer'
  };

  const workflowTemplates = [
    { value: 'software-engineer', label: 'Software Engineer (17 steps)', duration: '90 days' },
    { value: 'generic', label: 'Generic Employee (4 steps)', duration: '30 days' },
    { value: 'sales-representative', label: 'Sales Representative (12 steps)', duration: '60 days' },
    { value: 'manager', label: 'Manager (15 steps)', duration: '90 days' }
  ];

  const taskCategories = [
    { value: 'documentation', label: 'Documentation', icon: '📄' },
    { value: 'equipment', label: 'Equipment', icon: '💻' },
    { value: 'training', label: 'Training', icon: '📚' },
    { value: 'hr', label: 'HR', icon: '👥' },
    { value: 'it', label: 'IT', icon: '🔧' },
    { value: 'compliance', label: 'Compliance', icon: '✓' }
  ];

  // Computed
  $: filteredTasks = tasks.filter(task => {
    const matchesStatus = filterStatus === 'all' || task.status === filterStatus;
    const matchesCategory = filterCategory === 'all' || task.category === filterCategory;
    return matchesStatus && matchesCategory;
  });

  $: filteredWorkflows = workflows.filter(workflow => {
    return workflowStatusFilter === 'all' || workflow.status === workflowStatusFilter;
  });

  $: taskStats = {
    total: tasks.length,
    pending: tasks.filter(t => t.status === 'pending').length,
    in_progress: tasks.filter(t => t.status === 'in_progress').length,
    completed: tasks.filter(t => t.status === 'completed').length
  };

  $: workflowStats = {
    total: workflows.length,
    in_progress: workflows.filter(w => w.status === 'in_progress').length,
    completed: workflows.filter(w => w.status === 'completed').length
  };

  // API Calls - Tasks
  async function loadEmployees() {
    try {
      loading = true;
      error = '';
      
      const response = await fetch('/api/employees', {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });

      if (!response.ok) throw new Error('Failed to load employees');
      
      const data = await response.json();
      employees = (data || []).filter((emp: Employee) => 
        emp.status === 'active'
      );
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function loadTasks(employeeId: string) {
    try {
      tasksLoading = true;
      error = '';
      
      const response = await fetch(`/api/onboarding/${employeeId}`, {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });

      if (!response.ok) throw new Error('Failed to load tasks');
      
      tasks = await response.json();
    } catch (err: any) {
      error = err.message;
      tasks = [];
    } finally {
      tasksLoading = false;
    }
  }

  function selectEmployee(employee: Employee) {
    selectedEmployee = employee;
    loadTasks(employee.id);
  }

  async function createTask() {
    if (!selectedEmployee) return;

    try {
      loading = true;
      error = '';
      success = '';

      const response = await fetch('/api/onboarding', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          employee_id: selectedEmployee.id,
          ...newTask
        })
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to create task');
      }

      success = 'Task created successfully';
      showAddTaskModal = false;
      resetTaskForm();
      await loadTasks(selectedEmployee.id);
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function updateTaskStatus(taskId: string, status: string) {
    try {
      loading = true;
      error = '';
      success = '';

      const response = await fetch(`/api/onboarding/${taskId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ status })
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to update task');
      }

      success = 'Task updated successfully';
      if (selectedEmployee) {
        await loadTasks(selectedEmployee.id);
      }
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function deleteTask(taskId: string) {
    if (!confirm('Are you sure you want to delete this task?')) return;

    try {
      loading = true;
      error = '';
      success = '';

      const response = await fetch(`/api/onboarding/${taskId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to delete task');
      }

      success = 'Task deleted successfully';
      if (selectedEmployee) {
        await loadTasks(selectedEmployee.id);
      }
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function resetTaskForm() {
    newTask = {
      task_name: '',
      description: '',
      category: 'documentation',
      due_date: '',
      documents_required: false,
      assigned_to: ''
    };
  }

  // API Calls - Workflows
  async function loadWorkflows() {
    try {
      workflowsLoading = true;
      error = '';

      let url = '/api/workflows';
      if (workflowStatusFilter !== 'all') {
        url += `?status=${workflowStatusFilter}`;
      }

      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });

      if (!response.ok) throw new Error('Failed to load workflows');

      const data = await response.json();
      workflows = Array.isArray(data) ? data : [];
    } catch (err: any) {
      error = err.message;
      workflows = [];
    } finally {
      workflowsLoading = false;
    }
  }

  async function createWorkflow() {
    try {
      loading = true;
      error = '';
      success = '';

      const response = await fetch('/api/workflows', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(newWorkflow)
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to create workflow');
      }

      success = 'Workflow created successfully';
      showCreateWorkflowModal = false;
      resetWorkflowForm();
      await loadWorkflows();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function viewWorkflowDetails(workflow: Workflow) {
    try {
      loading = true;
      error = '';
      selectedWorkflow = workflow;

      const response = await fetch(`/api/workflows/${workflow.id}`, {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });

      if (!response.ok) throw new Error('Failed to load workflow details');

      const data = await response.json();
      workflowSteps = data.steps || [];
      showWorkflowDetailModal = true;
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function updateStepStatus(stepId: string, status: string) {
    try {
      loading = true;
      error = '';

      const response = await fetch(`/api/workflows/steps/${stepId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ status })
      });

      if (!response.ok) throw new Error('Failed to update step');

      success = 'Step updated successfully';
      if (selectedWorkflow) {
        await viewWorkflowDetails(selectedWorkflow);
        await loadWorkflows();
      }
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function resetWorkflowForm() {
    newWorkflow = {
      employee_id: '',
      template_name: 'software-engineer'
    };
  }

  // Utility Functions
  function formatDate(dateStr: string): string {
    if (!dateStr) return 'Not set';
    return new Date(dateStr).toLocaleDateString();
  }

  function getStatusBadgeClass(status: string): string {
    const classes: Record<string, string> = {
      pending: 'badge-warning',
      in_progress: 'badge-info',
      completed: 'badge-success',
      cancelled: 'badge-error',
      blocked: 'badge-error'
    };
    return classes[status] || 'badge-ghost';
  }

  function getCategoryIcon(category: string): string {
    const cat = taskCategories.find(c => c.value === category);
    return cat?.icon || '📋';
  }

  function getProgressColor(percentage: number): string {
    if (percentage >= 75) return '#10b981';
    if (percentage >= 50) return '#3b82f6';
    if (percentage >= 25) return '#f59e0b';
    return '#ef4444';
  }

  onMount(() => {
    loadEmployees();
    loadWorkflows();
  });
</script>

<div class="onboarding-container">
  <!-- Header -->
  <div class="onboarding-header">
    <h1>🎯 Employee Onboarding</h1>
    <p class="text-muted">Manage onboarding tasks and workflows for new hires</p>
  </div>

  <!-- Alerts -->
  {#if error}
    <div class="alert alert-error">
      <span>{error}</span>
      <button on:click={() => error = ''}>✕</button>
    </div>
  {/if}

  {#if success}
    <div class="alert alert-success">
      <span>{success}</span>
      <button on:click={() => success = ''}>✕</button>
    </div>
  {/if}

  <!-- Tabs -->
  <div class="tabs">
    <button 
      class="tab {activeTab === 'tasks' ? 'tab-active' : ''}"
      on:click={() => activeTab = 'tasks'}>
      ✓ Tasks
    </button>
    <button 
      class="tab {activeTab === 'workflows' ? 'tab-active' : ''}"
      on:click={() => activeTab = 'workflows'}>
      🔄 Workflows
    </button>
  </div>

  <!-- Content -->
  <div class="tab-content">
    <!-- Tasks Tab -->
    {#if activeTab === 'tasks'}
      <div class="tasks-section">
        <div class="two-column-layout">
          <!-- Left: Employee List -->
          <div class="employee-panel">
            <div class="panel-header">
              <h2>Select Employee</h2>
            </div>

            {#if loading}
              <div class="loading">Loading employees...</div>
            {:else if employees.length === 0}
              <div class="empty-state">
                <span class="empty-icon">👥</span>
                <p>No employees found</p>
              </div>
            {:else}
              <div class="employee-list">
                {#each employees as employee}
                  <button
                    class="employee-card {selectedEmployee?.id === employee.id ? 'active' : ''}"
                    on:click={() => selectEmployee(employee)}>
                    <div class="employee-avatar">{employee.first_name[0]}{employee.last_name[0]}</div>
                    <div class="employee-info">
                      <div class="employee-name">{employee.first_name} {employee.last_name}</div>
                      <div class="employee-meta">{employee.position || 'N/A'}</div>
                      <div class="employee-meta">{employee.department || 'N/A'}</div>
                    </div>
                  </button>
                {/each}
              </div>
            {/if}
          </div>

          <!-- Right: Tasks -->
          <div class="tasks-panel">
            {#if selectedEmployee}
              <div class="panel-header">
                <div>
                  <h2>Onboarding Tasks</h2>
                  <p class="text-muted">{selectedEmployee.first_name} {selectedEmployee.last_name}</p>
                </div>
                <button class="btn btn-primary" on:click={() => showAddTaskModal = true}>
                  + Add Task
                </button>
              </div>

              <!-- Stats -->
              <div class="stats-grid">
                <div class="stat-card">
                  <span class="stat-icon">📋</span>
                  <div class="stat-info">
                    <div class="stat-value">{taskStats.total}</div>
                    <div class="stat-label">Total Tasks</div>
                  </div>
                </div>
                <div class="stat-card">
                  <span class="stat-icon">⏳</span>
                  <div class="stat-info">
                    <div class="stat-value">{taskStats.pending + taskStats.in_progress}</div>
                    <div class="stat-label">In Progress</div>
                  </div>
                </div>
                <div class="stat-card">
                  <span class="stat-icon">✅</span>
                  <div class="stat-info">
                    <div class="stat-value">{taskStats.completed}</div>
                    <div class="stat-label">Completed</div>
                  </div>
                </div>
              </div>

              <!-- Filters -->
              <div class="filters">
                <select bind:value={filterStatus} class="select select-sm">
                  <option value="all">All Status</option>
                  <option value="pending">Pending</option>
                  <option value="in_progress">In Progress</option>
                  <option value="completed">Completed</option>
                </select>

                <select bind:value={filterCategory} class="select select-sm">
                  <option value="all">All Categories</option>
                  {#each taskCategories as cat}
                    <option value={cat.value}>{cat.icon} {cat.label}</option>
                  {/each}
                </select>
              </div>

              <!-- Tasks List -->
              {#if tasksLoading}
                <div class="loading">Loading tasks...</div>
              {:else if filteredTasks.length === 0}
                <div class="empty-state">
                  <span class="empty-icon">📝</span>
                  <p>No tasks found</p>
                  <button class="btn btn-primary" on:click={() => showAddTaskModal = true}>
                    Create First Task
                  </button>
                </div>
              {:else}
                <div class="tasks-list">
                  {#each filteredTasks as task}
                    <div class="task-card">
                      <div class="task-header">
                        <div class="task-title-row">
                          <span class="task-icon">{getCategoryIcon(task.category)}</span>
                          <h3 class="task-title">{task.task_name}</h3>
                        </div>
                        <div class="task-actions">
                          <select
                            value={task.status}
                            on:change={(e) => updateTaskStatus(task.id, e.currentTarget.value)}
                            class="select select-sm">
                            <option value="pending">Pending</option>
                            <option value="in_progress">In Progress</option>
                            <option value="completed">Completed</option>
                          </select>
                          <button class="btn btn-sm btn-ghost" on:click={() => deleteTask(task.id)}>
                            🗑️
                          </button>
                        </div>
                      </div>

                      {#if task.description}
                        <p class="task-description">{task.description}</p>
                      {/if}

                      <div class="task-meta">
                        <span class="badge {getStatusBadgeClass(task.status)}">
                          {task.status.replace('_', ' ')}
                        </span>
                        {#if task.category}
                          <span class="badge badge-ghost">{task.category}</span>
                        {/if}
                        {#if task.documents_required}
                          <span class="badge badge-info">📄 Docs Required</span>
                        {/if}
                      </div>

                      <div class="task-footer">
                        {#if task.due_date}
                          <span class="task-date">📅 Due: {formatDate(task.due_date)}</span>
                        {/if}
                        {#if task.completed_at}
                          <span class="task-date text-success">✓ Completed: {formatDate(task.completed_at)}</span>
                        {/if}
                      </div>
                    </div>
                  {/each}
                </div>
              {/if}
            {:else}
              <div class="empty-state">
                <span class="empty-icon">👈</span>
                <p>Select an employee to view their onboarding tasks</p>
              </div>
            {/if}
          </div>
        </div>
      </div>
    {/if}

    <!-- Workflows Tab -->
    {#if activeTab === 'workflows'}
      <div class="workflows-section">
        <div class="section-header">
          <div>
            <h2>Onboarding Workflows</h2>
            <p class="text-muted">Automated multi-step onboarding processes</p>
          </div>
          <button class="btn btn-primary" on:click={() => showCreateWorkflowModal = true}>
            + Create Workflow
          </button>
        </div>

        <!-- Workflow Stats -->
        <div class="stats-grid">
          <div class="stat-card">
            <span class="stat-icon">🔄</span>
            <div class="stat-info">
              <div class="stat-value">{workflowStats.total}</div>
              <div class="stat-label">Total Workflows</div>
            </div>
          </div>
          <div class="stat-card">
            <span class="stat-icon">⚡</span>
            <div class="stat-info">
              <div class="stat-value">{workflowStats.in_progress}</div>
              <div class="stat-label">In Progress</div>
            </div>
          </div>
          <div class="stat-card">
            <span class="stat-icon">✅</span>
            <div class="stat-info">
              <div class="stat-value">{workflowStats.completed}</div>
              <div class="stat-label">Completed</div>
            </div>
          </div>
        </div>

        <!-- Filter -->
        <div class="filters">
          <select bind:value={workflowStatusFilter} on:change={loadWorkflows} class="select select-sm">
            <option value="all">All Status</option>
            <option value="in_progress">In Progress</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>

        <!-- Workflows List -->
        {#if workflowsLoading}
          <div class="loading">Loading workflows...</div>
        {:else if filteredWorkflows.length === 0}
          <div class="empty-state">
            <span class="empty-icon">🔄</span>
            <p>No workflows found</p>
            <button class="btn btn-primary" on:click={() => showCreateWorkflowModal = true}>
              Create First Workflow
            </button>
          </div>
        {:else}
          <div class="workflows-grid">
            {#each filteredWorkflows as workflow}
              <div class="workflow-card" on:click={() => viewWorkflowDetails(workflow)}>
                <div class="workflow-header">
                  <span class="workflow-icon">🔄</span>
                  <span class="badge {getStatusBadgeClass(workflow.status)}">
                    {workflow.status.replace('_', ' ')}
                  </span>
                </div>

                <h3 class="workflow-title">{workflow.template_name.replace('-', ' ').toUpperCase()}</h3>
                <p class="workflow-employee">{workflow.employee_name || 'Employee'}</p>

                <div class="workflow-progress">
                  <div class="progress-bar">
                    <div 
                      class="progress-fill" 
                      style="width: {workflow.progress_percentage}%; background: {getProgressColor(workflow.progress_percentage)}">
                    </div>
                  </div>
                  <span class="progress-text">{workflow.progress_percentage}% Complete</span>
                </div>

                <div class="workflow-stage">
                  <span class="stage-label">Current Stage:</span>
                  <span class="stage-value">{workflow.current_stage}</span>
                </div>

                <div class="workflow-footer">
                  <span class="workflow-date">📅 Started: {formatDate(workflow.started_at)}</span>
                  {#if workflow.actual_completion}
                    <span class="workflow-date text-success">✓ {formatDate(workflow.actual_completion)}</span>
                  {:else}
                    <span class="workflow-date">⏳ Due: {formatDate(workflow.expected_completion)}</span>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>

<!-- Add Task Modal -->
{#if showAddTaskModal}
  <div class="modal" on:click={() => showAddTaskModal = false}>
    <div class="modal-box" on:click|stopPropagation>
      <h2>Add Onboarding Task</h2>

      <form on:submit|preventDefault={createTask} class="form">
        <div class="form-control">
          <label class="label">
            <span class="label-text">Task Name</span>
          </label>
          <input 
            type="text" 
            bind:value={newTask.task_name} 
            class="input w-full" 
            required 
            placeholder="e.g., Complete I-9 Form"
          />
        </div>

        <div class="form-control">
          <label class="label">
            <span class="label-text">Description</span>
          </label>
          <textarea 
            bind:value={newTask.description} 
            class="textarea w-full" 
            rows="3"
            placeholder="Task details..."
          ></textarea>
        </div>

        <div class="form-row">
          <div class="form-control">
            <label class="label">
              <span class="label-text">Category</span>
            </label>
            <select bind:value={newTask.category} class="select w-full">
              {#each taskCategories as cat}
                <option value={cat.value}>{cat.icon} {cat.label}</option>
              {/each}
            </select>
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">Due Date</span>
            </label>
            <input 
              type="date" 
              bind:value={newTask.due_date} 
              class="input w-full"
            />
          </div>
        </div>

        <div class="form-control">
          <label class="label cursor-pointer">
            <span class="label-text">Documents Required</span>
            <input 
              type="checkbox" 
              bind:checked={newTask.documents_required} 
              class="checkbox"
            />
          </label>
        </div>

        <div class="modal-actions">
          <button type="submit" class="btn btn-primary" disabled={loading}>
            {loading ? 'Creating...' : 'Create Task'}
          </button>
          <button type="button" class="btn btn-ghost" on:click={() => showAddTaskModal = false}>
            Cancel
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Create Workflow Modal -->
{#if showCreateWorkflowModal}
  <div class="modal" on:click={() => showCreateWorkflowModal = false}>
    <div class="modal-box" on:click|stopPropagation>
      <h2>Create Onboarding Workflow</h2>

      <form on:submit|preventDefault={createWorkflow} class="form">
        <div class="form-control">
          <label class="label">
            <span class="label-text">Select Employee</span>
          </label>
          <select bind:value={newWorkflow.employee_id} class="select w-full" required>
            <option value="">Choose employee...</option>
            {#each employees as employee}
              <option value={employee.id}>
                {employee.first_name} {employee.last_name} - {employee.position || 'N/A'}
              </option>
            {/each}
          </select>
        </div>

        <div class="form-control">
          <label class="label">
            <span class="label-text">Workflow Template</span>
          </label>
          <select bind:value={newWorkflow.template_name} class="select w-full" required>
            {#each workflowTemplates as template}
              <option value={template.value}>{template.label}</option>
            {/each}
          </select>
          <label class="label">
            <span class="text-muted">
              {workflowTemplates.find(t => t.value === newWorkflow.template_name)?.duration || ''}
            </span>
          </label>
        </div>

        <div class="template-preview">
          <h4>Template Details</h4>
          {#if newWorkflow.template_name === 'software-engineer'}
            <p>Comprehensive 17-step process covering documentation, equipment setup, training, team introductions, and development environment configuration.</p>
          {:else if newWorkflow.template_name === 'generic'}
            <p>Basic 4-step onboarding including documentation, equipment, workspace setup, and team introduction.</p>
          {:else if newWorkflow.template_name === 'sales-representative'}
            <p>Sales-focused 12-step process including CRM training, product knowledge, sales methodology, and territory assignment.</p>
          {:else if newWorkflow.template_name === 'manager'}
            <p>Leadership-focused 15-step process including team introductions, policy training, budget review, and strategic planning sessions.</p>
          {/if}
        </div>

        <div class="modal-actions">
          <button type="submit" class="btn btn-primary" disabled={loading}>
            {loading ? 'Creating...' : 'Create Workflow'}
          </button>
          <button type="button" class="btn btn-ghost" on:click={() => showCreateWorkflowModal = false}>
            Cancel
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Workflow Detail Modal -->
{#if showWorkflowDetailModal && selectedWorkflow}
  <div class="modal" on:click={() => showWorkflowDetailModal = false}>
    <div class="modal-box max-w-4xl" on:click|stopPropagation>
      <div class="modal-header">
        <div>
          <h2>{selectedWorkflow.template_name.replace('-', ' ').toUpperCase()}</h2>
          <p class="text-muted">{selectedWorkflow.employee_name}</p>
        </div>
        <button class="btn btn-circle btn-sm" on:click={() => showWorkflowDetailModal = false}>✕</button>
      </div>

      <!-- Progress Overview -->
      <div class="workflow-detail-progress">
        <div class="progress-bar-large">
          <div 
            class="progress-fill" 
            style="width: {selectedWorkflow.progress_percentage}%; background: {getProgressColor(selectedWorkflow.progress_percentage)}">
          </div>
        </div>
        <div class="progress-stats">
          <span>{selectedWorkflow.progress_percentage}% Complete</span>
          <span class="badge {getStatusBadgeClass(selectedWorkflow.status)}">
            {selectedWorkflow.status.replace('_', ' ')}
          </span>
        </div>
      </div>

      <!-- Steps List -->
      <div class="steps-list">
        <h3>Workflow Steps</h3>
        {#each workflowSteps as step, index}
          <div class="step-item">
            <div class="step-number">{index + 1}</div>
            <div class="step-content">
              <div class="step-header">
                <h4>{step.step_name}</h4>
                <select
                  value={step.status}
                  on:change={(e) => updateStepStatus(step.id, e.currentTarget.value)}
                  class="select select-sm">
                  <option value="pending">Pending</option>
                  <option value="in_progress">In Progress</option>
                  <option value="completed">Completed</option>
                  <option value="blocked">Blocked</option>
                </select>
              </div>
              {#if step.notes}
                <p class="step-notes">{step.notes}</p>
              {/if}
              <div class="step-meta">
                <span class="badge {getStatusBadgeClass(step.status)}">
                  {step.status.replace('_', ' ')}
                </span>
                {#if step.due_date}
                  <span class="text-muted">Due: {formatDate(step.due_date)}</span>
                {/if}
                {#if step.completed_at}
                  <span class="text-success">Completed: {formatDate(step.completed_at)}</span>
                {/if}
              </div>
            </div>
          </div>
        {/each}
      </div>

      <div class="modal-actions">
        <button class="btn btn-ghost" on:click={() => showWorkflowDetailModal = false}>
          Close
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .onboarding-container {
    padding: 2rem;
    max-width: 1600px;
    margin: 0 auto;
  }

  .onboarding-header {
    margin-bottom: 2rem;
  }

  .onboarding-header h1 {
    font-size: 2rem;
    font-weight: 700;
    color: #111827;
    margin-bottom: 0.5rem;
  }

  .text-muted {
    color: #6b7280;
    font-size: 0.875rem;
  }

  .text-success {
    color: #059669;
  }

  .alert {
    padding: 1rem;
    border-radius: 0.5rem;
    margin-bottom: 1.5rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .alert-error {
    background: #fef2f2;
    color: #991b1b;
    border: 1px solid #fecaca;
  }

  .alert-success {
    background: #f0fdf4;
    color: #166534;
    border: 1px solid #bbf7d0;
  }

  .alert button {
    background: none;
    border: none;
    font-size: 1.25rem;
    cursor: pointer;
  }

  .tabs {
    display: flex;
    gap: 0.5rem;
    border-bottom: 2px solid #e5e7eb;
    margin-bottom: 2rem;
  }

  .tab {
    padding: 0.75rem 1.5rem;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    color: #6b7280;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    margin-bottom: -2px;
  }

  .tab:hover {
    color: #111827;
  }

  .tab-active {
    color: #3b82f6;
    border-bottom-color: #3b82f6;
  }

  .two-column-layout {
    display: grid;
    grid-template-columns: 350px 1fr;
    gap: 2rem;
  }

  .employee-panel,
  .tasks-panel {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    padding: 1.5rem;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
  }

  .panel-header h2 {
    font-size: 1.25rem;
    font-weight: 700;
    color: #111827;
  }

  .loading {
    text-align: center;
    padding: 3rem;
    color: #6b7280;
  }

  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #6b7280;
  }

  .empty-icon {
    font-size: 4rem;
    display: block;
    margin-bottom: 1rem;
  }

  .employee-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .employee-card {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    cursor: pointer;
    transition: all 0.2s;
    text-align: left;
    width: 100%;
  }

  .employee-card:hover {
    background: #f9fafb;
    border-color: #3b82f6;
  }

  .employee-card.active {
    background: #eff6ff;
    border-color: #3b82f6;
  }

  .employee-avatar {
    width: 3rem;
    height: 3rem;
    border-radius: 50%;
    background: #3b82f6;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 1.125rem;
  }

  .employee-info {
    flex: 1;
  }

  .employee-name {
    font-weight: 600;
    color: #111827;
    margin-bottom: 0.25rem;
  }

  .employee-meta {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .stat-card {
    display: flex;
    align-items: center;
    gap: 1rem;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    padding: 1rem;
  }

  .stat-icon {
    font-size: 2rem;
  }

  .stat-info {
    display: flex;
    flex-direction: column;
  }

  .stat-value {
    font-size: 1.5rem;
    font-weight: 700;
    color: #111827;
  }

  .stat-label {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .filters {
    display: flex;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .select,
  .input {
    padding: 0.5rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 1rem;
  }

  .select-sm {
    padding: 0.375rem;
    font-size: 0.875rem;
  }

  .select:focus,
  .input:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .tasks-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .task-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    padding: 1.5rem;
    transition: all 0.2s;
  }

  .task-card:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  }

  .task-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1rem;
  }

  .task-title-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex: 1;
  }

  .task-icon {
    font-size: 1.5rem;
  }

  .task-title {
    font-size: 1rem;
    font-weight: 600;
    color: #111827;
  }

  .task-actions {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .task-description {
    color: #6b7280;
    font-size: 0.875rem;
    margin-bottom: 1rem;
  }

  .task-meta {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
    margin-bottom: 1rem;
  }

  .task-footer {
    display: flex;
    gap: 1rem;
    font-size: 0.875rem;
    color: #6b7280;
  }

  .task-date {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  .badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
  }

  .badge-success {
    background: #d1fae5;
    color: #065f46;
  }

  .badge-info {
    background: #dbeafe;
    color: #1e40af;
  }

  .badge-warning {
    background: #fef3c7;
    color: #92400e;
  }

  .badge-error {
    background: #fee2e2;
    color: #991b1b;
  }

  .badge-ghost {
    background: #f3f4f6;
    color: #4b5563;
  }

  .btn {
    padding: 0.5rem 1rem;
    border-radius: 0.375rem;
    font-weight: 500;
    cursor: pointer;
    border: 1px solid transparent;
    transition: all 0.2s;
    background: none;
  }

  .btn-primary {
    background: #3b82f6;
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background: #2563eb;
  }

  .btn-sm {
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
  }

  .btn-ghost {
    color: #6b7280;
    border-color: #d1d5db;
  }

  .btn-ghost:hover {
    background: #f9fafb;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-circle {
    width: 2rem;
    height: 2rem;
    border-radius: 50%;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
  }

  .section-header h2 {
    font-size: 1.5rem;
    font-weight: 700;
    color: #111827;
  }

  .workflows-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 1.5rem;
  }

  .workflow-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    padding: 1.5rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .workflow-card:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
    border-color: #3b82f6;
  }

  .workflow-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .workflow-icon {
    font-size: 2rem;
  }

  .workflow-title {
    font-size: 1.125rem;
    font-weight: 700;
    color: #111827;
    margin-bottom: 0.5rem;
    text-transform: capitalize;
  }

  .workflow-employee {
    color: #6b7280;
    font-size: 0.875rem;
    margin-bottom: 1rem;
  }

  .workflow-progress {
    margin-bottom: 1rem;
  }

  .progress-bar {
    height: 0.5rem;
    background: #e5e7eb;
    border-radius: 9999px;
    overflow: hidden;
    margin-bottom: 0.5rem;
  }

  .progress-fill {
    height: 100%;
    transition: width 0.3s;
  }

  .progress-text {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .workflow-stage {
    background: #f9fafb;
    padding: 0.75rem;
    border-radius: 0.375rem;
    margin-bottom: 1rem;
  }

  .stage-label {
    font-size: 0.75rem;
    color: #6b7280;
    text-transform: uppercase;
    font-weight: 600;
  }

  .stage-value {
    display: block;
    margin-top: 0.25rem;
    color: #111827;
    font-weight: 600;
  }

  .workflow-footer {
    display: flex;
    justify-content: space-between;
    font-size: 0.875rem;
    color: #6b7280;
  }

  .workflow-date {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  .modal {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
  }

  .modal-box {
    background: white;
    border-radius: 0.75rem;
    padding: 2rem;
    max-height: 90vh;
    overflow-y: auto;
    width: 100%;
    max-width: 32rem;
  }

  .max-w-4xl {
    max-width: 56rem;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
  }

  .modal-header h2 {
    font-size: 1.5rem;
    font-weight: 700;
    color: #111827;
  }

  .form {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .form-control {
    display: flex;
    flex-direction: column;
  }

  .label {
    margin-bottom: 0.5rem;
  }

  .label-text {
    font-weight: 500;
    color: #374151;
  }

  .label.cursor-pointer {
    cursor: pointer;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }

  .w-full {
    width: 100%;
  }

  .textarea {
    padding: 0.5rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 1rem;
    font-family: inherit;
    resize: vertical;
  }

  .textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .checkbox {
    width: 1.25rem;
    height: 1.25rem;
    border-radius: 0.25rem;
    border: 1px solid #d1d5db;
    cursor: pointer;
  }

  .form-row {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .template-preview {
    background: #f0f9ff;
    border: 1px solid #bae6fd;
    border-radius: 0.5rem;
    padding: 1rem;
  }

  .template-preview h4 {
    font-size: 0.875rem;
    font-weight: 600;
    color: #0c4a6e;
    margin-bottom: 0.5rem;
  }

  .template-preview p {
    font-size: 0.875rem;
    color: #075985;
  }

  .modal-actions {
    display: flex;
    gap: 0.75rem;
    justify-content: flex-end;
    margin-top: 1.5rem;
  }

  .workflow-detail-progress {
    background: #f9fafb;
    padding: 1.5rem;
    border-radius: 0.75rem;
    margin-bottom: 2rem;
  }

  .progress-bar-large {
    height: 1rem;
    background: #e5e7eb;
    border-radius: 9999px;
    overflow: hidden;
    margin-bottom: 1rem;
  }

  .progress-stats {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .steps-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .steps-list h3 {
    font-size: 1.125rem;
    font-weight: 700;
    color: #111827;
    margin-bottom: 1rem;
  }

  .step-item {
    display: flex;
    gap: 1rem;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    padding: 1rem;
  }

  .step-number {
    width: 2rem;
    height: 2rem;
    border-radius: 50%;
    background: #3b82f6;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    flex-shrink: 0;
  }

  .step-content {
    flex: 1;
  }

  .step-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .step-header h4 {
    font-size: 1rem;
    font-weight: 600;
    color: #111827;
  }

  .step-notes {
    color: #6b7280;
    font-size: 0.875rem;
    margin-bottom: 0.75rem;
  }

  .step-meta {
    display: flex;
    gap: 1rem;
    font-size: 0.875rem;
  }

  @media (max-width: 1024px) {
    .two-column-layout {
      grid-template-columns: 1fr;
    }

    .workflows-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 768px) {
    .onboarding-container {
      padding: 1rem;
    }

    .stats-grid {
      grid-template-columns: 1fr;
    }

    .form-row {
      grid-template-columns: 1fr;
    }
  }
</style>