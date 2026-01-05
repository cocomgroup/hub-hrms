<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();

  interface Props {
    employee?: any;
  }
  
  let { employee }: Props = $props();

  interface Project {
    id: string;
    name: string;
    description?: string;
    status: string;
    start_date?: string;
    end_date?: string;
    budget?: number;
    created_at: string;
  }

  interface ProjectMember {
    id: string;
    employee_id: string;
    employee_name: string;
    email: string;
    role: string;
    assigned_at: string;
    employment_type?: string;
  }

  interface TeamMember {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    employment_type: string;
    position?: string;
    status: string;
  }

  let loading = $state(false);
  let activeTab = $state('projects'); // 'projects' | 'team'
  let projects = $state<Project[]>([]);
  let teamMembers = $state<TeamMember[]>([]);
  let selectedProject = $state<Project | null>(null);
  let projectMembers = $state<ProjectMember[]>([]);
  let availableEmployees = $state<TeamMember[]>([]);
  
  // Stats
  let projectCount = $state(0);
  let teamMemberCount = $state(0);
  
  // Modal states
  let showMembersModal = $state(false);
  let showProjectForm = $state(false);
  let modalLoading = $state(false);
  let modalProcessing = $state(false);
  let selectedEmployeeId = $state('');
  let selectedRole = $state('member');
  
  // Messages
  let successMessage = $state('');
  let errorMessage = $state('');

  // Form states for new/edit project
  let projectForm = $state({
    name: '',
    description: '',
    status: 'active',
    start_date: '',
    end_date: '',
    budget: 0
  });
  let editingProjectId = $state<string | null>(null);

  let currentManager = $derived(employee || $authStore.employee);

  onMount(() => {
    loadProjects();
    loadTeamMembers();
  });

  async function loadProjects() {
    loading = true;
    try {
      const response = await fetch(`${API_BASE_URL}/projects`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        const data = await response.json();
        projects = Array.isArray(data) ? data : [];
        projectCount = projects.length;
      }
    } catch (err) {
      console.error('Error loading projects:', err);
      projects = [];
    } finally {
      loading = false;
    }
  }

  async function loadTeamMembers() {
    try {
      // Load only MY team members (manager_id = current user)
      const response = await fetch(`${API_BASE_URL}/employees/team`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        const data = await response.json();
        teamMembers = Array.isArray(data) ? data : [];
        teamMemberCount = teamMembers.length;
      }
    } catch (err) {
      console.error('Error loading team members:', err);
      teamMembers = [];
    }
  }

  async function openMembersModal(project: Project) {
    selectedProject = project;
    showMembersModal = true;
    modalLoading = true;
    errorMessage = '';
    successMessage = '';

    try {
      // Load project members
      const membersResponse = await fetch(`${API_BASE_URL}/projects/${project.id}/members`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (membersResponse.ok) {
        const data = await membersResponse.json();
        projectMembers = Array.isArray(data) ? data : [];
        
        // Filter available employees (from MY team only, not already assigned)
        const assignedEmployeeIds = new Set(projectMembers.map(m => m.employee_id));
        availableEmployees = teamMembers.filter(e => !assignedEmployeeIds.has(e.id));
      } else {
        projectMembers = [];
        availableEmployees = [...teamMembers];
      }
    } catch (err) {
      console.error('Error loading project members:', err);
      projectMembers = [];
      availableEmployees = [...teamMembers];
    } finally {
      modalLoading = false;
    }
  }

  function closeMembersModal() {
    showMembersModal = false;
    selectedProject = null;
    projectMembers = [];
    availableEmployees = [];
    selectedEmployeeId = '';
    selectedRole = 'member';
  }

  async function assignMember() {
    if (!selectedProject || !selectedEmployeeId) {
      errorMessage = 'Please select an employee';
      return;
    }

    modalProcessing = true;
    errorMessage = '';

    try {
      const response = await fetch(`${API_BASE_URL}/projects/${selectedProject.id}/members`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          employee_id: selectedEmployeeId,
          role: selectedRole
        })
      });

      if (response.ok) {
        successMessage = 'Team member assigned successfully!';
        selectedEmployeeId = '';
        selectedRole = 'member';
        // Reload modal data
        await openMembersModal(selectedProject);
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to assign team member';
      }
    } catch (err) {
      console.error('Error assigning member:', err);
      errorMessage = 'An error occurred while assigning the team member';
    } finally {
      modalProcessing = false;
    }
  }

  async function revokeMember(memberId: string, memberName: string) {
    if (!selectedProject) return;

    if (!confirm(`Remove ${memberName} from ${selectedProject.name}?`)) {
      return;
    }

    modalProcessing = true;
    errorMessage = '';

    try {
      const response = await fetch(`${API_BASE_URL}/projects/${selectedProject.id}/members/${memberId}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });

      if (response.ok) {
        successMessage = 'Team member removed successfully!';
        // Reload modal data
        await openMembersModal(selectedProject);
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to remove team member';
      }
    } catch (err) {
      console.error('Error removing member:', err);
      errorMessage = 'An error occurred while removing the team member';
    } finally {
      modalProcessing = false;
    }
  }

  function openProjectForm(project?: Project) {
    if (project) {
      // Edit mode
      editingProjectId = project.id;
      projectForm = {
        name: project.name,
        description: project.description || '',
        status: project.status,
        start_date: project.start_date ? project.start_date.split('T')[0] : '',
        end_date: project.end_date ? project.end_date.split('T')[0] : '',
        budget: project.budget || 0
      };
    } else {
      // Create mode
      editingProjectId = null;
      projectForm = {
        name: '',
        description: '',
        status: 'active',
        start_date: '',
        end_date: '',
        budget: 0
      };
    }
    showProjectForm = true;
    errorMessage = '';
    successMessage = '';
  }

  function closeProjectForm() {
    showProjectForm = false;
    editingProjectId = null;
    errorMessage = '';
    successMessage = '';
  }

  async function saveProject() {
    if (!projectForm.name.trim()) {
      errorMessage = 'Project name is required';
      return;
    }

    modalProcessing = true;
    errorMessage = '';

    try {
      const url = editingProjectId 
        ? `${API_BASE_URL}/projects/${editingProjectId}`
        : `${API_BASE_URL}/projects`;
      
      const method = editingProjectId ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(projectForm)
      });

      if (response.ok) {
        successMessage = editingProjectId 
          ? 'Project updated successfully!' 
          : 'Project created successfully!';
        closeProjectForm();
        await loadProjects();
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to save project';
      }
    } catch (err) {
      console.error('Error saving project:', err);
      errorMessage = 'An error occurred while saving the project';
    } finally {
      modalProcessing = false;
    }
  }

  async function deleteProject(project: Project) {
    if (!confirm(`Are you sure you want to delete "${project.name}"? This action cannot be undone.`)) {
      return;
    }

    try {
      const response = await fetch(`${API_BASE_URL}/projects/${project.id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });

      if (response.ok) {
        successMessage = 'Project deleted successfully!';
        await loadProjects();
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to delete project';
      }
    } catch (err) {
      console.error('Error deleting project:', err);
      errorMessage = 'An error occurred while deleting the project';
    }
  }

  function getStatusBadgeClass(status: string): string {
    const classes: Record<string, string> = {
      'active': 'status-active',
      'completed': 'status-completed',
      'on-hold': 'status-hold',
      'cancelled': 'status-cancelled'
    };
    return classes[status] || 'status-default';
  }

  function formatDate(dateString?: string): string {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function getInitials(name: string): string {
    return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2);
  }
</script>

<div class="project-manager">
  <div class="header">
    <div class="header-content">
      <h1>Project Management</h1>
      <p class="subtitle">Welcome back, {currentManager?.first_name || 'Manager'}</p>
    </div>
    <button type="button" class="btn btn-primary" onclick={() => openProjectForm()}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="12" y1="5" x2="12" y2="19"></line>
        <line x1="5" y1="12" x2="19" y2="12"></line>
      </svg>
      New Project
    </button>
  </div>

  {#if successMessage}
    <div class="alert alert-success">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
        <polyline points="22 4 12 14.01 9 11.01"></polyline>
      </svg>
      {successMessage}
      <button type="button" class="alert-close" onclick={() => successMessage = ''}>×</button>
    </div>
  {/if}

  {#if errorMessage}
    <div class="alert alert-error">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"></circle>
        <line x1="12" y1="8" x2="12" y2="12"></line>
        <line x1="12" y1="16" x2="12.01" y2="16"></line>
      </svg>
      {errorMessage}
      <button type="button" class="alert-close" onclick={() => errorMessage = ''}>×</button>
    </div>
  {/if}

  <!-- Stats Cards -->
  <div class="stats-grid">
    <button 
      type="button"
      class="stat-card {activeTab === 'projects' ? 'active' : ''}"
      onclick={() => activeTab = 'projects'}
    >
      <div class="stat-icon projects">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="3" width="7" height="7"></rect>
          <rect x="14" y="3" width="7" height="7"></rect>
          <rect x="14" y="14" width="7" height="7"></rect>
          <rect x="3" y="14" width="7" height="7"></rect>
        </svg>
      </div>
      <div class="stat-content">
        <div class="stat-value">{projectCount}</div>
        <div class="stat-label">Projects</div>
      </div>
    </button>

    <button 
      type="button"
      class="stat-card {activeTab === 'team' ? 'active' : ''}"
      onclick={() => activeTab = 'team'}
    >
      <div class="stat-icon team">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"></path>
          <circle cx="9" cy="7" r="4"></circle>
          <path d="M23 21v-2a4 4 0 00-3-3.87"></path>
          <path d="M16 3.13a4 4 0 010 7.75"></path>
        </svg>
      </div>
      <div class="stat-content">
        <div class="stat-value">{teamMemberCount}</div>
        <div class="stat-label">My Team Members</div>
      </div>
    </button>
  </div>

  <!-- Tabs -->
  <div class="tabs">
    <button
      type="button"
      class="tab-btn {activeTab === 'projects' ? 'active' : ''}"
      onclick={() => activeTab = 'projects'}
    >
      📋 Projects
    </button>
    <button
      type="button"
      class="tab-btn {activeTab === 'team' ? 'active' : ''}"
      onclick={() => activeTab = 'team'}
    >
      👥 Team Overview
    </button>
  </div>

  <!-- Content -->
  <div class="content">
    {#if loading}
      <div class="loading">
        <div class="spinner"></div>
        <p>Loading...</p>
      </div>
    {:else if activeTab === 'projects'}
      {#if projects.length === 0}
        <div class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="7" height="7"></rect>
            <rect x="14" y="3" width="7" height="7"></rect>
            <rect x="14" y="14" width="7" height="7"></rect>
            <rect x="3" y="14" width="7" height="7"></rect>
          </svg>
          <h3>No projects yet</h3>
          <p>Create your first project to get started</p>
          <button type="button" class="btn btn-primary" onclick={() => openProjectForm()}>
            Create Project
          </button>
        </div>
      {:else}
        <div class="projects-grid">
          {#each projects as project}
            <div class="project-card">
              <div class="project-header">
                <h3>{project.name}</h3>
                <span class="status-badge {getStatusBadgeClass(project.status)}">
                  {project.status}
                </span>
              </div>

              {#if project.description}
                <p class="project-description">{project.description}</p>
              {/if}

              <div class="project-meta">
                {#if project.start_date}
                  <div class="meta-item">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                      <line x1="16" y1="2" x2="16" y2="6"></line>
                      <line x1="8" y1="2" x2="8" y2="6"></line>
                      <line x1="3" y1="10" x2="21" y2="10"></line>
                    </svg>
                    {formatDate(project.start_date)}
                    {#if project.end_date}
                      → {formatDate(project.end_date)}
                    {/if}
                  </div>
                {/if}
                {#if project.budget}
                  <div class="meta-item">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <line x1="12" y1="1" x2="12" y2="23"></line>
                      <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
                    </svg>
                    ${project.budget.toLocaleString()}
                  </div>
                {/if}
              </div>

              <div class="project-actions">
                <button 
                  type="button" 
                  class="btn btn-sm btn-secondary"
                  onclick={() => openMembersModal(project)}
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"></path>
                    <circle cx="9" cy="7" r="4"></circle>
                    <path d="M23 21v-2a4 4 0 00-3-3.87"></path>
                    <path d="M16 3.13a4 4 0 010 7.75"></path>
                  </svg>
                  Manage Team
                </button>
                <button 
                  type="button" 
                  class="btn btn-sm btn-secondary"
                  onclick={() => openProjectForm(project)}
                >
                  Edit
                </button>
                <button 
                  type="button" 
                  class="btn btn-sm btn-danger"
                  onclick={() => deleteProject(project)}
                >
                  Delete
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    {:else if activeTab === 'team'}
      {#if teamMembers.length === 0}
        <div class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"></path>
            <circle cx="9" cy="7" r="4"></circle>
            <path d="M23 21v-2a4 4 0 00-3-3.87"></path>
            <path d="M16 3.13a4 4 0 010 7.75"></path>
          </svg>
          <h3>No team members assigned</h3>
          <p>You don't have any employees or contractors reporting to you yet.</p>
          <p class="help-text">Team members will appear here when their manager_id is set to your employee ID.</p>
        </div>
      {:else}
        <div class="team-grid">
          {#each teamMembers as member}
            <div class="team-card">
              <div class="team-avatar">
                {getInitials(`${member.first_name} ${member.last_name}`)}
              </div>
              <h4>{member.first_name} {member.last_name}</h4>
              <p class="team-email">{member.email}</p>
              {#if member.position}
                <p class="team-position">{member.position}</p>
              {/if}
              <span class="employment-badge">
                {member.employment_type}
              </span>
            </div>
          {/each}
        </div>
      {/if}
    {/if}
  </div>
</div>

<!-- Project Members Modal -->
{#if showMembersModal && selectedProject}
  <div class="modal-overlay" onclick={closeMembersModal}>
    <div class="modal large" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h3>Manage Team - {selectedProject.name}</h3>
        <button type="button" class="modal-close" onclick={closeMembersModal}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>

      <div class="modal-body">
        {#if successMessage}
          <div class="alert alert-success">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="20 6 9 17 4 12"></polyline>
            </svg>
            {successMessage}
          </div>
        {/if}

        {#if errorMessage}
          <div class="alert alert-error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="8" x2="12" y2="12"></line>
              <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
            {errorMessage}
          </div>
        {/if}

        <!-- Assign Team Member -->
        <div class="assign-section">
          <h4>Assign Team Member</h4>
          <p class="help-text">Only your direct reports can be assigned to this project</p>
          <div class="assign-form">
            <select 
              bind:value={selectedEmployeeId} 
              class="employee-select"
              disabled={modalProcessing}
            >
              <option value="">Select team member...</option>
              {#each availableEmployees as emp}
                <option value={emp.id}>{emp.first_name} {emp.last_name} ({emp.employment_type})</option>
              {/each}
            </select>

            <select 
              bind:value={selectedRole} 
              class="role-select"
              disabled={modalProcessing}
            >
              <option value="member">Member</option>
              <option value="lead">Lead</option>
              <option value="manager">Manager</option>
            </select>

            <button 
              type="button"
              class="btn btn-primary"
              onclick={assignMember}
              disabled={modalProcessing || !selectedEmployeeId}
            >
              {modalProcessing ? 'Assigning...' : 'Assign'}
            </button>
          </div>
          {#if availableEmployees.length === 0}
            <p class="no-members">All your team members are already assigned to this project</p>
          {/if}
        </div>

        <!-- Current Team Members -->
        <div class="current-members">
          <h4>Current Team Members</h4>
          {#if modalLoading}
            <div class="modal-loading">
              <div class="spinner-sm"></div>
              <span>Loading team...</span>
            </div>
          {:else if projectMembers.length === 0}
            <p class="no-members">No team members assigned yet</p>
          {:else}
            <div class="members-list">
              {#each projectMembers as member}
                <div class="member-item">
                  <div class="member-avatar-sm">
                    {getInitials(member.employee_name)}
                  </div>
                  <div class="member-info">
                    <h5>{member.employee_name}</h5>
                    <p>{member.email}</p>
                    <div class="member-meta">
                      <span class="role-badge">{member.role}</span>
                      {#if member.employment_type}
                        <span class="type-badge">{member.employment_type}</span>
                      {/if}
                      <span class="date-assigned">Assigned {formatDate(member.assigned_at)}</span>
                    </div>
                  </div>
                  <button 
                    type="button"
                    class="btn-revoke"
                    onclick={() => revokeMember(member.employee_id, member.employee_name)}
                    disabled={modalProcessing}
                  >
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <line x1="18" y1="6" x2="6" y2="18"></line>
                      <line x1="6" y1="6" x2="18" y2="18"></line>
                    </svg>
                    Remove
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" onclick={closeMembersModal}>
          Close
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Project Form Modal -->
{#if showProjectForm}
  <div class="modal-overlay" onclick={closeProjectForm}>
    <div class="modal" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h3>{editingProjectId ? 'Edit Project' : 'Create New Project'}</h3>
        <button type="button" class="modal-close" onclick={closeProjectForm}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>

      <div class="modal-body">
        {#if errorMessage}
          <div class="alert alert-error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="8" x2="12" y2="12"></line>
              <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
            {errorMessage}
          </div>
        {/if}

        <form class="project-form" onsubmit={(e) => { e.preventDefault(); saveProject(); }}>
          <div class="form-group">
            <label for="project-name">Project Name <span class="required">*</span></label>
            <input 
              id="project-name"
              type="text" 
              bind:value={projectForm.name}
              placeholder="Enter project name"
              required
            />
          </div>

          <div class="form-group">
            <label for="project-description">Description</label>
            <textarea 
              id="project-description"
              bind:value={projectForm.description}
              placeholder="Enter project description"
              rows="3"
            ></textarea>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="project-status">Status</label>
              <select id="project-status" bind:value={projectForm.status}>
                <option value="active">Active</option>
                <option value="completed">Completed</option>
                <option value="on-hold">On Hold</option>
                <option value="cancelled">Cancelled</option>
              </select>
            </div>

            <div class="form-group">
              <label for="project-budget">Budget ($)</label>
              <input 
                id="project-budget"
                type="number" 
                bind:value={projectForm.budget}
                min="0"
                step="1000"
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="project-start">Start Date</label>
              <input 
                id="project-start"
                type="date" 
                bind:value={projectForm.start_date}
              />
            </div>

            <div class="form-group">
              <label for="project-end">End Date</label>
              <input 
                id="project-end"
                type="date" 
                bind:value={projectForm.end_date}
              />
            </div>
          </div>
        </form>
      </div>

      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" onclick={closeProjectForm}>
          Cancel
        </button>
        <button 
          type="button" 
          class="btn btn-primary" 
          onclick={saveProject}
          disabled={modalProcessing}
        >
          {modalProcessing ? 'Saving...' : (editingProjectId ? 'Update' : 'Create')}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .project-manager {
    padding: 24px;
    max-width: 1400px;
    margin: 0 auto;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 32px;
  }

  .header-content h1 {
    font-size: 32px;
    font-weight: 700;
    color: #111827;
    margin: 0 0 8px 0;
  }

  .subtitle {
    font-size: 16px;
    color: #6b7280;
    margin: 0;
  }

  /* Alerts */
  .alert {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    border-radius: 8px;
    margin-bottom: 20px;
    position: relative;
  }

  .alert svg {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
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

  .alert-close {
    margin-left: auto;
    background: none;
    border: none;
    font-size: 24px;
    cursor: pointer;
    color: inherit;
    padding: 0 8px;
  }

  /* Stats Grid */
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 20px;
    margin-bottom: 32px;
  }

  .stat-card {
    background: white;
    border: 2px solid #e5e7eb;
    border-radius: 16px;
    padding: 24px;
    display: flex;
    align-items: center;
    gap: 20px;
    transition: all 0.3s ease;
    cursor: pointer;
    text-align: left;
  }

  .stat-card:hover {
    border-color: #3b82f6;
    transform: translateY(-2px);
    box-shadow: 0 8px 16px rgba(59, 130, 246, 0.15);
  }

  .stat-card.active {
    border-color: #3b82f6;
    background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  }

  .stat-icon {
    width: 64px;
    height: 64px;
    border-radius: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .stat-icon svg {
    width: 32px;
    height: 32px;
  }

  .stat-icon.projects {
    background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
    color: #1e40af;
  }

  .stat-icon.team {
    background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
    color: #92400e;
  }

  .stat-content {
    flex: 1;
  }

  .stat-value {
    font-size: 36px;
    font-weight: 700;
    color: #111827;
    line-height: 1;
    margin-bottom: 8px;
  }

  .stat-label {
    font-size: 14px;
    color: #6b7280;
    font-weight: 500;
  }

  /* Tabs */
  .tabs {
    display: flex;
    gap: 8px;
    margin-bottom: 24px;
    border-bottom: 2px solid #e5e7eb;
  }

  .tab-btn {
    padding: 16px 24px;
    background: transparent;
    border: none;
    border-bottom: 3px solid transparent;
    font-size: 16px;
    font-weight: 600;
    color: #6b7280;
    cursor: pointer;
    transition: all 0.2s;
    margin-bottom: -2px;
  }

  .tab-btn:hover {
    color: #3b82f6;
    background: #f9fafb;
  }

  .tab-btn.active {
    color: #3b82f6;
    border-bottom-color: #3b82f6;
  }

  /* Content */
  .content {
    background: white;
    border-radius: 16px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    padding: 24px;
    min-height: 400px;
  }

  /* Projects Grid */
  .projects-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 20px;
  }

  .project-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
    transition: all 0.2s;
  }

  .project-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  }

  .project-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 12px;
  }

  .project-header h3 {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0;
    flex: 1;
  }

  .status-badge {
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
    text-transform: capitalize;
    white-space: nowrap;
  }

  .status-active {
    background: #d1fae5;
    color: #065f46;
  }

  .status-completed {
    background: #dbeafe;
    color: #1e40af;
  }

  .status-hold {
    background: #fef3c7;
    color: #92400e;
  }

  .status-cancelled {
    background: #fee2e2;
    color: #991b1b;
  }

  .project-description {
    font-size: 14px;
    color: #6b7280;
    margin: 0 0 16px 0;
    line-height: 1.5;
  }

  .project-meta {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 16px;
    padding-top: 16px;
    border-top: 1px solid #e5e7eb;
  }

  .meta-item {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: #6b7280;
  }

  .meta-item svg {
    width: 16px;
    height: 16px;
  }

  .project-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  /* Team Grid */
  .team-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 20px;
  }

  .team-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
    text-align: center;
    transition: all 0.2s;
  }

  .team-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  }

  .team-avatar {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;
    font-weight: 600;
    margin: 0 auto 16px;
  }

  .team-card h4 {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }

  .team-email {
    font-size: 13px;
    color: #6b7280;
    margin: 0 0 8px 0;
  }

  .team-position {
    font-size: 14px;
    color: #374151;
    margin: 0 0 12px 0;
  }

  .employment-badge {
    display: inline-block;
    padding: 4px 12px;
    background: #fef3c7;
    color: #92400e;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
    text-transform: capitalize;
  }

  /* Empty State */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 20px;
    text-align: center;
  }

  .empty-state svg {
    width: 64px;
    height: 64px;
    color: #d1d5db;
    margin-bottom: 16px;
  }

  .empty-state h3 {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 8px 0;
  }

  .empty-state p {
    font-size: 14px;
    color: #6b7280;
    margin: 0 0 8px 0;
  }

  .help-text {
    font-size: 13px;
    color: #9ca3af;
    font-style: italic;
    margin-bottom: 20px;
  }

  /* Loading */
  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 20px;
    color: #6b7280;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    margin-bottom: 16px;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* Buttons */
  .btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 10px 20px;
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
    width: 16px;
    height: 16px;
  }

  .btn-sm {
    padding: 8px 14px;
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

  .btn-danger {
    background: #fee2e2;
    color: #991b1b;
    border: 1px solid #fecaca;
  }

  .btn-danger:hover:not(:disabled) {
    background: #fecaca;
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
    border-radius: 16px;
    width: 90%;
    max-width: 600px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }

  .modal.large {
    max-width: 800px;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px;
    border-bottom: 1px solid #e5e7eb;
  }

  .modal-header h3 {
    font-size: 20px;
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
    overflow-y: auto;
    flex: 1;
  }

  .modal-footer {
    padding: 16px 24px;
    border-top: 1px solid #e5e7eb;
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  /* Assign Section */
  .assign-section {
    margin-bottom: 32px;
  }

  .assign-section h4 {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 8px 0;
  }

  .assign-form {
    display: flex;
    gap: 12px;
    align-items: center;
    margin-top: 12px;
  }

  .employee-select,
  .role-select {
    padding: 10px 14px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    background: white;
    cursor: pointer;
  }

  .employee-select {
    flex: 2;
  }

  .role-select {
    flex: 1;
  }

  .employee-select:focus,
  .role-select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  /* Current Members */
  .current-members {
    margin-top: 32px;
  }

  .current-members h4 {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 16px 0;
  }

  .members-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .member-item {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
  }

  .member-avatar-sm {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .member-info {
    flex: 1;
  }

  .member-info h5 {
    font-size: 15px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }

  .member-info p {
    font-size: 13px;
    color: #6b7280;
    margin: 0 0 8px 0;
  }

  .member-meta {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .role-badge,
  .type-badge {
    padding: 3px 10px;
    background: #dbeafe;
    color: #1e40af;
    border-radius: 10px;
    font-size: 11px;
    font-weight: 600;
    text-transform: capitalize;
  }

  .type-badge {
    background: #fef3c7;
    color: #92400e;
  }

  .date-assigned {
    font-size: 12px;
    color: #9ca3af;
  }

  .btn-revoke {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 8px 14px;
    background: #fee2e2;
    color: #991b1b;
    border: 1px solid #fecaca;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    flex-shrink: 0;
  }

  .btn-revoke:hover:not(:disabled) {
    background: #fecaca;
  }

  .btn-revoke:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-revoke svg {
    width: 14px;
    height: 14px;
  }

  .no-members {
    text-align: center;
    color: #9ca3af;
    font-size: 14px;
    padding: 20px;
    margin: 0;
  }

  .modal-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 20px;
    color: #6b7280;
  }

  .spinner-sm {
    width: 20px;
    height: 20px;
    border: 3px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  /* Form Styles */
  .project-form {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .form-group label {
    font-size: 14px;
    font-weight: 600;
    color: #374151;
  }

  .required {
    color: #ef4444;
  }

  .form-group input,
  .form-group select,
  .form-group textarea {
    padding: 10px 14px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    font-family: inherit;
  }

  .form-group input:focus,
  .form-group select:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .form-group textarea {
    resize: vertical;
  }

  .form-row {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }

  @media (max-width: 768px) {
    .projects-grid,
    .team-grid {
      grid-template-columns: 1fr;
    }

    .header {
      flex-direction: column;
      gap: 16px;
    }

    .stats-grid {
      grid-template-columns: 1fr;
    }

    .form-row {
      grid-template-columns: 1fr;
    }

    .assign-form {
      flex-direction: column;
      align-items: stretch;
    }

    .employee-select,
    .role-select {
      width: 100%;
    }
  }
</style>