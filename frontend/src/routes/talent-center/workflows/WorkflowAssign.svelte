<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  
  const dispatch = createEventDispatcher();
  
  interface Employee {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    department: string;
    position: string;
    hire_date: string;
    status: string;
  }
  
  interface WorkflowTemplate {
    id: string;
    name: string;
    description: string;
    type: string;
    estimated_days: number;
    steps: any[];
  }
  
  interface AssignedWorkflow {
    id: string;
    employee_id: string;
    employee_name: string;
    template_name: string;
    status: string;
    progress: number;
    start_date: string;
    due_date: string;
  }
  
  let employees: Employee[] = [];
  let templates: WorkflowTemplate[] = [];
  let recentAssignments: AssignedWorkflow[] = [];
  let loading = true;
  let showAssignModal = false;
  
  let selectedEmployee: Employee | null = null;
  let selectedTemplate: WorkflowTemplate | null = null;
  let startDate = new Date().toISOString().split('T')[0];
  let notes = '';
  let assignToManager = false;
  
  let searchQuery = '';
  let filterDepartment = '';
  let filterStatus = '';
  
  onMount(async () => {
    await Promise.all([
      loadEmployees(),
      loadTemplates(),
      loadRecentAssignments()
    ]);
  });
  
  async function loadEmployees() {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/employees', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        employees = await response.json();
      }
    } catch (err) {
      console.error('Failed to load employees:', err);
    }
  }
  
  async function loadTemplates() {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/workflows/templates?active=true', {
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
  
  async function loadRecentAssignments() {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/workflows/assignments/recent', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        recentAssignments = await response.json();
      } else {
        // API endpoint might not exist yet, just log and continue
        console.log('Recent assignments endpoint not available');
        recentAssignments = [];
      }
    } catch (err) {
      console.error('Failed to load recent assignments:', err);
      recentAssignments = [];
    }
  }
  
  function openAssignModal(employee: Employee) {
    selectedEmployee = employee;
    selectedTemplate = null;
    startDate = new Date().toISOString().split('T')[0];
    notes = '';
    assignToManager = false;
    showAssignModal = true;
  }
  
  function selectTemplate(template: WorkflowTemplate) {
    selectedTemplate = template;
  }
  
  async function assignWorkflow() {
    if (!selectedEmployee || !selectedTemplate) {
      alert('Please select an employee and template');
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      // Fixed: Use correct endpoint /api/workflows instead of /api/onboarding/workflows
      const response = await fetch('/api/workflows', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          employee_id: selectedEmployee.id,
          template_id: selectedTemplate.id,
          start_date: startDate,
          notes: notes,
          assign_to_manager: assignToManager
        })
      });
      
      if (response.ok) {
        showAssignModal = false;
        await loadRecentAssignments();
        dispatch('assigned');
        alert(`Workflow "${selectedTemplate.name}" assigned to ${selectedEmployee.first_name} ${selectedEmployee.last_name}`);
      } else {
        const error = await response.text();
        alert(`Failed to assign workflow: ${error}`);
      }
    } catch (err) {
      console.error('Failed to assign workflow:', err);
      alert('Failed to assign workflow');
    }
  }
  
  $: filteredEmployees = employees.filter(emp => {
    const matchesSearch = searchQuery === '' || 
      `${emp.first_name} ${emp.last_name}`.toLowerCase().includes(searchQuery.toLowerCase()) ||
      emp.email.toLowerCase().includes(searchQuery.toLowerCase());
    
    const matchesDepartment = !filterDepartment || emp.department === filterDepartment;
    const matchesStatus = !filterStatus || emp.status === filterStatus;
    
    return matchesSearch && matchesDepartment && matchesStatus;
  });
  
  $: departments = [...new Set(employees.map(e => e.department).filter(Boolean))];
  
  function getStatusColor(status: string) {
    const colors: Record<string, string> = {
      'active': 'green',
      'pending': 'yellow',
      'inactive': 'gray',
      'completed': 'blue'
    };
    return colors[status.toLowerCase()] || 'gray';
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
</script>

<div class="workflow-assign">
  <div class="assign-header">
    <div>
      <h2>Assign Workflows to Employees</h2>
      <p>Select an employee and assign a workflow template for onboarding or other processes</p>
    </div>
  </div>

  <!-- Recent Assignments -->
  {#if recentAssignments.length > 0}
    <div class="recent-section">
      <h3>Recent Assignments</h3>
      <div class="recent-list">
        {#each recentAssignments.slice(0, 5) as assignment}
          <div class="recent-item">
            <div class="recent-info">
              <div class="recent-employee">{assignment.employee_name}</div>
              <div class="recent-template">{assignment.template_name}</div>
              <div class="recent-meta">
                Started: {formatDate(assignment.start_date)} • Due: {formatDate(assignment.due_date)}
              </div>
            </div>
            <div class="recent-status">
              <div class="progress-circle">
                <svg viewBox="0 0 36 36">
                  <path
                    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                    fill="none"
                    stroke="#e2e8f0"
                    stroke-width="3"
                  />
                  <path
                    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                    fill="none"
                    stroke="#4f46e5"
                    stroke-width="3"
                    stroke-dasharray="{assignment.progress}, 100"
                  />
                </svg>
                <span class="progress-text">{Math.round(assignment.progress)}%</span>
              </div>
              <span class="status-badge" class:active={assignment.status === 'active'} class:completed={assignment.status === 'completed'}>
                {assignment.status}
              </span>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Filters -->
  <div class="filters">
    <input
      type="text"
      class="search-input"
      placeholder="🔍 Search employees..."
      bind:value={searchQuery}
    />
    
    <select class="filter-select" bind:value={filterDepartment}>
      <option value="">All Departments</option>
      {#each departments as dept}
        <option value={dept}>{dept}</option>
      {/each}
    </select>
    
    <select class="filter-select" bind:value={filterStatus}>
      <option value="">All Statuses</option>
      <option value="active">Active</option>
      <option value="pending">Pending</option>
      <option value="inactive">Inactive</option>
    </select>
  </div>

  <!-- Employee List -->
  {#if loading}
    <div class="loading">Loading employees...</div>
  {:else if filteredEmployees.length === 0}
    <div class="empty-state">
      <div class="empty-icon">👥</div>
      <h3>No Employees Found</h3>
      <p>Try adjusting your search or filters</p>
    </div>
  {:else}
    <div class="employees-table">
      <table>
        <thead>
          <tr>
            <th>Employee</th>
            <th>Department</th>
            <th>Position</th>
            <th>Hire Date</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each filteredEmployees as employee}
            <tr>
              <td>
                <div class="employee-cell">
                  <div class="employee-avatar">{employee.first_name[0]}{employee.last_name[0]}</div>
                  <div class="employee-info">
                    <div class="employee-name">{employee.first_name} {employee.last_name}</div>
                    <div class="employee-email">{employee.email}</div>
                  </div>
                </div>
              </td>
              <td>{employee.department || '-'}</td>
              <td>{employee.position || '-'}</td>
              <td>{formatDate(employee.hire_date)}</td>
              <td>
                <span class="status-pill {getStatusColor(employee.status)}">
                  {employee.status}
                </span>
              </td>
              <td>
                <button class="btn-assign" on:click={() => openAssignModal(employee)}>
                  Assign Workflow
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<!-- Assign Workflow Modal -->
{#if showAssignModal && selectedEmployee}
  <div class="modal-overlay" on:click={() => showAssignModal = false}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <h2>Assign Workflow</h2>
        <button class="close-btn" on:click={() => showAssignModal = false}>✕</button>
      </div>
      
      <div class="modal-body">
        <!-- Employee Info -->
        <div class="employee-summary">
          <div class="employee-avatar large">{selectedEmployee.first_name[0]}{selectedEmployee.last_name[0]}</div>
          <div>
            <h3>{selectedEmployee.first_name} {selectedEmployee.last_name}</h3>
            <p>{selectedEmployee.position} • {selectedEmployee.department}</p>
          </div>
        </div>

        <!-- Template Selection -->
        <div class="form-group">
          <label>Select Workflow Template</label>
          <div class="template-grid">
            {#each templates as template}
              <button 
                class="template-option"
                class:selected={selectedTemplate?.id === template.id}
                on:click={() => selectTemplate(template)}
              >
                <div class="template-option-header">
                  <span class="template-type">{template.type}</span>
                  <span class="template-duration">{template.estimated_days} days</span>
                </div>
                <h4>{template.name}</h4>
                <p>{template.description}</p>
                <div class="template-steps-count">
                  {template.steps?.length || 0} steps
                </div>
              </button>
            {/each}
          </div>
        </div>

        <!-- Workflow Preview -->
        {#if selectedTemplate}
          <div class="workflow-preview">
            <h4>Workflow Steps Preview</h4>
            <div class="steps-preview-list">
              {#each selectedTemplate.steps?.slice(0, 3) || [] as step, index}
                <div class="step-preview-item">
                  <span class="step-num">{index + 1}</span>
                  <span class="step-name">{step.name}</span>
                  <span class="step-duration">{step.estimated_days}d</span>
                </div>
              {/each}
              {#if (selectedTemplate.steps?.length || 0) > 3}
                <div class="step-preview-item more">
                  +{selectedTemplate.steps.length - 3} more steps
                </div>
              {/if}
            </div>
          </div>
        {/if}

        <!-- Additional Options -->
        <div class="form-group">
          <label for="startDate">Start Date</label>
          <input type="date" id="startDate" bind:value={startDate} />
        </div>

        <div class="form-group">
          <label for="notes">Notes (Optional)</label>
          <textarea id="notes" bind:value={notes} placeholder="Add any special instructions or notes..."></textarea>
        </div>

        <div class="form-group">
          <label class="checkbox-label">
            <input type="checkbox" bind:checked={assignToManager} />
            Also notify employee's manager
          </label>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-secondary" on:click={() => showAssignModal = false}>
          Cancel
        </button>
        <button 
          class="btn-primary" 
          on:click={assignWorkflow}
          disabled={!selectedTemplate}
        >
          Assign Workflow
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .workflow-assign {
    padding: 24px;
    max-width: 1400px;
    margin: 0 auto;
  }

  .assign-header {
    margin-bottom: 32px;
  }

  .assign-header h2 {
    font-size: 28px;
    font-weight: 700;
    color: #1a202c;
    margin: 0 0 8px 0;
  }

  .assign-header p {
    font-size: 16px;
    color: #4a5568;
    margin: 0;
  }

  /* Recent Assignments */
  .recent-section {
    background: white;
    border-radius: 12px;
    padding: 24px;
    margin-bottom: 24px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .recent-section h3 {
    font-size: 18px;
    font-weight: 600;
    color: #1a202c;
    margin: 0 0 16px 0;
  }

  .recent-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .recent-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    background: #f7fafc;
    border-radius: 8px;
    border: 1px solid #e2e8f0;
  }

  .recent-info {
    flex: 1;
  }

  .recent-employee {
    font-size: 15px;
    font-weight: 600;
    color: #1a202c;
    margin-bottom: 4px;
  }

  .recent-template {
    font-size: 14px;
    color: #4a5568;
    margin-bottom: 4px;
  }

  .recent-meta {
    font-size: 12px;
    color: #718096;
  }

  .recent-status {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .progress-circle {
    position: relative;
    width: 44px;
    height: 44px;
  }

  .progress-circle svg {
    transform: rotate(-90deg);
  }

  .progress-text {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 11px;
    font-weight: 600;
    color: #2d3748;
  }

  .status-badge {
    padding: 6px 12px;
    border-radius: 6px;
    font-size: 12px;
    font-weight: 600;
    text-transform: capitalize;
    background: #edf2f7;
    color: #4a5568;
  }

  .status-badge.active {
    background: #c6f6d5;
    color: #22543d;
  }

  .status-badge.completed {
    background: #bee3f8;
    color: #2c5282;
  }

  /* Filters */
  .filters {
    display: grid;
    grid-template-columns: 1fr auto auto;
    gap: 12px;
    margin-bottom: 24px;
  }

  .search-input,
  .filter-select {
    padding: 10px 14px;
    border: 1px solid #cbd5e0;
    border-radius: 8px;
    font-size: 14px;
    background: white;
    color: #2d3748;
  }

  .search-input {
    min-width: 300px;
  }

  .search-input::placeholder {
    color: #a0aec0;
  }

  .filter-select {
    min-width: 180px;
    cursor: pointer;
  }

  /* Table */
  .employees-table {
    background: white;
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  thead {
    background: #f7fafc;
  }

  th {
    padding: 14px 16px;
    text-align: left;
    font-size: 13px;
    font-weight: 600;
    color: #4a5568;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    border-bottom: 1px solid #e2e8f0;
  }

  tbody tr {
    border-bottom: 1px solid #e2e8f0;
    transition: background-color 0.15s;
  }

  tbody tr:hover {
    background: #f7fafc;
  }

  td {
    padding: 16px;
    font-size: 14px;
    color: #2d3748;
  }

  .employee-cell {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .employee-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 14px;
    flex-shrink: 0;
  }

  .employee-avatar.large {
    width: 56px;
    height: 56px;
    font-size: 20px;
  }

  .employee-name {
    font-weight: 600;
    color: #1a202c;
    margin-bottom: 2px;
  }

  .employee-email {
    font-size: 13px;
    color: #718096;
  }

  /* Status Pills */
  .status-pill {
    display: inline-block;
    padding: 4px 10px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
    text-transform: capitalize;
  }

  .status-pill.green {
    background: #c6f6d5;
    color: #22543d;
  }

  .status-pill.yellow {
    background: #fef3c7;
    color: #78350f;
  }

  .status-pill.gray {
    background: #e2e8f0;
    color: #4a5568;
  }

  .status-pill.blue {
    background: #bfdbfe;
    color: #1e40af;
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
    max-width: 700px;
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
    color: #1a202c;
    margin: 0;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 24px;
    color: #718096;
    cursor: pointer;
    transition: color 0.15s;
  }

  .close-btn:hover {
    color: #2d3748;
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

  .employee-summary {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px;
    background: #f7fafc;
    border-radius: 12px;
    margin-bottom: 24px;
  }

  .employee-summary h3 {
    margin: 0 0 4px 0;
    font-size: 18px;
    font-weight: 600;
    color: #1a202c;
  }

  .employee-summary p {
    margin: 0;
    color: #718096;
    font-size: 14px;
  }

  /* Template Selection */
  .template-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .template-option {
    text-align: left;
    padding: 16px;
    background: white;
    border: 2px solid #e2e8f0;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .template-option:hover {
    border-color: #cbd5e0;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  }

  .template-option.selected {
    border-color: #4f46e5;
    background: #eef2ff;
    box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
  }

  .template-option-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .template-type {
    padding: 4px 10px;
    background: #4f46e5;
    color: white;
    border-radius: 6px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .template-duration {
    font-size: 13px;
    color: #6b7280;
    font-weight: 600;
  }

  .template-option h4 {
    font-size: 16px;
    font-weight: 600;
    color: #1a202c;
    margin: 0 0 8px 0;
  }

  .template-option p {
    font-size: 13px;
    color: #6b7280;
    margin: 0 0 10px 0;
    line-height: 1.5;
  }

  .template-steps-count {
    font-size: 12px;
    color: #4a5568;
    font-weight: 600;
  }

  /* Workflow Preview */
  .workflow-preview {
    margin-top: 24px;
    padding: 20px;
    background: #f7fafc;
    border-radius: 8px;
  }

  .workflow-preview h4 {
    font-size: 14px;
    font-weight: 600;
    color: #1a202c;
    margin: 0 0 16px 0;
  }

  .steps-preview-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .step-preview-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: white;
    border-radius: 6px;
    font-size: 13px;
  }

  .step-preview-item.more {
    color: #718096;
    font-style: italic;
    justify-content: center;
  }

  .step-num {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    background: #4f46e5;
    color: white;
    border-radius: 50%;
    font-weight: 600;
    font-size: 12px;
    flex-shrink: 0;
  }

  .step-name {
    flex: 1;
    color: #2d3748;
  }

  .step-duration {
    color: #718096;
    font-weight: 600;
    font-size: 12px;
  }

  /* Form Elements */
  .form-group {
    margin-bottom: 20px;
  }

  .form-group label {
    display: block;
    font-size: 14px;
    font-weight: 600;
    color: #1a202c;
    margin-bottom: 8px;
  }

  .form-group input[type="date"],
  .form-group textarea {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid #cbd5e0;
    border-radius: 6px;
    font-size: 14px;
    color: #2d3748;
    transition: border-color 0.15s;
  }

  .form-group input[type="date"]:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: #4f46e5;
    box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
  }

  .form-group textarea {
    min-height: 80px;
    resize: vertical;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    cursor: pointer;
    color: #2d3748;
  }

  /* Buttons */
  .btn-assign {
    padding: 8px 16px;
    background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-assign:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
  }

  .btn-primary {
    padding: 11px 24px;
    background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 6px 20px rgba(79, 70, 229, 0.3);
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-secondary {
    padding: 11px 24px;
    background: white;
    color: #4a5568;
    border: 1px solid #cbd5e0;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s;
  }

  .btn-secondary:hover {
    background: #f7fafc;
    border-color: #a0aec0;
  }

  .empty-state,
  .loading {
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
    color: #1a202c;
    margin: 0 0 8px 0;
  }

  .empty-state p {
    font-size: 14px;
    color: #718096;
    margin: 0;
  }

  @media (max-width: 768px) {
    .filters {
      grid-template-columns: 1fr;
    }

    .search-input {
      min-width: auto;
    }

    .employees-table {
      overflow-x: auto;
    }

    table {
      min-width: 800px;
    }
  }
</style>