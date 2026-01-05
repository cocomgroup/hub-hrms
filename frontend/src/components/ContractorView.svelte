<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();

  interface Task {
    id: string;
    title: string;
    description: string;
    type: string;
    priority: string;
    due_date: string;
    status: string;
    created_at: string;
  }

  interface Project {
    id: string;
    name: string;
    description?: string;
  }

  interface ProjectRow {
    project_id: string;
    project_name: string;
    hours: { [date: string]: number }; // date -> hours mapping
    descriptions: { [date: string]: string }; // date -> description mapping
    entry_ids: { [date: string]: string }; // date -> entry_id mapping
    total: number;
  }

  interface Props {
    employee?: any;
  }
  
  let { employee }: Props = $props();

  let loading = $state(false);
  let saving = $state(false);
  let activeTab = $state('timesheet');
  let tasks = $state<Task[]>([]);
  let projects = $state<Project[]>([]);
  let projectRows = $state<ProjectRow[]>([]);
  let pendingTasksCount = $state(0);
  let currentWeekStart = $state(getWeekStart(new Date()));
  let successMessage = $state('');
  let errorMessage = $state('');
  let editingDescription = $state<string | null>(null);
  let timesheetStatus = $state<string>('draft'); // Track current week's timesheet status
  let timesheetId = $state<string | null>(null); // Track current week's timesheet ID
  
  // Get employee data from prop or auth store
  let currentEmployee = $derived(employee || $authStore.employee);
  
  // Check if timesheet can be edited
  const canEditTimesheet = $derived(
    timesheetStatus === 'draft' || timesheetStatus === 'rejected'
  );
  
  // Calculate week dates and total
  const weekDates = $derived(getWeekDates(currentWeekStart));
  const weekTotal = $derived(
    projectRows.reduce((sum, row) => sum + row.total, 0)
  );
  
  // Move console logs to $effect for reactive tracking
  $effect(() => {
    console.log('[ContractorView] State:', {
      employeeProp: employee,
      authStoreEmployee: $authStore.employee,
      currentEmployee
    });
  });

  onMount(() => {
    console.log('[ContractorView] onMount called');
    loadProjects();
    loadContractorData();
    loadWeekEntries();
  });

  function getWeekStart(date: Date): Date {
    const d = new Date(date);
    const day = d.getDay();
    const diff = d.getDate() - day + (day === 0 ? -6 : 1); // Monday
    d.setDate(diff);
    d.setHours(0, 0, 0, 0);
    return d;
  }

  function getWeekDates(weekStart: Date): Date[] {
    const dates = [];
    for (let i = 0; i < 7; i++) {
      const date = new Date(weekStart);
      date.setDate(weekStart.getDate() + i);
      dates.push(date);
    }
    return dates;
  }

  function formatDate(date: Date): string {
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  function formatDateISO(date: Date): string {
    return date.toISOString().split('T')[0];
  }

  function getDayName(date: Date): string {
    return date.toLocaleDateString('en-US', { weekday: 'short' });
  }

  function previousWeek() {
    const newStart = new Date(currentWeekStart);
    newStart.setDate(newStart.getDate() - 7);
    currentWeekStart = newStart;
    loadWeekEntries();
  }

  function nextWeek() {
    const newStart = new Date(currentWeekStart);
    newStart.setDate(newStart.getDate() + 7);
    currentWeekStart = newStart;
    loadWeekEntries();
  }

  function currentWeek() {
    currentWeekStart = getWeekStart(new Date());
    loadWeekEntries();
  }

  async function loadProjects() {
    try {
      // Call /timesheet/projects to get only projects the contractor is assigned to
      const response = await fetch(`${API_BASE_URL}/timesheet/projects`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        projects = await response.json();
      } else {
        console.error('Failed to load projects:', response.status);
        projects = [];
      }
    } catch (err) {
      console.error('Error loading projects:', err);
      projects = [];
    }
  }

  async function loadContractorData() {
    loading = true;
    try {
      await loadTasks();
    } catch (err) {
      console.error('Error loading contractor data:', err);
    } finally {
      loading = false;
    }
  }

  async function loadTasks() {
    try {
      const response = await fetch(`${API_BASE_URL}/onboarding/tasks`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      if (response.ok) {
        const data = await response.json();
        tasks = Array.isArray(data) ? data : [];
        pendingTasksCount = tasks.filter(t => t.status === 'pending').length;
      }
    } catch (err) {
      console.error('Error loading tasks:', err);
      tasks = [];
    }
  }

  async function loadWeekEntries() {
    try {
      const startDate = formatDateISO(weekDates[0]);
      const endDate = formatDateISO(weekDates[6]);
      
      // Load time entries
      const entriesResponse = await fetch(
        `${API_BASE_URL}/timesheet/entries?start_date=${startDate}&end_date=${endDate}`,
        { headers: { 'Authorization': `Bearer ${$authStore.token}` } }
      );
      
      if (entriesResponse.ok) {
        const entries = await entriesResponse.json();
        projectRows = convertToProjectRows(entries);
      } else {
        projectRows = [];
      }
      
      // Load timesheet status for the week
      const summaryResponse = await fetch(
        `${API_BASE_URL}/timesheet/summary?start_date=${startDate}&end_date=${endDate}`,
        { headers: { 'Authorization': `Bearer ${$authStore.token}` } }
      );
      
      if (summaryResponse.ok) {
        const summary = await summaryResponse.json();
        timesheetStatus = summary.status || 'draft';
        timesheetId = summary.timesheet_id || null;
      } else {
        timesheetStatus = 'draft';
        timesheetId = null;
      }
      
    } catch (err) {
      console.error('Error loading week entries:', err);
      projectRows = [];
      timesheetStatus = 'draft';
      timesheetId = null;
    }
  }

  function convertToProjectRows(entries: any[]): ProjectRow[] {
    const projectMap = new Map<string, ProjectRow>();
    
    if (Array.isArray(entries)) {
      entries.forEach(entry => {
        const projectId = entry.project_id || 'default';
        const projectName = entry.project_name || 'General Hours';
        const date = entry.date;
        
        if (!projectMap.has(projectId)) {
          projectMap.set(projectId, {
            project_id: projectId,
            project_name: projectName,
            hours: {},
            descriptions: {},
            entry_ids: {},
            total: 0
          });
        }
        
        const row = projectMap.get(projectId)!;
        row.hours[date] = entry.hours || entry.total_hours || 0;
        row.descriptions[date] = entry.description || entry.notes || '';
        row.entry_ids[date] = entry.id;
        row.total = Object.values(row.hours).reduce((sum, h) => sum + h, 0);
      });
    }
    
    return Array.from(projectMap.values());
  }

  function addProjectRow() {
    if (projects.length === 0) {
      errorMessage = 'No projects assigned to you. Please ask your manager to assign you to a project before logging hours.';
      return;
    }
    
    // Find a project not already in the list, or use the first one
    const usedProjectIds = new Set(projectRows.map(r => r.project_id));
    const availableProject = projects.find(p => !usedProjectIds.has(p.id)) || projects[0];
    
    projectRows = [...projectRows, {
      project_id: availableProject.id,
      project_name: availableProject.name,
      hours: {},
      descriptions: {},
      entry_ids: {},
      total: 0
    }];
  }

  function deleteRow(index: number) {
    projectRows = projectRows.filter((_, i) => i !== index);
  }

  function updateHours(rowIndex: number, date: string, value: string) {
    const hours = parseFloat(value) || 0;
    projectRows[rowIndex].hours[date] = hours;
    projectRows[rowIndex].total = Object.values(projectRows[rowIndex].hours).reduce((sum, h) => sum + h, 0);
    projectRows = [...projectRows]; // Trigger reactivity
  }

  function updateProject(rowIndex: number, projectId: string) {
    const project = projects.find(p => p.id === projectId);
    if (project) {
      projectRows[rowIndex].project_id = projectId;
      projectRows[rowIndex].project_name = project.name;
      projectRows = [...projectRows];
    }
  }

  function updateDescription(rowIndex: number, date: string, description: string) {
    projectRows[rowIndex].descriptions[date] = description;
  }

  function openDescriptionModal(rowIndex: number, date: string) {
    editingDescription = `${rowIndex}-${date}`;
  }

  function closeDescriptionModal() {
    editingDescription = null;
  }

  async function saveDraft() {
    await saveEntries('draft');
  }

  async function submitTimesheet() {
    if (weekTotal === 0) {
      errorMessage = 'Please enter at least some hours before submitting.';
      return;
    }
    await saveEntries('submitted');
  }

  async function saveEntries(status: 'draft' | 'submitted') {
    saving = true;
    errorMessage = '';
    successMessage = '';
    
    try {
      // Step 1: Save/update all time entries
      const entriesToSave: any[] = [];
      
      projectRows.forEach(row => {
        weekDates.forEach(date => {
          const dateStr = formatDateISO(date);
          const hours = row.hours[dateStr] || 0;
          
          if (hours > 0) {
            entriesToSave.push({
              date: dateStr,
              project_id: row.project_id,
              hours,
              type: 'regular',
              notes: row.descriptions[dateStr] || ''
            });
          }
        });
      });
      
      // Save entries using bulk endpoint
      if (entriesToSave.length > 0) {
        const bulkResponse = await fetch(`${API_BASE_URL}/timesheet/entries/bulk`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${$authStore.token}`,
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ entries: entriesToSave })
        });
        
        if (!bulkResponse.ok) {
          const error = await bulkResponse.json();
          throw new Error(error.error || 'Failed to save time entries');
        }
      }
      
      // Step 2: If submitting, submit the timesheet for the week
      if (status === 'submitted') {
        const startDate = formatDateISO(weekDates[0]);
        const endDate = formatDateISO(weekDates[6]);
        
        const submitResponse = await fetch(`${API_BASE_URL}/timesheet/submit`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${$authStore.token}`,
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            start_date: startDate,
            end_date: endDate
          })
        });
        
        if (!submitResponse.ok) {
          const error = await submitResponse.json();
          throw new Error(error.error || 'Failed to submit timesheet');
        }
        
        successMessage = 'Timesheet submitted for manager approval!';
      } else {
        successMessage = 'Timesheet saved as draft.';
      }
      
      // Reload the entries to show updated data
      await loadWeekEntries();
      
    } catch (err: any) {
      console.error('Error saving timesheet:', err);
      errorMessage = err.message || 'An error occurred while saving.';
    } finally {
      saving = false;
    }
  }

  function getTaskPriorityClass(priority: string): string {
    const classes = {
      'high': 'priority-high',
      'medium': 'priority-medium',
      'low': 'priority-low'
    };
    return classes[priority] || 'priority-medium';
  }

  function getTaskStatusClass(status: string): string {
    const classes = {
      'completed': 'status-completed',
      'in-progress': 'status-progress',
      'pending': 'status-pending'
    };
    return classes[status] || 'status-pending';
  }

  function formatTaskDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString();
  }
</script>

<div class="contractor-view">
  <div class="header">
    <div class="header-content">
      <h1>👷 Contractor Portal</h1>
      <p class="subtitle">Welcome back, {currentEmployee?.first_name || 'Contractor'}!</p>
    </div>
    <div class="header-badge">
      <span class="badge badge-contractor">Contractor</span>
    </div>
  </div>

  {#if loading}
    <div class="loading">Loading your information...</div>
  {:else}
    <!-- Stats Cards -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">⏱️</div>
        <div class="stat-content">
          <div class="stat-value">{weekTotal.toFixed(1)}</div>
          <div class="stat-label">Hours This Week</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">📊</div>
        <div class="stat-content">
          <div class="stat-value">{projectRows.length}</div>
          <div class="stat-label">Active Projects</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">📋</div>
        <div class="stat-content">
          <div class="stat-value">{pendingTasksCount}</div>
          <div class="stat-label">Pending Tasks</div>
        </div>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button
        type="button"
        class="tab-btn {activeTab === 'timesheet' ? 'active' : ''}"
        onclick={() => activeTab = 'timesheet'}
      >
        ⏱️ Timesheet
      </button>
      <button
        type="button"
        class="tab-btn {activeTab === 'tasks' ? 'active' : ''}"
        onclick={() => activeTab = 'tasks'}
      >
        📋 Assigned Tasks
      </button>
    </div>

    <!-- Content -->
    <div class="content">
      {#if activeTab === 'timesheet'}
        <div class="timesheet-section">
          <div class="timesheet-header">
            <div class="header-left">
              <h2>Weekly Timesheet</h2>
              {#if timesheetStatus !== 'draft'}
                <span class="status-badge status-{timesheetStatus}">
                  {timesheetStatus === 'submitted' ? '⏳ Pending Approval' : 
                   timesheetStatus === 'approved' ? '✓ Approved' : 
                   timesheetStatus === 'rejected' ? '✗ Rejected' : timesheetStatus}
                </span>
              {/if}
            </div>
            <div class="week-nav">
              <button type="button" class="btn btn-sm btn-secondary" onclick={previousWeek}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="15 18 9 12 15 6"></polyline>
                </svg>
                Previous
              </button>
              <button type="button" class="btn btn-sm btn-secondary" onclick={currentWeek}>
                This Week
              </button>
              <button type="button" class="btn btn-sm btn-secondary" onclick={nextWeek}>
                Next
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="9 18 15 12 9 6"></polyline>
                </svg>
              </button>
            </div>
          </div>

          <div class="week-display">
            Week of {formatDate(weekDates[0])} - {formatDate(weekDates[6])}
          </div>

          {#if successMessage}
            <div class="alert alert-success">{successMessage}</div>
          {/if}

          {#if errorMessage}
            <div class="alert alert-error">{errorMessage}</div>
          {/if}

          <!-- Timesheet Table -->
          <div class="table-container">
            <table class="timesheet-table">
              <thead>
                <tr>
                  <th class="project-col">Project</th>
                  {#each weekDates as date}
                    <th class="date-col">
                      <div class="date-header">
                        <span class="day-name">{getDayName(date)}</span>
                        <span class="day-date">{formatDate(date)}</span>
                      </div>
                    </th>
                  {/each}
                  <th class="total-col">Total</th>
                  <th class="actions-col"></th>
                </tr>
              </thead>
              <tbody>
                {#each projectRows as row, rowIndex}
                  <tr class="project-row">
                    <td class="project-cell">
                      <select 
                        value={row.project_id}
                        onchange={(e) => updateProject(rowIndex, e.currentTarget.value)}
                        class="project-select"
                        disabled={!canEditTimesheet}
                      >
                        {#each projects as project}
                          <option value={project.id}>{project.name}</option>
                        {/each}
                      </select>
                    </td>
                    {#each weekDates as date}
                      {@const dateStr = formatDateISO(date)}
                      {@const hours = row.hours[dateStr] || 0}
                      {@const hasDescription = row.descriptions[dateStr]}
                      <td class="hours-cell">
                        <input 
                          type="number" 
                          step="0.25" 
                          min="0" 
                          max="24" 
                          value={hours || ''}
                          oninput={(e) => updateHours(rowIndex, dateStr, e.currentTarget.value)}
                          class="hours-input"
                          placeholder="0"
                          disabled={!canEditTimesheet}
                        />
                        <button 
                          type="button"
                          class="note-btn {hasDescription ? 'has-note' : ''}"
                          onclick={() => openDescriptionModal(rowIndex, dateStr)}
                          title={hasDescription ? 'Edit note' : 'Add note'}
                          disabled={!canEditTimesheet}
                        >
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                          </svg>
                        </button>
                      </td>
                    {/each}
                    <td class="total-cell">{row.total.toFixed(2)}</td>
                    <td class="actions-cell">
                      <button 
                        type="button" 
                        class="btn-icon btn-delete" 
                        onclick={() => deleteRow(rowIndex)}
                        title="Delete row"
                        disabled={!canEditTimesheet}
                      >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <polyline points="3 6 5 6 21 6"></polyline>
                          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                        </svg>
                      </button>
                    </td>
                  </tr>
                {/each}
                
                {#if projectRows.length === 0}
                  <tr>
                    <td colspan="10" class="empty-row">
                      No projects added yet. Click "Add Project" to get started.
                    </td>
                  </tr>
                {/if}

                <!-- Total Row -->
                <tr class="total-row">
                  <td class="total-label">Week Total</td>
                  {#each weekDates as date}
                    {@const dateStr = formatDateISO(date)}
                    {@const dayTotal = projectRows.reduce((sum, row) => sum + (row.hours[dateStr] || 0), 0)}
                    <td class="day-total">{dayTotal > 0 ? dayTotal.toFixed(2) : ''}</td>
                  {/each}
                  <td class="grand-total">{weekTotal.toFixed(2)}</td>
                  <td></td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="timesheet-actions">
            <button 
              type="button" 
              class="btn btn-secondary" 
              onclick={addProjectRow}
              disabled={!canEditTimesheet}
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
              </svg>
              Add Project
            </button>
            
            <div class="action-group">
              {#if timesheetStatus === 'rejected'}
                <div class="rejection-notice">
                  <span>⚠️ Timesheet was rejected. Please make changes and resubmit.</span>
                </div>
              {/if}
              
              {#if canEditTimesheet}
                <button 
                  type="button" 
                  class="btn btn-secondary" 
                  onclick={saveDraft} 
                  disabled={saving}
                >
                  {saving ? 'Saving...' : 'Save Draft'}
                </button>
                <button 
                  type="button" 
                  class="btn btn-primary" 
                  onclick={submitTimesheet} 
                  disabled={saving || weekTotal === 0}
                >
                  {saving ? 'Submitting...' : timesheetStatus === 'rejected' ? 'Resubmit Timesheet' : 'Submit Timesheet'}
                </button>
              {:else if timesheetStatus === 'submitted'}
                <div class="status-info">
                  ⏳ Awaiting manager approval
                </div>
              {:else if timesheetStatus === 'approved'}
                <div class="status-info success">
                  ✓ Approved by manager
                </div>
              {/if}
            </div>
          </div>
        </div>
      {/if}

      {#if activeTab === 'tasks'}
        <div class="section">
          <h2>Assigned Tasks</h2>
          
          {#if tasks.length === 0}
            <div class="empty-state">
              <p>No tasks assigned yet.</p>
              <p class="text-muted">Your assigned tasks will appear here.</p>
            </div>
          {:else}
            <div class="tasks-grid">
              {#each tasks as task}
                <div class="task-card">
                  <div class="task-header">
                    <h3>{task.title}</h3>
                    <span class="task-priority {getTaskPriorityClass(task.priority)}">
                      {task.priority}
                    </span>
                  </div>
                  
                  <p class="task-description">{task.description || 'No description'}</p>
                  
                  <div class="task-footer">
                    <span class="task-type">{task.type}</span>
                    <span class="task-status {getTaskStatusClass(task.status)}">
                      {task.status}
                    </span>
                  </div>
                  
                  {#if task.due_date}
                    <div class="task-due-date">
                      Due: {formatTaskDate(task.due_date)}
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Description Modal -->
{#if editingDescription}
  {@const [rowIndex, dateStr] = editingDescription.split('-')}
  {@const row = projectRows[parseInt(rowIndex)]}
  <div 
    class="modal-overlay" 
    onclick={closeDescriptionModal}
    onkeydown={(e) => e.key === 'Escape' && closeDescriptionModal()}
    role="button"
    tabindex="0"
    aria-label="Close modal"
  >
    <div 
      class="modal" 
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="dialog"
      aria-labelledby="modal-title"
      aria-modal="true"
      tabindex="-1"
    >
      <div class="modal-header">
        <h3 id="modal-title">Add Note</h3>
        <button 
          type="button" 
          class="modal-close" 
          onclick={closeDescriptionModal}
          aria-label="Close modal"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>
      <div class="modal-body">
        <p class="modal-info">{row.project_name} - {formatDate(new Date(dateStr))}</p>
        <textarea 
          value={row.descriptions[dateStr] || ''}
          oninput={(e) => updateDescription(parseInt(rowIndex), dateStr, e.currentTarget.value)}
          placeholder="What did you work on?"
          rows="4"
          class="description-textarea"
        ></textarea>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-primary" onclick={closeDescriptionModal}>
          Done
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .contractor-view {
    max-width: 1600px;
    margin: 0 auto;
    padding: 24px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 32px;
    padding: 24px;
    background: white;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .header-content h1 {
    font-size: 32px;
    font-weight: 700;
    color: #111827;
    margin: 0 0 8px 0;
  }

  .subtitle {
    color: #6b7280;
    font-size: 16px;
    margin: 0;
  }

  .header-badge {
    display: flex;
    gap: 8px;
  }

  .badge {
    padding: 6px 12px;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
  }

  .badge-contractor {
    background: #fef3c7;
    color: #92400e;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 20px;
    margin-bottom: 32px;
  }

  .stat-card {
    background: white;
    padding: 24px;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .stat-icon {
    font-size: 36px;
    width: 60px;
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f3f4f6;
    border-radius: 12px;
  }

  .stat-content {
    flex: 1;
  }

  .stat-value {
    font-size: 28px;
    font-weight: 600;
    color: #111827;
    margin-bottom: 4px;
  }

  .stat-label {
    font-size: 14px;
    color: #6b7280;
  }

  .tabs {
    display: flex;
    gap: 8px;
    margin-bottom: 24px;
    background: white;
    padding: 8px;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .tab-btn {
    flex: 1;
    padding: 12px 24px;
    background: transparent;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    font-size: 16px;
    font-weight: 500;
    color: #6b7280;
    transition: all 0.2s;
  }

  .tab-btn:hover {
    background: #f3f4f6;
  }

  .tab-btn.active {
    background: #3b82f6;
    color: white;
  }

  .content {
    background: white;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    overflow: hidden;
  }

  /* Timesheet Styles */
  .timesheet-section {
    padding: 24px;
  }

  .timesheet-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }
  
  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .timesheet-header h2 {
    font-size: 20px;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }
  
  .status-badge {
    padding: 6px 12px;
    border-radius: 12px;
    font-size: 13px;
    font-weight: 500;
    white-space: nowrap;
  }
  
  .status-badge.status-submitted {
    background: #fef3c7;
    color: #92400e;
  }
  
  .status-badge.status-approved {
    background: #d1fae5;
    color: #065f46;
  }
  
  .status-badge.status-rejected {
    background: #fee2e2;
    color: #991b1b;
  }

  .week-nav {
    display: flex;
    gap: 8px;
  }

  .week-display {
    text-align: center;
    font-size: 16px;
    font-weight: 500;
    color: #4b5563;
    margin-bottom: 24px;
    padding: 12px;
    background: #f9fafb;
    border-radius: 8px;
  }

  .alert {
    padding: 12px 16px;
    border-radius: 8px;
    margin-bottom: 16px;
  }

  .alert-success {
    background: #d1fae5;
    color: #065f46;
    border: 1px solid #a7f3d0;
  }

  .alert-error {
    background: #fee2e2;
    color: #991b1b;
    border: 1px solid #fecaca;
  }

  /* Table Styles */
  .table-container {
    overflow-x: auto;
    margin-bottom: 24px;
    border-radius: 8px;
    border: 1px solid #e5e7eb;
  }

  .timesheet-table {
    width: 100%;
    border-collapse: collapse;
    min-width: 1000px;
  }

  thead {
    background: #f9fafb;
    border-bottom: 2px solid #e5e7eb;
  }

  th {
    padding: 16px 12px;
    text-align: center;
    font-weight: 600;
    font-size: 14px;
    color: #374151;
    white-space: nowrap;
  }

  .project-col {
    text-align: left;
    min-width: 200px;
  }

  .date-col {
    min-width: 100px;
  }

  .date-header {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .day-name {
    font-weight: 600;
    color: #111827;
  }

  .day-date {
    font-size: 12px;
    font-weight: 400;
    color: #6b7280;
  }

  .total-col {
    min-width: 80px;
    background: #eff6ff;
  }

  .actions-col {
    width: 60px;
  }

  td {
    padding: 8px;
    border-bottom: 1px solid #e5e7eb;
  }

  .project-row:hover {
    background: #f9fafb;
  }

  .project-cell {
    padding: 12px;
  }

  .project-select {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    background: white;
    color: #111827;
  }

  .project-select:focus {
    outline: none;
    border-color: #3b82f6;
    ring: 2px;
    ring-color: #dbeafe;
  }

  .hours-cell {
    text-align: center;
    position: relative;
  }

  .hours-input {
    width: 60px;
    padding: 8px;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    text-align: center;
    font-size: 14px;
    font-weight: 500;
  }

  .hours-input:focus {
    outline: none;
    border-color: #3b82f6;
    ring: 2px;
    ring-color: #dbeafe;
  }

  .hours-input::placeholder {
    color: #d1d5db;
  }

  /* Remove spinner */
  .hours-input::-webkit-inner-spin-button,
  .hours-input::-webkit-outer-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  .note-btn {
    position: absolute;
    top: 4px;
    right: 4px;
    padding: 4px;
    background: transparent;
    border: none;
    cursor: pointer;
    border-radius: 4px;
    color: #9ca3af;
    transition: all 0.2s;
  }

  .note-btn:hover {
    background: #f3f4f6;
    color: #374151;
  }

  .note-btn.has-note {
    color: #3b82f6;
  }

  .note-btn svg {
    width: 14px;
    height: 14px;
  }

  .total-cell {
    font-weight: 600;
    color: #3b82f6;
    text-align: center;
    background: #eff6ff;
  }

  .actions-cell {
    text-align: center;
  }

  .btn-icon {
    padding: 8px;
    border: none;
    background: transparent;
    cursor: pointer;
    border-radius: 6px;
    transition: all 0.2s;
  }

  .btn-icon svg {
    width: 16px;
    height: 16px;
  }

  .btn-delete {
    color: #ef4444;
  }

  .btn-delete:hover {
    background: #fee2e2;
  }

  .empty-row {
    text-align: center;
    padding: 48px !important;
    color: #9ca3af;
  }

  .total-row {
    background: #f3f4f6;
    font-weight: 600;
    border-top: 2px solid #e5e7eb;
  }

  .total-label {
    text-align: right;
    padding-right: 16px !important;
    color: #374151;
  }

  .day-total {
    text-align: center;
    color: #6b7280;
  }

  .grand-total {
    text-align: center;
    font-size: 16px;
    color: #1f2937;
    background: #dbeafe;
  }

  .timesheet-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .action-group {
    display: flex;
    gap: 12px;
    align-items: center;
  }
  
  .status-info {
    padding: 8px 16px;
    background: #fef3c7;
    color: #92400e;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
  }
  
  .status-info.success {
    background: #d1fae5;
    color: #065f46;
  }
  
  .rejection-notice {
    padding: 12px 16px;
    background: #fee2e2;
    color: #991b1b;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    border: 1px solid #fecaca;
  }

  /* Modal Styles */
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
    border-radius: 12px;
    width: 90%;
    max-width: 500px;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid #e5e7eb;
  }

  .modal-header h3 {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  .modal-close {
    padding: 8px;
    background: transparent;
    border: none;
    cursor: pointer;
    border-radius: 6px;
    color: #6b7280;
    transition: all 0.2s;
  }

  .modal-close:hover {
    background: #f3f4f6;
    color: #111827;
  }

  .modal-close svg {
    width: 20px;
    height: 20px;
  }

  .modal-body {
    padding: 24px;
  }

  .modal-info {
    font-size: 14px;
    color: #6b7280;
    margin: 0 0 16px 0;
  }

  .description-textarea {
    width: 100%;
    padding: 12px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    font-family: inherit;
    resize: vertical;
  }

  .description-textarea:focus {
    outline: none;
    border-color: #3b82f6;
    ring: 2px;
    ring-color: #dbeafe;
  }

  .modal-footer {
    padding: 16px 24px;
    border-top: 1px solid #e5e7eb;
    display: flex;
    justify-content: flex-end;
  }

  /* Task Styles */
  .section {
    padding: 24px;
  }

  .section h2 {
    font-size: 20px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 20px 0;
  }

  .empty-state {
    text-align: center;
    padding: 48px;
    color: #6b7280;
  }

  .empty-state p {
    margin: 8px 0;
  }

  .text-muted {
    font-size: 14px;
    color: #9ca3af;
  }

  .tasks-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 16px;
  }

  .task-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
    transition: all 0.2s;
  }

  .task-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  }

  .task-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 12px;
  }

  .task-header h3 {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0;
    flex: 1;
  }

  .task-priority {
    padding: 4px 8px;
    border-radius: 6px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .priority-high {
    background: #fee2e2;
    color: #991b1b;
  }

  .priority-medium {
    background: #fef3c7;
    color: #92400e;
  }

  .priority-low {
    background: #e0e7ff;
    color: #3730a3;
  }

  .task-description {
    font-size: 14px;
    color: #6b7280;
    margin: 0 0 16px 0;
    line-height: 1.5;
  }

  .task-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .task-type {
    font-size: 12px;
    color: #6b7280;
    text-transform: capitalize;
  }

  .task-status {
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 500;
    text-transform: capitalize;
  }

  .status-completed {
    background: #d1fae5;
    color: #065f46;
  }

  .status-progress {
    background: #dbeafe;
    color: #1e40af;
  }

  .status-pending {
    background: #fef3c7;
    color: #92400e;
  }

  .task-due-date {
    font-size: 12px;
    color: #6b7280;
    padding-top: 8px;
    border-top: 1px solid #e5e7eb;
  }

  .loading {
    text-align: center;
    padding: 48px;
    color: #6b7280;
  }

  /* Buttons */
  .btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 12px 24px;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn svg {
    width: 18px;
    height: 18px;
  }

  .btn-sm {
    padding: 8px 16px;
    font-size: 13px;
  }

  .btn-primary {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
  }

  .btn-secondary {
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }

  .btn-secondary:hover:not(:disabled) {
    background: #f9fafb;
    border-color: #9ca3af;
  }

  @media (max-width: 1200px) {
    .table-container {
      font-size: 12px;
    }

    .hours-input {
      width: 50px;
      font-size: 12px;
    }
  }

  @media (max-width: 768px) {
    .contractor-view {
      padding: 16px;
    }

    .header {
      flex-direction: column;
      gap: 16px;
    }

    .header-content h1 {
      font-size: 24px;
    }

    .stats-grid {
      grid-template-columns: 1fr;
    }

    .tabs {
      flex-direction: column;
    }

    .timesheet-header {
      flex-direction: column;
      gap: 16px;
      align-items: stretch;
    }

    .week-nav {
      justify-content: center;
    }

    .timesheet-actions {
      flex-direction: column;
    }

    .action-group {
      width: 100%;
      flex-direction: column;
    }

    .tasks-grid {
      grid-template-columns: 1fr;
    }
  }
</style>