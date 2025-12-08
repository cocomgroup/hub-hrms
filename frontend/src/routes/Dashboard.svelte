<script lang="ts">
  import { authStore } from '../stores/auth';

  let stats = $state({
    totalEmployees: 0,
    activeOnboarding: 0,
    pendingPTO: 0,
    upcomingPayroll: 0
  });

  $effect(() => {
    // Simulate loading stats
    stats = {
      totalEmployees: 127,
      activeOnboarding: 8,
      pendingPTO: 12,
      upcomingPayroll: 3
    };
  });
</script>

<div class="dashboard">
  <header class="page-header">
    <div>
      <h1>Dashboard</h1>
      <p>Welcome back! Here's what's happening with your team today.</p>
    </div>
  </header>

  <div class="stats-grid">
    <div class="stat-card">
      <div class="stat-icon" style="background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
          <circle cx="9" cy="7" r="4"></circle>
          <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
          <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
        </svg>
      </div>
      <div class="stat-content">
        <div class="stat-value">{stats.totalEmployees}</div>
        <div class="stat-label">Total Employees</div>
      </div>
    </div>

    <div class="stat-card">
      <div class="stat-icon" style="background: linear-gradient(135deg, #10b981 0%, #059669 100%);">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
          <polyline points="14 2 14 8 20 8"></polyline>
        </svg>
      </div>
      <div class="stat-content">
        <div class="stat-value">{stats.activeOnboarding}</div>
        <div class="stat-label">Active Onboarding</div>
      </div>
    </div>

    <div class="stat-card">
      <div class="stat-icon" style="background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
          <line x1="16" y1="2" x2="16" y2="6"></line>
          <line x1="8" y1="2" x2="8" y2="6"></line>
          <line x1="3" y1="10" x2="21" y2="10"></line>
        </svg>
      </div>
      <div class="stat-content">
        <div class="stat-value">{stats.pendingPTO}</div>
        <div class="stat-label">Pending PTO Requests</div>
      </div>
    </div>

    <div class="stat-card">
      <div class="stat-icon" style="background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="1" x2="12" y2="23"></line>
          <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
        </svg>
      </div>
      <div class="stat-content">
        <div class="stat-value">{stats.upcomingPayroll}</div>
        <div class="stat-label">Days to Payroll</div>
      </div>
    </div>
  </div>

  <div class="dashboard-grid">
    <div class="dashboard-card">
      <h2>Recent Activity</h2>
      <div class="activity-list">
        <div class="activity-item">
          <div class="activity-icon">👤</div>
          <div class="activity-content">
            <p class="activity-text"><strong>Sarah Johnson</strong> completed onboarding</p>
            <p class="activity-time">2 hours ago</p>
          </div>
        </div>
        <div class="activity-item">
          <div class="activity-icon">📅</div>
          <div class="activity-content">
            <p class="activity-text"><strong>Mike Chen</strong> requested PTO for Dec 20-24</p>
            <p class="activity-time">5 hours ago</p>
          </div>
        </div>
        <div class="activity-item">
          <div class="activity-icon">⏰</div>
          <div class="activity-content">
            <p class="activity-text"><strong>Emily Davis</strong> submitted timesheet</p>
            <p class="activity-time">1 day ago</p>
          </div>
        </div>
      </div>
    </div>

    <div class="dashboard-card">
      <h2>Upcoming Events</h2>
      <div class="events-list">
        <div class="event-item">
          <div class="event-date">
            <div class="event-month">DEC</div>
            <div class="event-day">15</div>
          </div>
          <div class="event-content">
            <p class="event-title">Payroll Processing</p>
            <p class="event-desc">Bi-weekly payroll run</p>
          </div>
        </div>
        <div class="event-item">
          <div class="event-date">
            <div class="event-month">JAN</div>
            <div class="event-day">01</div>
          </div>
          <div class="event-content">
            <p class="event-title">Benefits Enrollment</p>
            <p class="event-desc">Open enrollment period begins</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .dashboard {
    max-width: 1400px;
    animation: fadeIn 0.6s ease-out;
  }

  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(20px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .page-header {
    margin-bottom: 2rem;
  }

  h1 {
    font-size: 2rem;
    font-weight: 700;
    color: #e4e7eb;
    margin-bottom: 0.5rem;
  }

  .page-header p {
    color: #94a3b8;
    font-size: 1rem;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .stat-card {
    background: rgba(17, 24, 39, 0.8);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(99, 102, 241, 0.1);
    border-radius: 16px;
    padding: 1.5rem;
    display: flex;
    align-items: center;
    gap: 1rem;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .stat-card:hover {
    transform: translateY(-4px);
    border-color: rgba(99, 102, 241, 0.3);
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.3);
  }

  .stat-icon {
    width: 56px;
    height: 56px;
    min-width: 56px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }

  .stat-icon svg {
    width: 28px;
    height: 28px;
    color: white;
  }

  .stat-content {
    flex: 1;
  }

  .stat-value {
    font-size: 2rem;
    font-weight: 700;
    color: #e4e7eb;
    line-height: 1;
    margin-bottom: 0.25rem;
  }

  .stat-label {
    font-size: 0.875rem;
    color: #94a3b8;
  }

  .dashboard-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    gap: 1.5rem;
  }

  .dashboard-card {
    background: rgba(17, 24, 39, 0.8);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(99, 102, 241, 0.1);
    border-radius: 16px;
    padding: 1.5rem;
  }

  .dashboard-card h2 {
    font-size: 1.25rem;
    font-weight: 600;
    color: #e4e7eb;
    margin-bottom: 1.5rem;
  }

  .activity-list, .events-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .activity-item {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    padding: 1rem;
    background: rgba(15, 23, 42, 0.6);
    border-radius: 12px;
    transition: background 0.2s;
  }

  .activity-item:hover {
    background: rgba(15, 23, 42, 0.8);
  }

  .activity-icon {
    font-size: 1.5rem;
    width: 40px;
    height: 40px;
    min-width: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(99, 102, 241, 0.1);
    border-radius: 10px;
  }

  .activity-content {
    flex: 1;
  }

  .activity-text {
    color: #e4e7eb;
    font-size: 0.9375rem;
    margin-bottom: 0.25rem;
  }

  .activity-time {
    color: #64748b;
    font-size: 0.8125rem;
  }

  .event-item {
    display: flex;
    gap: 1rem;
    padding: 1rem;
    background: rgba(15, 23, 42, 0.6);
    border-radius: 12px;
  }

  .event-date {
    width: 60px;
    min-width: 60px;
    text-align: center;
    background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
    border-radius: 10px;
    padding: 0.5rem;
  }

  .event-month {
    font-size: 0.75rem;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.8);
  }

  .event-day {
    font-size: 1.5rem;
    font-weight: 700;
    color: white;
  }

  .event-content {
    flex: 1;
  }

  .event-title {
    font-weight: 600;
    color: #e4e7eb;
    margin-bottom: 0.25rem;
  }

  .event-desc {
    color: #94a3b8;
    font-size: 0.875rem;
  }

  @media (max-width: 768px) {
    .stats-grid {
      grid-template-columns: 1fr;
    }

    .dashboard-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
