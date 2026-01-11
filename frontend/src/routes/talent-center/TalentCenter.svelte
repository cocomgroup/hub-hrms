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
  <div class="header">
    <div>
      <h1>🎯 Talent Center</h1>
      <p class="subtitle">Recruiting, Onboarding & Workflow Management</p>
    </div>
  </div>
  
  <!-- Tab Navigation -->
  <div class="tabs">
    <button
      class="tab"
      class:active={activeTab === 'recruiting'}
      on:click={() => activeTab = 'recruiting'}
    >
      <span class="tab-icon">📋</span>
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
        <span class="tab-text">Workflows</span>
        {#if stats.activeOnboardings > 0}
          <span class="tab-badge">{stats.activeOnboardings}</span>
        {/if}
      </button>
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
    padding: 32px 24px;
  }
  
  .header {
    max-width: 1400px;
    margin: 0 auto 32px;
  }
  
  .header h1 {
    font-size: 28px;
    font-weight: 700;
    color: #111827;
    margin: 0 0 8px 0;
  }
  
  .subtitle {
    font-size: 14px;
    color: #6b7280;
    margin: 0;
  }
  
  .tabs {
    max-width: 1400px;
    margin: 0 auto;
    display: flex;
    gap: 8px;
    border-bottom: 2px solid #e5e7eb;
    margin-bottom: 32px;
  }
  
  .tab {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 20px;
    background: none;
    border: none;
    border-bottom: 3px solid transparent;
    color: #6b7280;
    font-size: 15px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    position: relative;
    margin-bottom: -2px;
  }
  
  .tab:hover {
    color: #111827;
    background: #f9fafb;
  }
  
  .tab.active {
    color: #2563eb;
    border-bottom-color: #2563eb;
  }
  
  .tab-icon {
    font-size: 18px;
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