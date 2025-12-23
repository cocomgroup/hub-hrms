<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';

  type Props = {
    navigate: (page: string) => void;
  };
  
  let { navigate }: Props = $props();

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

  interface PTORequest {
    id: string;
    employeeName: string;
    type: string;
    startDate: string;
    endDate: string;
    days: number;
    status: string;
  }

  let stats = $state<TeamStats>({
    teamMembers: 0,
    pendingPTO: 0,
    upcomingReviews: 0,
    openTasks: 0
  });

  let teamMembers = $state<TeamMember[]>([]);
  let ptoRequests = $state<PTORequest[]>([]);
  let loading = $state(true);
  let error = $state('');
  let currentUser = $state<any>(null);

  async function fetchDashboardData() {
    try {
      loading = true;
      error = '';

      // Get current user
      authStore.subscribe(value => {
        currentUser = value.user;
      });

      // Fetch all employees
      const employeesRes = await fetch('http://localhost:8080/api/employees', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (employeesRes.ok) {
        const allEmployees = await employeesRes.json();
        
        // Filter team members (employees managed by current user)
        if (currentUser?.employee_id) {
          teamMembers = allEmployees.filter((e: any) => 
            e.manager_id === currentUser.employee_id
          ).map((e: any) => ({
            id: e.id,
            name: `${e.first_name} ${e.last_name}`,
            email: e.email,
            position: e.position,
            status: e.status
          }));
        }

        stats.teamMembers = teamMembers.length;
      }

      // TODO: Fetch PTO requests when PTO API is ready
      stats.pendingPTO = 0;
      stats.upcomingReviews = 0;
      stats.openTasks = 0;

    } catch (err) {
      error = 'Failed to load dashboard data';
      console.error('Dashboard error:', err);
    } finally {
      loading = false;
    }
  }

  async function handlePTOAction(requestId: string, action: 'approve' | 'deny') {
    try {
      // TODO: Implement PTO approval when API is ready
      console.log(`${action} PTO request ${requestId}`);
    } catch (err) {
      console.error('PTO action error:', err);
    }
  }

  onMount(() => {
    fetchDashboardData();
  });
</script>

<div class="manager-dashboard">
  <div class="dashboard-header">
    <div>
      <h1>Manager Dashboard</h1>
      <p class="subtitle">Manage your team and responsibilities</p>
    </div>
  </div>

  {#if loading}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading dashboard...</p>
    </div>
  {:else if error}
    <div class="error-state">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"></circle>
        <line x1="12" y1="8" x2="12" y2="12"></line>
        <line x1="12" y1="16" x2="12.01" y2="16"></line>
      </svg>
      <p>{error}</p>
      <button onclick={fetchDashboardData}>Retry</button>
    </div>
  {:else}
    <!-- Stats Grid -->
    <div class="stats-grid">
      <!-- Team Members Card -->
      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="9" cy="7" r="4"></circle>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
          </svg>
        </div>
        <div class="stat-info">
          <div class="stat-value">{stats.teamMembers}</div>
          <div class="stat-label">Team Members</div>
        </div>
      </div>

      <!-- Pending PTO Card -->
      <div class="stat-card" onclick={() => navigate('pto')}>
        <div class="stat-icon" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
            <line x1="16" y1="2" x2="16" y2="6"></line>
            <line x1="8" y1="2" x2="8" y2="6"></line>
            <line x1="3" y1="10" x2="21" y2="10"></line>
          </svg>
        </div>
        <div class="stat-info">
          <div class="stat-value">{stats.pendingPTO}</div>
          <div class="stat-label">Pending PTO Requests</div>
        </div>
        <div class="stat-arrow">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="5" y1="12" x2="19" y2="12"></line>
            <polyline points="12 5 19 12 12 19"></polyline>
          </svg>
        </div>
      </div>

      <!-- Upcoming Reviews Card -->
      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 11l3 3L22 4"></path>
            <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"></path>
          </svg>
        </div>
        <div class="stat-info">
          <div class="stat-value">{stats.upcomingReviews}</div>
          <div class="stat-label">Upcoming Reviews</div>
        </div>
      </div>

      <!-- Open Tasks Card -->
      <div class="stat-card" onclick={() => navigate('timesheet')}>
        <div class="stat-icon" style="background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"></circle>
            <polyline points="12 6 12 12 16 14"></polyline>
          </svg>
        </div>
        <div class="stat-info">
          <div class="stat-value">{stats.openTasks}</div>
          <div class="stat-label">Open Tasks</div>
        </div>
        <div class="stat-arrow">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="5" y1="12" x2="19" y2="12"></line>
            <polyline points="12 5 19 12 12 19"></polyline>
          </svg>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="quick-actions">
      <h2>Quick Actions</h2>
      <div class="actions-grid">
        <button class="action-btn" onclick={() => navigate('pto')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
            <line x1="16" y1="2" x2="16" y2="6"></line>
            <line x1="8" y1="2" x2="8" y2="6"></line>
            <line x1="3" y1="10" x2="21" y2="10"></line>
          </svg>
          <span>Review PTO</span>
        </button>

        <button class="action-btn" onclick={() => navigate('timesheet')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"></circle>
            <polyline points="12 6 12 12 16 14"></polyline>
          </svg>
          <span>Approve Timesheets</span>
        </button>

        <button class="action-btn" onclick={() => navigate('employees')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="9" cy="7" r="4"></circle>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
          </svg>
          <span>View Team</span>
        </button>

        <button class="action-btn" onclick={() => navigate('onboarding')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="8.5" cy="7" r="4"></circle>
            <line x1="20" y1="8" x2="20" y2="14"></line>
            <line x1="23" y1="11" x2="17" y2="11"></line>
          </svg>
          <span>Onboarding</span>
        </button>
      </div>
    </div>

    <div class="dashboard-content">
      <!-- Team Members Section -->
      <div class="section team-section">
        <div class="section-header">
          <h2>My Team</h2>
          <button class="link-btn" onclick={() => navigate('employees')}>View All</button>
        </div>
        
        {#if teamMembers.length === 0}
          <div class="empty-state">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
            </svg>
            <p>No team members found</p>
            <span class="empty-hint">Team members you manage will appear here</span>
          </div>
        {:else}
          <div class="team-grid">
            {#each teamMembers.slice(0, 6) as member}
              <div class="team-card">
                <div class="team-member-avatar">
                  {member.name.split(' ').map(n => n[0]).join('')}
                </div>
                <div class="team-member-info">
                  <h3>{member.name}</h3>
                  <p class="position">{member.position}</p>
                  <p class="email">{member.email}</p>
                  <span class="status status-{member.status}">{member.status}</span>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Pending PTO Requests Section -->
      <div class="section pto-section">
        <div class="section-header">
          <h2>Pending PTO Requests</h2>
          <button class="link-btn" onclick={() => navigate('pto')}>View All</button>
        </div>
        
        {#if ptoRequests.length === 0}
          <div class="empty-state">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
              <line x1="16" y1="2" x2="16" y2="6"></line>
              <line x1="8" y1="2" x2="8" y2="6"></line>
              <line x1="3" y1="10" x2="21" y2="10"></line>
            </svg>
            <p>No pending PTO requests</p>
            <span class="empty-hint">Time off requests awaiting approval will appear here</span>
          </div>
        {:else}
          <div class="pto-list">
            {#each ptoRequests as request}
              <div class="pto-item">
                <div class="pto-info">
                  <h3>{request.employeeName}</h3>
                  <p class="pto-type">{request.type}</p>
                  <p class="pto-dates">{request.startDate} - {request.endDate} ({request.days} days)</p>
                </div>
                <div class="pto-actions">
                  <button 
                    class="btn-approve" 
                    onclick={() => handlePTOAction(request.id, 'approve')}
                  >
                    Approve
                  </button>
                  <button 
                    class="btn-deny" 
                    onclick={() => handlePTOAction(request.id, 'deny')}
                  >
                    Deny
                  </button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
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

  h1 {
    font-size: 2rem;
    font-weight: 700;
    color: #1f2937;
    margin: 0 0 0.5rem 0;
  }

  .subtitle {
    color: #6b7280;
    margin: 0;
  }

  .loading {
    text-align: center;
    padding: 4rem 2rem;
  }

  .spinner {
    width: 50px;
    height: 50px;
    border: 4px solid #f3f4f6;
    border-top-color: #4f46e5;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 1rem;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .error-state {
    text-align: center;
    padding: 3rem 2rem;
    background: #fef2f2;
    border-radius: 8px;
    border: 1px solid #fecaca;
  }

  .error-state svg {
    width: 48px;
    height: 48px;
    color: #dc2626;
    margin-bottom: 1rem;
  }

  .error-state button {
    margin-top: 1rem;
    padding: 0.5rem 1.5rem;
    background: #4f46e5;
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
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
    cursor: default;
    transition: all 0.2s;
  }

  .stat-card:has(.stat-arrow) {
    cursor: pointer;
  }

  .stat-card:has(.stat-arrow):hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }

  .stat-icon {
    width: 56px;
    height: 56px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .stat-icon svg {
    width: 28px;
    height: 28px;
    color: white;
  }

  .stat-info {
    flex: 1;
  }

  .stat-value {
    font-size: 1.75rem;
    font-weight: 700;
    color: #1f2937;
    margin-bottom: 0.25rem;
  }

  .stat-label {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .stat-arrow {
    width: 24px;
    height: 24px;
    color: #d1d5db;
  }

  .stat-arrow svg {
    width: 100%;
    height: 100%;
  }

  .quick-actions {
    background: white;
    border-radius: 12px;
    padding: 1.5rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    margin-bottom: 2rem;
  }

  .quick-actions h2 {
    font-size: 1.25rem;
    font-weight: 600;
    margin-bottom: 1rem;
    color: #1f2937;
  }

  .actions-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }

  .action-btn {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
  }

  .action-btn:hover {
    background: #4f46e5;
    border-color: #4f46e5;
    color: white;
  }

  .action-btn svg {
    width: 20px;
    height: 20px;
  }

  .dashboard-content {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem;
  }

  @media (max-width: 1024px) {
    .dashboard-content {
      grid-template-columns: 1fr;
    }
  }

  .section {
    background: white;
    border-radius: 12px;
    padding: 1.5rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
  }

  .section-header h2 {
    font-size: 1.125rem;
    font-weight: 600;
    color: #1f2937;
    margin: 0;
  }

  .link-btn {
    background: none;
    border: none;
    color: #4f46e5;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    padding: 0;
  }

  .link-btn:hover {
    text-decoration: underline;
  }

  .empty-state {
    text-align: center;
    padding: 3rem 2rem;
    color: #9ca3af;
  }

  .empty-state svg {
    width: 48px;
    height: 48px;
    margin-bottom: 1rem;
  }

  .empty-state p {
    font-size: 0.875rem;
    font-weight: 500;
    color: #6b7280;
    margin: 0 0 0.5rem 0;
  }

  .empty-hint {
    font-size: 0.75rem;
    color: #9ca3af;
  }

  .team-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 1rem;
  }

  .team-card {
    padding: 1rem;
    background: #f9fafb;
    border-radius: 8px;
    border: 1px solid #e5e7eb;
  }

  .team-member-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 0.875rem;
    margin-bottom: 0.75rem;
  }

  .team-member-info h3 {
    font-size: 0.875rem;
    font-weight: 600;
    color: #1f2937;
    margin: 0 0 0.25rem 0;
  }

  .team-member-info .position {
    font-size: 0.75rem;
    color: #6b7280;
    margin: 0 0 0.25rem 0;
  }

  .team-member-info .email {
    font-size: 0.75rem;
    color: #9ca3af;
    margin: 0 0 0.5rem 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .status {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    border-radius: 9999px;
    font-size: 0.625rem;
    font-weight: 500;
    text-transform: uppercase;
  }

  .status-active {
    background: #d1fae5;
    color: #065f46;
  }

  .status-inactive {
    background: #fee2e2;
    color: #991b1b;
  }

  .pto-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .pto-item {
    padding: 1rem;
    background: #f9fafb;
    border-radius: 8px;
    border: 1px solid #e5e7eb;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
  }

  .pto-info h3 {
    font-size: 0.875rem;
    font-weight: 600;
    color: #1f2937;
    margin: 0 0 0.25rem 0;
  }

  .pto-type {
    font-size: 0.75rem;
    color: #4f46e5;
    font-weight: 500;
    margin: 0 0 0.25rem 0;
    text-transform: capitalize;
  }

  .pto-dates {
    font-size: 0.75rem;
    color: #6b7280;
    margin: 0;
  }

  .pto-actions {
    display: flex;
    gap: 0.5rem;
  }

  .btn-approve,
  .btn-deny {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-approve {
    background: #d1fae5;
    color: #065f46;
  }

  .btn-approve:hover {
    background: #a7f3d0;
  }

  .btn-deny {
    background: #fee2e2;
    color: #991b1b;
  }

  .btn-deny:hover {
    background: #fecaca;
  }
</style>