<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();

  interface TeamMember {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    phone?: string;
    employment_type: string;
    position?: string;
    department?: string;
    hire_date?: string;
    status: string;
  }

  interface Project {
    id: string;
    name: string;
    status: string;
    start_date?: string;
    end_date?: string;
  }

  interface ProjectAssignment {
    project_id: string;
    project_name: string;
    role: string;
    assigned_at: string;
  }

  let loading = $state(false);
  let teamMembers = $state<TeamMember[]>([]);
  let filteredMembers = $state<TeamMember[]>([]);
  let searchQuery = $state('');
  let filterType = $state('all');
  let allProjects = $state<Project[]>([]);
  
  // Modal states
  let showProjectModal = $state(false);
  let selectedEmployee = $state<TeamMember | null>(null);
  let employeeProjects = $state<ProjectAssignment[]>([]);
  let availableProjects = $state<Project[]>([]);
  let selectedProjectId = $state('');
  let selectedRole = $state('member');
  let modalLoading = $state(false);
  let modalProcessing = $state(false);
  let successMessage = $state('');
  let errorMessage = $state('');

  onMount(() => {
    loadTeamMembers();
    loadAllProjects();
  });

  async function loadTeamMembers() {
    loading = true;
    try {
      // Load only MY team members (manager_id = current user)
      const response = await fetch(`${API_BASE_URL}/employees/team`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        const data = await response.json();
        teamMembers = Array.isArray(data) ? data : [];
        applyFilters();
      } else {
        console.error('Failed to load team members:', response.status);
        teamMembers = [];
      }
    } catch (err) {
      console.error('Error loading team members:', err);
      teamMembers = [];
    } finally {
      loading = false;
    }
  }

  async function loadAllProjects() {
    try {
      const response = await fetch(`${API_BASE_URL}/projects`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        const data = await response.json();
        allProjects = Array.isArray(data) ? data.filter((p: Project) => p.status === 'active') : [];
      }
    } catch (err) {
      console.error('Error loading projects:', err);
      allProjects = [];
    }
  }

  function applyFilters() {
    let filtered = [...teamMembers];

    if (filterType !== 'all') {
      filtered = filtered.filter(member => {
        if (filterType === '1099' || filterType === 'contractor') {
          return member.employment_type === '1099' || 
                 member.employment_type === 'contractor';
        }
        return member.employment_type === filterType;
      });
    }

    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(member =>
        member.first_name.toLowerCase().includes(query) ||
        member.last_name.toLowerCase().includes(query) ||
        member.email.toLowerCase().includes(query) ||
        (member.position?.toLowerCase().includes(query)) ||
        (member.department?.toLowerCase().includes(query))
      );
    }

    filteredMembers = filtered;
  }

  function handleSearch(e: Event) {
    searchQuery = (e.target as HTMLInputElement).value;
    applyFilters();
  }

  function handleFilterChange(e: Event) {
    filterType = (e.target as HTMLSelectElement).value;
    applyFilters();
  }

  async function openProjectModal(member: TeamMember) {
    selectedEmployee = member;
    showProjectModal = true;
    modalLoading = true;
    errorMessage = '';
    successMessage = '';

    try {
      // Load employee's current projects
      const response = await fetch(`${API_BASE_URL}/projects/employee/${member.id}`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        const data = await response.json();
        employeeProjects = Array.isArray(data) ? data : [];
        
        // Filter available projects (not already assigned)
        const assignedProjectIds = new Set(employeeProjects.map(p => p.project_id));
        availableProjects = allProjects.filter(p => !assignedProjectIds.has(p.id));
      } else {
        employeeProjects = [];
        availableProjects = [...allProjects];
      }
    } catch (err) {
      console.error('Error loading employee projects:', err);
      employeeProjects = [];
      availableProjects = [...allProjects];
    } finally {
      modalLoading = false;
    }
  }

  function closeProjectModal() {
    showProjectModal = false;
    selectedEmployee = null;
    employeeProjects = [];
    availableProjects = [];
    selectedProjectId = '';
    selectedRole = 'member';
    errorMessage = '';
    successMessage = '';
  }

  async function assignProject() {
    if (!selectedEmployee || !selectedProjectId) {
      errorMessage = 'Please select a project';
      return;
    }

    modalProcessing = true;
    errorMessage = '';

    try {
      const response = await fetch(`${API_BASE_URL}/projects/${selectedProjectId}/members`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          employee_id: selectedEmployee.id,
          role: selectedRole
        })
      });

      if (response.ok) {
        successMessage = `Project assigned successfully!`;
        // Reload the modal data
        await openProjectModal(selectedEmployee);
        selectedProjectId = '';
        selectedRole = 'member';
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to assign project';
      }
    } catch (err) {
      console.error('Error assigning project:', err);
      errorMessage = 'An error occurred while assigning the project';
    } finally {
      modalProcessing = false;
    }
  }

  async function revokeProject(projectId: string) {
    if (!selectedEmployee) return;

    if (!confirm('Are you sure you want to revoke this project assignment?')) {
      return;
    }

    modalProcessing = true;
    errorMessage = '';
    successMessage = '';

    try {
      const response = await fetch(`${API_BASE_URL}/projects/${projectId}/members/${selectedEmployee.id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });

      if (response.ok) {
        successMessage = 'Project revoked successfully!';
        // Reload the modal data
        await openProjectModal(selectedEmployee);
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to revoke project';
      }
    } catch (err) {
      console.error('Error revoking project:', err);
      errorMessage = 'An error occurred while revoking the project';
    } finally {
      modalProcessing = false;
    }
  }

  function getEmploymentTypeBadge(type: string): string {
    const types: Record<string, string> = {
      'W2': 'badge-w2',
      'full-time': 'badge-fulltime',
      'part-time': 'badge-parttime',
      '1099': 'badge-contractor',
      'contractor': 'badge-contractor'
    };
    return types[type] || 'badge-default';
  }

  function getEmploymentTypeLabel(type: string): string {
    const labels: Record<string, string> = {
      'W2': 'W-2 Employee',
      'full-time': 'Full Time',
      'part-time': 'Part Time',
      '1099': 'Contractor',
      'contractor': 'Contractor'
    };
    return labels[type] || type;
  }

  function formatDate(dateString?: string): string {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function getInitials(firstName: string, lastName: string): string {
    return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase();
  }
</script>

<div class="team-management">
  <div class="section-header">
    <h2>My Team</h2>
    <p class="section-subtitle">Manage your assigned employees and contractors</p>
  </div>

  {#if successMessage}
    <div class="alert alert-success">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
        <polyline points="22 4 12 14.01 9 11.01"></polyline>
      </svg>
      {successMessage}
    </div>
  {/if}

  <!-- Filters and Search -->
  <div class="controls">
    <div class="search-box">
      <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"></circle>
        <path d="m21 21-4.35-4.35"></path>
      </svg>
      <input
        type="text"
        placeholder="Search by name, email, position..."
        value={searchQuery}
        oninput={handleSearch}
        class="search-input"
      />
    </div>

    <select value={filterType} onchange={handleFilterChange} class="filter-select">
      <option value="all">All Types</option>
      <option value="W2">W-2 Employees</option>
      <option value="full-time">Full Time</option>
      <option value="part-time">Part Time</option>
      <option value="contractor">Contractors</option>
    </select>
  </div>

  {#if loading}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading team members...</p>
    </div>
  {:else if filteredMembers.length === 0}
    <div class="empty-state">
      {#if searchQuery || filterType !== 'all'}
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"></circle>
          <path d="m21 21-4.35-4.35"></path>
        </svg>
        <h3>No team members found</h3>
        <p>Try adjusting your search or filter</p>
        <button type="button" class="btn btn-secondary" onclick={() => { searchQuery = ''; filterType = 'all'; applyFilters(); }}>
          Clear Filters
        </button>
      {:else}
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"></path>
          <circle cx="9" cy="7" r="4"></circle>
          <path d="M23 21v-2a4 4 0 00-3-3.87"></path>
          <path d="M16 3.13a4 4 0 010 7.75"></path>
        </svg>
        <h3>No team members assigned</h3>
        <p>You don't have any employees or contractors reporting to you yet.</p>
        <p class="help-text">Team members will appear here when their manager_id is set to your employee ID.</p>
      {/if}
    </div>
  {:else}
    <div class="results-header">
      <span class="results-count">{filteredMembers.length} team member{filteredMembers.length !== 1 ? 's' : ''}</span>
    </div>

    <div class="team-grid">
      {#each filteredMembers as member}
        <div class="member-card">
          <div class="member-header">
            <div class="member-avatar">
              {getInitials(member.first_name, member.last_name)}
            </div>
            <div class="member-info">
              <h3>{member.first_name} {member.last_name}</h3>
              <p class="member-email">{member.email}</p>
            </div>
          </div>

          <div class="member-details">
            <div class="detail-row">
              <span class="detail-label">Type</span>
              <span class="badge {getEmploymentTypeBadge(member.employment_type)}">
                {getEmploymentTypeLabel(member.employment_type)}
              </span>
            </div>

            {#if member.position}
              <div class="detail-row">
                <span class="detail-label">Position</span>
                <span class="detail-value">{member.position}</span>
              </div>
            {/if}

            {#if member.department}
              <div class="detail-row">
                <span class="detail-label">Department</span>
                <span class="detail-value">{member.department}</span>
              </div>
            {/if}

            {#if member.phone}
              <div class="detail-row">
                <span class="detail-label">Phone</span>
                <span class="detail-value">{member.phone}</span>
              </div>
            {/if}

            <div class="detail-row">
              <span class="detail-label">Hire Date</span>
              <span class="detail-value">{formatDate(member.hire_date)}</span>
            </div>

            <div class="detail-row">
              <span class="detail-label">Status</span>
              <span class="status-badge status-{member.status}">
                {member.status}
              </span>
            </div>
          </div>

          <div class="member-actions">
            <button 
              type="button" 
              class="btn btn-sm btn-primary"
              onclick={() => openProjectModal(member)}
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="3" width="7" height="7"></rect>
                <rect x="14" y="3" width="7" height="7"></rect>
                <rect x="14" y="14" width="7" height="7"></rect>
                <rect x="3" y="14" width="7" height="7"></rect>
              </svg>
              Manage Projects
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Project Assignment Modal -->
{#if showProjectModal && selectedEmployee}
  <div class="modal-overlay" onclick={closeProjectModal}>
    <div class="modal" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h3>Manage Projects - {selectedEmployee.first_name} {selectedEmployee.last_name}</h3>
        <button type="button" class="modal-close" onclick={closeProjectModal}>
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

        <!-- Assign New Project -->
        <div class="assign-section">
          <h4>Assign Project</h4>
          <div class="assign-form">
            <select 
              bind:value={selectedProjectId} 
              class="project-select"
              disabled={modalProcessing}
            >
              <option value="">Select a project...</option>
              {#each availableProjects as project}
                <option value={project.id}>{project.name}</option>
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
              onclick={assignProject}
              disabled={modalProcessing || !selectedProjectId}
            >
              {modalProcessing ? 'Assigning...' : 'Assign'}
            </button>
          </div>
          {#if availableProjects.length === 0}
            <p class="no-projects">All active projects have been assigned</p>
          {/if}
        </div>

        <!-- Current Projects -->
        <div class="current-projects">
          <h4>Current Projects</h4>
          {#if modalLoading}
            <div class="modal-loading">
              <div class="spinner-sm"></div>
              <span>Loading projects...</span>
            </div>
          {:else if employeeProjects.length === 0}
            <p class="no-projects">No projects assigned yet</p>
          {:else}
            <div class="projects-list">
              {#each employeeProjects as assignment}
                <div class="project-item">
                  <div class="project-info">
                    <h5>{assignment.project_name}</h5>
                    <div class="project-meta">
                      <span class="role-badge">{assignment.role}</span>
                      <span class="date-assigned">Assigned {formatDate(assignment.assigned_at)}</span>
                    </div>
                  </div>
                  <button 
                    type="button"
                    class="btn-revoke"
                    onclick={() => revokeProject(assignment.project_id)}
                    disabled={modalProcessing}
                  >
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <line x1="18" y1="6" x2="6" y2="18"></line>
                      <line x1="6" y1="6" x2="18" y2="18"></line>
                    </svg>
                    Revoke
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" onclick={closeProjectModal}>
          Close
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .team-management {
    padding: 24px;
  }

  .section-header {
    margin-bottom: 24px;
  }

  .section-header h2 {
    font-size: 24px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 8px 0;
  }

  .section-subtitle {
    font-size: 14px;
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

  /* Controls */
  .controls {
    display: flex;
    gap: 16px;
    margin-bottom: 24px;
  }

  .search-box {
    flex: 1;
    position: relative;
  }

  .search-icon {
    position: absolute;
    left: 12px;
    top: 50%;
    transform: translateY(-50%);
    width: 20px;
    height: 20px;
    color: #9ca3af;
  }

  .search-input {
    width: 100%;
    padding: 12px 12px 12px 44px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
  }

  .search-input:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .filter-select {
    padding: 12px 16px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    background: white;
    cursor: pointer;
  }

  .filter-select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  /* Results */
  .results-header {
    margin-bottom: 16px;
  }

  .results-count {
    font-size: 14px;
    color: #6b7280;
    font-weight: 500;
  }

  /* Team Grid */
  .team-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 20px;
  }

  .member-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
    transition: all 0.2s;
  }

  .member-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  }

  .member-header {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 20px;
    padding-bottom: 20px;
    border-bottom: 1px solid #e5e7eb;
  }

  .member-avatar {
    width: 56px;
    height: 56px;
    border-radius: 50%;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .member-info {
    flex: 1;
    min-width: 0;
  }

  .member-info h3 {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .member-email {
    font-size: 14px;
    color: #6b7280;
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .member-details {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 20px;
  }

  .detail-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .detail-label {
    font-size: 13px;
    color: #6b7280;
    font-weight: 500;
  }

  .detail-value {
    font-size: 14px;
    color: #111827;
    font-weight: 500;
  }

  /* Badges */
  .badge {
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
    text-transform: capitalize;
  }

  .badge-w2 {
    background: #dbeafe;
    color: #1e40af;
  }

  .badge-fulltime {
    background: #d1fae5;
    color: #065f46;
  }

  .badge-parttime {
    background: #e0e7ff;
    color: #3730a3;
  }

  .badge-contractor {
    background: #fef3c7;
    color: #92400e;
  }

  .badge-default {
    background: #f3f4f6;
    color: #4b5563;
  }

  .status-badge {
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
    text-transform: capitalize;
  }

  .status-active {
    background: #d1fae5;
    color: #065f46;
  }

  .status-inactive {
    background: #fee2e2;
    color: #991b1b;
  }

  /* Actions */
  .member-actions {
    display: flex;
    gap: 8px;
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
    max-width: 700px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
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
  }

  /* Assign Section */
  .assign-section {
    margin-bottom: 32px;
  }

  .assign-section h4 {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 16px 0;
  }

  .assign-form {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .project-select,
  .role-select {
    padding: 10px 14px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    background: white;
    cursor: pointer;
  }

  .project-select {
    flex: 2;
  }

  .role-select {
    flex: 1;
  }

  .project-select:focus,
  .role-select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .project-select:disabled,
  .role-select:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Current Projects */
  .current-projects {
    margin-top: 32px;
  }

  .current-projects h4 {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 16px 0;
  }

  .projects-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .project-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
  }

  .project-info {
    flex: 1;
  }

  .project-info h5 {
    font-size: 15px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 8px 0;
  }

  .project-meta {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .role-badge {
    padding: 3px 10px;
    background: #dbeafe;
    color: #1e40af;
    border-radius: 10px;
    font-size: 11px;
    font-weight: 600;
    text-transform: capitalize;
  }

  .date-assigned {
    font-size: 13px;
    color: #6b7280;
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

  .no-projects {
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

  @media (max-width: 768px) {
    .team-grid {
      grid-template-columns: 1fr;
    }

    .controls {
      flex-direction: column;
    }

    .assign-form {
      flex-direction: column;
      align-items: stretch;
    }

    .project-select,
    .role-select {
      width: 100%;
    }
  }
</style>