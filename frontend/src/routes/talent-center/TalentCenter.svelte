<script lang="ts">
  import { onMount } from 'svelte';
  import RecruitingDashboard from './recruiting/RecruitingDashboard.svelte';
  import WorkflowManagement from './workflows/WorkflowManagement.svelte';
  
  export let initialTab: 'recruiting' | 'workflows' = 'recruiting';
  
  let activeTab: 'recruiting' | 'workflows' = initialTab;
  let stats = {
    activeJobs: 0,
    totalApplicants: 0,
    activeOnboardings: 0,
    workflowTemplates: 0
  };
  
  onMount(async () => {
    await loadStats();
  });
  
  async function loadStats() {
    try {
      const token = localStorage.getItem('token');
      
      // Load recruiting stats
      const recruitingRes = await fetch('/api/recruiting/stats', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (recruitingRes.ok) {
        const recruitingStats = await recruitingRes.json();
        stats.activeJobs = recruitingStats.active_jobs || 0;
        stats.totalApplicants = recruitingStats.total_applicants || 0;
      }
      
      // Load workflow stats
      const workflowRes = await fetch('/api/workflows/stats', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (workflowRes.ok) {
        const workflowStats = await workflowRes.json();
        stats.activeOnboardings = workflowStats.active_onboardings || 0;
        stats.workflowTemplates = workflowStats.templates_count || 0;
      }
    } catch (err) {
      console.error('Failed to load stats:', err);
    }
  }
</script>

<div class="talent-center">
  <!-- Header -->
  <div class="talent-header">
    <div class="header-content">
      <div class="title-section">
        <h1>🎯 Talent Center</h1>
        <p class="subtitle">Recruiting, Onboarding & Workflow Management</p>
      </div>
      
      <!-- Quick Stats -->
      <div class="quick-stats">
        <div class="stat-pill">
          <span class="stat-icon">📋</span>
          <span class="stat-number">{stats.activeJobs}</span>
          <span class="stat-label">Active Jobs</span>
        </div>
        <div class="stat-pill">
          <span class="stat-icon">👥</span>
          <span class="stat-number">{stats.totalApplicants}</span>
          <span class="stat-label">Applicants</span>
        </div>
        <div class="stat-pill">
          <span class="stat-icon">🚀</span>
          <span class="stat-number">{stats.activeOnboardings}</span>
          <span class="stat-label">Onboarding</span>
        </div>
        <div class="stat-pill">
          <span class="stat-icon">⚙️</span>
          <span class="stat-number">{stats.workflowTemplates}</span>
          <span class="stat-label">Templates</span>
        </div>
      </div>
    </div>
  </div>
  
  <!-- Tab Navigation -->
  <div class="tabs-container">
    <div class="tabs">
      <button
        class="tab"
        class:active={activeTab === 'recruiting'}
        on:click={() => activeTab = 'recruiting'}
      >
        <span class="tab-icon">🎯</span>
        <span class="tab-text">Recruiting</span>
        {#if stats.activeJobs > 0}
          <span class="tab-badge">{stats.activeJobs}</span>
        {/if}
      </button>
      
      <button
        class="tab"
        class:active={activeTab === 'workflows'}
        on:click={() => activeTab = 'workflows'}
      >
        <span class="tab-icon">⚙️</span>
        <span class="tab-text">Workflow Management</span>
        {#if stats.activeOnboardings > 0}
          <span class="tab-badge">{stats.activeOnboardings}</span>
        {/if}
      </button>
    </div>
  </div>
  
  <!-- Tab Content -->
  <div class="tab-content">
    {#if activeTab === 'recruiting'}
      <RecruitingDashboard on:statsUpdated={loadStats} />
    {:else if activeTab === 'workflows'}
      <WorkflowManagement on:statsUpdated={loadStats} />
    {/if}
  </div>
</div>

<style>
  .talent-center {
    min-height: 100vh;
    background: #f9fafb;
  }
  
  .talent-header {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 32px 24px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }
  
  .header-content {
    max-width: 1400px;
    margin: 0 auto;
  }
  
  .title-section h1 {
    font-size: 32px;
    font-weight: 700;
    margin: 0 0 8px 0;
  }
  
  .subtitle {
    font-size: 16px;
    opacity: 0.9;
    margin: 0 0 24px 0;
  }
  
  .quick-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;
  }
  
  .stat-pill {
    background: rgba(255, 255, 255, 0.15);
    backdrop-filter: blur(10px);
    border-radius: 12px;
    padding: 16px;
    display: flex;
    align-items: center;
    gap: 12px;
    border: 1px solid rgba(255, 255, 255, 0.2);
  }
  
  .stat-icon {
    font-size: 24px;
  }
  
  .stat-number {
    font-size: 24px;
    font-weight: 700;
  }
  
  .stat-label {
    font-size: 13px;
    opacity: 0.9;
    margin-left: auto;
  }
  
  .tabs-container {
    background: white;
    border-bottom: 1px solid #e5e7eb;
    position: sticky;
    top: 0;
    z-index: 10;
  }
  
  .tabs {
    max-width: 1400px;
    margin: 0 auto;
    display: flex;
    gap: 8px;
    padding: 0 24px;
  }
  
  .tab {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 16px 24px;
    background: none;
    border: none;
    border-bottom: 3px solid transparent;
    color: #6b7280;
    font-size: 15px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    position: relative;
  }
  
  .tab:hover {
    color: #111827;
    background: #f9fafb;
  }
  
  .tab.active {
    color: #667eea;
    border-bottom-color: #667eea;
  }
  
  .tab-icon {
    font-size: 20px;
  }
  
  .tab-badge {
    background: #ef4444;
    color: white;
    font-size: 11px;
    font-weight: 600;
    padding: 2px 8px;
    border-radius: 12px;
    min-width: 20px;
    text-align: center;
  }
  
  .tab-content {
    max-width: 1400px;
    margin: 0 auto;
    padding: 24px;
  }
  
  @media (max-width: 768px) {
    .talent-header {
      padding: 24px 16px;
    }
    
    .title-section h1 {
      font-size: 24px;
    }
    
    .quick-stats {
      grid-template-columns: repeat(2, 1fr);
    }
    
    .tabs {
      padding: 0 16px;
      overflow-x: auto;
    }
    
    .tab {
      padding: 12px 16px;
      white-space: nowrap;
    }
    
    .tab-text {
      display: none;
    }
    
    .tab-content {
      padding: 16px;
    }
  }
</style>
