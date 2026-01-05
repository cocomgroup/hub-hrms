<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  import TeamManagement from '../components/TeamManagement.svelte';
  import TimesheetApproval from '../components/TimesheetApproval.svelte';
  import ProjectManager from '../components/ProjectManager.svelte';
  
  const API_BASE_URL = getApiBaseUrl();

  interface Props {
    employee?: any;
  }
  
  let { employee }: Props = $props();

  let loading = $state(false);
  let activeTab = $state('timesheets'); // 'timesheets' | 'team' | 'projects'
  let pendingCount = $state(0);
  let teamCount = $state(0);
  let projectCount = $state(0);
  
  let currentManager = $derived(employee || $authStore.employee);

  onMount(() => {
    loadDashboardStats();
  });

  async function loadDashboardStats() {
    loading = true;
    try {
      // Load pending timesheets count
      const timesheetResponse = await fetch(`${API_BASE_URL}/timesheet/pending`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (timesheetResponse.ok) {
        const timesheets = await timesheetResponse.json();
        pendingCount = Array.isArray(timesheets) ? timesheets.length : 0;
      }

      // Load MY team members count (manager_id = current user)
      const teamResponse = await fetch(`${API_BASE_URL}/employees/team`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (teamResponse.ok) {
        const team = await teamResponse.json();
        teamCount = Array.isArray(team) ? team.length : 0;
      }

      // Load projects count
      const projectResponse = await fetch(`${API_BASE_URL}/projects`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (projectResponse.ok) {
        const projects = await projectResponse.json();
        projectCount = Array.isArray(projects) ? projects.length : 0;
      }
    } catch (err) {
      console.error('Error loading dashboard stats:', err);
    } finally {
      loading = false;
    }
  }

  function handleTabChange(tab: string) {
    activeTab = tab;
  }

  function handleTimesheetAction() {
    loadDashboardStats();
  }
</script>

<div class="manager-dashboard">
  <div class="header">
    <div class="header-content">
      <h1>Manager Dashboard</h1>
      <p class="subtitle">Welcome back, {currentManager?.first_name || 'Manager'}</p>
    </div>
  </div>

  <!-- Stats Cards -->
  <div class="stats-grid">
    <button 
      type="button"
      class="stat-card {activeTab === 'timesheets' ? 'active' : ''}"
      onclick={() => handleTabChange('timesheets')}
    >
      <div class="stat-icon pending">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>
        </svg>
      </div>
      <div class="stat-content">
        <div class="stat-value">{pendingCount}</div>
        <div class="stat-label">Pending Timesheets</div>
      </div>
      {#if pendingCount > 0}
        <div class="stat-badge">{pendingCount} need review</div>
      {/if}
    </button>

    <button 
      type="button"
      class="stat-card {activeTab === 'team' ? 'active' : ''}"
      onclick={() => handleTabChange('team')}
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
        <div class="stat-value">{teamCount}</div>
        <div class="stat-label">My Team Members</div>
      </div>
    </button>

    <button 
      type="button"
      class="stat-card {activeTab === 'projects' ? 'active' : ''}"
      onclick={() => handleTabChange('projects')}
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
  </div>

  <!-- Tabs -->
  <div class="tabs">
    <button
      type="button"
      class="tab-btn {activeTab === 'timesheets' ? 'active' : ''}"
      onclick={() => handleTabChange('timesheets')}
    >
      ⏰ Timesheet Approvals
      {#if pendingCount > 0}
        <span class="badge">{pendingCount}</span>
      {/if}
    </button>
    <button
      type="button"
      class="tab-btn {activeTab === 'team' ? 'active' : ''}"
      onclick={() => handleTabChange('team')}
    >
      👥 Manage Team
    </button>
    <button
      type="button"
      class="tab-btn {activeTab === 'projects' ? 'active' : ''}"
      onclick={() => handleTabChange('projects')}
    >
      📋 Projects
    </button>
  </div>

  <!-- Content -->
  <div class="content">
    {#if loading}
      <div class="loading">
        <div class="spinner"></div>
        <p>Loading...</p>
      </div>
    {:else if activeTab === 'timesheets'}
      <TimesheetApproval onAction={handleTimesheetAction} />
    {:else if activeTab === 'team'}
      <TeamManagement />
    {:else if activeTab === 'projects'}
      <ProjectManager employee={currentManager} />
    {/if}
  </div>
</div>

<style>
  .manager-dashboard {
    padding: 24px;
    max-width: 1400px;
    margin: 0 auto;
  }

  .header {
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

  /* Stats Grid */
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 20px;
    margin-bottom: 32px;
  }

  .stat-card {
    position: relative;
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

  .stat-icon.pending {
    background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
    color: #92400e;
  }

  .stat-icon.team {
    background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
    color: #1e40af;
  }

  .stat-icon.projects {
    background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
    color: #065f46;
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

  .stat-badge {
    position: absolute;
    top: 12px;
    right: 12px;
    background: #fef3c7;
    color: #92400e;
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
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
    display: flex;
    align-items: center;
    gap: 8px;
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

  .badge {
    background: #3b82f6;
    color: white;
    padding: 2px 8px;
    border-radius: 10px;
    font-size: 12px;
    font-weight: 600;
  }

  /* Content */
  .content {
    background: white;
    border-radius: 16px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    min-height: 400px;
  }

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

  @media (max-width: 768px) {
    .manager-dashboard {
      padding: 16px;
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

    .tab-btn {
      justify-content: center;
    }
  }
</style>