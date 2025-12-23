<script lang="ts">
  import { onMount } from 'svelte';

  interface DashboardStats {
    totalEmployees: number;
    activeEmployees: number;
    pendingOnboarding: number;
    pendingPTO: number;
    pendingTimesheets: number;
    organizationCount: number;
    recentHires: number;
  }

  interface RecentActivity {
    id: string;
    type: string;
    description: string;
    timestamp: string;
    status: string;
  }

  let stats = $state<DashboardStats>({
    totalEmployees: 0,
    activeEmployees: 0,
    pendingOnboarding: 0,
    pendingPTO: 0,
    pendingTimesheets: 0,
    organizationCount: 0,
    recentHires: 0
  });

  let recentActivities = $state<RecentActivity[]>([]);
  let loading = $state(true);
  let error = $state('');

  onMount(async () => {
    await loadDashboardData();
  });

  async function loadDashboardData() {
    try {
      loading = true;
      error = '';

      const token = localStorage.getItem('token');
      if (!token) {
        console.error('No auth token - redirecting to login');
        window.location.href = '/'; // or your login route
        return;
      }
      
      const headers = { 
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      };

      // Fetch data with proper error handling for each endpoint
      let employees: any[] = [];
      let organizations: any[] = [];
      let ptoRequests: any[] = [];
      let timesheets: any[] = [];
      let onboarding: any[] = [];

      // Fetch employees (usually exists)
      try {
        const employeesRes = await fetch('/api/employees', { headers });
        if (employeesRes.ok) {
          employees = await employeesRes.json();
        }
      } catch (err) {
        console.warn('Failed to load employees:', err);
      }

      // Fetch organizations (may not exist)
      try {
        const organizationsRes = await fetch('/api/organizations', { headers });
        if (organizationsRes.ok) {
          const data = await organizationsRes.json();
          organizations = Array.isArray(data) ? data : [];
        }
      } catch (err) {
        console.warn('Organizations endpoint not available:', err);
      }

      // Fetch PTO requests (may not exist)
      try {
        const ptoRes = await fetch('/api/pto/requests?status=pending', { headers });
        if (ptoRes.ok) {
          const data = await ptoRes.json();
          ptoRequests = Array.isArray(data) ? data : [];
        }
      } catch (err) {
        console.warn('PTO endpoint not available:', err);
      }

      // Fetch timesheets (may not exist)
      try {
        const timesheetsRes = await fetch('/api/timesheets?status=pending', { headers });
        if (timesheetsRes.ok) {
          const data = await timesheetsRes.json();
          timesheets = Array.isArray(data) ? data : [];
        }
      } catch (err) {
        console.warn('Timesheets endpoint not available:', err);
      }

      // Fetch onboarding (may not exist)
      try {
        const onboardingRes = await fetch('/api/onboarding?status=pending', { headers });
        if (onboardingRes.ok) {
          const data = await onboardingRes.json();
          onboarding = Array.isArray(data) ? data : [];
        }
      } catch (err) {
        console.warn('Onboarding endpoint not available:', err);
      }

      // Calculate stats
      stats = {
        totalEmployees: employees.length || 0,
        activeEmployees: employees.filter((e: any) => e.employment_status === 'active').length || 0,
        pendingOnboarding: onboarding.length || 0,
        pendingPTO: ptoRequests.length || 0,
        pendingTimesheets: timesheets.length || 0,
        organizationCount: organizations.length || 0,
        recentHires: employees.filter((e: any) => {
          const hireDate = new Date(e.hire_date);
          const thirtyDaysAgo = new Date();
          thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
          return hireDate > thirtyDaysAgo;
        }).length || 0
      };

      // Build recent activities
      const activities: RecentActivity[] = [];
      
      ptoRequests.slice(0, 5).forEach((req: any) => {
        activities.push({
          id: req.id,
          type: 'PTO Request',
          description: `${req.employee_name} requested ${req.days} days PTO`,
          timestamp: req.created_at,
          status: 'pending'
        });
      });

      timesheets.slice(0, 5).forEach((ts: any) => {
        activities.push({
          id: ts.id,
          type: 'Timesheet',
          description: `${ts.employee_name} submitted timesheet`,
          timestamp: ts.created_at,
          status: 'pending'
        });
      });

      // Sort by timestamp
      recentActivities = activities.sort((a, b) => 
        new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
      ).slice(0, 10);

    } catch (err) {
      error = 'Failed to load dashboard data';
      console.error(err);
    } finally {
      loading = false;
    }
  }

  function formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const hours = Math.floor(diff / (1000 * 60 * 60));
    
    if (hours < 1) return 'Just now';
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    if (days < 7) return `${days}d ago`;
    return date.toLocaleDateString();
  }

  interface Props {
    navigate: (page: string) => void;
  }

  let { navigate }: Props = $props();
</script>

<div class="hr-dashboard">
  <div class="dashboard-header">
    <h1>HR Manager Dashboard</h1>
    <p class="subtitle">Overview of your organization</p>
  </div>

  {#if loading}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading dashboard...</p>
    </div>
  {:else if error}
    <div class="error-banner">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"></circle>
        <line x1="12" y1="8" x2="12" y2="12"></line>
        <line x1="12" y1="16" x2="12.01" y2="16"></line>
      </svg>
      <span>{error}</span>
    </div>
  {:else}
    <!-- Stats Grid -->
    <div class="stats-grid">
      <div class="stat-card" onclick={() => navigate('employees')}>
        <div class="stat-icon employees">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="9" cy="7" r="4"></circle>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">Total Employees</div>
          <div class="stat-value">{stats.totalEmployees}</div>
          <div class="stat-detail">{stats.activeEmployees} active</div>
        </div>
      </div>

      <div class="stat-card" onclick={() => navigate('organizations')}>
        <div class="stat-icon organizations">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
            <polyline points="9 22 9 12 15 12 15 22"></polyline>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">Organizations</div>
          <div class="stat-value">{stats.organizationCount}</div>
          <div class="stat-detail">Active departments</div>
        </div>
      </div>

      <div class="stat-card" onclick={() => navigate('onboarding')}>
        <div class="stat-icon onboarding">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
            <polyline points="14 2 14 8 20 8"></polyline>
            <line x1="12" y1="18" x2="12" y2="12"></line>
            <line x1="9" y1="15" x2="15" y2="15"></line>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">Pending Onboarding</div>
          <div class="stat-value">{stats.pendingOnboarding}</div>
          <div class="stat-detail">{stats.recentHires} recent hires</div>
        </div>
      </div>

      <div class="stat-card" onclick={() => navigate('pto')}>
        <div class="stat-icon pto">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
            <line x1="16" y1="2" x2="16" y2="6"></line>
            <line x1="8" y1="2" x2="8" y2="6"></line>
            <line x1="3" y1="10" x2="21" y2="10"></line>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">PTO Requests</div>
          <div class="stat-value">{stats.pendingPTO}</div>
          <div class="stat-detail">Awaiting approval</div>
        </div>
      </div>

      <div class="stat-card" onclick={() => navigate('timesheet')}>
        <div class="stat-icon timesheets">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"></circle>
            <polyline points="12 6 12 12 16 14"></polyline>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">Timesheets</div>
          <div class="stat-value">{stats.pendingTimesheets}</div>
          <div class="stat-detail">Pending review</div>
        </div>
      </div>

      <div class="stat-card" onclick={() => navigate('payroll')}>
        <div class="stat-icon payroll">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="1" x2="12" y2="23"></line>
            <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">Payroll</div>
          <div class="stat-value">Ready</div>
          <div class="stat-detail">Next run scheduled</div>
        </div>
      </div>
      <div class="stat-card" onclick={() => navigate('recruiting')}>
        <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="8.5" cy="7" r="4"></circle>
            <line x1="20" y1="8" x2="20" y2="14"></line>
            <line x1="23" y1="11" x2="17" y2="11"></line>
          </svg>
        </div>
        <div class="stat-info">
          <div class="stat-value">Recruiting</div>
          <div class="stat-label">Manage Candidates</div>
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
      <div class="action-grid">
        <button class="action-btn" onclick={() => navigate('employees')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="8.5" cy="7" r="4"></circle>
            <line x1="20" y1="8" x2="20" y2="14"></line>
            <line x1="23" y1="11" x2="17" y2="11"></line>
          </svg>
          <span>Add Employee</span>
        </button>

        <button class="action-btn" onclick={() => navigate('organizations')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
            <polyline points="9 22 9 12 15 12 15 22"></polyline>
          </svg>
          <span>Create Organization</span>
        </button>

        <button class="action-btn" onclick={() => navigate('users')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
            <circle cx="12" cy="7" r="4"></circle>
          </svg>
          <span>Create User</span>
        </button>

        <button class="action-btn" onclick={() => navigate('payroll')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="1" x2="12" y2="23"></line>
            <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
          </svg>
          <span>Run Payroll</span>
        </button>

        <button class="action-btn" onclick={() => navigate('recruiting')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="8.5" cy="7" r="4"></circle>
            <line x1="20" y1="8" x2="20" y2="14"></line>
            <line x1="23" y1="11" x2="17" y2="11"></line>
          </svg>
          <span>Recruiting</span>
        </button>
      </div>
    </div>

    <!-- Recent Activity -->
    <div class="recent-activity">
      <h2>Recent Activity</h2>
      {#if recentActivities.length === 0}
        <div class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="8" x2="12" y2="12"></line>
            <line x1="12" y1="16" x2="12.01" y2="16"></line>
          </svg>
          <p>No recent activity</p>
        </div>
      {:else}
        <div class="activity-list">
          {#each recentActivities as activity}
            <div class="activity-item">
              <div class="activity-icon" class:pto={activity.type === 'PTO Request'} class:timesheet={activity.type === 'Timesheet'}>
                {#if activity.type === 'PTO Request'}
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                    <line x1="16" y1="2" x2="16" y2="6"></line>
                    <line x1="8" y1="2" x2="8" y2="6"></line>
                    <line x1="3" y1="10" x2="21" y2="10"></line>
                  </svg>
                {:else}
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"></circle>
                    <polyline points="12 6 12 12 16 14"></polyline>
                  </svg>
                {/if}
              </div>
              <div class="activity-content">
                <div class="activity-type">{activity.type}</div>
                <div class="activity-description">{activity.description}</div>
                <div class="activity-time">{formatDate(activity.timestamp)}</div>
              </div>
              <div class="activity-status">
                <span class="status-badge pending">Pending</span>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .hr-dashboard {
    padding: 2rem;
    max-width: 1400px;
    margin: 0 auto;
  }

  .dashboard-header {
    margin-bottom: 2rem;
  }

  .dashboard-header h1 {
    font-size: 2rem;
    font-weight: 600;
    color: #1a1a1a;
    margin: 0 0 0.5rem 0;
  }

  .subtitle {
    color: #666;
    font-size: 0.95rem;
    margin: 0;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem;
    gap: 1rem;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid #f0f0f0;
    border-top-color: #4f46e5;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .error-banner {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem;
    background: #fee;
    border: 1px solid #fcc;
    border-radius: 8px;
    color: #c00;
  }

  .error-banner svg {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .stat-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 1.5rem;
    display: flex;
    gap: 1rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .stat-card:hover {
    border-color: #4f46e5;
    box-shadow: 0 4px 6px -1px rgba(79, 70, 229, 0.1);
    transform: translateY(-2px);
  }

  .stat-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .stat-icon svg {
    width: 24px;
    height: 24px;
    stroke-width: 2;
  }

  .stat-icon.employees {
    background: #ede9fe;
    color: #7c3aed;
  }

  .stat-icon.organizations {
    background: #dbeafe;
    color: #2563eb;
  }

  .stat-icon.onboarding {
    background: #d1fae5;
    color: #059669;
  }

  .stat-icon.pto {
    background: #fed7aa;
    color: #d97706;
  }

  .stat-icon.timesheets {
    background: #fce7f3;
    color: #db2777;
  }

  .stat-icon.payroll {
    background: #ccfbf1;
    color: #0d9488;
  }

  .stat-content {
    flex: 1;
  }

  .stat-label {
    font-size: 0.875rem;
    color: #6b7280;
    margin-bottom: 0.25rem;
  }

  .stat-value {
    font-size: 1.875rem;
    font-weight: 600;
    color: #1f2937;
    line-height: 1.2;
  }

  .stat-detail {
    font-size: 0.813rem;
    color: #9ca3af;
    margin-top: 0.25rem;
  }

  .quick-actions {
    margin-bottom: 2rem;
  }

  .quick-actions h2 {
    font-size: 1.25rem;
    font-weight: 600;
    margin-bottom: 1rem;
    color: #1f2937;
  }

  .action-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }

  .action-btn {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem 1.25rem;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    cursor: pointer;
    font-size: 0.938rem;
    font-weight: 500;
    color: #374151;
    transition: all 0.2s;
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

  .recent-activity h2 {
    font-size: 1.25rem;
    font-weight: 600;
    margin-bottom: 1rem;
    color: #1f2937;
  }

  .activity-list {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    overflow: hidden;
  }

  .activity-item {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem 1.25rem;
    border-bottom: 1px solid #f3f4f6;
  }

  .activity-item:last-child {
    border-bottom: none;
  }

  .activity-icon {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .activity-icon.pto {
    background: #fef3c7;
    color: #d97706;
  }

  .activity-icon.timesheet {
    background: #fce7f3;
    color: #db2777;
  }

  .activity-icon svg {
    width: 20px;
    height: 20px;
  }

  .activity-content {
    flex: 1;
  }

  .activity-type {
    font-size: 0.813rem;
    font-weight: 600;
    color: #4f46e5;
    margin-bottom: 0.25rem;
  }

  .activity-description {
    font-size: 0.938rem;
    color: #374151;
    margin-bottom: 0.25rem;
  }

  .activity-time {
    font-size: 0.813rem;
    color: #9ca3af;
  }

  .activity-status {
    flex-shrink: 0;
  }

  .status-badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 500;
  }

  .status-badge.pending {
    background: #fef3c7;
    color: #92400e;
  }

  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #9ca3af;
  }

  .empty-state svg {
    width: 48px;
    height: 48px;
    margin: 0 auto 1rem;
    opacity: 0.5;
  }

  .empty-state p {
    margin: 0;
    font-size: 0.938rem;
  }
</style>