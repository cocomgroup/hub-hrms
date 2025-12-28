<script lang="ts">
  import { onMount } from 'svelte';
  import { getApiBaseUrl } from '../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();

  interface Project {
    id: string;
    name: string;
    description: string;
    status: string;
    priority: string;
    manager_id?: string;
    manager_name?: string;
    manager_email?: string;
    start_date?: string;
    end_date?: string;
    budget?: number;
    member_count: number;
    members?: ProjectMember[];
    created_at: string;
  }

  interface ProjectMember {
    id: string;
    project_id: string;
    employee_id: string;
    employee_name: string;
    employee_email: string;
    employee_position: string;
    role: string;
  }

  interface Employee {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    position: string;
    department: string;
    manager_id?: string;
  }

  // Props
  let { navigate }: { navigate: (page: string) => void } = $props();

  // State
  let projects = $state<Project[]>([]);
  let employees = $state<Employee[]>([]);
  let managers = $state<Employee[]>([]);
  let loading = $state(false);
  let activeTab = $state<'projects' | 'assignments'>('projects');
  let showCreateModal = $state(false);
  let showAssignModal = $state(false);
  let showManagerAssignModal = $state(false);
  let selectedProject = $state<Project | null>(null);
  let filterStatus = $state('all');
  let searchTerm = $state('');
  let error = $state('');
  let success = $state('');

  // Form states
  let projectForm = $state({
    name: '',
    description: '',
    status: 'active',
    priority: 'medium',
    manager_id: '',
    start_date: '',
    end_date: '',
    budget: ''
  });

  let memberForm = $state({
    employee_id: '',
    role: 'member'
  });

  let managerAssignForm = $state({
    employee_id: '',
    manager_id: ''
  });

  onMount(() => {
    loadProjects();
    loadEmployees();
  });

  async function loadProjects() {
    loading = true;
    try {
      const token = localStorage.getItem('token');
      const params = new URLSearchParams();
      if (filterStatus !== 'all') params.append('status', filterStatus);
      
      const url = `${API_BASE_URL}/projects?${params.toString()}`;
      const response = await fetch(url, {
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (response.ok) {
        const data = await response.json();
        projects = Array.isArray(data) ? data : [];
      } else {
        projects = [];
      }
    } catch (err) {
      console.error('Error loading projects:', err);
      projects = [];
    } finally {
      loading = false;
    }
  }

  async function loadEmployees() {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${API_BASE_URL}/employees`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (response.ok) {
        const data = await response.json();
        employees = Array.isArray(data) ? data : [];
        // Filter managers (employees with manager in title or who manage others)
        managers = employees.filter((e: Employee) => 
          e.position?.toLowerCase().includes('manager') || 
          e.position?.toLowerCase().includes('director') ||
          e.position?.toLowerCase().includes('lead')
        );
      }
    } catch (err) {
      console.error('Error loading employees:', err);
      employees = [];
    }
  }

  async function createProject() {
    error = '';
    success = '';

    if (!projectForm.name.trim()) {
      error = 'Project name is required';
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const body: any = {
        name: projectForm.name,
        description: projectForm.description,
        status: projectForm.status,
        priority: projectForm.priority
      };

      // Add optional fields with proper formatting
      if (projectForm.manager_id) {
        body.manager_id = projectForm.manager_id;
      }
      
      if (projectForm.start_date) {
        // Convert date string to RFC3339 format
        body.start_date = new Date(projectForm.start_date).toISOString();
      }
      
      if (projectForm.end_date) {
        // Convert date string to RFC3339 format
        body.end_date = new Date(projectForm.end_date).toISOString();
      }
      
      if (projectForm.budget) {
        body.budget = parseFloat(projectForm.budget);
      }

      console.log('Creating project with data:', body);

      const response = await fetch(`${API_BASE_URL}/projects`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(body)
      });

      if (response.ok) {
        success = 'Project created successfully';
        showCreateModal = false;
        resetProjectForm();
        loadProjects();
      } else {
        const data = await response.json();
        console.error('Backend error:', data);
        error = data.error || 'Failed to create project';
      }
    } catch (err) {
      console.error('Error creating project:', err);
      error = 'Error creating project';
    }
  }

  async function assignMemberToProject(projectId: string) {
    error = '';
    success = '';

    if (!memberForm.employee_id) {
      error = 'Please select an employee';
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${API_BASE_URL}/projects/${projectId}/members`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(memberForm)
      });

      if (response.ok) {
        success = 'Member assigned successfully';
        showAssignModal = false;
        memberForm = { employee_id: '', role: 'member' };
        loadProjectDetails(projectId);
      } else {
        const data = await response.json();
        error = data.error || 'Failed to assign member';
      }
    } catch (err) {
      console.error('Error assigning member:', err);
      error = 'Error assigning member';
    }
  }

  async function removeMemberFromProject(projectId: string, employeeId: string) {
    if (!confirm('Remove this member from the project?')) return;

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${API_BASE_URL}/projects/${projectId}/members/${employeeId}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (response.ok) {
        success = 'Member removed successfully';
        loadProjectDetails(projectId);
      }
    } catch (err) {
      console.error('Error removing member:', err);
      error = 'Error removing member';
    }
  }

  async function assignEmployeeToManager() {
    error = '';
    success = '';

    if (!managerAssignForm.employee_id || !managerAssignForm.manager_id) {
      error = 'Please select both employee and manager';
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${API_BASE_URL}/managers/assign`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(managerAssignForm)
      });

      if (response.ok) {
        success = 'Employee assigned to manager successfully';
        showManagerAssignModal = false;
        managerAssignForm = { employee_id: '', manager_id: '' };
        loadEmployees();
      } else {
        const data = await response.json();
        error = data.error || 'Failed to assign employee to manager';
      }
    } catch (err) {
      console.error('Error assigning to manager:', err);
      error = 'Error assigning to manager';
    }
  }

  async function loadProjectDetails(projectId: string) {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${API_BASE_URL}/projects/${projectId}`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (response.ok) {
        const project = await response.json();
        const index = projects.findIndex(p => p.id === projectId);
        if (index !== -1) {
          projects[index] = project;
        }
        if (selectedProject?.id === projectId) {
          selectedProject = project;
        }
      }
    } catch (err) {
      console.error('Error loading project details:', err);
    }
  }

  async function deleteProject(projectId: string) {
    if (!confirm('Are you sure you want to delete this project?')) return;

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${API_BASE_URL}/projects/${projectId}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (response.ok) {
        success = 'Project deleted successfully';
        loadProjects();
        if (selectedProject?.id === projectId) {
          selectedProject = null;
        }
      }
    } catch (err) {
      console.error('Error deleting project:', err);
      error = 'Error deleting project';
    }
  }

  function resetProjectForm() {
    projectForm = {
      name: '',
      description: '',
      status: 'active',
      priority: 'medium',
      manager_id: '',
      start_date: '',
      end_date: '',
      budget: ''
    };
  }

  function openAssignModal(project: Project) {
    selectedProject = project;
    showAssignModal = true;
    error = '';
    success = '';
  }

  function getStatusColor(status: string): string {
    const colors: Record<string, string> = {
      'active': 'status-active',
      'on-hold': 'status-onhold',
      'completed': 'status-completed',
      'archived': 'status-archived'
    };
    return colors[status] || '';
  }

  function getPriorityColor(priority: string): string {
    const colors: Record<string, string> = {
      'low': 'priority-low',
      'medium': 'priority-medium',
      'high': 'priority-high',
      'critical': 'priority-critical'
    };
    return colors[priority] || '';
  }

  let filteredProjects = $derived(
    projects.filter(p => {
      const matchesSearch = p.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                           p.description?.toLowerCase().includes(searchTerm.toLowerCase());
      return matchesSearch;
    })
  );

  let unassignedEmployees = $derived(
    employees.filter(e => !e.manager_id)
  );
</script>

<div class="project-management">
  <!-- Header -->
  <div class="header">
    <div>
      <h1>Project Management</h1>
      <p class="subtitle">Create projects, assign managers, and manage teams</p>
    </div>
    <button class="btn-primary" onclick={() => { showCreateModal = true; error = ''; success = ''; }}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M12 5v14M5 12h14"></path>
      </svg>
      Create Project
    </button>
  </div>

  <!-- Messages -->
  {#if error}
    <div class="alert alert-error">
      {error}
      <button onclick={() => error = ''}>×</button>
    </div>
  {/if}

  {#if success}
    <div class="alert alert-success">
      {success}
      <button onclick={() => success = ''}>×</button>
    </div>
  {/if}

  <!-- Tabs -->
  <div class="tabs">
    <button 
      class="tab-btn" 
      class:active={activeTab === 'projects'} 
      onclick={() => activeTab = 'projects'}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
      </svg>
      Projects ({projects.length})
    </button>
    <button 
      class="tab-btn" 
      class:active={activeTab === 'assignments'} 
      onclick={() => activeTab = 'assignments'}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
        <circle cx="9" cy="7" r="4"></circle>
        <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"></path>
      </svg>
      Manager Assignments
    </button>
  </div>

  <!-- Projects Tab -->
  {#if activeTab === 'projects'}
    <div class="projects-section">
      <!-- Filters -->
      <div class="filters">
        <input 
          type="text" 
          placeholder="Search projects..." 
          bind:value={searchTerm}
          class="search-input"
        />
        <select bind:value={filterStatus} onchange={() => loadProjects()} class="filter-select">
          <option value="all">All Status</option>
          <option value="active">Active</option>
          <option value="on-hold">On Hold</option>
          <option value="completed">Completed</option>
          <option value="archived">Archived</option>
        </select>
      </div>

      {#if loading}
        <div class="loading">Loading projects...</div>
      {:else if filteredProjects.length === 0}
        <div class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
          </svg>
          <h3>No projects yet</h3>
          <p>Create your first project to get started</p>
          <button class="btn-primary" onclick={() => showCreateModal = true}>
            Create Project
          </button>
        </div>
      {:else}
        <div class="projects-grid">
          {#each filteredProjects as project}
            <div class="project-card">
              <div class="project-header">
                <h3>{project.name}</h3>
                <div class="project-meta">
                  <span class="status {getStatusColor(project.status)}">
                    {project.status}
                  </span>
                  <span class="priority {getPriorityColor(project.priority)}">
                    {project.priority}
                  </span>
                </div>
              </div>

              <p class="project-description">{project.description || 'No description'}</p>

              {#if project.manager_name}
                <div class="project-info">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                    <circle cx="12" cy="7" r="4"></circle>
                  </svg>
                  <span>Manager: {project.manager_name}</span>
                </div>
              {:else}
                <div class="project-info warning">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                  </svg>
                  <span>No manager assigned</span>
                </div>
              {/if}

              <div class="project-info">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                  <circle cx="9" cy="7" r="4"></circle>
                  <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"></path>
                </svg>
                <span>{project.member_count} team members</span>
              </div>

              <div class="project-actions">
                <button class="btn-secondary" onclick={() => openAssignModal(project)}>
                  Add Member
                </button>
                <button class="btn-icon" onclick={() => selectedProject = project}>
                  View Details
                </button>
                <button class="btn-icon danger" onclick={() => deleteProject(project.id)}>
                  Delete
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  <!-- Manager Assignments Tab -->
  {#if activeTab === 'assignments'}
    <div class="assignments-section">
      <div class="section-header">
        <h2>Manager Assignments</h2>
        <button class="btn-primary" onclick={() => { showManagerAssignModal = true; error = ''; success = ''; }}>
          Assign Employee to Manager
        </button>
      </div>

      <!-- Unassigned Employees -->
      <div class="unassigned-section">
        <h3>Unassigned Employees ({unassignedEmployees.length})</h3>
        {#if unassignedEmployees.length === 0}
          <p class="text-muted">All employees have been assigned to managers</p>
        {:else}
          <div class="employees-grid">
            {#each unassignedEmployees as employee}
              <div class="employee-card">
                <div class="employee-avatar">
                  {employee.first_name[0]}{employee.last_name[0]}
                </div>
                <div class="employee-info">
                  <div class="employee-name">{employee.first_name} {employee.last_name}</div>
                  <div class="employee-position">{employee.position}</div>
                  <div class="employee-dept">{employee.department}</div>
                </div>
                <button 
                  class="btn-secondary btn-sm" 
                  onclick={() => {
                    managerAssignForm.employee_id = employee.id;
                    showManagerAssignModal = true;
                  }}>
                  Assign Manager
                </button>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Assigned Employees by Manager -->
      <div class="assigned-section">
        <h3>Employees by Manager</h3>
        {#each managers as manager}
          {@const managedEmployees = employees.filter(e => e.manager_id === manager.id)}
          {#if managedEmployees.length > 0}
            <div class="manager-group">
              <div class="manager-header">
                <div class="manager-info">
                  <div class="manager-avatar">
                    {manager.first_name[0]}{manager.last_name[0]}
                  </div>
                  <div>
                    <div class="manager-name">{manager.first_name} {manager.last_name}</div>
                    <div class="manager-title">{manager.position}</div>
                  </div>
                </div>
                <div class="team-count">{managedEmployees.length} team members</div>
              </div>
              <div class="team-members">
                {#each managedEmployees as employee}
                  <div class="team-member">
                    <span class="member-name">{employee.first_name} {employee.last_name}</span>
                    <span class="member-position">{employee.position}</span>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        {/each}
      </div>
    </div>
  {/if}

  <!-- Create Project Modal -->
  {#if showCreateModal}
    <div class="modal" onclick={(e) => { if (e.target.classList.contains('modal')) showCreateModal = false; }}>
      <div class="modal-content">
        <div class="modal-header">
          <h2>Create New Project</h2>
          <button class="close-btn" onclick={() => showCreateModal = false}>×</button>
        </div>

        <form onsubmit={(e) => { e.preventDefault(); createProject(); }}>
          <div class="form-group">
            <label>Project Name *</label>
            <input 
              type="text" 
              bind:value={projectForm.name} 
              placeholder="Enter project name" 
              required 
            />
          </div>

          <div class="form-group">
            <label>Description</label>
            <textarea 
              bind:value={projectForm.description} 
              placeholder="Project description"
              rows="3"
            />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Status</label>
              <select bind:value={projectForm.status}>
                <option value="active">Active</option>
                <option value="on-hold">On Hold</option>
                <option value="completed">Completed</option>
                <option value="archived">Archived</option>
              </select>
            </div>

            <div class="form-group">
              <label>Priority</label>
              <select bind:value={projectForm.priority}>
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label>Project Manager</label>
            <select bind:value={projectForm.manager_id}>
              <option value="">Select Manager (Optional)</option>
              {#each managers as manager}
                <option value={manager.id}>
                  {manager.first_name} {manager.last_name} - {manager.position}
                </option>
              {/each}
            </select>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Start Date</label>
              <input type="date" bind:value={projectForm.start_date} />
            </div>

            <div class="form-group">
              <label>End Date</label>
              <input type="date" bind:value={projectForm.end_date} />
            </div>
          </div>

          <div class="form-group">
            <label>Budget</label>
            <input 
              type="number" 
              bind:value={projectForm.budget} 
              placeholder="0.00"
              step="0.01"
              min="0"
            />
          </div>

          <div class="modal-actions">
            <button type="button" class="btn-secondary" onclick={() => showCreateModal = false}>
              Cancel
            </button>
            <button type="submit" class="btn-primary">
              Create Project
            </button>
          </div>
        </form>
      </div>
    </div>
  {/if}

  <!-- Assign Member Modal -->
  {#if showAssignModal && selectedProject}
    <div class="modal" onclick={(e) => { if (e.target.classList.contains('modal')) showAssignModal = false; }}>
      <div class="modal-content">
        <div class="modal-header">
          <h2>Assign Member to {selectedProject.name}</h2>
          <button class="close-btn" onclick={() => showAssignModal = false}>×</button>
        </div>

        <form onsubmit={(e) => { e.preventDefault(); assignMemberToProject(selectedProject.id); }}>
          <div class="form-group">
            <label>Employee *</label>
            <select bind:value={memberForm.employee_id} required>
              <option value="">Select Employee</option>
              {#each employees as employee}
                <option value={employee.id}>
                  {employee.first_name} {employee.last_name} - {employee.position}
                </option>
              {/each}
            </select>
          </div>

          <div class="form-group">
            <label>Role</label>
            <input 
              type="text" 
              bind:value={memberForm.role} 
              placeholder="e.g., Developer, Designer, Analyst" 
            />
          </div>

          <div class="modal-actions">
            <button type="button" class="btn-secondary" onclick={() => showAssignModal = false}>
              Cancel
            </button>
            <button type="submit" class="btn-primary">
              Assign Member
            </button>
          </div>
        </form>
      </div>
    </div>
  {/if}

  <!-- Assign to Manager Modal -->
  {#if showManagerAssignModal}
    <div class="modal" onclick={(e) => { if (e.target.classList.contains('modal')) showManagerAssignModal = false; }}>
      <div class="modal-content">
        <div class="modal-header">
          <h2>Assign Employee to Manager</h2>
          <button class="close-btn" onclick={() => showManagerAssignModal = false}>×</button>
        </div>

        <form onsubmit={(e) => { e.preventDefault(); assignEmployeeToManager(); }}>
          <div class="form-group">
            <label>Employee *</label>
            <select bind:value={managerAssignForm.employee_id} required>
              <option value="">Select Employee</option>
              {#each employees as employee}
                <option value={employee.id}>
                  {employee.first_name} {employee.last_name} - {employee.position}
                </option>
              {/each}
            </select>
          </div>

          <div class="form-group">
            <label>Manager *</label>
            <select bind:value={managerAssignForm.manager_id} required>
              <option value="">Select Manager</option>
              {#each managers as manager}
                <option value={manager.id}>
                  {manager.first_name} {manager.last_name} - {manager.position}
                </option>
              {/each}
            </select>
          </div>

          <div class="modal-actions">
            <button type="button" class="btn-secondary" onclick={() => showManagerAssignModal = false}>
              Cancel
            </button>
            <button type="submit" class="btn-primary">
              Assign to Manager
            </button>
          </div>
        </form>
      </div>
    </div>
  {/if}
</div>

<style>
  .project-management {
    padding: 2rem;
    max-width: 1400px;
    margin: 0 auto;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 2rem;
  }

  .header h1 {
    margin: 0 0 0.5rem 0;
    color: #1f2937;
  }

  .subtitle {
    color: #6b7280;
    margin: 0;
  }

  .alert {
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 1.5rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .alert-error {
    background: #fee2e2;
    color: #991b1b;
    border: 1px solid #fecaca;
  }

  .alert-success {
    background: #d1fae5;
    color: #065f46;
    border: 1px solid #a7f3d0;
  }

  .alert button {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    color: inherit;
    padding: 0;
    width: 24px;
    height: 24px;
  }

  .tabs {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 2rem;
    border-bottom: 2px solid #e5e7eb;
  }

  .tab-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 1rem 1.5rem;
    border: none;
    background: none;
    cursor: pointer;
    font-weight: 500;
    color: #6b7280;
    border-bottom: 2px solid transparent;
    margin-bottom: -2px;
    transition: all 0.2s;
  }

  .tab-btn svg {
    width: 20px;
    height: 20px;
  }

  .tab-btn.active {
    color: #4f46e5;
    border-bottom-color: #4f46e5;
  }

  .tab-btn:hover {
    color: #4f46e5;
  }

  .filters {
    display: flex;
    gap: 1rem;
    margin-bottom: 2rem;
  }

  .search-input {
    flex: 1;
    padding: 0.75rem 1rem;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 1rem;
  }

  .filter-select {
    padding: 0.75rem 1rem;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 1rem;
    background: white;
    min-width: 150px;
  }

  .loading {
    text-align: center;
    padding: 3rem;
    color: #6b7280;
  }

  .empty-state {
    text-align: center;
    padding: 4rem 2rem;
    background: white;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .empty-state svg {
    width: 64px;
    height: 64px;
    color: #d1d5db;
    margin-bottom: 1rem;
  }

  .empty-state h3 {
    margin: 0 0 0.5rem 0;
    color: #1f2937;
  }

  .empty-state p {
    color: #6b7280;
    margin-bottom: 1.5rem;
  }

  .projects-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 1.5rem;
  }

  .project-card {
    background: white;
    border-radius: 12px;
    padding: 1.5rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    transition: box-shadow 0.2s;
  }

  .project-card:hover {
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }

  .project-header {
    margin-bottom: 1rem;
  }

  .project-header h3 {
    margin: 0 0 0.75rem 0;
    color: #1f2937;
    font-size: 1.25rem;
  }

  .project-meta {
    display: flex;
    gap: 0.5rem;
  }

  .status, .priority {
    padding: 0.25rem 0.75rem;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
  }

  .status-active { background: #d1fae5; color: #065f46; }
  .status-onhold { background: #fef3c7; color: #92400e; }
  .status-completed { background: #dbeafe; color: #1e40af; }
  .status-archived { background: #f3f4f6; color: #4b5563; }

  .priority-low { background: #e0e7ff; color: #3730a3; }
  .priority-medium { background: #fef3c7; color: #92400e; }
  .priority-high { background: #fed7aa; color: #9a3412; }
  .priority-critical { background: #fecaca; color: #991b1b; }

  .project-description {
    color: #6b7280;
    margin-bottom: 1rem;
    line-height: 1.5;
  }

  .project-info {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
    color: #6b7280;
    font-size: 0.875rem;
  }

  .project-info svg {
    width: 16px;
    height: 16px;
  }

  .project-info.warning {
    color: #d97706;
  }

  .project-actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #e5e7eb;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
  }

  .section-header h2 {
    margin: 0;
  }

  .unassigned-section,
  .assigned-section {
    background: white;
    border-radius: 12px;
    padding: 2rem;
    margin-bottom: 2rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .unassigned-section h3,
  .assigned-section h3 {
    margin: 0 0 1.5rem 0;
    color: #1f2937;
  }

  .text-muted {
    color: #9ca3af;
    font-style: italic;
  }

  .employees-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1rem;
  }

  .employee-card {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background: #f9fafb;
    border-radius: 8px;
    border: 1px solid #e5e7eb;
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
    font-size: 1rem;
    flex-shrink: 0;
  }

  .employee-info {
    flex: 1;
    min-width: 0;
  }

  .employee-name {
    font-weight: 600;
    color: #1f2937;
    margin-bottom: 0.25rem;
  }

  .employee-position {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .employee-dept {
    font-size: 0.75rem;
    color: #9ca3af;
    margin-top: 0.125rem;
  }

  .manager-group {
    background: #f9fafb;
    border-radius: 8px;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
    border: 1px solid #e5e7eb;
  }

  .manager-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .manager-info {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .manager-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 1rem;
  }

  .manager-name {
    font-weight: 600;
    color: #1f2937;
  }

  .manager-title {
    font-size: 0.875rem;
    color: #6b7280;
    margin-top: 0.125rem;
  }

  .team-count {
    font-size: 0.875rem;
    color: #6b7280;
    font-weight: 500;
  }

  .team-members {
    display: grid;
    gap: 0.75rem;
  }

  .team-member {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 1rem;
    background: white;
    border-radius: 6px;
    border: 1px solid #e5e7eb;
  }

  .member-name {
    font-weight: 500;
    color: #1f2937;
  }

  .member-position {
    font-size: 0.875rem;
    color: #6b7280;
  }

  /* Buttons */
  .btn-primary {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.5rem;
    background: #4f46e5;
    color: white;
    border: none;
    border-radius: 8px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
  }

  .btn-primary:hover {
    background: #4338ca;
  }

  .btn-primary svg {
    width: 20px;
    height: 20px;
  }

  .btn-secondary {
    padding: 0.5rem 1rem;
    background: white;
    color: #4f46e5;
    border: 1px solid #4f46e5;
    border-radius: 6px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-secondary:hover {
    background: #4f46e5;
    color: white;
  }

  .btn-sm {
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
  }

  .btn-icon {
    padding: 0.5rem 0.75rem;
    background: #f3f4f6;
    color: #6b7280;
    border: none;
    border-radius: 6px;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-icon:hover {
    background: #e5e7eb;
    color: #1f2937;
  }

  .btn-icon.danger {
    color: #dc2626;
  }

  .btn-icon.danger:hover {
    background: #fee2e2;
  }

  /* Modal */
  .modal {
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
    padding: 1rem;
  }

  .modal-content {
    background: white;
    border-radius: 12px;
    max-width: 600px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.5rem;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 2rem;
    color: #9ca3af;
    cursor: pointer;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    transition: all 0.2s;
  }

  .close-btn:hover {
    background: #f3f4f6;
    color: #1f2937;
  }

  form {
    padding: 1.5rem;
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  .form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
    color: #374151;
  }

  .form-group input,
  .form-group select,
  .form-group textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 1rem;
    font-family: inherit;
  }

  .form-group input:focus,
  .form-group select:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: #4f46e5;
    ring: 2px;
    ring-color: #e0e7ff;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  .modal-actions {
    display: flex;
    gap: 1rem;
    justify-content: flex-end;
    padding-top: 1.5rem;
    border-top: 1px solid #e5e7eb;
  }

  @media (max-width: 768px) {
    .header {
      flex-direction: column;
      gap: 1rem;
    }

    .filters {
      flex-direction: column;
    }

    .form-row {
      grid-template-columns: 1fr;
    }

    .projects-grid {
      grid-template-columns: 1fr;
    }

    .employees-grid {
      grid-template-columns: 1fr;
    }
  }
</style>