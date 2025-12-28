<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';

  type Props = {
    navigate: (page: string) => void;
  };
  
  let { navigate }: Props = $props();

  const API_BASE_URL = getApiBaseUrl();

  interface TeamStats {
    teamMembers: number;
    pendingPTO: number;
    upcomingReviews: number;
    openTasks: number;
  }

  interface TeamMember {
    id: string;
    name: string;
    email: string;
    position: string;
    status: string;
  }

  interface Manager {
    id: string;
    name: string;
    email: string;
    position: string;
  }

  let stats = $state<TeamStats>({
    teamMembers: 0,
    pendingPTO: 0,
    upcomingReviews: 0,
    openTasks: 0
  });

  let teamMembers = $state<TeamMember[]>([]);
  let loading = $state(true);
  let error = $state('');
  let currentUser = $state<any>(null);
  
  // Manager selector state (for Admin/HR)
  let isAdminOrHR = $state(false);
  let managers = $state<Manager[]>([]);
  let selectedManagerId = $state<string>('');
  let viewingManagerName = $state<string>('');

  async function loadManagers() {
    try {
      const response = await fetch(`${API_BASE_URL}/employees`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to fetch employees');
      }

      const allEmployees = await response.json();
      
      // Get all employees who are managers (have direct reports)
      const managerIds = new Set(
        allEmployees
          .filter((e: any) => e.manager_id)
          .map((e: any) => e.manager_id)
      );

      managers = allEmployees
        .filter((e: any) => managerIds.has(e.id))
        .map((e: any) => ({
          id: e.id,
          name: `${e.first_name} ${e.last_name}`,
          email: e.email,
          position: e.position
        }))
        .sort((a: Manager, b: Manager) => a.name.localeCompare(b.name));

    } catch (err) {
      console.error('Error loading managers:', err);
    }
  }

  async function fetchDashboardData(managerId?: string) {
    try {
      loading = true;
      error = '';

      // Get current user
      authStore.subscribe(value => {
        currentUser = value.user;
        isAdminOrHR = value.user?.role === 'admin' || value.user?.role === 'hr_manager';
      });

      // Determine which manager's data to show
      let targetManagerId = managerId || currentUser?.employee_id;

      // If admin/HR and no manager selected yet, use their own employee_id as default
      if (isAdminOrHR && !selectedManagerId && currentUser?.employee_id) {
        selectedManagerId = currentUser.employee_id;
        targetManagerId = currentUser.employee_id;
      }

      // Fetch all employees
      const employeesRes = await fetch(`${API_BASE_URL}/employees`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (!employeesRes.ok) {
        throw new Error(`Failed to fetch employees: ${employeesRes.status}`);
      }

      const allEmployees = await employeesRes.json();
      
      // Find the manager we're viewing
      const viewingManager = allEmployees.find((e: any) => e.id === targetManagerId);
      if (viewingManager) {
        viewingManagerName = `${viewingManager.first_name} ${viewingManager.last_name}`;
      }

      // Filter team members (employees managed by target manager)
      if (targetManagerId) {
        teamMembers = allEmployees.filter((e: any) => 
          e.manager_id === targetManagerId
        ).map((e: any) => ({
          id: e.id,
          name: `${e.first_name} ${e.last_name}`,
          email: e.email,
          position: e.position,
          status: e.status
        }));
      }

      stats.teamMembers = teamMembers.length;
      stats.pendingPTO = 0;
      stats.upcomingReviews = 0;
      stats.openTasks = 0;

    } catch (err: any) {
      error = 'Failed to load dashboard data';
      console.error('Dashboard error:', err);
    } finally {
      loading = false;
    }
  }

  function handleManagerChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    selectedManagerId = target.value;
    fetchDashboardData(selectedManagerId);
  }

  onMount(async () => {
    // Check if user is admin/HR
    authStore.subscribe(value => {
      isAdminOrHR = value.user?.role === 'admin' || value.user?.role === 'hr_manager';
    });

    if (isAdminOrHR) {
      await loadManagers();
    }
    
    await fetchDashboardData();
  });
</script>

<div class="manager-dashboard">
  <div class="dashboard-header">
    <h1>Manager Dashboard</h1>
    {#if viewingManagerName && isAdminOrHR}
      <p class="viewing-as">Viewing: <strong>{viewingManagerName}'s Team</strong></p>
    {/if}
  </div>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  <!-- Manager Selector for Admin/HR -->
  {#if isAdminOrHR && managers.length > 0}
    <div class="manager-selector">
      <div class="selector-card">
        <div class="selector-header">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="9" cy="7" r="4"></circle>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
          </svg>
          <div class="selector-info">
            <h3>Select Manager</h3>
            <p>View dashboard for any manager</p>
          </div>
        </div>
        
        <select 
          bind:value={selectedManagerId}
          onchange={handleManagerChange}
          class="manager-select"
        >
          <option value="">Select a manager...</option>
          {#each managers as manager}
            <option value={manager.id}>
              {manager.name} - {manager.position}
            </option>
          {/each}
        </select>
      </div>
    </div>
  {/if}

  {#if loading}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading dashboard...</p>
    </div>
  {:else if !selectedManagerId && isAdminOrHR}
    <div class="empty-state">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
        <circle cx="9" cy="7" r="4"></circle>
        <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
        <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
      </svg>
      <h3>Select a Manager</h3>
      <p>Choose a manager from the dropdown above to view their team dashboard</p>
    </div>
  {:else}
    <!-- Stats Cards -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon team">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="9" cy="7" r="4"></circle>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
          </svg>
        </div>
        <div class="stat-content">
          <p class="stat-label">Team Members</p>
          <p class="stat-value">{stats.teamMembers}</p>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon pto">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
            <line x1="16" y1="2" x2="16" y2="6"></line>
            <line x1="8" y1="2" x2="8" y2="6"></line>
            <line x1="3" y1="10" x2="21" y2="10"></line>
          </svg>
        </div>
        <div class="stat-content">
          <p class="stat-label">Pending PTO</p>
          <p class="stat-value">{stats.pendingPTO}</p>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon reviews">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
            <polyline points="14 2 14 8 20 8"></polyline>
            <line x1="16" y1="13" x2="8" y2="13"></line>
            <line x1="16" y1="17" x2="8" y2="17"></line>
          </svg>
        </div>
        <div class="stat-content">
          <p class="stat-label">Upcoming Reviews</p>
          <p class="stat-value">{stats.upcomingReviews}</p>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon tasks">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="9 11 12 14 22 4"></polyline>
            <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"></path>
          </svg>
        </div>
        <div class="stat-content">
          <p class="stat-label">Open Tasks</p>
          <p class="stat-value">{stats.openTasks}</p>
        </div>
      </div>
    </div>

    <!-- Team Members Section -->
    <div class="section">
      <div class="section-header">
        <h2>Team Members</h2>
      </div>

      {#if teamMembers.length === 0}
        <div class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="9" cy="7" r="4"></circle>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
          </svg>
          <h3>No Team Members</h3>
          <p>This manager currently has no direct reports</p>
        </div>
      {:else}
        <div class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Email</th>
                <th>Position</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {#each teamMembers as member}
                <tr>
                  <td>
                    <div class="member-name">
                      <div class="avatar">
                        {member.name.split(' ').map(n => n[0]).join('')}
                      </div>
                      {member.name}
                    </div>
                  </td>
                  <td>{member.email}</td>
                  <td>{member.position}</td>
                  <td>
                    <span class="status-badge status-{member.status}">
                      {member.status}
                    </span>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>

    <!-- Quick Actions -->
    <div class="section">
      <div class="section-header">
        <h2>Quick Actions</h2>
      </div>

      <div class="actions-grid">
        <button class="action-card" onclick={() => navigate('pto')}>
          <div class="action-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
              <line x1="16" y1="2" x2="16" y2="6"></line>
              <line x1="8" y1="2" x2="8" y2="6"></line>
              <line x1="3" y1="10" x2="21" y2="10"></line>
            </svg>
          </div>
          <div class="action-content">
            <h3>Review PTO Requests</h3>
            <p>Approve or deny team time off</p>
          </div>
        </button>

        <button class="action-card" onclick={() => navigate('timesheet')}>
          <div class="action-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <polyline points="12 6 12 12 16 14"></polyline>
            </svg>
          </div>
          <div class="action-content">
            <h3>View Timesheets</h3>
            <p>Review team hours and attendance</p>
          </div>
        </button>

        <button class="action-card" onclick={() => navigate('employees')}>
          <div class="action-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
          </div>
          <div class="action-content">
            <h3>Manage Team</h3>
            <p>View and update employee info</p>
          </div>
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .manager-dashboard {
    padding: 2rem;
    max-width: 1400px;
    margin: 0 auto;
  }

  .dashboard-header {
    margin-bottom: 2rem;
  }

  .dashboard-header h1 {
    margin: 0 0 0.5rem 0;
    font-size: 2rem;
    color: #1f2937;
  }

  .viewing-as {
    margin: 0;
    color: #6b7280;
    font-size: 0.95rem;
  }

  .viewing-as strong {
    color: #3b82f6;
    font-weight: 600;
  }

  /* Manager Selector */
  .manager-selector {
    margin-bottom: 2rem;
  }

  .selector-card {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 12px;
    padding: 1.5rem;
    color: white;
  }

  .selector-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .selector-header svg {
    width: 48px;
    height: 48px;
    opacity: 0.9;
  }

  .selector-info h3 {
    margin: 0 0 0.25rem 0;
    font-size: 1.25rem;
    font-weight: 600;
  }

  .selector-info p {
    margin: 0;
    font-size: 0.875rem;
    opacity: 0.9;
  }

  .manager-select {
    width: 100%;
    padding: 0.75rem 1rem;
    font-size: 1rem;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.95);
    color: #1f2937;
    cursor: pointer;
    transition: all 0.2s;
  }

  .manager-select:hover {
    background: white;
    border-color: rgba(255, 255, 255, 0.5);
  }

  .manager-select:focus {
    outline: none;
    border-color: white;
    box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.2);
  }

  .error-banner {
    background: #fee2e2;
    border: 1px solid #fecaca;
    color: #991b1b;
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 1.5rem;
  }

  .loading {
    text-align: center;
    padding: 4rem 2rem;
  }

  .spinner {
    width: 48px;
    height: 48px;
    border: 4px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 1rem;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .empty-state {
    text-align: center;
    padding: 4rem 2rem;
    color: #6b7280;
  }

  .empty-state svg {
    width: 64px;
    height: 64px;
    margin: 0 auto 1rem;
    opacity: 0.5;
  }

  .empty-state h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1.25rem;
    color: #374151;
  }

  .empty-state p {
    margin: 0;
    font-size: 0.95rem;
  }

  /* Stats Grid */
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .stat-card {
    background: white;
    border-radius: 12px;
    padding: 1.5rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .stat-icon {
    width: 56px;
    height: 56px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .stat-icon svg {
    width: 28px;
    height: 28px;
    color: white;
  }

  .stat-icon.team { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
  .stat-icon.pto { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
  .stat-icon.reviews { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
  .stat-icon.tasks { background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); }

  .stat-content {
    flex: 1;
  }

  .stat-label {
    margin: 0 0 0.25rem 0;
    font-size: 0.875rem;
    color: #6b7280;
    font-weight: 500;
  }

  .stat-value {
    margin: 0;
    font-size: 2rem;
    font-weight: 700;
    color: #1f2937;
  }

  /* Sections */
  .section {
    background: white;
    border-radius: 12px;
    padding: 1.5rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    margin-bottom: 2rem;
  }

  .section-header {
    margin-bottom: 1.5rem;
  }

  .section-header h2 {
    margin: 0;
    font-size: 1.5rem;
    color: #1f2937;
  }

  /* Table */
  .table-container {
    overflow-x: auto;
  }

  .data-table {
    width: 100%;
    border-collapse: collapse;
  }

  .data-table th {
    text-align: left;
    padding: 0.75rem 1rem;
    background: #f9fafb;
    color: #6b7280;
    font-weight: 600;
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .data-table td {
    padding: 1rem;
    border-top: 1px solid #e5e7eb;
  }

  .data-table tr:hover {
    background: #f9fafb;
  }

  .member-name {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    font-weight: 500;
    color: #1f2937;
  }

  .avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .status-badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .status-active {
    background: #d1fae5;
    color: #065f46;
  }

  .status-onboarding {
    background: #dbeafe;
    color: #1e40af;
  }

  .status-inactive {
    background: #f3f4f6;
    color: #6b7280;
  }

  /* Actions Grid */
  .actions-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1rem;
  }

  .action-card {
    background: white;
    border: 2px solid #e5e7eb;
    border-radius: 12px;
    padding: 1.5rem;
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .action-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.1);
    transform: translateY(-2px);
  }

  .action-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .action-icon svg {
    width: 24px;
    height: 24px;
    color: white;
  }

  .action-content {
    text-align: left;
  }

  .action-content h3 {
    margin: 0 0 0.25rem 0;
    font-size: 1rem;
    color: #1f2937;
  }

  .action-content p {
    margin: 0;
    font-size: 0.875rem;
    color: #6b7280;
  }
</style>