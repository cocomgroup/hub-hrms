<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { API_BASE_URL } from '../config';
  
  interface QuickAction {
    id: string;
    title: string;
    description: string;
    icon: string;
    action: () => void;
    color: string;
  }

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

  interface TimeEntry {
    id: string;
    date: string;
    clock_in: string;
    clock_out: string;
    total_hours: number;
    status: string;
  }

  interface PTOBalance {
    vacation_days: number;
    sick_days: number;
    personal_days: number;
  }

  interface PayStub {
    id: string;
    pay_period: string;
    pay_date: string;
    gross_pay: number;
    net_pay: number;
    status: string;
  }

  let loading = false;
  let activeTab = 'overview';
  let employee = $authStore.employee;
  
  // Data
  let tasks: Task[] = [];
  let recentTimeEntries: TimeEntry[] = [];
  let ptoBalance: PTOBalance | null = null;
  let recentPayStubs: PayStub[] = [];
  let upcomingPTO: any[] = [];

  // Stats
  let pendingTasksCount = 0;
  let hoursThisWeek = 0;
  let nextPayDate = '';

  onMount(() => {
    loadDashboardData();
  });

  async function loadDashboardData() {
    loading = true;
    try {
      await Promise.all([
        loadTasks(),
        loadRecentTimeEntries(),
        loadPTOBalance(),
        loadRecentPayStubs()
      ]);
    } catch (err) {
      console.error('Error loading dashboard:', err);
    } finally {
      loading = false;
    }
  }

  async function loadTasks() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/workflows/my-tasks`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        tasks = await response.json();
        pendingTasksCount = tasks.filter(t => t.status === 'pending').length;
      }
    } catch (err) {
      console.error('Error loading tasks:', err);
    }
  }

  async function loadRecentTimeEntries() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/timesheet/entries`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        const entries = await response.json();
        recentTimeEntries = entries.slice(0, 5);
        
        // Calculate hours this week
        const weekStart = getWeekStart();
        hoursThisWeek = entries
          .filter((e: TimeEntry) => new Date(e.date) >= weekStart)
          .reduce((sum: number, e: TimeEntry) => sum + (e.total_hours || 0), 0);
      }
    } catch (err) {
      console.error('Error loading time entries:', err);
    }
  }

  async function loadPTOBalance() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/pto/balance`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        ptoBalance = await response.json();
      }
    } catch (err) {
      console.error('Error loading PTO balance:', err);
    }
  }

  async function loadRecentPayStubs() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/payroll/paystubs`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        const stubs = await response.json();
        recentPayStubs = stubs.slice(0, 3);
        
        // Find next pay date
        if (stubs.length > 0) {
          nextPayDate = stubs[0].pay_date;
        }
      }
    } catch (err) {
      console.error('Error loading pay stubs:', err);
    }
  }

  function getWeekStart(): Date {
    const now = new Date();
    const day = now.getDay();
    const diff = now.getDate() - day + (day === 0 ? -6 : 1);
    return new Date(now.setDate(diff));
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  function formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount);
  }

  // Quick Actions
  const quickActions: QuickAction[] = [
    {
      id: 'timesheet',
      title: 'Submit Timesheet',
      description: 'Log your hours for the week',
      icon: '⏱️',
      action: () => window.location.href = '/timesheet',
      color: 'purple'
    },
    {
      id: 'pto',
      title: 'Request Time Off',
      description: 'Submit a PTO request',
      icon: '🏖️',
      action: () => window.location.href = '/pto',
      color: 'blue'
    },
    {
      id: 'paystub',
      title: 'View Pay Stubs',
      description: 'Access your payment history',
      icon: '💰',
      action: () => activeTab = 'payroll',
      color: 'green'
    },
    {
      id: 'benefits',
      title: 'Manage Benefits',
      description: 'Healthcare and retirement',
      icon: '🏥',
      action: () => activeTab = 'benefits',
      color: 'red'
    },
    {
      id: 'profile',
      title: 'Update Profile',
      description: 'Personal information',
      icon: '👤',
      action: () => activeTab = 'profile',
      color: 'gray'
    },
    {
      id: 'tasks',
      title: 'My Tasks',
      description: `${pendingTasksCount} pending`,
      icon: '✓',
      action: () => activeTab = 'tasks',
      color: 'yellow'
    }
  ];

  function getPriorityColor(priority: string): string {
    switch (priority) {
      case 'high': return 'priority-high';
      case 'medium': return 'priority-medium';
      case 'low': return 'priority-low';
      default: return '';
    }
  }

  function getStatusColor(status: string): string {
    switch (status) {
      case 'completed': return 'status-completed';
      case 'pending': return 'status-pending';
      case 'in_progress': return 'status-progress';
      default: return '';
    }
  }
</script>

<div class="dashboard-container">
  <!-- Header -->
  <div class="dashboard-header">
    <div>
      <h1>Welcome back, {employee?.first_name || 'Employee'}! 👋</h1>
      <p class="subtitle">{employee?.position || 'Position'} • {employee?.department || 'Department'}</p>
    </div>
    <div class="header-stats">
      <div class="stat-card">
        <span class="stat-icon">📋</span>
        <div>
          <div class="stat-value">{pendingTasksCount}</div>
          <div class="stat-label">Pending Tasks</div>
        </div>
      </div>
      <div class="stat-card">
        <span class="stat-icon">⏰</span>
        <div>
          <div class="stat-value">{hoursThisWeek.toFixed(1)}h</div>
          <div class="stat-label">This Week</div>
        </div>
      </div>
    </div>
  </div>

  <!-- Tab Navigation -->
  <div class="tab-nav">
    <button 
      class="tab-btn {activeTab === 'overview' ? 'active' : ''}"
      on:click={() => activeTab = 'overview'}
    >
      Overview
    </button>
    <button 
      class="tab-btn {activeTab === 'tasks' ? 'active' : ''}"
      on:click={() => activeTab = 'tasks'}
    >
      My Tasks {#if pendingTasksCount > 0}<span class="badge">{pendingTasksCount}</span>{/if}
    </button>
    <button 
      class="tab-btn {activeTab === 'payroll' ? 'active' : ''}"
      on:click={() => activeTab = 'payroll'}
    >
      Pay Stubs
    </button>
    <button 
      class="tab-btn {activeTab === 'benefits' ? 'active' : ''}"
      on:click={() => activeTab = 'benefits'}
    >
      Benefits
    </button>
    <button 
      class="tab-btn {activeTab === 'profile' ? 'active' : ''}"
      on:click={() => activeTab = 'profile'}
    >
      Profile
    </button>
  </div>

  <!-- Tab Content -->
  <div class="tab-content">
    {#if activeTab === 'overview'}
      <!-- Quick Actions -->
      <section class="section">
        <h2>Quick Actions</h2>
        <div class="quick-actions-grid">
          {#each quickActions as action}
            <button 
              class="quick-action-card action-{action.color}"
              on:click={action.action}
            >
              <span class="action-icon">{action.icon}</span>
              <div class="action-content">
                <h3>{action.title}</h3>
                <p>{action.description}</p>
              </div>
              <span class="action-arrow">→</span>
            </button>
          {/each}
        </div>
      </section>

      <!-- Two Column Layout -->
      <div class="two-column">
        <!-- Left Column -->
        <div class="column">
          <!-- Pending Tasks -->
          <section class="section card">
            <div class="section-header">
              <h2>Pending Tasks</h2>
              <button class="link-btn" on:click={() => activeTab = 'tasks'}>
                View All →
              </button>
            </div>
            {#if tasks.length === 0}
              <div class="empty-state-small">
                <span class="empty-icon">✓</span>
                <p>No pending tasks</p>
              </div>
            {:else}
              <div class="task-list">
                {#each tasks.slice(0, 3) as task}
                  <div class="task-item">
                    <div class="task-header">
                      <span class="task-title">{task.title}</span>
                      <span class="priority-badge {getPriorityColor(task.priority)}">
                        {task.priority}
                      </span>
                    </div>
                    <p class="task-description">{task.description}</p>
                    <div class="task-footer">
                      <span class="task-due">Due: {formatDate(task.due_date)}</span>
                      <span class="status-badge {getStatusColor(task.status)}">
                        {task.status}
                      </span>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </section>

          <!-- Recent Time Entries -->
          <section class="section card">
            <div class="section-header">
              <h2>Recent Time Entries</h2>
              <button class="link-btn" on:click={() => window.location.href = '/timesheet'}>
                View All →
              </button>
            </div>
            {#if recentTimeEntries.length === 0}
              <div class="empty-state-small">
                <span class="empty-icon">⏱️</span>
                <p>No time entries yet</p>
              </div>
            {:else}
              <div class="time-entries">
                {#each recentTimeEntries as entry}
                  <div class="time-entry-row">
                    <div class="entry-date">{formatDate(entry.date)}</div>
                    <div class="entry-time">
                      {entry.clock_in} - {entry.clock_out}
                    </div>
                    <div class="entry-hours">{entry.total_hours}h</div>
                  </div>
                {/each}
              </div>
            {/if}
          </section>
        </div>

        <!-- Right Column -->
        <div class="column">
          <!-- PTO Balance -->
          {#if ptoBalance}
            <section class="section card">
              <h2>PTO Balance</h2>
              <div class="pto-summary">
                <div class="pto-item">
                  <span class="pto-icon">🏖️</span>
                  <div>
                    <div class="pto-value">{ptoBalance.vacation_days}</div>
                    <div class="pto-label">Vacation Days</div>
                  </div>
                </div>
                <div class="pto-item">
                  <span class="pto-icon">🏥</span>
                  <div>
                    <div class="pto-value">{ptoBalance.sick_days}</div>
                    <div class="pto-label">Sick Days</div>
                  </div>
                </div>
                <div class="pto-item">
                  <span class="pto-icon">⏰</span>
                  <div>
                    <div class="pto-value">{ptoBalance.personal_days}</div>
                    <div class="pto-label">Personal Days</div>
                  </div>
                </div>
              </div>
              <button class="btn btn-primary btn-block" on:click={() => window.location.href = '/pto'}>
                Request Time Off
              </button>
            </section>
          {/if}

          <!-- Upcoming Pay -->
          {#if nextPayDate}
            <section class="section card highlight">
              <div class="pay-preview">
                <span class="pay-icon">💰</span>
                <div class="pay-info">
                  <div class="pay-label">Next Pay Date</div>
                  <div class="pay-date">{formatDate(nextPayDate)}</div>
                </div>
              </div>
              <button class="btn btn-secondary btn-block" on:click={() => activeTab = 'payroll'}>
                View Pay Stubs
              </button>
            </section>
          {/if}

          <!-- Benefits Quick Links -->
          <section class="section card">
            <h2>Benefits Portal</h2>
            <div class="benefits-links">
              <a href="#" class="benefit-link" on:click|preventDefault={() => activeTab = 'benefits'}>
                <span class="benefit-icon">🏥</span>
                <div>
                  <div class="benefit-title">Healthcare</div>
                  <div class="benefit-subtitle">Medical, Dental, Vision</div>
                </div>
                <span class="link-arrow">→</span>
              </a>
              <a href="#" class="benefit-link" on:click|preventDefault={() => activeTab = 'benefits'}>
                <span class="benefit-icon">💼</span>
                <div>
                  <div class="benefit-title">Retirement</div>
                  <div class="benefit-subtitle">401(k) Plan</div>
                </div>
                <span class="link-arrow">→</span>
              </a>
            </div>
          </section>
        </div>
      </div>
    {/if}

    {#if activeTab === 'tasks'}
      <section class="section">
        <h2>My Tasks</h2>
        {#if tasks.length === 0}
          <div class="empty-state">
            <div class="empty-icon">✓</div>
            <h3>All Caught Up!</h3>
            <p>You don't have any pending tasks.</p>
          </div>
        {:else}
          <div class="tasks-grid">
            {#each tasks as task}
              <div class="task-card">
                <div class="task-card-header">
                  <span class="priority-badge {getPriorityColor(task.priority)}">
                    {task.priority}
                  </span>
                  <span class="status-badge {getStatusColor(task.status)}">
                    {task.status}
                  </span>
                </div>
                <h3 class="task-card-title">{task.title}</h3>
                <p class="task-card-description">{task.description}</p>
                <div class="task-card-footer">
                  <span class="task-type">{task.type}</span>
                  <span class="task-due">Due: {formatDate(task.due_date)}</span>
                </div>
                <button class="btn btn-primary btn-sm">Complete Task</button>
              </div>
            {/each}
          </div>
        {/if}
      </section>
    {/if}

    {#if activeTab === 'payroll'}
      <section class="section">
        <h2>Pay Stubs</h2>
        {#if recentPayStubs.length === 0}
          <div class="empty-state">
            <div class="empty-icon">💰</div>
            <h3>No Pay Stubs Available</h3>
            <p>Pay stubs will appear here once payroll is processed.</p>
          </div>
        {:else}
          <div class="paystubs-list">
            {#each recentPayStubs as stub}
              <div class="paystub-card">
                <div class="paystub-header">
                  <div>
                    <h3>Pay Period: {stub.pay_period}</h3>
                    <p class="paystub-date">Pay Date: {formatDate(stub.pay_date)}</p>
                  </div>
                  <span class="status-badge {getStatusColor(stub.status)}">
                    {stub.status}
                  </span>
                </div>
                <div class="paystub-amounts">
                  <div class="amount-item">
                    <span class="amount-label">Gross Pay</span>
                    <span class="amount-value">{formatCurrency(stub.gross_pay)}</span>
                  </div>
                  <div class="amount-item">
                    <span class="amount-label">Net Pay</span>
                    <span class="amount-value net">{formatCurrency(stub.net_pay)}</span>
                  </div>
                </div>
                <button class="btn btn-secondary btn-sm">
                  📄 Download PDF
                </button>
              </div>
            {/each}
          </div>
        {/if}
      </section>
    {/if}

    {#if activeTab === 'benefits'}
      <section class="section">
        <h2>Benefits Overview</h2>
        
        <div class="benefits-grid">
          <!-- Healthcare -->
          <div class="benefit-card">
            <div class="benefit-card-header">
              <span class="benefit-card-icon">🏥</span>
              <h3>Healthcare Benefits</h3>
            </div>
            <p class="benefit-description">
              Medical, dental, and vision insurance plans
            </p>
            <div class="benefit-links-list">
              <a href="https://example.com/healthcare" target="_blank" class="benefit-external-link">
                <span>Login to Healthcare Portal</span>
                <span class="external-icon">↗</span>
              </a>
              <a href="https://example.com/find-doctor" target="_blank" class="benefit-external-link">
                <span>Find a Doctor</span>
                <span class="external-icon">↗</span>
              </a>
              <a href="https://example.com/claims" target="_blank" class="benefit-external-link">
                <span>View Claims</span>
                <span class="external-icon">↗</span>
              </a>
            </div>
          </div>

          <!-- Retirement -->
          <div class="benefit-card">
            <div class="benefit-card-header">
              <span class="benefit-card-icon">💼</span>
              <h3>Retirement Plan</h3>
            </div>
            <p class="benefit-description">
              401(k) with company match up to 6%
            </p>
            <div class="benefit-links-list">
              <a href="https://example.com/401k" target="_blank" class="benefit-external-link">
                <span>Login to 401(k) Portal</span>
                <span class="external-icon">↗</span>
              </a>
              <a href="https://example.com/contributions" target="_blank" class="benefit-external-link">
                <span>Manage Contributions</span>
                <span class="external-icon">↗</span>
              </a>
              <a href="https://example.com/investments" target="_blank" class="benefit-external-link">
                <span>View Investments</span>
                <span class="external-icon">↗</span>
              </a>
            </div>
          </div>

          <!-- Life Insurance -->
          <div class="benefit-card">
            <div class="benefit-card-header">
              <span class="benefit-card-icon">🛡️</span>
              <h3>Life Insurance</h3>
            </div>
            <p class="benefit-description">
              Basic and supplemental life insurance coverage
            </p>
            <div class="benefit-links-list">
              <a href="https://example.com/life-insurance" target="_blank" class="benefit-external-link">
                <span>View Coverage Details</span>
                <span class="external-icon">↗</span>
              </a>
              <a href="https://example.com/beneficiaries" target="_blank" class="benefit-external-link">
                <span>Update Beneficiaries</span>
                <span class="external-icon">↗</span>
              </a>
            </div>
          </div>

          <!-- FSA/HSA -->
          <div class="benefit-card">
            <div class="benefit-card-header">
              <span class="benefit-card-icon">💳</span>
              <h3>FSA / HSA</h3>
            </div>
            <p class="benefit-description">
              Flexible spending and health savings accounts
            </p>
            <div class="benefit-links-list">
              <a href="https://example.com/fsa" target="_blank" class="benefit-external-link">
                <span>Manage FSA/HSA</span>
                <span class="external-icon">↗</span>
              </a>
              <a href="https://example.com/submit-claim" target="_blank" class="benefit-external-link">
                <span>Submit Claim</span>
                <span class="external-icon">↗</span>
              </a>
            </div>
          </div>
        </div>

        <div class="help-section">
          <h3>Need Help with Benefits?</h3>
          <p>Contact HR at <a href="mailto:benefits@company.com">benefits@company.com</a> or call (555) 123-4567</p>
        </div>
      </section>
    {/if}

    {#if activeTab === 'profile'}
      <section class="section">
        <h2>My Profile</h2>
        
        <div class="profile-sections">
          <!-- Personal Information -->
          <div class="profile-section card">
            <div class="section-header">
              <h3>Personal Information</h3>
              <button class="btn btn-secondary btn-sm">Edit</button>
            </div>
            <div class="info-grid">
              <div class="info-item">
                <label>Full Name</label>
                <div class="info-value">{employee?.first_name} {employee?.last_name}</div>
              </div>
              <div class="info-item">
                <label>Email</label>
                <div class="info-value">{employee?.email}</div>
              </div>
              <div class="info-item">
                <label>Phone</label>
                <div class="info-value">{employee?.phone || 'Not provided'}</div>
              </div>
              <div class="info-item">
                <label>Date of Birth</label>
                <div class="info-value">{employee?.date_of_birth ? formatDate(employee.date_of_birth) : 'Not provided'}</div>
              </div>
            </div>
          </div>

          <!-- Employment Information -->
          <div class="profile-section card">
            <h3>Employment Information</h3>
            <div class="info-grid">
              <div class="info-item">
                <label>Position</label>
                <div class="info-value">{employee?.position}</div>
              </div>
              <div class="info-item">
                <label>Department</label>
                <div class="info-value">{employee?.department}</div>
              </div>
              <div class="info-item">
                <label>Employment Type</label>
                <div class="info-value">{employee?.employment_type}</div>
              </div>
              <div class="info-item">
                <label>Hire Date</label>
                <div class="info-value">{employee?.hire_date ? formatDate(employee.hire_date) : 'N/A'}</div>
              </div>
            </div>
          </div>

          <!-- Address -->
          <div class="profile-section card">
            <div class="section-header">
              <h3>Address</h3>
              <button class="btn btn-secondary btn-sm">Update</button>
            </div>
            <div class="info-grid">
              <div class="info-item full-width">
                <label>Street Address</label>
                <div class="info-value">{employee?.street_address || 'Not provided'}</div>
              </div>
              <div class="info-item">
                <label>City</label>
                <div class="info-value">{employee?.city || 'Not provided'}</div>
              </div>
              <div class="info-item">
                <label>State</label>
                <div class="info-value">{employee?.state || 'Not provided'}</div>
              </div>
              <div class="info-item">
                <label>ZIP Code</label>
                <div class="info-value">{employee?.zip_code || 'Not provided'}</div>
              </div>
            </div>
          </div>

          <!-- Emergency Contact -->
          <div class="profile-section card">
            <div class="section-header">
              <h3>Emergency Contact</h3>
              <button class="btn btn-secondary btn-sm">Update</button>
            </div>
            <div class="info-grid">
              <div class="info-item">
                <label>Name</label>
                <div class="info-value">{employee?.emergency_contact_name || 'Not provided'}</div>
              </div>
              <div class="info-item">
                <label>Phone</label>
                <div class="info-value">{employee?.emergency_contact_phone || 'Not provided'}</div>
              </div>
            </div>
          </div>
        </div>
      </section>
    {/if}
  </div>
</div>

<style>
  .dashboard-container {
    padding: 2rem;
    max-width: 1400px;
    margin: 0 auto;
  }

  .dashboard-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 2rem;
    gap: 2rem;
  }

  .dashboard-header h1 {
    font-size: 2rem;
    color: #e4e7eb;
    margin: 0 0 0.5rem 0;
  }

  .subtitle {
    color: #999;
    margin: 0;
  }

  .header-stats {
    display: flex;
    gap: 1rem;
  }

  .stat-card {
    background: #1e2128;
    border-radius: 12px;
    padding: 1rem 1.5rem;
    display: flex;
    align-items: center;
    gap: 1rem;
    border: 1px solid #2d3139;
  }

  .stat-icon {
    font-size: 2rem;
  }

  .stat-value {
    font-size: 1.5rem;
    font-weight: bold;
    color: #667eea;
  }

  .stat-label {
    font-size: 0.875rem;
    color: #999;
  }

  /* Tabs */
  .tab-nav {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 2rem;
    border-bottom: 2px solid #2d3139;
  }

  .tab-btn {
    background: none;
    border: none;
    padding: 1rem 1.5rem;
    color: #999;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    border-bottom: 2px solid transparent;
    margin-bottom: -2px;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .tab-btn:hover {
    color: #e4e7eb;
  }

  .tab-btn.active {
    color: #667eea;
    border-bottom-color: #667eea;
  }

  .badge {
    background: #667eea;
    color: white;
    padding: 0.125rem 0.5rem;
    border-radius: 12px;
    font-size: 0.75rem;
  }

  /* Quick Actions */
  .quick-actions-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1rem;
  }

  .quick-action-card {
    background: #1e2128;
    border: 2px solid #2d3139;
    border-radius: 12px;
    padding: 1.5rem;
    display: flex;
    align-items: center;
    gap: 1rem;
    cursor: pointer;
    transition: all 0.2s;
    text-align: left;
  }

  .quick-action-card:hover {
    transform: translateY(-2px);
    border-color: #667eea;
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.2);
  }

  .action-icon {
    font-size: 2.5rem;
  }

  .action-content {
    flex: 1;
  }

  .action-content h3 {
    margin: 0 0 0.25rem 0;
    color: #e4e7eb;
    font-size: 1rem;
  }

  .action-content p {
    margin: 0;
    color: #999;
    font-size: 0.875rem;
  }

  .action-arrow {
    color: #667eea;
    font-size: 1.5rem;
    opacity: 0;
    transition: opacity 0.2s;
  }

  .quick-action-card:hover .action-arrow {
    opacity: 1;
  }

  /* Section */
  .section {
    margin-bottom: 2rem;
  }

  .section h2 {
    font-size: 1.5rem;
    color: #e4e7eb;
    margin-bottom: 1rem;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .section-header h2,
  .section-header h3 {
    margin: 0;
  }

  .card {
    background: #1e2128;
    border-radius: 12px;
    padding: 1.5rem;
    border: 1px solid #2d3139;
  }

  .card.highlight {
    border-color: #667eea;
    background: linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%);
  }

  /* Two Column */
  .two-column {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
  }

  .column {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  /* Tasks */
  .task-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .task-item {
    background: #2d3139;
    border-radius: 8px;
    padding: 1rem;
  }

  .task-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .task-title {
    font-weight: 500;
    color: #e4e7eb;
  }

  .task-description {
    color: #999;
    font-size: 0.875rem;
    margin: 0.5rem 0;
  }

  .task-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 0.75rem;
    font-size: 0.875rem;
  }

  .task-due {
    color: #999;
  }

  .tasks-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1.5rem;
  }

  .task-card {
    background: #1e2128;
    border: 1px solid #2d3139;
    border-radius: 12px;
    padding: 1.5rem;
  }

  .task-card-header {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .task-card-title {
    color: #e4e7eb;
    margin: 0 0 0.75rem 0;
    font-size: 1.125rem;
  }

  .task-card-description {
    color: #999;
    margin: 0 0 1rem 0;
    font-size: 0.875rem;
    line-height: 1.5;
  }

  .task-card-footer {
    display: flex;
    justify-content: space-between;
    margin-bottom: 1rem;
    font-size: 0.875rem;
    color: #999;
  }

  /* Badges */
  .priority-badge,
  .status-badge {
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 500;
    text-transform: capitalize;
  }

  .priority-high {
    background: rgba(244, 67, 54, 0.15);
    color: #f44336;
  }

  .priority-medium {
    background: rgba(255, 193, 7, 0.15);
    color: #ffc107;
  }

  .priority-low {
    background: rgba(76, 175, 80, 0.15);
    color: #4caf50;
  }

  .status-completed {
    background: rgba(76, 175, 80, 0.15);
    color: #4caf50;
  }

  .status-pending {
    background: rgba(255, 193, 7, 0.15);
    color: #ffc107;
  }

  .status-progress {
    background: rgba(33, 150, 243, 0.15);
    color: #2196f3;
  }

  /* Time Entries */
  .time-entries {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .time-entry-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem;
    background: #2d3139;
    border-radius: 6px;
    font-size: 0.875rem;
  }

  .entry-date {
    color: #e4e7eb;
    font-weight: 500;
  }

  .entry-time {
    color: #999;
  }

  .entry-hours {
    color: #667eea;
    font-weight: 600;
  }

  /* PTO */
  .pto-summary {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .pto-item {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background: #2d3139;
    border-radius: 8px;
  }

  .pto-icon {
    font-size: 2rem;
  }

  .pto-value {
    font-size: 1.5rem;
    font-weight: bold;
    color: #667eea;
  }

  .pto-label {
    color: #999;
    font-size: 0.875rem;
  }

  /* Pay */
  .pay-preview {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .pay-icon {
    font-size: 3rem;
  }

  .pay-label {
    color: #999;
    font-size: 0.875rem;
  }

  .pay-date {
    font-size: 1.25rem;
    font-weight: bold;
    color: #e4e7eb;
  }

  /* Pay Stubs */
  .paystubs-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .paystub-card {
    background: #1e2128;
    border: 1px solid #2d3139;
    border-radius: 12px;
    padding: 1.5rem;
  }

  .paystub-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1rem;
  }

  .paystub-header h3 {
    margin: 0 0 0.25rem 0;
    color: #e4e7eb;
  }

  .paystub-date {
    color: #999;
    font-size: 0.875rem;
    margin: 0;
  }

  .paystub-amounts {
    display: flex;
    gap: 2rem;
    margin-bottom: 1rem;
    padding: 1rem;
    background: #2d3139;
    border-radius: 8px;
  }

  .amount-item {
    flex: 1;
  }

  .amount-label {
    display: block;
    color: #999;
    font-size: 0.875rem;
    margin-bottom: 0.25rem;
  }

  .amount-value {
    display: block;
    font-size: 1.25rem;
    font-weight: bold;
    color: #e4e7eb;
  }

  .amount-value.net {
    color: #4caf50;
  }

  /* Benefits */
  .benefits-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .benefit-card {
    background: #1e2128;
    border: 1px solid #2d3139;
    border-radius: 12px;
    padding: 1.5rem;
  }

  .benefit-card-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .benefit-card-icon {
    font-size: 2.5rem;
  }

  .benefit-card-header h3 {
    margin: 0;
    color: #e4e7eb;
  }

  .benefit-description {
    color: #999;
    margin: 0 0 1.5rem 0;
    line-height: 1.5;
  }

  .benefits-links,
  .benefit-links-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .benefit-link,
  .benefit-external-link {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 1rem;
    background: #2d3139;
    border-radius: 8px;
    text-decoration: none;
    color: #e4e7eb;
    transition: all 0.2s;
  }

  .benefit-link:hover,
  .benefit-external-link:hover {
    background: #3d4149;
    transform: translateX(4px);
  }

  .benefit-icon {
    font-size: 1.5rem;
  }

  .benefit-title {
    font-weight: 500;
    margin-bottom: 0.25rem;
  }

  .benefit-subtitle {
    font-size: 0.875rem;
    color: #999;
  }

  .link-arrow,
  .external-icon {
    color: #667eea;
    font-size: 1.25rem;
  }

  .help-section {
    background: #2d3139;
    padding: 1.5rem;
    border-radius: 8px;
    text-align: center;
  }

  .help-section h3 {
    color: #e4e7eb;
    margin: 0 0 0.5rem 0;
  }

  .help-section p {
    color: #999;
    margin: 0;
  }

  .help-section a {
    color: #667eea;
    text-decoration: none;
  }

  /* Profile */
  .profile-sections {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .profile-section {
    background: #1e2128;
    border: 1px solid #2d3139;
    border-radius: 12px;
    padding: 1.5rem;
  }

  .profile-section h3 {
    color: #e4e7eb;
    margin: 0 0 1rem 0;
  }

  .info-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1.5rem;
  }

  .info-item {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .info-item.full-width {
    grid-column: 1 / -1;
  }

  .info-item label {
    color: #999;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .info-value {
    color: #e4e7eb;
    font-size: 1rem;
  }

  /* Buttons */
  .btn {
    padding: 0.625rem 1.25rem;
    border: none;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
  }

  .btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  }

  .btn-secondary {
    background: #2d3139;
    color: #e4e7eb;
  }

  .btn-secondary:hover {
    background: #3d4149;
  }

  .btn-sm {
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
  }

  .btn-block {
    width: 100%;
  }

  .link-btn {
    background: none;
    border: none;
    color: #667eea;
    font-size: 0.875rem;
    cursor: pointer;
    padding: 0;
  }

  .link-btn:hover {
    text-decoration: underline;
  }

  /* Empty States */
  .empty-state {
    text-align: center;
    padding: 3rem 2rem;
    color: #999;
  }

  .empty-state .empty-icon {
    font-size: 4rem;
    margin-bottom: 1rem;
    opacity: 0.5;
  }

  .empty-state h3 {
    color: #e4e7eb;
    margin-bottom: 0.5rem;
  }

  .empty-state-small {
    text-align: center;
    padding: 2rem;
    color: #999;
  }

  .empty-state-small .empty-icon {
    font-size: 3rem;
    margin-bottom: 0.5rem;
    opacity: 0.5;
  }

  /* Responsive */
  @media (max-width: 968px) {
    .dashboard-container {
      padding: 1rem;
    }

    .dashboard-header {
      flex-direction: column;
    }

    .header-stats {
      width: 100%;
      justify-content: space-between;
    }

    .stat-card {
      flex: 1;
    }

    .two-column {
      grid-template-columns: 1fr;
    }

    .quick-actions-grid {
      grid-template-columns: 1fr;
    }

    .tab-nav {
      overflow-x: auto;
      flex-wrap: nowrap;
    }

    .benefits-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
