<script lang="ts">
  interface Employee {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    department: string;
    position: string;
    status: string;
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
    documents_required: boolean;
    created_at: string;
    updated_at: string;
  }

  let employees = $state<Employee[]>([]);
  let selectedEmployee = $state<Employee | null>(null);
  let tasks = $state<OnboardingTask[]>([]);
  let loading = $state(true);
  let tasksLoading = $state(false);
  let error = $state('');
  let showAddTaskModal = $state(false);

  let newTask = $state({
    task_name: '',
    description: '',
    category: 'documentation',
    due_date: '',
    documents_required: false
  });

  let filterStatus = $state('all');
  let filterCategory = $state('all');

  $effect(() => {
    loadEmployees();
  });

  async function loadEmployees() {
    loading = true;
    error = '';
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${import.meta.env.VITE_API_URL}/api/employees`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (!response.ok) throw new Error('Failed to load employees');
      const data = await response.json();
      // Filter out system administrator and only show active employees
      employees = (data || []).filter((emp: Employee) => 
        emp.status === 'active' && emp.email !== 'admin@hub-hrms.local'
      );
    } catch (err: any) {
      error = err.message;
      employees = [];
    } finally {
      loading = false;
    }
  }

  async function loadTasks(employeeId: string) {
    tasksLoading = true;
    error = '';
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${import.meta.env.VITE_API_URL}/api/onboarding/${employeeId}`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (!response.ok) throw new Error('Failed to load tasks');
      const data = await response.json();
      tasks = data || [];
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
      const token = localStorage.getItem('token');
      const taskData: any = {
        task_name: newTask.task_name,
        description: newTask.description || undefined,
        category: newTask.category || undefined,
        status: 'pending',
        documents_required: newTask.documents_required
      };

      if (newTask.due_date) {
        taskData.due_date = newTask.due_date;
      }

      const response = await fetch(`${import.meta.env.VITE_API_URL}/api/onboarding/${selectedEmployee.id}/tasks`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(taskData)
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to create task');
      }

      await loadTasks(selectedEmployee.id);
      closeAddTaskModal();
    } catch (err: any) {
      error = err.message;
    }
  }

  async function updateTaskStatus(taskId: string, status: string) {
    if (!selectedEmployee) return;

    try {
      const token = localStorage.getItem('token');
      const updateData: any = { status };
      if (status === 'completed') {
        updateData.completed_at = new Date().toISOString();
      }

      const response = await fetch(`${import.meta.env.VITE_API_URL}/api/onboarding/${selectedEmployee.id}/tasks/${taskId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(updateData)
      });

      if (!response.ok) throw new Error('Failed to update task');
      await loadTasks(selectedEmployee.id);
    } catch (err: any) {
      error = err.message;
    }
  }

  function openAddTaskModal() {
    newTask = {
      task_name: '',
      description: '',
      category: 'documentation',
      due_date: '',
      documents_required: false
    };
    showAddTaskModal = true;
  }

  function closeAddTaskModal() {
    showAddTaskModal = false;
  }

  let filteredTasks = $derived(() => {
    let result = tasks;
    if (filterStatus !== 'all') result = result.filter(t => t.status === filterStatus);
    if (filterCategory !== 'all') result = result.filter(t => t.category === filterCategory);
    return result;
  });

  let completionRate = $derived(() => {
    if (tasks.length === 0) return 0;
    const completed = tasks.filter(t => t.status === 'completed').length;
    return Math.round((completed / tasks.length) * 100);
  });

  function getStatusColor(status: string): string {
    const colors: Record<string, string> = {
      completed: '#10b981',
      'in-progress': '#f59e0b',
      pending: '#6366f1'
    };
    return colors[status] || '#6b7280';
  }

  function formatDate(dateString?: string): string {
    if (!dateString) return '-';
    return new Date(dateString).toLocaleDateString();
  }
</script>

<div class="onboarding-container">
  <div class="header">
    <h1>Onboarding Management</h1>
    <p>Track and manage employee onboarding tasks</p>
  </div>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  <div class="content-grid">
    <div class="employee-sidebar">
      <h2>Active Employees</h2>
      
      {#if loading}
        <div class="loading-spinner">Loading...</div>
      {:else if employees.length === 0}
        <div class="empty-state"><p>No active employees</p></div>
      {:else}
        <div class="employee-list">
          {#each employees as employee}
            <button
              class="employee-card"
              class:selected={selectedEmployee?.id === employee.id}
              onclick={() => selectEmployee(employee)}
            >
              <div class="employee-avatar">
                {employee.first_name.charAt(0)}{employee.last_name.charAt(0)}
              </div>
              <div class="employee-info">
                <div class="employee-name">{employee.first_name} {employee.last_name}</div>
                <div class="employee-details">{employee.position}</div>
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </div>

    <div class="tasks-panel">
      {#if !selectedEmployee}
        <div class="empty-state">
          <h3>Select an Employee</h3>
          <p>Choose an employee to view their onboarding tasks</p>
        </div>
      {:else}
        <div class="tasks-header">
          <div class="employee-banner">
            <h2>{selectedEmployee.first_name} {selectedEmployee.last_name}</h2>
            <p>{selectedEmployee.position} - {selectedEmployee.department}</p>
          </div>

          {#if tasks.length > 0}
            <div class="progress-card">
              <div class="progress-label">
                <span>Completion Rate</span>
                <span class="progress-value">{completionRate()}%</span>
              </div>
              <div class="progress-bar">
                <div class="progress-fill" style="width: {completionRate()}%"></div>
              </div>
              <div class="task-stats">
                {tasks.filter(t => t.status === 'completed').length} of {tasks.length} tasks completed
              </div>
            </div>
          {/if}

          <div class="actions-bar">
            <button class="btn-primary" onclick={openAddTaskModal}>+ Add Task</button>
            <div class="filters">
              <select bind:value={filterStatus} class="filter-select">
                <option value="all">All Status</option>
                <option value="pending">Pending</option>
                <option value="in-progress">In Progress</option>
                <option value="completed">Completed</option>
              </select>
              <select bind:value={filterCategory} class="filter-select">
                <option value="all">All Categories</option>
                <option value="documentation">Documentation</option>
                <option value="training">Training</option>
                <option value="equipment">Equipment</option>
                <option value="it-setup">IT Setup</option>
                <option value="hr-paperwork">HR Paperwork</option>
              </select>
            </div>
          </div>
        </div>

        {#if tasksLoading}
          <div class="loading-spinner">Loading tasks...</div>
        {:else if filteredTasks().length === 0}
          <div class="empty-state">
            <h3>No Tasks</h3>
            <p>{tasks.length === 0 ? 'Create the first task!' : 'No tasks match filters'}</p>
          </div>
        {:else}
          <div class="tasks-grid">
            {#each filteredTasks() as task}
              <div class="task-card">
                <div class="task-header">
                  <div class="task-title-row">
                    <h3>{task.task_name}</h3>
                    {#if task.category}
                      <span class="task-category">{task.category}</span>
                    {/if}
                  </div>
                  <select
                    class="status-select"
                    value={task.status}
                    onchange={(e) => updateTaskStatus(task.id, e.currentTarget.value)}
                    style="background-color: {getStatusColor(task.status)}20; color: {getStatusColor(task.status)}"
                  >
                    <option value="pending">Pending</option>
                    <option value="in-progress">In Progress</option>
                    <option value="completed">Completed</option>
                  </select>
                </div>
                {#if task.description}
                  <p class="task-description">{task.description}</p>
                {/if}
                <div class="task-meta">
                  {#if task.due_date}
                    <div class="meta-item">
                      <span class="meta-label">Due:</span>
                      <span class="meta-value">{formatDate(task.due_date)}</span>
                    </div>
                  {/if}
                  {#if task.completed_at}
                    <div class="meta-item">
                      <span class="meta-label">Completed:</span>
                      <span class="meta-value">{formatDate(task.completed_at)}</span>
                    </div>
                  {/if}
                  {#if task.documents_required}
                    <span class="badge">📄 Docs Required</span>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      {/if}
    </div>
  </div>
</div>

{#if showAddTaskModal && selectedEmployee}
  <div 
    class="modal-overlay"
    onclick={(e) => e.target === e.currentTarget && closeAddTaskModal()}
    onkeydown={(e) => e.key === 'Escape' && closeAddTaskModal()}
    role="button"
    tabindex="0"
    aria-label="Close modal"
  >
    <div class="modal" role="dialog" aria-modal="true" tabindex="-1">
      <div class="modal-header">
        <h2>Add Onboarding Task</h2>
        <button class="close-btn" onclick={closeAddTaskModal}>×</button>
      </div>
      <form onsubmit={(e) => { e.preventDefault(); createTask(); }} class="modal-body">
        <div class="form-group">
          <label>Task Name *<input type="text" bind:value={newTask.task_name} required /></label>
        </div>
        <div class="form-group">
          <label>Description<textarea bind:value={newTask.description} rows="3"></textarea></label>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>Category
              <select bind:value={newTask.category}>
                <option value="documentation">Documentation</option>
                <option value="training">Training</option>
                <option value="equipment">Equipment</option>
                <option value="it-setup">IT Setup</option>
                <option value="hr-paperwork">HR Paperwork</option>
              </select>
            </label>
          </div>
          <div class="form-group">
            <label>Due Date<input type="date" bind:value={newTask.due_date} /></label>
          </div>
        </div>
        <div class="form-group">
          <label class="checkbox-label">
            <input type="checkbox" bind:checked={newTask.documents_required} />
            Documents Required
          </label>
        </div>
        <div class="modal-actions">
          <button type="button" class="btn-secondary" onclick={closeAddTaskModal}>Cancel</button>
          <button type="submit" class="btn-primary">Create Task</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<style>
  .onboarding-container { padding: 2rem; max-width: 1400px; margin: 0 auto; }
  .header h1 { font-size: 2rem; font-weight: 700; color: #e4e7eb; margin-bottom: 0.5rem; }
  .header p { color: #9ca3af; }
  .error-banner { background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); color: #fca5a5; padding: 1rem; border-radius: 8px; margin-bottom: 1.5rem; }
  .content-grid { display: grid; grid-template-columns: 320px 1fr; gap: 2rem; min-height: calc(100vh - 250px); }
  
  .employee-sidebar { background: rgba(15, 23, 42, 0.6); backdrop-filter: blur(10px); border: 1px solid rgba(99, 102, 241, 0.2); border-radius: 12px; padding: 1.5rem; height: fit-content; max-height: calc(100vh - 250px); overflow-y: auto; }
  .employee-sidebar h2 { font-size: 1.125rem; font-weight: 600; color: #e4e7eb; margin-bottom: 1rem; }
  .employee-list { display: flex; flex-direction: column; gap: 0.75rem; }
  .employee-card { display: flex; align-items: center; gap: 0.75rem; padding: 0.875rem; background: rgba(30, 41, 59, 0.4); border: 1px solid rgba(99, 102, 241, 0.1); border-radius: 8px; cursor: pointer; transition: all 0.2s; text-align: left; width: 100%; }
  .employee-card:hover { background: rgba(99, 102, 241, 0.1); border-color: rgba(99, 102, 241, 0.3); transform: translateX(4px); }
  .employee-card.selected { background: rgba(99, 102, 241, 0.2); border-color: rgba(99, 102, 241, 0.5); }
  .employee-avatar { width: 40px; height: 40px; border-radius: 50%; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); display: flex; align-items: center; justify-content: center; font-size: 0.875rem; font-weight: 600; color: white; }
  .employee-info { flex: 1; min-width: 0; }
  .employee-name { font-weight: 600; color: #e4e7eb; font-size: 0.875rem; }
  .employee-details { font-size: 0.75rem; color: #9ca3af; }
  
  .tasks-panel { background: rgba(15, 23, 42, 0.6); backdrop-filter: blur(10px); border: 1px solid rgba(99, 102, 241, 0.2); border-radius: 12px; padding: 2rem; }
  .tasks-header { margin-bottom: 2rem; }
  .employee-banner h2 { font-size: 1.5rem; font-weight: 700; color: #e4e7eb; margin-bottom: 0.25rem; }
  .employee-banner p { color: #9ca3af; font-size: 0.875rem; }
  
  .progress-card { background: rgba(99, 102, 241, 0.1); border: 1px solid rgba(99, 102, 241, 0.2); border-radius: 8px; padding: 1.25rem; margin: 1.5rem 0; }
  .progress-label { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.75rem; font-size: 0.875rem; color: #9ca3af; }
  .progress-value { font-size: 1.5rem; font-weight: 700; color: #6366f1; }
  .progress-bar { height: 8px; background: rgba(30, 41, 59, 0.6); border-radius: 4px; overflow: hidden; margin-bottom: 0.5rem; }
  .progress-fill { height: 100%; background: linear-gradient(90deg, #6366f1 0%, #8b5cf6 100%); border-radius: 4px; transition: width 0.3s; }
  .task-stats { font-size: 0.75rem; color: #9ca3af; }
  
  .actions-bar { display: flex; justify-content: space-between; align-items: center; gap: 1rem; flex-wrap: wrap; }
  .filters { display: flex; gap: 0.75rem; }
  .filter-select { padding: 0.5rem 0.875rem; background: rgba(30, 41, 59, 0.6); border: 1px solid rgba(99, 102, 241, 0.2); border-radius: 6px; color: #e4e7eb; font-size: 0.875rem; cursor: pointer; }
  
  .tasks-grid { display: grid; gap: 1rem; margin-top: 1.5rem; }
  .task-card { background: rgba(30, 41, 59, 0.4); border: 1px solid rgba(99, 102, 241, 0.2); border-radius: 8px; padding: 1.25rem; transition: all 0.2s; }
  .task-card:hover { border-color: rgba(99, 102, 241, 0.4); transform: translateY(-2px); }
  .task-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 0.75rem; gap: 1rem; }
  .task-title-row { display: flex; align-items: center; gap: 0.75rem; flex: 1; }
  .task-card h3 { font-size: 1rem; font-weight: 600; color: #e4e7eb; margin: 0; }
  .task-category { font-size: 0.75rem; padding: 0.25rem 0.625rem; background: rgba(99, 102, 241, 0.2); color: #a5b4fc; border-radius: 4px; text-transform: capitalize; }
  .status-select { padding: 0.375rem 0.75rem; border-radius: 6px; border: 1px solid rgba(99, 102, 241, 0.3); font-size: 0.75rem; font-weight: 600; cursor: pointer; }
  .task-description { color: #9ca3af; font-size: 0.875rem; line-height: 1.5; margin-bottom: 0.75rem; }
  .task-meta { display: flex; flex-wrap: wrap; gap: 1rem; font-size: 0.75rem; }
  .meta-item { display: flex; gap: 0.25rem; }
  .meta-label { color: #6b7280; }
  .meta-value { color: #9ca3af; font-weight: 500; }
  .badge { padding: 0.25rem 0.625rem; border-radius: 4px; font-weight: 500; background: rgba(245, 158, 11, 0.2); color: #fbbf24; }
  
  .btn-primary { padding: 0.625rem 1.25rem; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; border: none; border-radius: 8px; font-weight: 600; font-size: 0.875rem; cursor: pointer; transition: all 0.2s; }
  .btn-primary:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4); }
  .btn-secondary { padding: 0.625rem 1.25rem; background: rgba(30, 41, 59, 0.6); color: #e4e7eb; border: 1px solid rgba(99, 102, 241, 0.2); border-radius: 8px; font-weight: 600; font-size: 0.875rem; cursor: pointer; }
  
  .loading-spinner, .empty-state { text-align: center; padding: 3rem; color: #9ca3af; }
  .empty-state h3 { font-size: 1.25rem; color: #e4e7eb; margin-bottom: 0.5rem; }
  
  .modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0, 0, 0, 0.75); display: flex; align-items: center; justify-content: center; z-index: 1000; padding: 1rem; }
  .modal { background: linear-gradient(135deg, rgba(15, 23, 42, 0.95) 0%, rgba(30, 41, 59, 0.95) 100%); border: 1px solid rgba(99, 102, 241, 0.3); border-radius: 16px; width: 100%; max-width: 600px; max-height: 90vh; overflow-y: auto; box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5); }
  .modal-header { display: flex; justify-content: space-between; align-items: center; padding: 1.5rem 2rem; border-bottom: 1px solid rgba(99, 102, 241, 0.2); }
  .modal-header h2 { font-size: 1.5rem; font-weight: 700; color: #e4e7eb; margin: 0; }
  .close-btn { width: 32px; height: 32px; border-radius: 6px; border: none; background: rgba(99, 102, 241, 0.1); color: #e4e7eb; font-size: 1.5rem; cursor: pointer; }
  .modal-body { padding: 2rem; }
  .form-group { display: flex; flex-direction: column; gap: 0.5rem; margin-bottom: 1.25rem; }
  .form-group label { display: flex; flex-direction: column; gap: 0.5rem; font-size: 0.875rem; font-weight: 600; color: #e4e7eb; }
  .form-group input, .form-group select, .form-group textarea { padding: 0.875rem 1rem; background: rgba(15, 23, 42, 0.6); border: 1px solid rgba(99, 102, 241, 0.2); border-radius: 8px; color: #e4e7eb; font-size: 0.875rem; }
  .form-row { display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem; }
  .checkbox-label { flex-direction: row !important; align-items: center; gap: 0.5rem !important; cursor: pointer; }
  .checkbox-label input[type="checkbox"] { width: 18px; height: 18px; cursor: pointer; }
  .modal-actions { display: flex; justify-content: flex-end; gap: 0.75rem; margin-top: 2rem; }
  
  @media (max-width: 1024px) {
    .content-grid { grid-template-columns: 1fr; }
    .employee-sidebar { max-height: 300px; }
  }
</style>
